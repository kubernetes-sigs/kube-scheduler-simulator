/*
Copyright 2016 The Kubernetes Authors.

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

package persistentvolume

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	coreinformers "k8s.io/client-go/informers/core/v1"
	storageinformers "k8s.io/client-go/informers/storage/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	cloudprovider "k8s.io/cloud-provider"
	storagehelpers "k8s.io/component-helpers/storage/volume"
	csitrans "k8s.io/csi-translation-lib"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/volume/common"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume/metrics"
	"k8s.io/kubernetes/pkg/features"
	proxyutil "k8s.io/kubernetes/pkg/proxy/util"
	"k8s.io/kubernetes/pkg/util/goroutinemap"
	"k8s.io/kubernetes/pkg/util/slice"
	vol "k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/csimigration"

	"k8s.io/klog/v2"
)

// This file contains the controller base functionality, i.e. framework to
// process PV/PVC added/updated/deleted events. The real binding, provisioning,
// recycling and deleting is done in pv_controller.go

// ControllerParameters contains arguments for creation of a new
// PersistentVolume controller.
type ControllerParameters struct {
	KubeClient                clientset.Interface
	SyncPeriod                time.Duration
	VolumePlugins             []vol.VolumePlugin
	Cloud                     cloudprovider.Interface
	ClusterName               string
	VolumeInformer            coreinformers.PersistentVolumeInformer
	ClaimInformer             coreinformers.PersistentVolumeClaimInformer
	ClassInformer             storageinformers.StorageClassInformer
	PodInformer               coreinformers.PodInformer
	NodeInformer              coreinformers.NodeInformer
	EventRecorder             record.EventRecorder
	EnableDynamicProvisioning bool
	FilteredDialOptions       *proxyutil.FilteredDialOptions
}

// NewController creates a new PersistentVolume controller
func NewController(p ControllerParameters) (*PersistentVolumeController, error) {
	eventRecorder := p.EventRecorder
	if eventRecorder == nil {
		broadcaster := record.NewBroadcaster()
		broadcaster.StartStructuredLogging(0)
		broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: p.KubeClient.CoreV1().Events("")})
		eventRecorder = broadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: "persistentvolume-controller"})
	}

	controller := &PersistentVolumeController{
		volumes:                       newPersistentVolumeOrderedIndex(),
		claims:                        cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc),
		kubeClient:                    p.KubeClient,
		eventRecorder:                 eventRecorder,
		runningOperations:             goroutinemap.NewGoRoutineMap(true /* exponentialBackOffOnError */),
		cloud:                         p.Cloud,
		enableDynamicProvisioning:     p.EnableDynamicProvisioning,
		clusterName:                   p.ClusterName,
		createProvisionedPVRetryCount: createProvisionedPVRetryCount,
		createProvisionedPVInterval:   createProvisionedPVInterval,
		claimQueue:                    workqueue.NewNamed("claims"),
		volumeQueue:                   workqueue.NewNamed("volumes"),
		resyncPeriod:                  p.SyncPeriod,
		operationTimestamps:           metrics.NewOperationStartTimeCache(),
	}

	// Prober is nil because PV is not aware of Flexvolume.
	if err := controller.volumePluginMgr.InitPlugins(p.VolumePlugins, nil /* prober */, controller); err != nil {
		return nil, fmt.Errorf("could not initialize volume plugins for PersistentVolume Controller: %w", err)
	}

	p.VolumeInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { controller.enqueueWork(controller.volumeQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { controller.enqueueWork(controller.volumeQueue, newObj) },
			DeleteFunc: func(obj interface{}) { controller.enqueueWork(controller.volumeQueue, obj) },
		},
	)
	controller.volumeLister = p.VolumeInformer.Lister()
	controller.volumeListerSynced = p.VolumeInformer.Informer().HasSynced

	p.ClaimInformer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    func(obj interface{}) { controller.enqueueWork(controller.claimQueue, obj) },
			UpdateFunc: func(oldObj, newObj interface{}) { controller.enqueueWork(controller.claimQueue, newObj) },
			DeleteFunc: func(obj interface{}) { controller.enqueueWork(controller.claimQueue, obj) },
		},
	)
	controller.claimLister = p.ClaimInformer.Lister()
	controller.claimListerSynced = p.ClaimInformer.Informer().HasSynced

	controller.classLister = p.ClassInformer.Lister()
	controller.classListerSynced = p.ClassInformer.Informer().HasSynced
	controller.podLister = p.PodInformer.Lister()
	controller.podIndexer = p.PodInformer.Informer().GetIndexer()
	controller.podListerSynced = p.PodInformer.Informer().HasSynced
	controller.NodeLister = p.NodeInformer.Lister()
	controller.NodeListerSynced = p.NodeInformer.Informer().HasSynced

	// This custom indexer will index pods by its PVC keys. Then we don't need
	// to iterate all pods every time to find pods which reference given PVC.
	if err := common.AddPodPVCIndexerIfNotPresent(controller.podIndexer); err != nil {
		return nil, fmt.Errorf("could not initialize attach detach controller: %w", err)
	}

	csiTranslator := csitrans.New()
	controller.translator = csiTranslator
	controller.csiMigratedPluginManager = csimigration.NewPluginManager(csiTranslator, utilfeature.DefaultFeatureGate)

	controller.filteredDialOptions = p.FilteredDialOptions

	return controller, nil
}

// initializeCaches fills all controller caches with initial data from etcd in
// order to have the caches already filled when first addClaim/addVolume to
// perform initial synchronization of the controller.
func (ctrl *PersistentVolumeController) initializeCaches(volumeLister corelisters.PersistentVolumeLister, claimLister corelisters.PersistentVolumeClaimLister) {
	volumeList, err := volumeLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("PersistentVolumeController can't initialize caches: %v", err)
		return
	}
	for _, volume := range volumeList {
		volumeClone := volume.DeepCopy()
		if _, err = ctrl.storeVolumeUpdate(volumeClone); err != nil {
			klog.Errorf("error updating volume cache: %v", err)
		}
	}

	claimList, err := claimLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("PersistentVolumeController can't initialize caches: %v", err)
		return
	}
	for _, claim := range claimList {
		if _, err = ctrl.storeClaimUpdate(claim.DeepCopy()); err != nil {
			klog.Errorf("error updating claim cache: %v", err)
		}
	}
	klog.V(4).Infof("controller initialized")
}

// enqueueWork adds volume or claim to given work queue.
func (ctrl *PersistentVolumeController) enqueueWork(queue workqueue.Interface, obj interface{}) {
	// Beware of "xxx deleted" events
	if unknown, ok := obj.(cache.DeletedFinalStateUnknown); ok && unknown.Obj != nil {
		obj = unknown.Obj
	}
	objName, err := controller.KeyFunc(obj)
	if err != nil {
		klog.Errorf("failed to get key from object: %v", err)
		return
	}
	klog.V(5).Infof("enqueued %q for sync", objName)
	queue.Add(objName)
}

func (ctrl *PersistentVolumeController) storeVolumeUpdate(volume interface{}) (bool, error) {
	return storeObjectUpdate(ctrl.volumes.store, volume, "volume")
}

func (ctrl *PersistentVolumeController) storeClaimUpdate(claim interface{}) (bool, error) {
	return storeObjectUpdate(ctrl.claims, claim, "claim")
}

// updateVolume runs in worker thread and handles "volume added",
// "volume updated" and "periodic sync" events.
func (ctrl *PersistentVolumeController) updateVolume(ctx context.Context, volume *v1.PersistentVolume) {
	// Store the new volume version in the cache and do not process it if this
	// is an old version.
	new, err := ctrl.storeVolumeUpdate(volume)
	if err != nil {
		klog.Errorf("%v", err)
	}
	if !new {
		return
	}

	err = ctrl.syncVolume(ctx, volume)
	if err != nil {
		if errors.IsConflict(err) {
			// Version conflict error happens quite often and the controller
			// recovers from it easily.
			klog.V(3).Infof("could not sync volume %q: %+v", volume.Name, err)
		} else {
			klog.Errorf("could not sync volume %q: %+v", volume.Name, err)
		}
	}
}

// deleteVolume runs in worker thread and handles "volume deleted" event.
func (ctrl *PersistentVolumeController) deleteVolume(volume *v1.PersistentVolume) {
	if err := ctrl.volumes.store.Delete(volume); err != nil {
		klog.Errorf("volume %q deletion encountered : %v", volume.Name, err)
	} else {
		klog.V(4).Infof("volume %q deleted", volume.Name)
	}
	// record deletion metric if a deletion start timestamp is in the cache
	// the following calls will be a no-op if there is nothing for this volume in the cache
	// end of timestamp cache entry lifecycle, "RecordMetric" will do the clean
	metrics.RecordMetric(volume.Name, &ctrl.operationTimestamps, nil)

	if volume.Spec.ClaimRef == nil {
		return
	}
	// sync the claim when its volume is deleted. Explicitly syncing the
	// claim here in response to volume deletion prevents the claim from
	// waiting until the next sync period for its Lost status.
	claimKey := claimrefToClaimKey(volume.Spec.ClaimRef)
	klog.V(5).Infof("deleteVolume[%s]: scheduling sync of claim %q", volume.Name, claimKey)
	ctrl.claimQueue.Add(claimKey)
}

// updateClaim runs in worker thread and handles "claim added",
// "claim updated" and "periodic sync" events.
func (ctrl *PersistentVolumeController) updateClaim(ctx context.Context, claim *v1.PersistentVolumeClaim) {
	// Store the new claim version in the cache and do not process it if this is
	// an old version.
	new, err := ctrl.storeClaimUpdate(claim)
	if err != nil {
		klog.Errorf("%v", err)
	}
	if !new {
		return
	}
	err = ctrl.syncClaim(ctx, claim)
	if err != nil {
		if errors.IsConflict(err) {
			// Version conflict error happens quite often and the controller
			// recovers from it easily.
			klog.V(3).Infof("could not sync claim %q: %+v", claimToClaimKey(claim), err)
		} else {
			klog.Errorf("could not sync volume %q: %+v", claimToClaimKey(claim), err)
		}
	}
}

// Unit test [5-5] [5-6] [5-7]
// deleteClaim runs in worker thread and handles "claim deleted" event.
func (ctrl *PersistentVolumeController) deleteClaim(claim *v1.PersistentVolumeClaim) {
	if err := ctrl.claims.Delete(claim); err != nil {
		klog.Errorf("claim %q deletion encountered : %v", claim.Name, err)
	}
	claimKey := claimToClaimKey(claim)
	klog.V(4).Infof("claim %q deleted", claimKey)
	// clean any possible unfinished provision start timestamp from cache
	// Unit test [5-8] [5-9]
	ctrl.operationTimestamps.Delete(claimKey)

	volumeName := claim.Spec.VolumeName
	if volumeName == "" {
		klog.V(5).Infof("deleteClaim[%q]: volume not bound", claimKey)
		return
	}

	// sync the volume when its claim is deleted.  Explicitly sync'ing the
	// volume here in response to claim deletion prevents the volume from
	// waiting until the next sync period for its Release.
	klog.V(5).Infof("deleteClaim[%q]: scheduling sync of volume %s", claimKey, volumeName)
	ctrl.volumeQueue.Add(volumeName)
}

// Run starts all of this controller's control loops
func (ctrl *PersistentVolumeController) Run(ctx context.Context) {
	defer utilruntime.HandleCrash()
	defer ctrl.claimQueue.ShutDown()
	defer ctrl.volumeQueue.ShutDown()

	klog.Infof("Starting persistent volume controller")
	defer klog.Infof("Shutting down persistent volume controller")

	if !cache.WaitForNamedCacheSync("persistent volume", ctx.Done(), ctrl.volumeListerSynced, ctrl.claimListerSynced, ctrl.classListerSynced, ctrl.podListerSynced, ctrl.NodeListerSynced) {
		return
	}

	ctrl.initializeCaches(ctrl.volumeLister, ctrl.claimLister)

	go wait.Until(ctrl.resync, ctrl.resyncPeriod, ctx.Done())
	go wait.UntilWithContext(ctx, ctrl.volumeWorker, time.Second)
	go wait.UntilWithContext(ctx, ctrl.claimWorker, time.Second)

	metrics.Register(ctrl.volumes.store, ctrl.claims, &ctrl.volumePluginMgr)

	<-ctx.Done()
}

func (ctrl *PersistentVolumeController) updateClaimMigrationAnnotations(ctx context.Context,
	claim *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	// TODO: update[Claim|Volume]MigrationAnnotations can be optimized to not
	// copy the claim/volume if no modifications are required. Though this
	// requires some refactoring as well as an interesting change in the
	// semantics of the function which may be undesirable. If no copy is made
	// when no modifications are required this function could sometimes return a
	// copy of the volume and sometimes return a ref to the original
	claimClone := claim.DeepCopy()
	modified := updateMigrationAnnotations(ctrl.csiMigratedPluginManager, ctrl.translator, claimClone.Annotations, true)
	if !modified {
		return claimClone, nil
	}
	newClaim, err := ctrl.kubeClient.CoreV1().PersistentVolumeClaims(claimClone.Namespace).Update(ctx, claimClone, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("persistent Volume Controller can't anneal migration annotations: %v", err)
	}
	_, err = ctrl.storeClaimUpdate(newClaim)
	if err != nil {
		return nil, fmt.Errorf("persistent Volume Controller can't anneal migration annotations: %v", err)
	}
	return newClaim, nil
}

func (ctrl *PersistentVolumeController) updateVolumeMigrationAnnotationsAndFinalizers(ctx context.Context,
	volume *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	volumeClone := volume.DeepCopy()
	annModified := updateMigrationAnnotations(ctrl.csiMigratedPluginManager, ctrl.translator, volumeClone.Annotations, false)
	modifiedFinalizers, finalizersModified := modifyDeletionFinalizers(ctrl.csiMigratedPluginManager, volumeClone)
	if !annModified && !finalizersModified {
		return volumeClone, nil
	}
	if finalizersModified {
		volumeClone.ObjectMeta.SetFinalizers(modifiedFinalizers)
	}
	newVol, err := ctrl.kubeClient.CoreV1().PersistentVolumes().Update(ctx, volumeClone, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("persistent Volume Controller can't anneal migration annotations or finalizer: %v", err)
	}
	_, err = ctrl.storeVolumeUpdate(newVol)
	if err != nil {
		return nil, fmt.Errorf("persistent Volume Controller can't anneal migration annotations or finalizer: %v", err)
	}
	return newVol, nil
}

// modifyDeletionFinalizers updates the finalizers based on the reclaim policy and if it is a in-tree volume or not.
// The in-tree PV deletion protection finalizer is only added if the reclaimPolicy associated with the PV is `Delete`.
// The in-tree PV deletion protection finalizer is removed if the reclaimPolicy associated with the PV is `Retain` or
// `Recycle`, removing the finalizer is necessary to reflect the recalimPolicy updates on the PV.
// The method also removes any external PV Deletion Protection finalizers added on the PV, this represents CSI migration
// rollback/disable scenarios.
func modifyDeletionFinalizers(cmpm CSIMigratedPluginManager, volume *v1.PersistentVolume) ([]string, bool) {
	modified := false
	var outFinalizers []string
	if !utilfeature.DefaultFeatureGate.Enabled(features.HonorPVReclaimPolicy) {
		return volume.Finalizers, false
	}
	if !metav1.HasAnnotation(volume.ObjectMeta, storagehelpers.AnnDynamicallyProvisioned) {
		// PV deletion protection finalizer is currently supported only for dynamically
		// provisioned volumes.
		return volume.Finalizers, false
	}
	if volume.Finalizers != nil {
		outFinalizers = append(outFinalizers, volume.Finalizers...)
	}
	provisioner := volume.Annotations[storagehelpers.AnnDynamicallyProvisioned]
	if cmpm.IsMigrationEnabledForPlugin(provisioner) {
		// Remove in-tree delete finalizer on the PV as migration is enabled.
		if slice.ContainsString(outFinalizers, storagehelpers.PVDeletionInTreeProtectionFinalizer, nil) {
			outFinalizers = slice.RemoveString(outFinalizers, storagehelpers.PVDeletionInTreeProtectionFinalizer, nil)
			modified = true
		}
		return outFinalizers, modified
	}
	// Check if it is a in-tree volume.
	if !strings.HasPrefix(provisioner, "kubernetes.io/") {
		// The provision plugin does not begin with known in-tree plugin volume prefix annotation.
		return volume.Finalizers, false
	}
	reclaimPolicy := volume.Spec.PersistentVolumeReclaimPolicy
	// Add back the in-tree PV deletion protection finalizer if does not already exists
	if reclaimPolicy == v1.PersistentVolumeReclaimDelete && !slice.ContainsString(outFinalizers, storagehelpers.PVDeletionInTreeProtectionFinalizer, nil) {
		klog.V(4).Infof("Adding in-tree pv deletion protection finalizer on %s", volume.Name)
		outFinalizers = append(outFinalizers, storagehelpers.PVDeletionInTreeProtectionFinalizer)
		modified = true
	} else if (reclaimPolicy == v1.PersistentVolumeReclaimRetain || reclaimPolicy == v1.PersistentVolumeReclaimRecycle) && slice.ContainsString(outFinalizers, storagehelpers.PVDeletionInTreeProtectionFinalizer, nil) {
		// Remove the in-tree PV deletion protection finalizer if the reclaim policy is 'Retain' or 'Recycle'
		klog.V(4).Infof("Removing in-tree pv deletion protection finalizer on %s", volume.Name)
		outFinalizers = slice.RemoveString(outFinalizers, storagehelpers.PVDeletionInTreeProtectionFinalizer, nil)
		modified = true
	}
	// Remove the external PV deletion protection finalizer
	if slice.ContainsString(outFinalizers, storagehelpers.PVDeletionProtectionFinalizer, nil) {
		klog.V(4).Infof("Removing external pv deletion protection finalizer on %s", volume.Name)
		outFinalizers = slice.RemoveString(outFinalizers, storagehelpers.PVDeletionProtectionFinalizer, nil)
		modified = true
	}
	return outFinalizers, modified
}

// updateMigrationAnnotations takes an Annotations map and checks for a
// provisioner name using the provisionerKey. It will then add a
// "pv.kubernetes.io/migrated-to" annotation if migration with the CSI
// driver name for that provisioner is "on" based on feature flags, it will also
// remove the annotation is migration is "off" for that provisioner in rollback
// scenarios. Returns true if the annotations map was modified and false otherwise.
func updateMigrationAnnotations(cmpm CSIMigratedPluginManager, translator CSINameTranslator, ann map[string]string, claim bool) bool {
	var csiDriverName string
	var err error

	if ann == nil {
		// No annotations so we can't get the provisioner and don't know whether
		// this is migrated - no change
		return false
	}
	var provisionerKey string
	if claim {
		provisionerKey = storagehelpers.AnnStorageProvisioner
	} else {
		provisionerKey = storagehelpers.AnnDynamicallyProvisioned
	}
	provisioner, ok := ann[provisionerKey]
	if !ok {
		if claim {
			// Also check beta AnnStorageProvisioner annontation to make sure
			provisioner, ok = ann[storagehelpers.AnnBetaStorageProvisioner]
			if !ok {
				return false
			}
		} else {
			// Volume Statically provisioned.
			return false
		}
	}

	migratedToDriver := ann[storagehelpers.AnnMigratedTo]
	if cmpm.IsMigrationEnabledForPlugin(provisioner) {
		csiDriverName, err = translator.GetCSINameFromInTreeName(provisioner)
		if err != nil {
			klog.Errorf("Could not update volume migration annotations. Migration enabled for plugin %s but could not find corresponding driver name: %v", provisioner, err)
			return false
		}
		if migratedToDriver != csiDriverName {
			ann[storagehelpers.AnnMigratedTo] = csiDriverName
			return true
		}
	} else {
		if migratedToDriver != "" {
			// Migration annotation exists but the driver isn't migrated currently
			delete(ann, storagehelpers.AnnMigratedTo)
			return true
		}
	}
	return false
}

// volumeWorker processes items from volumeQueue. It must run only once,
// syncVolume is not assured to be reentrant.
func (ctrl *PersistentVolumeController) volumeWorker(ctx context.Context) {
	workFunc := func(ctx context.Context) bool {
		keyObj, quit := ctrl.volumeQueue.Get()
		if quit {
			return true
		}
		defer ctrl.volumeQueue.Done(keyObj)
		key := keyObj.(string)
		klog.V(5).Infof("volumeWorker[%s]", key)

		_, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			klog.V(4).Infof("error getting name of volume %q to get volume from informer: %v", key, err)
			return false
		}
		volume, err := ctrl.volumeLister.Get(name)
		if err == nil {
			// The volume still exists in informer cache, the event must have
			// been add/update/sync
			ctrl.updateVolume(ctx, volume)
			return false
		}
		if !errors.IsNotFound(err) {
			klog.V(2).Infof("error getting volume %q from informer: %v", key, err)
			return false
		}

		// The volume is not in informer cache, the event must have been
		// "delete"
		volumeObj, found, err := ctrl.volumes.store.GetByKey(key)
		if err != nil {
			klog.V(2).Infof("error getting volume %q from cache: %v", key, err)
			return false
		}
		if !found {
			// The controller has already processed the delete event and
			// deleted the volume from its cache
			klog.V(2).Infof("deletion of volume %q was already processed", key)
			return false
		}
		volume, ok := volumeObj.(*v1.PersistentVolume)
		if !ok {
			klog.Errorf("expected volume, got %+v", volumeObj)
			return false
		}
		ctrl.deleteVolume(volume)
		return false
	}
	for {
		if quit := workFunc(ctx); quit {
			klog.Infof("volume worker queue shutting down")
			return
		}
	}
}

// claimWorker processes items from claimQueue. It must run only once,
// syncClaim is not reentrant.
func (ctrl *PersistentVolumeController) claimWorker(ctx context.Context) {
	workFunc := func() bool {
		keyObj, quit := ctrl.claimQueue.Get()
		if quit {
			return true
		}
		defer ctrl.claimQueue.Done(keyObj)
		key := keyObj.(string)
		klog.V(5).Infof("claimWorker[%s]", key)

		namespace, name, err := cache.SplitMetaNamespaceKey(key)
		if err != nil {
			klog.V(4).Infof("error getting namespace & name of claim %q to get claim from informer: %v", key, err)
			return false
		}
		claim, err := ctrl.claimLister.PersistentVolumeClaims(namespace).Get(name)
		if err == nil {
			// The claim still exists in informer cache, the event must have
			// been add/update/sync
			ctrl.updateClaim(ctx, claim)
			return false
		}
		if !errors.IsNotFound(err) {
			klog.V(2).Infof("error getting claim %q from informer: %v", key, err)
			return false
		}

		// The claim is not in informer cache, the event must have been "delete"
		claimObj, found, err := ctrl.claims.GetByKey(key)
		if err != nil {
			klog.V(2).Infof("error getting claim %q from cache: %v", key, err)
			return false
		}
		if !found {
			// The controller has already processed the delete event and
			// deleted the claim from its cache
			klog.V(2).Infof("deletion of claim %q was already processed", key)
			return false
		}
		claim, ok := claimObj.(*v1.PersistentVolumeClaim)
		if !ok {
			klog.Errorf("expected claim, got %+v", claimObj)
			return false
		}
		ctrl.deleteClaim(claim)
		return false
	}
	for {
		if quit := workFunc(); quit {
			klog.Infof("claim worker queue shutting down")
			return
		}
	}
}

// resync supplements short resync period of shared informers - we don't want
// all consumers of PV/PVC shared informer to have a short resync period,
// therefore we do our own.
func (ctrl *PersistentVolumeController) resync() {
	klog.V(4).Infof("resyncing PV controller")

	pvcs, err := ctrl.claimLister.List(labels.NewSelector())
	if err != nil {
		klog.Warningf("cannot list claims: %s", err)
		return
	}
	for _, pvc := range pvcs {
		ctrl.enqueueWork(ctrl.claimQueue, pvc)
	}

	pvs, err := ctrl.volumeLister.List(labels.NewSelector())
	if err != nil {
		klog.Warningf("cannot list persistent volumes: %s", err)
		return
	}
	for _, pv := range pvs {
		ctrl.enqueueWork(ctrl.volumeQueue, pv)
	}
}

// setClaimProvisioner saves
// claim.Annotations["volume.kubernetes.io/storage-provisioner"] = class.Provisioner
func (ctrl *PersistentVolumeController) setClaimProvisioner(ctx context.Context, claim *v1.PersistentVolumeClaim, provisionerName string) (*v1.PersistentVolumeClaim, error) {
	if val, ok := claim.Annotations[storagehelpers.AnnStorageProvisioner]; ok && val == provisionerName {
		// annotation is already set, nothing to do
		return claim, nil
	}

	// The volume from method args can be pointing to watcher cache. We must not
	// modify these, therefore create a copy.
	claimClone := claim.DeepCopy()
	// TODO: remove the beta storage provisioner anno after the deprecation period
	metav1.SetMetaDataAnnotation(&claimClone.ObjectMeta, storagehelpers.AnnBetaStorageProvisioner, provisionerName)
	metav1.SetMetaDataAnnotation(&claimClone.ObjectMeta, storagehelpers.AnnStorageProvisioner, provisionerName)
	updateMigrationAnnotations(ctrl.csiMigratedPluginManager, ctrl.translator, claimClone.Annotations, true)
	newClaim, err := ctrl.kubeClient.CoreV1().PersistentVolumeClaims(claim.Namespace).Update(context.TODO(), claimClone, metav1.UpdateOptions{})
	if err != nil {
		return newClaim, err
	}
	_, err = ctrl.storeClaimUpdate(newClaim)
	if err != nil {
		return newClaim, err
	}
	return newClaim, nil
}

// Stateless functions

func getClaimStatusForLogging(claim *v1.PersistentVolumeClaim) string {
	bound := metav1.HasAnnotation(claim.ObjectMeta, storagehelpers.AnnBindCompleted)
	boundByController := metav1.HasAnnotation(claim.ObjectMeta, storagehelpers.AnnBoundByController)

	return fmt.Sprintf("phase: %s, bound to: %q, bindCompleted: %v, boundByController: %v", claim.Status.Phase, claim.Spec.VolumeName, bound, boundByController)
}

func getVolumeStatusForLogging(volume *v1.PersistentVolume) string {
	boundByController := metav1.HasAnnotation(volume.ObjectMeta, storagehelpers.AnnBoundByController)
	claimName := ""
	if volume.Spec.ClaimRef != nil {
		claimName = fmt.Sprintf("%s/%s (uid: %s)", volume.Spec.ClaimRef.Namespace, volume.Spec.ClaimRef.Name, volume.Spec.ClaimRef.UID)
	}
	return fmt.Sprintf("phase: %s, bound to: %q, boundByController: %v", volume.Status.Phase, claimName, boundByController)
}

// storeObjectUpdate updates given cache with a new object version from Informer
// callback (i.e. with events from etcd) or with an object modified by the
// controller itself. Returns "true", if the cache was updated, false if the
// object is an old version and should be ignored.
func storeObjectUpdate(store cache.Store, obj interface{}, className string) (bool, error) {
	objName, err := controller.KeyFunc(obj)
	if err != nil {
		return false, fmt.Errorf("couldn't get key for object %+v: %w", obj, err)
	}
	oldObj, found, err := store.Get(obj)
	if err != nil {
		return false, fmt.Errorf("error finding %s %q in controller cache: %w", className, objName, err)
	}

	objAccessor, err := meta.Accessor(obj)
	if err != nil {
		return false, err
	}

	if !found {
		// This is a new object
		klog.V(4).Infof("storeObjectUpdate: adding %s %q, version %s", className, objName, objAccessor.GetResourceVersion())
		if err = store.Add(obj); err != nil {
			return false, fmt.Errorf("error adding %s %q to controller cache: %w", className, objName, err)
		}
		return true, nil
	}

	oldObjAccessor, err := meta.Accessor(oldObj)
	if err != nil {
		return false, err
	}

	objResourceVersion, err := strconv.ParseInt(objAccessor.GetResourceVersion(), 10, 64)
	if err != nil {
		return false, fmt.Errorf("error parsing ResourceVersion %q of %s %q: %s", objAccessor.GetResourceVersion(), className, objName, err)
	}
	oldObjResourceVersion, err := strconv.ParseInt(oldObjAccessor.GetResourceVersion(), 10, 64)
	if err != nil {
		return false, fmt.Errorf("error parsing old ResourceVersion %q of %s %q: %s", oldObjAccessor.GetResourceVersion(), className, objName, err)
	}

	// Throw away only older version, let the same version pass - we do want to
	// get periodic sync events.
	if oldObjResourceVersion > objResourceVersion {
		klog.V(4).Infof("storeObjectUpdate: ignoring %s %q version %s", className, objName, objAccessor.GetResourceVersion())
		return false, nil
	}

	klog.V(4).Infof("storeObjectUpdate updating %s %q with version %s", className, objName, objAccessor.GetResourceVersion())
	if err = store.Update(obj); err != nil {
		return false, fmt.Errorf("error updating %s %q in controller cache: %w", className, objName, err)
	}
	return true, nil
}
