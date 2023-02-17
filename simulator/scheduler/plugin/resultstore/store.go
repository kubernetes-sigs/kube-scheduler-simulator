package resultstore

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/annotation"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/util"
)

// Store has results of scheduling.
// It manages all scheduling results and reflects all results on the pod annotation when the scheduling is finished.
type Store struct {
	mu *sync.Mutex

	client            clientset.Interface
	results           map[key]*result
	scorePluginWeight map[string]int32
}

const (
	// PassedFilterMessage is used when a node pass the filter plugin.
	PassedFilterMessage = "passed"
	// SuccessMessage is used when no error is retured from the plugin.
	SuccessMessage = "success"
	// WaitMessage is used when wait status is retured from the plugin.
	WaitMessage = "wait"
	// PostFilterNominatedMessage is used when a postFilter plugin returns success.
	PostFilterNominatedMessage = "preemption victim"
)

// result has a scheduling result of pod.
type result struct {
	// selectedNode is the scheduling result. It'll be filled when the Pod go through Reserve phase.
	selectedNode string

	// plugin name → pre score(string)
	// If success, SuccessMessage is shown.
	// If non success, status.Message() is shown.
	preScore map[string]string

	// node name → plugin name → score(string)
	score map[string]map[string]string

	// node name → plugin name → finalScore(string)
	// This score is normalized and applied weight for each plugins.
	finalScore map[string]map[string]string

	// plugin name → pre filter status.
	// If success, SuccessMessage is shown.
	// If non success, status.Message() is shown.
	preFilterStatus map[string]string

	// plugin name → pre filter result(framework.PreFilterResult)
	// NodeNames in framework.PreFilterResult is shown.
	preFilterResult map[string][]string

	// node name → plugin name → filtering result
	// When node pass the filter, filtering result will be PassedFilterMessage.
	// When node blocked by the filter, filtering result is blocked reason.
	filter map[string]map[string]string

	// node name → plugin name → post filtering result
	postFilter map[string]map[string]string

	// plugin name → permit result (framework.Status)
	permit map[string]string

	// plugin name → permit timeout(string)
	permitTimeout map[string]string

	// plugin name → reserve result(string)
	reserve map[string]string

	// plugin name → prebind result(string)
	prebind map[string]string

	// plugin name → bind result(string)
	bind map[string]string
}

func New(informerFactory informers.SharedInformerFactory, client clientset.Interface, scorePluginWeight map[string]int32) *Store {
	s := &Store{
		mu:                new(sync.Mutex),
		client:            client,
		results:           map[key]*result{},
		scorePluginWeight: scorePluginWeight,
	}

	// Store adds scheduling results when pod is updating.
	// This is because scheduling framework doesn’t have any phase to hook scheduling finished. (both successfully and non-successfully)
	informerFactory.Core().V1().Pods().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			UpdateFunc: s.addSchedulingResultToPod,
		},
	)

	return s
}

// key is the key of result map on Store.
// key is created from namespace and podName.
type key string

// newKey creates key with namespace and podName.
func newKey(namespace, podName string) key {
	k := namespace + "/" + podName
	return key(k)
}

func newData() *result {
	d := &result{
		score:           map[string]map[string]string{},
		finalScore:      map[string]map[string]string{},
		preFilterResult: map[string][]string{},
		preFilterStatus: map[string]string{},
		preScore:        map[string]string{},
		filter:          map[string]map[string]string{},
		postFilter:      map[string]map[string]string{},
		permit:          map[string]string{},
		permitTimeout:   map[string]string{},
		reserve:         map[string]string{},
		bind:            map[string]string{},
		prebind:         map[string]string{},
	}
	return d
}

//nolint:funlen,cyclop
func (s *Store) addSchedulingResultToPod(_, newObj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx := context.Background()

	pod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.ErrorS(nil, "Cannot convert to *v1.Pod", "obj", newObj)
		return
	}

	k := newKey(pod.Namespace, pod.Name)
	if _, ok := s.results[k]; !ok {
		// Store doesn't have scheduling result of pod.
		return
	}

	if err := s.addPreFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add prefilter result to pod: %+v", err)
		return
	}

	if err := s.addFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add filtering result to pod: %+v", err)
		return
	}

	if err := s.addPostFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add post filtering result to pod: %+v", err)
		return
	}

	if err := s.addPreScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add prescore result to pod: %+v", err)
		return
	}

	if err := s.addScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add scoring result to pod: %+v", err)
		return
	}

	if err := s.addFinalScoreResultToPod(pod); err != nil {
		klog.Errorf("failed to add final score result to pod: %+v", err)
		return
	}

	if err := s.addReserveResultToPod(pod); err != nil {
		klog.Errorf("failed to add reserve result to pod: %+v", err)
		return
	}

	if err := s.addPermitResultToPod(pod); err != nil {
		klog.Errorf("failed to add permit result to pod: %+v", err)
		return
	}

	if err := s.addPreBindResultToPod(pod); err != nil {
		klog.Errorf("failed to add prebind result to pod: %+v", err)
		return
	}

	if err := s.addBindResultToPod(pod); err != nil {
		klog.Errorf("failed to add bind result to pod: %+v", err)
		return
	}

	s.addSelectedNodeToPod(pod)

	updateFunc := func() (bool, error) {
		_, err := s.client.CoreV1().Pods(pod.Namespace).Update(ctx, pod, metav1.UpdateOptions{})
		if err != nil {
			return false, xerrors.Errorf("update pod: %v", err)
		}

		return true, nil
	}
	if err := util.RetryWithExponentialBackOff(updateFunc); err != nil {
		klog.Errorf("failed to update pod with retry to record score: %+v", err)
		return
	}

	// delete data from Store only if data is successfully added on pod's annotations.
	s.deleteData(k)
}

func (s *Store) addPreFilterResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)

	_, ok := pod.GetAnnotations()[annotation.PreFilterResultAnnotationKey]
	if !ok {
		if s.results[k].preFilterResult == nil {
			s.results[k].preFilterResult = map[string][]string{}
		}
		r, err := json.Marshal(s.results[k].preFilterResult)
		if err != nil {
			return xerrors.Errorf("encode json to record preFilter result: %w", err)
		}

		metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PreFilterResultAnnotationKey, string(r))
	}

	_, ok2 := pod.GetAnnotations()[annotation.PreFilterStatusResultAnnotationKey]
	if ok2 {
		return nil
	}

	if s.results[k].preFilterStatus == nil {
		s.results[k].preFilterStatus = map[string]string{}
	}
	sta, err := json.Marshal(s.results[k].preFilterStatus)
	if err != nil {
		return xerrors.Errorf("encode json to record preFilter status: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PreFilterStatusResultAnnotationKey, string(sta))

	return nil
}

func (s *Store) addPreScoreResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.PreScoreResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].preScore == nil {
		s.results[k].preScore = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].preScore)
	if err != nil {
		return xerrors.Errorf("encode json to record preScore status: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PreScoreResultAnnotationKey, string(status))
	return nil
}

func (s *Store) addBindResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.BindResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].bind == nil {
		s.results[k].bind = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].bind)
	if err != nil {
		return xerrors.Errorf("encode json to record bind status: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.BindResultAnnotationKey, string(status))
	return nil
}

func (s *Store) addPreBindResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.PreBindResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].prebind == nil {
		s.results[k].prebind = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].prebind)
	if err != nil {
		return xerrors.Errorf("encode json to record preBind status: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PreBindResultAnnotationKey, string(status))
	return nil
}

func (s *Store) addSelectedNodeToPod(pod *v1.Pod) {
	_, ok := pod.GetAnnotations()[annotation.SelectedNodeAnnotationKey]
	if ok {
		return
	}

	k := newKey(pod.Namespace, pod.Name)
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.SelectedNodeAnnotationKey, s.results[k].selectedNode)
}

func (s *Store) addReserveResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.ReserveResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].reserve == nil {
		s.results[k].reserve = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].reserve)
	if err != nil {
		return xerrors.Errorf("encode json to record reserve status: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ReserveResultAnnotationKey, string(status))
	return nil
}

func (s *Store) addPermitResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)

	_, ok := pod.GetAnnotations()[annotation.PermitTimeoutResultAnnotationKey]
	if !ok {
		if s.results[k].permitTimeout == nil {
			s.results[k].permitTimeout = map[string]string{}
		}
		timeout, err := json.Marshal(s.results[k].permitTimeout)
		if err != nil {
			return xerrors.Errorf("encode json to record permit timeout: %w", err)
		}
		metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PermitTimeoutResultAnnotationKey, string(timeout))
	}

	_, ok = pod.GetAnnotations()[annotation.PermitStatusResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].permit == nil {
		s.results[k].permit = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].permit)
	if err != nil {
		return xerrors.Errorf("encode json to record permit status: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PermitStatusResultAnnotationKey, string(status))
	return nil
}

func (s *Store) addFilterResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.FilterResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].filter == nil {
		s.results[k].filter = map[string]map[string]string{}
	}
	status, err := json.Marshal(s.results[k].filter)
	if err != nil {
		return xerrors.Errorf("encode json to record filter status: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.FilterResultAnnotationKey, string(status))
	return nil
}

func (s *Store) addPostFilterResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.PostFilterResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].postFilter == nil {
		s.results[k].postFilter = map[string]map[string]string{}
	}
	result, err := json.Marshal(s.results[k].postFilter)
	if err != nil {
		return xerrors.Errorf("encode json to record post filter results: %w", err)
	}
	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.PostFilterResultAnnotationKey, string(result))
	return nil
}

func (s *Store) addScoreResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.ScoreResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].score == nil {
		s.results[k].score = map[string]map[string]string{}
	}
	scores, err := json.Marshal(s.results[k].score)
	if err != nil {
		return xerrors.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ScoreResultAnnotationKey, string(scores))
	return nil
}

func (s *Store) addFinalScoreResultToPod(pod *v1.Pod) error {
	_, ok := pod.GetAnnotations()[annotation.FinalScoreResultAnnotationKey]
	if ok {
		return nil
	}

	k := newKey(pod.Namespace, pod.Name)
	if s.results[k].finalScore == nil {
		s.results[k].finalScore = map[string]map[string]string{}
	}
	scores, err := json.Marshal(s.results[k].finalScore)
	if err != nil {
		return xerrors.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.FinalScoreResultAnnotationKey, string(scores))
	return nil
}

// AddFilterResult adds filtering result to pod annotation.
func (s *Store) AddFilterResult(namespace, podName, nodeName, pluginName, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	if _, ok := s.results[k].filter[nodeName]; !ok {
		s.results[k].filter[nodeName] = map[string]string{}
	}

	s.results[k].filter[nodeName][pluginName] = reason
}

// AddPostFilterResult adds post filter result to the pod annotaiton.
//   - nominatedNodeName represents the node name which nominated by the postFilter plugin.
//     Otherwise, the string "" would be stored in this arg.
func (s *Store) AddPostFilterResult(namespace, podName, nominatedNodeName, pluginName string, nodeNames []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}
	for _, nodeName := range nodeNames {
		if _, ok := s.results[k].postFilter[nodeName]; !ok {
			s.results[k].postFilter[nodeName] = map[string]string{}
		}
		if nodeName == nominatedNodeName {
			s.results[k].postFilter[nodeName][pluginName] = PostFilterNominatedMessage
		}
	}
}

// AddScoreResult adds scoring result to pod annotation.
func (s *Store) AddScoreResult(namespace, podName, nodeName, pluginName string, score int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	if _, ok := s.results[k].score[nodeName]; !ok {
		s.results[k].score[nodeName] = map[string]string{}
	}

	s.results[k].score[nodeName][pluginName] = strconv.FormatInt(score, 10)

	// we already locked on first of this func
	s.addNormalizedScoreResultWithoutLock(namespace, podName, nodeName, pluginName, score)
}

// AddNormalizedScoreResult adds final score result to pod annotation.
func (s *Store) AddNormalizedScoreResult(namespace, podName, nodeName, pluginName string, normalizedscore int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.addNormalizedScoreResultWithoutLock(namespace, podName, nodeName, pluginName, normalizedscore)
}

func (s *Store) addNormalizedScoreResultWithoutLock(namespace, podName, nodeName, pluginName string, normalizedscore int64) {
	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	if _, ok := s.results[k].finalScore[nodeName]; !ok {
		s.results[k].finalScore[nodeName] = map[string]string{}
	}

	finalscore := s.applyWeightOnScore(pluginName, normalizedscore)

	// apply weight to calculate final score.
	s.results[k].finalScore[nodeName][pluginName] = strconv.FormatInt(finalscore, 10)
}

func (s *Store) applyWeightOnScore(pluginName string, score int64) int64 {
	weight := s.scorePluginWeight[pluginName]
	return score * int64(weight)
}

func (s *Store) DeleteData(k key) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteData(k)
}

// deleteData deletes the result stored with the given key.
// Note: we assume the store lock is already acquired.
func (s *Store) deleteData(k key) {
	delete(s.results, k)
}

func (s *Store) AddPreFilterResult(namespace, podName, pluginName, reason string, preFilterResult *framework.PreFilterResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].preFilterStatus[pluginName] = reason
	if preFilterResult != nil {
		s.results[k].preFilterResult[pluginName] = preFilterResult.NodeNames.List()
	}
}

func (s *Store) AddPreScoreResult(namespace, podName, pluginName, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].preScore[pluginName] = reason
}

func (s *Store) AddPermitResult(namespace, podName, pluginName, status string, timeout time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].permit[pluginName] = status
	s.results[k].permitTimeout[pluginName] = timeout.String()
}

func (s *Store) AddSelectedNode(namespace, podName, nodeName string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].selectedNode = nodeName
}

func (s *Store) AddReserveResult(namespace, podName, pluginName, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].reserve[pluginName] = status
}

func (s *Store) AddBindResult(namespace, podName, pluginName, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].bind[pluginName] = status
}

func (s *Store) AddPreBindResult(namespace, podName, pluginName, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}

	s.results[k].prebind[pluginName] = status
}
