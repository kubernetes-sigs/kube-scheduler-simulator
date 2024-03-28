package nodenumber

import (
	"context"
	"errors"
	"strconv"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
)

// NodeNumber is an example plugin that favors nodes that have the number suffix which is the same as the number suffix of the pod name.
// But if a reverse option is true, it favors nodes that have the number suffix which **isn't** the same as the number suffix of pod name.
//
// For example:
// With reverse option false, when schedule a pod named Pod1, a Node named Node1 gets a lower score than a node named Node9.
//
// NOTE: this plugin only handle single digit numbers only.
type NodeNumber struct {
	// if reverse is true, it favors nodes that doesn't have the same number suffix.
	//
	// For example:
	// When schedule a pod named Pod1, a Node named Node1 gets a lower score than a node named Node9.
	reverse bool
}

var (
	_ framework.ScorePlugin    = &NodeNumber{}
	_ framework.PreScorePlugin = &NodeNumber{}
)

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name             = "NodeNumber"
	preScoreStateKey = "PreScore" + Name
)

// Name returns the name of the plugin. It is used in logs, etc.
func (pl *NodeNumber) Name() string {
	return Name
}

// preScoreState computed at PreScore and used at Score.
type preScoreState struct {
	podSuffixNumber int
}

// Clone implements the mandatory Clone interface. We don't really copy the data since
// there is no need for that.
func (s *preScoreState) Clone() framework.StateData {
	return s
}

func (pl *NodeNumber) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	klog.InfoS("execute PreScore on NodeNumber plugin", "pod", klog.KObj(pod))

	podNameLastChar := pod.Name[len(pod.Name)-1:]
	podnum, err := strconv.Atoi(podNameLastChar)
	if err != nil {
		// return success even if its suffix is non-number.
		return nil
	}

	s := &preScoreState{
		podSuffixNumber: podnum,
	}
	state.Write(preScoreStateKey, s)

	return nil
}

func (pl *NodeNumber) EventsToRegister() []framework.ClusterEvent {
	return []framework.ClusterEvent{
		{Resource: framework.Node, ActionType: framework.Add},
	}
}

var ErrNotExpectedPreScoreState = errors.New("unexpected pre score state")

// Score invoked at the score extension point.
func (pl *NodeNumber) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	klog.InfoS("execute Score on NodeNumber plugin", "pod", klog.KObj(pod))
	data, err := state.Read(preScoreStateKey)
	if err != nil {
		// return success even if there is no value in preScoreStateKey, since the
		// suffix of pod name maybe non-number.
		return 0, nil
	}

	s, ok := data.(*preScoreState)
	if !ok {
		err = xerrors.Errorf("fetched pre score state is not *preScoreState, but %T, %w", data, ErrNotExpectedPreScoreState)
		return 0, framework.AsStatus(err)
	}

	nodeNameLastChar := nodeName[len(nodeName)-1:]

	nodenum, err := strconv.Atoi(nodeNameLastChar)
	if err != nil {
		// return success even if its suffix is non-number.
		return 0, nil
	}

	var matchScore int64 = 10
	var nonMatchScore int64 = 0 //nolint:revive // for better readability.
	if pl.reverse {
		matchScore = 0
		nonMatchScore = 10
	}

	if s.podSuffixNumber == nodenum {
		// if match, node get high score.
		return matchScore, nil
	}

	return nonMatchScore, nil
}

// ScoreExtensions of the Score plugin.
func (pl *NodeNumber) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// New initializes a new plugin and returns it.
func New(ctx context.Context, arg runtime.Object, h framework.Handle) (framework.Plugin, error) {
	typedArg := NodeNumberArgs{Reverse: false}
	if arg != nil {
		err := frameworkruntime.DecodeInto(arg, &typedArg)
		if err != nil {
			return nil, xerrors.Errorf("decode arg into NodeNumberArgs: %w", err)
		}
		klog.Info("NodeNumberArgs is successfully applied")
	}
	return &NodeNumber{reverse: typedArg.Reverse}, nil
}

// NodeNumberArgs is arguments for node number plugin.
//
//nolint:revive
type NodeNumberArgs struct {
	metav1.TypeMeta

	Reverse bool `json:"reverse"`
}
