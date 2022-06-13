# KEP-140: Scenario-based simulation

## Summary

A new scenario-based simulation feature is introduced to kube-scheduler-simulator by new `Scenario` CRD.

## Motivation

Nowadays, the scheduler is extendable with the multiple way:
- configure with [KubeSchedulerConfiguration](https://kubernetes.io/docs/reference/scheduling/config/)
- add Plugins of [Scheduling Framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)
- add [Extenders](https://github.com/kubernetes/enhancements/tree/5320deb4834c05ad9fb491dcd361f952727ece3e/keps/sig-scheduling/1819-scheduler-extender)
- etc...

But, unfortunately, not all expansions yield good results.
Those who customize the scheduler need to make sure their scheduler are working as expected, and doesn't have an unacceptably negative impact on the scheduling result or scheduling performance. And, usually, evaluating the scheduler is not so easy because there are many factors for the evaluation of the scheduler's ability.

The scenario-based simulation feature will be useful for those who customize the scheduler to evaluate their scheduler.

### Goals

Users can simulate their scheduler with some defined scenarios and can evaluate their scheduler with the result.

### Non-Goals

See the result of scenario-based simulation from Web UI. (may be implemented in the future, but out of scope of this proposal.)

## Proposal

### Implementation design details

#### the current simulator and proposal

The simulator was initially designed with a strong emphasis on Web UI. 
Then, thanks to so much contributions from everyone, we've expanded the simulator to be able to be used from other clients like kubectl, client-go, etc.

Now that simulators are no longer just for webUI, we need to think about how we can design scenario-based simulation to be easy to use from other clients as well.

Therefore, this kep proposes to define the scenario **as CRD**. All clients, including web UI, can use the scenario-based simulation feature by creating the Scenario resource.

#### Scenario CRD

The Scenario is a non-namespaced resource. 
This CRD will be applied to kube-apiserver started in kube-scheduler-simulator.

We may need to change etcd request-size limitation by --max-request-bytes since the scenario resource may be bigger than other normal resources.
https://etcd.io/docs/v3.4/dev-guide/limit/#request-size-limit

```go
type Scenario struct {
  metav1.TypeMeta 
  metav1.ObjectMeta // non namespaced.

  Spec ScenarioSpec
  Status ScenarioStatus
}

type ScenarioSpec struct {
  // Events field has all operations for a scenario.
  // Also you can add new events during the scenario is running.
  //
  // +patchMergeKey=ID
	// +patchStrategy=merge
  Events   []ScenarioEvent
}

type ScenarioEvent struct {
  // ID for this event. Normally, the system sets this field for you.
  ID string
  // Time indicates the time at which the event occurs.
  Time FakeTime
  // Operation describes which operation this event wants to do.
  // Only "Create", "Patch", "Delete", "Done" are valid operations in ScenarioEvent.
  Operation OperationType

  // One of the following four fields must be specified.
  // If more than one is specified or if all are empty, the event is invalid and the scenario will fail.

  // CreateOperation is the operation to create new resource.
  // When use CreateOperation, Operation should be "Create".
  CreateOperation *CreateOperation
  // PatchOperation is the operation to patch a resource.
  // When use PatchOperation, Operation should be "Patch".
  PatchOperation *PatchOperation
  // DeleteOperation indicates the operation to delete a resource.
  // When use DeleteOperation, Operation should be "Delete".
  DeleteOperation *DeleteOperation
  // DoneOperation indicates the operation to mark the scenario as DONE.
  // When use DoneOperation, Operation should be "Done".
  DoneOperation   *DoneOperation
}

// OperationType describes Operation.
// Please see the following defined OperationType, all operation types not listed below are invalid.
type OperationType string 

const (
  CreateOperation         OperationType = "Create"
  PatchOperation          OperationType = "Patch"
  DeleteOperation         OperationType = "Delete"
  DoneOperation           OperationType = "Done"
  PodScheduledOperation   OperationType = "PodScheduled"
  PodUnscheduledOperation OperationType = "PodUnscheduled"
  PodPreemptedOperation   OperationType = "PodPreempted"
)

type CreateOperation struct {
  // Object is the Object to be create.
  Object runtime.Object
}

type PatchOperation struct {
  TypeMeta   metav1.TypeMeta
  ObjectMeta metav1.ObjectMeta
​​  // Patch is the patch for target.
  Patch string
}

type DeleteOperation struct {
  TypeMeta   metav1.TypeMeta
  ObjectMeta metav1.ObjectMeta
}

type DoneOperation struct {
  Done bool   
}

// FakeTime is the time simply represented by numbers and used in the simulation.
// In FakeTime, time is moved to next time when it can no longer schedule any more Pods in that time.
// See [TODO: document here] for more information about FakeTime.
type FakeTime int32

type ScenarioStatus struct {
  Phase ScenarioPhase
	// A human readable message indicating details about why the scenario is in this phase.
  // optional
	Message *string 
  // Time indicates the current time.
  Time FakeTime
  // ScenarioResult has the result of the simulation.
  // Just before Time advances, this result is updated based on all occurrences at that time.
  ScenarioResult ScenarioResult
}

​​type ScenarioPhase string

const (
  // ScenarioPending phase indicates the scenario isn't started yet.
  // e.g. waiting for another scenario to finish running.
  ScenarioPending    ScenarioPhase = "Pending"
  // ScenarioRunning phase indicates the scenario is running.
  ScenarioRunning    ScenarioPhase = "Running"
  // ScenarioPaused phase indicates all ScenarioSpec.Events 
  // has been finished but has not been marked as done by ScenarioDone ScenarioEvent.
  ScenarioPaused     ScenarioPhase = "Paused"
  // ScenarioCompleted phase describes Scenario is fully compelted
  // by ScenarioDone ScenarioEvent. User 
  // can’t add any ScenarioEvent Once 
  // Sceanrio reached at the phase.
  ScenarioCompleted  ScenarioPhase = "Completed"
  // ScenarioFailed phase indicates something wrong happened during running scenario.
  // For example:
  // - the controller cannot create resource for some reason.
  // - users change the scheduler configuration via simulator API.
  ScenarioFailed     ScenarioPhase = "Failed"
  ScenarioUnknown    ScenarioPhase = "Unknown"
) 

type ScenarioResults struct {
  // SimulatorVersion represents the version of the simulator that runs this scenario.
  SimulatorVersion string
  // Timeline is a map of events keyed with FakeTime.
  // This may have many of the same events as .spec.events, but has additional PodScheduled and Delete events for Pods 
  // to represent a Pod is scheduled or preempted by the scheduler.
  Timeline         map[FakeTime][]ScenarioTimelineEvent
}

type ScenarioTimelineEvent struct {
  // Time indicates the time at which the event occurs.
  Time      FakeTime
  // Operation describes which operation this event wants to do.
  // Only "Create", "Patch", "Delete", "Done", "PodScheduled", "PodUnscheduled", "PodPreempted" are valid operations in ScenarioTimelineEvent.
  Operation OperationType

  // Only one of the following fields must be non-empty.

  // Create is the result of ScenarioSpec.Events.CreateOperation.
  // When Create is non nil, Operation should be "Create".
  Create        *CreateOperationResult
  // Patch is the result of ScenarioSpec.Events.PatchOperation.
  // When Patch is non nil, Operation should be "Patch".
  Patch         *PatchOperationResult
  // Delete is the result of ScenarioSpec.Events.DeleteOperation.
  // When Delete is non nil, Operation should be "Delete".
  Delete        *DeleteOperationResult
  // Done is the result of ScenarioSpec.Events.DoneOperation.
  // When Done is non nil, Operation should be "Done".
  Done          *DoneOperationResult
  // PodScheduled represents the Pod is scheduled to a Node.
  // When PodScheduled is non nil, Operation should be "PodScheduled".
  PodScheduled  *PodResult
  // PodUnscheduled represents the scheduler tried to schedule the Pod, but cannot schedule to any Node.
  // When PodUnscheduled is non nil, Operation should be "PodUnscheduled".
  PodUnscheduled  *PodResult
  // PodPreempted represents the scheduler preempted the Pod.
  // When PodPreempted is non nil, Operation should be "PodPreempted".
  PodPreempted  *PodResult
}

type CreateOperationResult struct {
  // Object is the Object to be create.
  Object runtime.Object
}

type PatchOperationResult struct {
  TypeMeta   metav1.TypeMeta
  ObjectMeta metav1.ObjectMeta
​​  // Patch is the patch for target.
  Patch string
  // Result is the resource after patch.
  Object runtime.Object
}

type DeleteOperationResult struct {
  TypeMeta   metav1.TypeMeta
  ObjectMeta metav1.ObjectMeta
}

type DoneOperationResult struct {
  Done bool   
}

// PodResult has the results related to the specific Pod.
// Depending on the status of the Pod, some fields may be empty.
type PodResult struct {
  Pod v1.Pod
  // BoundTo indicates to which Node the Pod was scheduled.
  BoundTo             *string
  // PreemptedBy indicates which Pod the Pod was deleted for.
  // This field may be nil if this Pod has not been preempted.
  PreemptedBy         *string
  // CreatedAt indicates when the Pod was created.
  CreatedAt           FakeTime
  // BoundAt indicates when the Pod was scheduled.
  // This field may be nil if this Pod has not been scheduled.
  BoundAt             *FakeTime
  // PreemptedAt indicates when the Pod was preempted.
  // This field may be nil if this Pod has not been preempted.
  PreemptedAt         *FakeTime
  // ScheduleResult has the results of all scheduling for the Pod.
  ScheduleResult      []ScenarioPodScheduleResult
}

type ScenarioPodScheduleResult struct {
  // Time indicates the time scheduling at which the scheduling is performed.
  Time                *FakeTime
  // AllCandidateNodes indicates all candidate Nodes before Filter.
  AllCandidateNodes   []string
  // AllFilteredNodes indicates all candidate Nodes after Filter.
  AllFilteredNodes    []string
  // PluginResults has each plugin’s result.
  PluginResults       ScenarioPluginsResults
}

type (
  NodeName   string
  PluginName string
)

type ScenarioPluginsResults struct {
  // Filter has each filter plugin’s result.
  Filter            map[NodeName]map[PluginName]string
  // Score has each score plugin’s score.
  Score             map[NodeName]map[PluginName]ScenarioPluginsScoreResult
}

type ScenarioPluginsScoreResult struct {
  // RawScore has the score from Score method of Score plugins.
  RawScore             int64
  // NormalizedScore has the score calculated by NormalizeScore method of Score plugins.
  NormalizedScore      int64
  // FinalScore has score plugin’s final score calculated by normalizing with NormalizedScore and applied Score plugin weight. 
  FinalScore           int64
}
```

#### The concept "FakeTime"

FakeTime is the concept to represent time in simulation, not a difficult one. 
- In FakeTime, time is simply represented by numbers. like 1, 2, 3…
- In FakeTime, time is moved to next time **when it can no longer schedule any more Pods**.

The following shows what happens at a single time in FakeTime:

1. run all operations defined for that time
2. scheduler starts scheduling
3. scheduler stops scheduling when it can no longer schedule any more Pods.
4. update status.scenarioResults.
5. move to next time.

##### Why scheduler needs to restarts/stops scheduling loop?

To ensure that the results of the simulation do not vary significantly from run to run. 

When users define multiple operations for the same time, users expect them to run concurrently, but in practice it is difficult to run them at the same time. This is because the scheduler is constantly attempting to schedule. 
For example, suppose a user has defined a scenario to create 1000 Nodes at the same time. Since it is strictly impossible to create 1000 Nodes at the same time, pending Pods will be scheduled to the Nodes created first. And depending on what order the Nodes were created, the results of the simulation may change.
To prevent this, the scheduler needs to be stopped scheduling until 1000 Nodes are created.

##### How to stop scheduling

We can prevent a scheduling queue from releasing next Pods by replacing `Scheduler.NextPod` function.
https://github.com/kubernetes/kubernetes/blob/867b5cc31b376c9f5d04cf9278112368b0337104/pkg/scheduler/scheduler.go#L75

##### adding events to running Scenario 

It is allowed to add events while the Scenario is running.

Note that it does not make sense to add past FakeTime events.
The scenario will continue to run until all events in .spec.Events are completed, and when all events are completed, the scenario will be "Paused" phase in the case .spec.Events doesn't have "Done" operation.

So, it is strongly recommended adding events to running Scenario only after Scenario have reached "Paused" phase. 
(since FakeTime has stopped moving forward in "Paused" phase as described above.)
Otherwise you may add the past FakeTime events and they are ignored by running Scenario.

##### configure when to update scenarioResults

As described in the above, the controller only update status.scenarioResults in Scenario resource when proceeding to the next FakeTime.

This is because kube-apiserver will be so busy if the controller update status.scenarioResults everytime it updated,
especially when the size of Scenario is so big.

> etcd is designed to handle small key value pairs typical for metadata. Larger requests will work, but may increase the latency of other requests. By default, the maximum size of any request is 1.5 MiB. This limit is configurable through --max-request-bytes flag for etcd server.
https://etcd.io/docs/v3.4/dev-guide/limit/#request-size-limit

And, sometimes, users may want to reduce this updating request more for some reasons. 

For example, 
- when watching Scenario for dynamic event addition 
- when using Scenario for accurate benchmark testing, users may want to reduce the request to update Scenario for kube-apiserver as much as possible

We can add a new configuration environment variable `UPDATE_SCENARIO_RESULTS_STRATEGY` and define some strategy like:
- `UPDATE_SCENARIO_RESULTS_STRATEGY=AtMovingNextTime`: default value. update status.scenarioResults in Scenario resource when proceeding to the next FakeTime.
- `UPDATE_SCENARIO_RESULTS_STRATEGY=OnPause`: update status.scenarioResults in Scenario resource when the Scenario's phase is `Paused` or `Failed`.
- `UPDATE_SCENARIO_RESULTS_STRATEGY=OnDone`: update status.scenarioResults in Scenario resource when the Scenario's phase is `Done` or `Failed`.

#### The result calculation packages

ScenarioResults only has the simple data that represent what was happen during the scenario.

So, we will provide useful functions and data structures to analize the result. 

For example:
- the function to aggregate changes in allocation rate of the entire cluster.
- the function to aggregate changes in resource utilization for each Node.
- the function to aggregate data by Pod.
- (Do you have any other idea? Tell us!)

By putting only the minimum simple information in ScenarioResults and providing functions to change it into a user-friendly structs, many data structures can be supported in the future without any changes to API.

#### Prohibitions and Restrictions

When Scenario is created, the scenario is started by the controller. The scenario is run one by one, and multiple scenarios are never run at the same time.

In addition, the following actions are prohibited during scenario execution The scenario result will be unstable or invalid if any of these actions are performed.
- change the scheduler configuration via simulator API.
- create/delete/edit any resources.

And all resources created before starting a scenario are deleted at the start of the scenario,
so that they don't affect the simulation results.

### User Stories 

#### Story 1

The company has added many features into scheduler via some custom plugins, 
and they want to make sure that their expansions are working as expected and has not negatively impacted the scheduling results.

##### Solution

They can define appropriate scenario and analize the results.

#### Story 2

The users want to see how their customized scheduler behaves in the worst case scenario.

##### Solution

Even when a scenario is running, users can add events to that scenario. 
So, in this case, they can add events that are most worst case for the scheduler by looking at the simulation results and the resources status at that point.

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->
