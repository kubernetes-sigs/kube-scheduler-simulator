package resultstore

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"

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
	// PassedFilterMessage is used when node pass the filter plugin.
	PassedFilterMessage = "passed"
)

// result has a scheduling result of pod.
type result struct {
	// node name → plugin name → score(string)
	score map[string]map[string]string

	// node name → plugin name → finalscore(string)
	// This score is normalized and applied weight for each plugins.
	finalscore map[string]map[string]string

	// node name → plugin name → filtering result
	// When node pass the filter, filtering result will be PassedFilterMessage.
	// When node blocked by the filter, filtering result is blocked reason.
	filter map[string]map[string]string
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
		score:      map[string]map[string]string{},
		finalscore: map[string]map[string]string{},
		filter:     map[string]map[string]string{},
	}
	return d
}

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

	if err := s.addFilterResultToPod(pod); err != nil {
		klog.Errorf("failed to add filtering result to pod: %+v", err)
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

func (s *Store) addFilterResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.results[k].filter)
	if err != nil {
		return xerrors.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.FilterResultAnnotationKey, string(scores))
	return nil
}

func (s *Store) addScoreResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.results[k].score)
	if err != nil {
		return xerrors.Errorf("encode json to record scores: %w", err)
	}

	metav1.SetMetaDataAnnotation(&pod.ObjectMeta, annotation.ScoreResultAnnotationKey, string(scores))
	return nil
}

func (s *Store) addFinalScoreResultToPod(pod *v1.Pod) error {
	k := newKey(pod.Namespace, pod.Name)
	scores, err := json.Marshal(s.results[k].finalscore)
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

	if _, ok := s.results[k].finalscore[nodeName]; !ok {
		s.results[k].finalscore[nodeName] = map[string]string{}
	}

	finalscore := s.applyWeightOnScore(pluginName, normalizedscore)

	// apply weight to calculate final score.
	s.results[k].finalscore[nodeName][pluginName] = strconv.FormatInt(finalscore, 10)
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
