/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	flowcontrolbootstrap "k8s.io/apiserver/pkg/apis/flowcontrol/bootstrap"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	serverstorage "k8s.io/apiserver/pkg/server/storage"
	"k8s.io/client-go/informers"
	flowcontrolclient "k8s.io/client-go/kubernetes/typed/flowcontrol/v1beta2"
	flowcontrollisters "k8s.io/client-go/listers/flowcontrol/v1beta2"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/flowcontrol"
	flowcontrolapisv1alpha1 "k8s.io/kubernetes/pkg/apis/flowcontrol/v1alpha1"
	flowcontrolapisv1beta1 "k8s.io/kubernetes/pkg/apis/flowcontrol/v1beta1"
	flowcontrolapisv1beta2 "k8s.io/kubernetes/pkg/apis/flowcontrol/v1beta2"
	"k8s.io/kubernetes/pkg/registry/flowcontrol/ensurer"
	flowschemastore "k8s.io/kubernetes/pkg/registry/flowcontrol/flowschema/storage"
	prioritylevelconfigurationstore "k8s.io/kubernetes/pkg/registry/flowcontrol/prioritylevelconfiguration/storage"
)

var _ genericapiserver.PostStartHookProvider = RESTStorageProvider{}

// RESTStorageProvider is a provider of REST storage
type RESTStorageProvider struct {
	InformerFactory informers.SharedInformerFactory
}

// PostStartHookName is the name of the post-start-hook provided by flow-control storage
const PostStartHookName = "priority-and-fairness-config-producer"

// NewRESTStorage creates a new rest storage for flow-control api models.
func (p RESTStorageProvider) NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (genericapiserver.APIGroupInfo, error) {
	apiGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(flowcontrol.GroupName, legacyscheme.Scheme, legacyscheme.ParameterCodec, legacyscheme.Codecs)

	if storageMap, err := p.storage(apiResourceConfigSource, restOptionsGetter, flowcontrolapisv1alpha1.SchemeGroupVersion); err != nil {
		return genericapiserver.APIGroupInfo{}, err
	} else if len(storageMap) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap[flowcontrolapisv1alpha1.SchemeGroupVersion.Version] = storageMap
	}

	if storageMap, err := p.storage(apiResourceConfigSource, restOptionsGetter, flowcontrolapisv1beta1.SchemeGroupVersion); err != nil {
		return genericapiserver.APIGroupInfo{}, err
	} else if len(storageMap) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap[flowcontrolapisv1beta1.SchemeGroupVersion.Version] = storageMap
	}

	if storageMap, err := p.storage(apiResourceConfigSource, restOptionsGetter, flowcontrolapisv1beta2.SchemeGroupVersion); err != nil {
		return genericapiserver.APIGroupInfo{}, err
	} else if len(storageMap) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap[flowcontrolapisv1beta2.SchemeGroupVersion.Version] = storageMap
	}

	return apiGroupInfo, nil
}

func (p RESTStorageProvider) storage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter, groupVersion schema.GroupVersion) (map[string]rest.Storage, error) {
	storage := map[string]rest.Storage{}

	// flow-schema
	if resource := "flowschemas"; apiResourceConfigSource.ResourceEnabled(groupVersion.WithResource(resource)) {
		flowSchemaStorage, flowSchemaStatusStorage, err := flowschemastore.NewREST(restOptionsGetter)
		if err != nil {
			return nil, err
		}
		storage[resource] = flowSchemaStorage
		storage[resource+"/status"] = flowSchemaStatusStorage
	}

	// priority-level-configuration
	if resource := "prioritylevelconfigurations"; apiResourceConfigSource.ResourceEnabled(groupVersion.WithResource(resource)) {
		priorityLevelConfigurationStorage, priorityLevelConfigurationStatusStorage, err := prioritylevelconfigurationstore.NewREST(restOptionsGetter)
		if err != nil {
			return nil, err
		}
		storage[resource] = priorityLevelConfigurationStorage
		storage[resource+"/status"] = priorityLevelConfigurationStatusStorage
	}

	return storage, nil
}

// GroupName returns group name of the storage
func (p RESTStorageProvider) GroupName() string {
	return flowcontrol.GroupName
}

// PostStartHook returns the hook func that launches the config provider
func (p RESTStorageProvider) PostStartHook() (string, genericapiserver.PostStartHookFunc, error) {
	bce := &bootstrapConfigurationEnsurer{
		informersSynced: []cache.InformerSynced{
			p.InformerFactory.Flowcontrol().V1beta2().PriorityLevelConfigurations().Informer().HasSynced,
			p.InformerFactory.Flowcontrol().V1beta2().FlowSchemas().Informer().HasSynced,
		},
		fsLister:  p.InformerFactory.Flowcontrol().V1beta2().FlowSchemas().Lister(),
		plcLister: p.InformerFactory.Flowcontrol().V1beta2().PriorityLevelConfigurations().Lister(),
	}
	return PostStartHookName, bce.ensureAPFBootstrapConfiguration, nil
}

type bootstrapConfigurationEnsurer struct {
	informersSynced []cache.InformerSynced
	fsLister        flowcontrollisters.FlowSchemaLister
	plcLister       flowcontrollisters.PriorityLevelConfigurationLister
}

func (bce *bootstrapConfigurationEnsurer) ensureAPFBootstrapConfiguration(hookContext genericapiserver.PostStartHookContext) error {
	clientset, err := flowcontrolclient.NewForConfig(hookContext.LoopbackClientConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize clientset for APF - %w", err)
	}

	// get a derived context that gets cancelled after 5m or
	// when the StopCh gets closed, whichever happens first.
	ctx, cancel := contextFromChannelAndMaxWaitDuration(hookContext.StopCh, 5*time.Minute)
	defer cancel()

	if !cache.WaitForCacheSync(ctx.Done(), bce.informersSynced...) {
		return fmt.Errorf("APF bootstrap ensurer timed out waiting for cache sync")
	}

	err = wait.PollImmediateUntilWithContext(
		ctx,
		time.Second,
		func(context.Context) (bool, error) {
			if err := ensure(clientset, bce.fsLister, bce.plcLister); err != nil {
				klog.ErrorS(err, "APF bootstrap ensurer ran into error, will retry later")
				return false, nil
			}
			return true, nil
		})
	if err != nil {
		return fmt.Errorf("unable to initialize APF bootstrap configuration")
	}

	// we have successfully initialized the bootstrap configuration, now we
	// spin up a goroutine which reconciles the bootstrap configuration periodically.
	go func() {
		wait.PollImmediateUntil(
			time.Minute,
			func() (bool, error) {
				if err := ensure(clientset, bce.fsLister, bce.plcLister); err != nil {
					klog.ErrorS(err, "APF bootstrap ensurer ran into error, will retry later")
				}
				// always auto update both suggested and mandatory configuration
				return false, nil
			}, hookContext.StopCh)
		klog.Info("APF bootstrap ensurer is exiting")
	}()

	return nil
}

func ensure(clientset flowcontrolclient.FlowcontrolV1beta2Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	if err := ensureSuggestedConfiguration(clientset, fsLister, plcLister); err != nil {
		// We should not attempt creation of mandatory objects if ensuring the suggested
		// configuration resulted in an error.
		// This only happens when the stop channel is closed.
		return fmt.Errorf("failed ensuring suggested settings - %w", err)
	}

	if err := ensureMandatoryConfiguration(clientset, fsLister, plcLister); err != nil {
		return fmt.Errorf("failed ensuring mandatory settings - %w", err)
	}

	if err := removeDanglingBootstrapConfiguration(clientset, fsLister, plcLister); err != nil {
		return fmt.Errorf("failed to delete removed settings - %w", err)
	}

	return nil
}

func ensureSuggestedConfiguration(clientset flowcontrolclient.FlowcontrolV1beta2Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	plEnsurer := ensurer.NewSuggestedPriorityLevelEnsurerEnsurer(clientset.PriorityLevelConfigurations(), plcLister)
	if err := plEnsurer.Ensure(flowcontrolbootstrap.SuggestedPriorityLevelConfigurations); err != nil {
		return err
	}

	fsEnsurer := ensurer.NewSuggestedFlowSchemaEnsurer(clientset.FlowSchemas(), fsLister)
	return fsEnsurer.Ensure(flowcontrolbootstrap.SuggestedFlowSchemas)
}

func ensureMandatoryConfiguration(clientset flowcontrolclient.FlowcontrolV1beta2Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	fsEnsurer := ensurer.NewMandatoryFlowSchemaEnsurer(clientset.FlowSchemas(), fsLister)
	if err := fsEnsurer.Ensure(flowcontrolbootstrap.MandatoryFlowSchemas); err != nil {
		return err
	}

	plEnsurer := ensurer.NewMandatoryPriorityLevelEnsurer(clientset.PriorityLevelConfigurations(), plcLister)
	return plEnsurer.Ensure(flowcontrolbootstrap.MandatoryPriorityLevelConfigurations)
}

func removeDanglingBootstrapConfiguration(clientset flowcontrolclient.FlowcontrolV1beta2Interface, fsLister flowcontrollisters.FlowSchemaLister, plcLister flowcontrollisters.PriorityLevelConfigurationLister) error {
	if err := removeDanglingBootstrapFlowSchema(clientset.FlowSchemas(), fsLister); err != nil {
		return err
	}

	return removeDanglingBootstrapPriorityLevel(clientset.PriorityLevelConfigurations(), plcLister)
}

func removeDanglingBootstrapFlowSchema(client flowcontrolclient.FlowSchemaInterface, lister flowcontrollisters.FlowSchemaLister) error {
	bootstrap := append(flowcontrolbootstrap.MandatoryFlowSchemas, flowcontrolbootstrap.SuggestedFlowSchemas...)
	candidates, err := ensurer.GetFlowSchemaRemoveCandidates(lister, bootstrap)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	fsRemover := ensurer.NewFlowSchemaRemover(client, lister)
	return fsRemover.RemoveAutoUpdateEnabledObjects(candidates)
}

func removeDanglingBootstrapPriorityLevel(client flowcontrolclient.PriorityLevelConfigurationInterface, lister flowcontrollisters.PriorityLevelConfigurationLister) error {
	bootstrap := append(flowcontrolbootstrap.MandatoryPriorityLevelConfigurations, flowcontrolbootstrap.SuggestedPriorityLevelConfigurations...)
	candidates, err := ensurer.GetPriorityLevelRemoveCandidates(lister, bootstrap)
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	plRemover := ensurer.NewPriorityLevelRemover(client, lister)
	return plRemover.RemoveAutoUpdateEnabledObjects(candidates)
}

// contextFromChannelAndMaxWaitDuration returns a Context that is bound to the
// specified channel and the wait duration. The derived context will be
// cancelled when the specified channel stopCh is closed or the maximum wait
// duration specified in maxWait elapses, whichever happens first.
//
// Note the caller must *always* call the CancelFunc, otherwise resources may be leaked.
func contextFromChannelAndMaxWaitDuration(stopCh <-chan struct{}, maxWait time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

		select {
		case <-stopCh:
		case <-time.After(maxWait):

		// the caller can explicitly cancel the context which is an
		// indication to us to exit the goroutine immediately.
		// Note that we are calling cancel more than once when we are here,
		// CancelFunc is idempotent and we expect no ripple effects here.
		case <-ctx.Done():
		}
	}()
	return ctx, cancel
}
