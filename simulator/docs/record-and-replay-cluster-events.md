# [Beta] Record your real cluster's events and replay them in the simulator

You can record events from your real cluster and replay them in the simulator. This feature is useful for reproducing issues that occur in your real cluster.

## Record events

To record events from your real cluster, you need to follow these steps:

1. Set `true` to `recorderEnabled`.
2. Set the path of the kubeconfig file for your cluster to `KubeConfig`.
   - This feature only requires the read permission for events.
3. Set the path of the file to save the recorded events to `recordedFilePath`.
4. Make sure the file path is mounted to the simulator container.

```yaml
recorderEnabled: true
kubeConfig: "/path/to/your-cluster-kubeconfig"
recordedFilePath: "/path/to/recorded-events.json"
```

```yaml
volumes:
  ...
  - ./path/to/recorded-events.json:/path/to/recorded-events.json
```

> [!NOTE]
> When a file already exists at `recordedFilePath`, it backs up the file in the same directory adding a timestamp to the filename and creates a new file for recording.

### Resources to record

It records the events of the following resources:

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

## Replay events

To replay the recorded events in the simulator, you need to follow these steps:

1. Set `true` to `replayerEnabled`.
2. Set the path of the file where the events are recorded to `recordedFilePath`.

```yaml
replayerEnabled: true
recordedFilePath: "/path/to/recorded-events.json"
```

### Resources to replay

It replays the events of the following resources:

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
