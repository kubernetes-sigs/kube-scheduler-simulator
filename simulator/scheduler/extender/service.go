package extender

//go:generate mockgen -package=mock_$GOPACKAGE -source=./resultstore/resultstore.go -destination=./mock_$GOPACKAGE/resultstore.go

import (
	"strconv"

	"golang.org/x/xerrors"
	clientset "k8s.io/client-go/kubernetes"
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"

	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/extender/resultstore"
	"sigs.k8s.io/kube-scheduler-simulator/simulator/scheduler/storereflector"
)

// Service manages Extenders and the result.
type Service struct {
	client    clientset.Interface
	extenders []Extender
	store     resultstore.Store
}

const ResultStoreKey = "ExtenderResultStoreKey"

// New initializes Service.
// `extenderCfgs` expect to receive an untouched config file(set by user).
func New(client clientset.Interface, extenderCfgs []configv1.Extender, storeReflector storereflector.Reflector) (*Service, error) {
	extenders, err := createExtenders(extenderCfgs)
	if err != nil {
		return nil, xerrors.Errorf("create HTTPExtenders: %w", err)
	}
	store := resultstore.New()
	// Register the result store of Extenders to the sharedStore.
	storeReflector.AddResultStore(store, ResultStoreKey)
	return &Service{
		client:    client,
		extenders: extenders,
		store:     store,
	}, nil
}

// Filter returns the result of the specified filter extender
// and store it.
func (s *Service) Filter(id int, args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	result, err := s.extenders[id].Filter(args)
	if err != nil {
		return nil, xerrors.Errorf("call filter of specified HTTPExtender: %w", err)
	}
	s.store.AddFilterResult(args, *result, s.extenders[id].Name())
	return result, nil
}

// Prioritize returns the result of the specified prioritize extender
// and store it.
func (s *Service) Prioritize(id int, args extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error) {
	result, err := s.extenders[id].Prioritize(args)
	if err != nil {
		return nil, xerrors.Errorf("call prioritize of specified HTTPExtender: %w", err)
	}
	s.store.AddPrioritizeResult(args, *result, s.extenders[id].Name())
	return result, nil
}

// Preempt returns the result of the specified preempt extender
// and store it.
func (s *Service) Preempt(id int, args extenderv1.ExtenderPreemptionArgs) (*extenderv1.ExtenderPreemptionResult, error) {
	result, err := s.extenders[id].Preempt(args)
	if err != nil {
		return nil, xerrors.Errorf("call preempt of specified HTTPExtender: %w", err)
	}
	s.store.AddPreemptResult(args, *result, s.extenders[id].Name())
	return result, nil
}

// Bind returns the result of the specified bind extender
// and store it.
func (s *Service) Bind(id int, args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error) {
	result, err := s.extenders[id].Bind(args)
	if err != nil {
		return nil, xerrors.Errorf("call bind of specified HTTPExtender: %w", err)
	}
	s.store.AddBindResult(args, *result, s.extenders[id].Name())
	return result, nil
}

// OverrideExtendersCfgToSimulator rewrites the scheduler config so that the extenders requests go through the simulator server.
func OverrideExtendersCfgToSimulator(cfg *configv1.KubeSchedulerConfiguration, simulatorPort int) {
	for i := range cfg.Extenders {
		// i will be the extender's index. That index is specified by request param as `id`.
		cfg.Extenders[i].EnableHTTPS = false
		cfg.Extenders[i].TLSConfig = nil
		// NOTE: We do not plan to launch the "HTTPS" simulator server with echo on our project.
		// If you customize the server to use HTTPS with echo, you need to fix this line.
		cfg.Extenders[i].URLPrefix = "http://localhost:" + strconv.Itoa(simulatorPort) + "/api/v1/extender/"
		if cfg.Extenders[i].FilterVerb != "" {
			cfg.Extenders[i].FilterVerb = "filter/" + strconv.Itoa(i)
		}
		if cfg.Extenders[i].PrioritizeVerb != "" {
			cfg.Extenders[i].PrioritizeVerb = "prioritize/" + strconv.Itoa(i)
		}
		if cfg.Extenders[i].PreemptVerb != "" {
			cfg.Extenders[i].PreemptVerb = "preempt/" + strconv.Itoa(i)
		}
		if cfg.Extenders[i].BindVerb != "" {
			cfg.Extenders[i].BindVerb = "bind/" + strconv.Itoa(i)
		}
	}
}
