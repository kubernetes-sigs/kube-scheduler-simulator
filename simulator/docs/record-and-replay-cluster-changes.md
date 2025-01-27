# [Beta] Record your real cluster's changes in resources and replay them in the simulator

You can record resource addition/update/deletion at your real cluster. This feature is useful for reproducing issues that occur in your real cluster.

## Record changes

To record changes from your real cluster, you need to follow these steps:

1. Set `true` to `recorderEnabled`.
2. Set the path of the kubeconfig file for your cluster to `KubeConfig`.
  - This feature only requires the read permission for resources.
3. Set the path of the file to save the recorded changes to `recordedFilePath`.
4. Make sure the file path is mounted to the simulator container.

```yaml
recorderEnabled: true
kubeConfig: "/path/to/your-cluster-kubeconfig"
recordedFilePath: "/path/to/recorded-changes.json"
```

```yaml
volumes:
  ...
  - ./path/to/recorded-changes.json:/path/to/recorded-changes.json
```

> [!NOTE]
> When a file already exists at `recordedFilePath`, it puts out an error.

### Resources to record

It records the changes of the following resources:

- Pods
- Nodes
- PersistentVolumes
- PersistentVolumeClaims
- StorageClasses

You can tweak which resources to record via the option in [/simulator/cmd/simulator/simulator.go](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/master/simulator/cmd/simulator/simulator.go):

```go
recorderOptions := recorder.Options{Path: cfg.RecordFilePath,
// GVRs is a list of GroupVersionResource that will be recorded.
// If it's nil, DefaultGVRs are used.
	GVRs: []schema.GroupVersionResource{
		{Group: "your-group", Version: "v1", Resource: "your-custom-resources"},
	},
}
```

## Replay changes

To replay the recorded changes in the simulator, you need to follow these steps:

1. Set `true` to `replayerEnabled`.
2. Set the path of the file where the changes are recorded to `recordedFilePath`.

```yaml
replayerEnabled: true
recordedFilePath: "/path/to/recorded-changes.json"
```

### Resources to replay

It replays the changes of the following resources:

- Pods
- Nodes
- PersistentVolumes
- PersistentVolumeClaims
- StorageClasses

You can tweak which resources to replay via the option in [/simulator/cmd/simulator/simulator.go](https://github.com/kubernetes-sigs/kube-scheduler-simulator/blob/master/simulator/cmd/simulator/simulator.go):

```go
resourceApplierOptions := resourceapplier.Options{
	// GVRsToApply is a list of GroupVersionResource that will be replayed.
	// If GVRsToApply is nil, defaultGVRs are used.
	GVRsToApply: []schema.GroupVersionResource{
		{Group: "your-group", Version: "v1", Resource: "your-custom-resources"},
	},

	// Actually, more options are available...

	// FilterBeforeCreating is a list of additional filtering functions that are applied before creating resources.
	FilterBeforeCreating: map[schema.GroupVersionResource][]resourceapplier.FilteringFunction{},
	// MutateBeforeCreating is a list of additional mutating functions that are applied before creating resources.
	MutateBeforeCreating: map[schema.GroupVersionResource][]resourceapplier.MutatingFunction{},
	// FilterBeforeUpdating is a list of additional filtering functions that are applied before updating resources.
	FilterBeforeUpdating: map[schema.GroupVersionResource][]resourceapplier.FilteringFunction{},
	// MutateBeforeUpdating is a list of additional mutating functions that are applied before updating resources.
	MutateBeforeUpdating: map[schema.GroupVersionResource][]resourceapplier.MutatingFunction{},
}
```
