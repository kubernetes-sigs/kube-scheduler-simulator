package extender

//go:generate mockgen -destination=./mock_$GOPACKAGE/extender.go . Extender

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"golang.org/x/xerrors"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/sets"
	restclient "k8s.io/client-go/rest"
	configv1 "k8s.io/kube-scheduler/config/v1"
	extenderv1 "k8s.io/kube-scheduler/extender/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

const (
	// DefaultExtenderTimeout defines the default extender timeout in second.
	DefaultExtenderTimeout = 5 * time.Second
)

// Extender provides methods to call the actual extender's endpoint set by user.
type Extender interface {
	Name() string
	Filter(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error)
	Prioritize(args extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error)
	Preempt(args extenderv1.ExtenderPreemptionArgs) (*extenderv1.ExtenderPreemptionResult, error)
	Bind(args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// extender represents an extender server.
type extender struct {
	extenderURL      string
	preemptVerb      string
	filterVerb       string
	prioritizeVerb   string
	bindVerb         string
	weight           int64
	client           httpClient
	nodeCacheCapable bool

	// https://github.com/kubernetes/kubernetes/blob/fc04e732bb3e7198d2fa44efa5457c7c6f8c0f5b/pkg/scheduler/extender.go#L51
	managedResources sets.Set[string]
}

// makeTransport makes http.Transport from the extender config.
func makeTransport(config *configv1.Extender) (http.RoundTripper, error) {
	var cfg restclient.Config
	if config.TLSConfig != nil {
		cfg.TLSClientConfig.Insecure = config.TLSConfig.Insecure
		cfg.TLSClientConfig.ServerName = config.TLSConfig.ServerName
		cfg.TLSClientConfig.CertFile = config.TLSConfig.CertFile
		cfg.TLSClientConfig.KeyFile = config.TLSConfig.KeyFile
		cfg.TLSClientConfig.CAFile = config.TLSConfig.CAFile
		cfg.TLSClientConfig.CertData = config.TLSConfig.CertData
		cfg.TLSClientConfig.KeyData = config.TLSConfig.KeyData
		cfg.TLSClientConfig.CAData = config.TLSConfig.CAData
	}
	if config.EnableHTTPS {
		hasCA := len(cfg.CAFile) > 0 || len(cfg.CAData) > 0
		if !hasCA {
			cfg.Insecure = true
		}
	}
	tlsConfig, err := restclient.TLSConfigFor(&cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		return utilnet.SetTransportDefaults(&http.Transport{
			TLSClientConfig: tlsConfig,
		}), nil
	}
	return utilnet.SetTransportDefaults(&http.Transport{}), nil
}

// newExtender creates an Extender object.
func newExtender(config *configv1.Extender) (Extender, error) {
	if config.HTTPTimeout.Duration.Nanoseconds() == 0 {
		config.HTTPTimeout.Duration = DefaultExtenderTimeout
	}

	transport, err := makeTransport(config)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   config.HTTPTimeout.Duration,
	}
	managedResources := sets.New[string]()
	for _, r := range config.ManagedResources {
		managedResources.Insert(r.Name)
	}
	return &extender{
		extenderURL:      config.URLPrefix,
		preemptVerb:      config.PreemptVerb,
		filterVerb:       config.FilterVerb,
		prioritizeVerb:   config.PrioritizeVerb,
		bindVerb:         config.BindVerb,
		weight:           config.Weight,
		client:           client,
		nodeCacheCapable: config.NodeCacheCapable,
		managedResources: managedResources,
	}, nil
}

// Name returns the extender URL as the server name.
func (e *extender) Name() string {
	return e.extenderURL
}

// Filter sends the request to the original extender server, and returns the response as is.
func (e *extender) Filter(args extenderv1.ExtenderArgs) (*extenderv1.ExtenderFilterResult, error) {
	var result extenderv1.ExtenderFilterResult
	if e.filterVerb == "" {
		return nil, xerrors.Errorf("filterVerb is empty")
	}
	if err := e.send(e.filterVerb, args, &result); err != nil {
		return nil, xerrors.Errorf("send filter request: %w", err)
	}
	return &result, nil
}

// Prioritize sends the request to the original extender server, and returns the response as is.
func (e *extender) Prioritize(args extenderv1.ExtenderArgs) (*extenderv1.HostPriorityList, error) {
	var result extenderv1.HostPriorityList
	if e.prioritizeVerb == "" {
		return nil, xerrors.Errorf("prioritizeVerb is empty")
	}
	if err := e.send(e.prioritizeVerb, args, &result); err != nil {
		return nil, xerrors.Errorf("send prioritize request: %w", err)
	}
	for i := range result {
		// MaxExtenderPriority may diverge from the max priority used in the scheduler and defined by MaxNodeScore,
		// therefore we need to scale the score returned by extenders to the score range used by the scheduler.
		result[i].Score = result[i].Score * e.weight * (framework.MaxNodeScore / extenderv1.MaxExtenderPriority)
	}
	return &result, nil
}

// Preempt sends the request to the original extender server, and returns the response as is.
func (e *extender) Preempt(args extenderv1.ExtenderPreemptionArgs) (*extenderv1.ExtenderPreemptionResult, error) {
	var result extenderv1.ExtenderPreemptionResult
	if e.preemptVerb == "" {
		return nil, xerrors.Errorf("preemptVerb is empty")
	}
	if err := e.send(e.preemptVerb, args, &result); err != nil {
		return nil, xerrors.Errorf("send preempt request: %w", err)
	}
	return &result, nil
}

// Bind sends the request to the original extender server, and returns the response as is.
func (e *extender) Bind(args extenderv1.ExtenderBindingArgs) (*extenderv1.ExtenderBindingResult, error) {
	var result extenderv1.ExtenderBindingResult
	if e.bindVerb == "" {
		return nil, xerrors.Errorf("bindVerb is empty")
	}
	if err := e.send(e.bindVerb, args, &result); err != nil {
		return nil, xerrors.Errorf("send bind request: %w", err)
	}
	return &result, nil
}

// Send is Helper function to send messages to the extender.
func (e *extender) send(action string, args interface{}, result interface{}) error {
	out, err := json.Marshal(args)
	if err != nil {
		return xerrors.Errorf("json Marshal: %w", err)
	}
	url := strings.TrimRight(e.extenderURL, "/") + "/" + action

	req, err := http.NewRequest("POST", url, bytes.NewReader(out))
	if err != nil {
		return xerrors.Errorf("http NewRequest: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return xerrors.Errorf("client Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return xerrors.Errorf("failed %v with extender at URL %v, code %v", action, url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// createExtenders creates Extender that represents actual extender's endpoint based on the config set by user.
func createExtenders(configs []configv1.Extender) ([]Extender, error) {
	if len(configs) == 0 {
		return nil, nil
	}
	extenders := make([]Extender, len(configs))
	for i := range configs {
		e, err := newExtender(&configs[i])
		if err != nil {
			return nil, xerrors.Errorf("failed newExtender: %w", err)
		}
		extenders[i] = e
	}
	return extenders, nil
}
