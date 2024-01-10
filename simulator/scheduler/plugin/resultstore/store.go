package resultstore

import (
	"encoding/json"
	"strconv"
	"sync"
	"time"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/plugin/annotation"
)

// Store has results of scheduling.
// It manages all scheduling results and reflects all results on the pod annotation when the scheduling is finished.
type Store struct {
	mu *sync.Mutex

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

	// customResults has the user defined custom results.
	// annotation key -> result(string)
	customResults map[string]string
}

func New(scorePluginWeight map[string]int32) *Store {
	s := &Store{
		mu:                new(sync.Mutex),
		results:           map[key]*result{},
		scorePluginWeight: scorePluginWeight,
	}

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
		customResults:   map[string]string{},
	}
	return d
}

// GetStoredResult get all stored result of a given Pod.
//
//nolint:cyclop,funlen
func (s *Store) GetStoredResult(pod *v1.Pod) map[string]string {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(pod.Namespace, pod.Name)
	if _, ok := s.results[k]; !ok {
		// Store doesn't have scheduling result of pod.
		return nil
	}

	annotation := map[string]string{}
	if err := s.addPreFilterResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add prefilter result to pod: %+v", err)
		return nil
	}

	if err := s.addFilterResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add filtering result to pod: %+v", err)
		return nil
	}

	if err := s.addPostFilterResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add post filtering result to pod: %+v", err)
		return nil
	}

	if err := s.addPreScoreResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add prescore result to pod: %+v", err)
		return nil
	}

	if err := s.addScoreResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add scoring result to pod: %+v", err)
		return nil
	}

	if err := s.addFinalScoreResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add final score result to pod: %+v", err)
		return nil
	}

	if err := s.addReserveResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add reserve result to pod: %+v", err)
		return nil
	}

	if err := s.addPermitResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add permit result to pod: %+v", err)
		return nil
	}

	if err := s.addPreBindResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add prebind result to pod: %+v", err)
		return nil
	}

	if err := s.addBindResultToMap(annotation, k); err != nil {
		klog.Errorf("failed to add bind result to pod: %+v", err)
		return nil
	}

	s.addCustomResultsToMap(annotation, k)
	s.addSelectedNodeToPod(annotation, k)

	return annotation
}

func (s *Store) addPreFilterResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.PreFilterResultAnnotationKey]
	if !ok {
		if s.results[k].preFilterResult == nil {
			s.results[k].preFilterResult = map[string][]string{}
		}
		r, err := json.Marshal(s.results[k].preFilterResult)
		if err != nil {
			return xerrors.Errorf("encode json to record preFilter result: %w", err)
		}

		anno[annotation.PreFilterResultAnnotationKey] = string(r)
	}

	_, ok2 := anno[annotation.PreFilterStatusResultAnnotationKey]
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

	anno[annotation.PreFilterStatusResultAnnotationKey] = string(sta)

	return nil
}

func (s *Store) addPreScoreResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.PreScoreResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].preScore == nil {
		s.results[k].preScore = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].preScore)
	if err != nil {
		return xerrors.Errorf("encode json to record preScore status: %w", err)
	}

	anno[annotation.PreScoreResultAnnotationKey] = string(status)
	return nil
}

func (s *Store) addBindResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.BindResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].bind == nil {
		s.results[k].bind = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].bind)
	if err != nil {
		return xerrors.Errorf("encode json to record bind status: %w", err)
	}

	anno[annotation.BindResultAnnotationKey] = string(status)
	return nil
}

func (s *Store) addPreBindResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.PreBindResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].prebind == nil {
		s.results[k].prebind = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].prebind)
	if err != nil {
		return xerrors.Errorf("encode json to record preBind status: %w", err)
	}

	anno[annotation.PreBindResultAnnotationKey] = string(status)
	return nil
}

func (s *Store) addSelectedNodeToPod(anno map[string]string, k key) {
	_, ok := anno[annotation.SelectedNodeAnnotationKey]
	if ok {
		return
	}

	anno[annotation.SelectedNodeAnnotationKey] = s.results[k].selectedNode
}

func (s *Store) addReserveResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.ReserveResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].reserve == nil {
		s.results[k].reserve = map[string]string{}
	}
	status, err := json.Marshal(s.results[k].reserve)
	if err != nil {
		return xerrors.Errorf("encode json to record reserve status: %w", err)
	}
	anno[annotation.ReserveResultAnnotationKey] = string(status)
	return nil
}

func (s *Store) addPermitResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.PermitTimeoutResultAnnotationKey]
	if !ok {
		if s.results[k].permitTimeout == nil {
			s.results[k].permitTimeout = map[string]string{}
		}
		timeout, err := json.Marshal(s.results[k].permitTimeout)
		if err != nil {
			return xerrors.Errorf("encode json to record permit timeout: %w", err)
		}
		anno[annotation.PermitTimeoutResultAnnotationKey] = string(timeout)
	}

	_, ok = anno[annotation.PermitStatusResultAnnotationKey]
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
	anno[annotation.PermitStatusResultAnnotationKey] = string(status)
	return nil
}

func (s *Store) addFilterResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.FilterResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].filter == nil {
		s.results[k].filter = map[string]map[string]string{}
	}
	status, err := json.Marshal(s.results[k].filter)
	if err != nil {
		return xerrors.Errorf("encode json to record filter status: %w", err)
	}

	anno[annotation.FilterResultAnnotationKey] = string(status)
	return nil
}

func (s *Store) addPostFilterResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.PostFilterResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].postFilter == nil {
		s.results[k].postFilter = map[string]map[string]string{}
	}
	result, err := json.Marshal(s.results[k].postFilter)
	if err != nil {
		return xerrors.Errorf("encode json to record post filter results: %w", err)
	}
	anno[annotation.PostFilterResultAnnotationKey] = string(result)
	return nil
}

func (s *Store) addScoreResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.ScoreResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].score == nil {
		s.results[k].score = map[string]map[string]string{}
	}
	scores, err := json.Marshal(s.results[k].score)
	if err != nil {
		return xerrors.Errorf("encode json to record scores: %w", err)
	}

	anno[annotation.ScoreResultAnnotationKey] = string(scores)
	return nil
}

func (s *Store) addFinalScoreResultToMap(anno map[string]string, k key) error {
	_, ok := anno[annotation.FinalScoreResultAnnotationKey]
	if ok {
		return nil
	}

	if s.results[k].finalScore == nil {
		s.results[k].finalScore = map[string]map[string]string{}
	}
	scores, err := json.Marshal(s.results[k].finalScore)
	if err != nil {
		return xerrors.Errorf("encode json to record scores: %w", err)
	}

	anno[annotation.FinalScoreResultAnnotationKey] = string(scores)
	return nil
}

func (s *Store) addCustomResultsToMap(anno map[string]string, k key) {
	for annokey, r := range s.results[k].customResults {
		_, ok := anno[annokey]
		if ok {
			continue
		}
		anno[annokey] = r
	}
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

// DeleteData deletes the data corresponding to the specified Pod.
func (s *Store) DeleteData(pod v1.Pod) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deleteData(newKey(pod.Namespace, pod.Name))
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
		s.results[k].preFilterResult[pluginName] = preFilterResult.NodeNames.UnsortedList()
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

// AddCustomResult adds user defined data.
// The results added through this func is reflected on the Pod's annotation eventually like other scheduling results.
// This function is intended to be called from the plugin.PluginExtender; allow users to export some internal state on Pods for debugging purpose.
// For example,
// Calling AddCustomResult in NodeAffinity's PreFilterPluginExtender:
// AddCustomResult("namespace", "incomingPod", "node-affinity-filter-internal-state-anno-key", "internal-state")
// Then, "incomingPod" Pod will get {"node-affinity-filter-internal-state-anno-key": "internal-state"} annotation after scheduling.
func (s *Store) AddCustomResult(namespace, podName, annotationKey, result string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := newKey(namespace, podName)
	if _, ok := s.results[k]; !ok {
		s.results[k] = newData()
	}
	s.results[k].customResults[annotationKey] = result
}
