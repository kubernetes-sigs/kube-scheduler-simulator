# KEP-140: Scenario-based simulation

## Summary

A new scenario-based simulation feature is introduced to kube-scheduler-simulator by new `Scenario` CRD.

## Motivation

Nowadays, the scheduler is extendable in the multiple ways:
- configure with [KubeSchedulerConfiguration](https://kubernetes.io/docs/reference/scheduling/config/)
- add Plugins of [Scheduling Framework](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/)
- add [Extenders](https://github.com/kubernetes/enhancements/tree/5320deb4834c05ad9fb491dcd361f952727ece3e/keps/sig-scheduling/1819-scheduler-extender)
- etc...

But, unfortunately, not all expansions yield good results.
Those who customize the scheduler need to make sure their scheduler is working as expected, and doesn't have an unacceptably negative impact on the scheduling result or scheduling performance. And, usually, evaluating the scheduler is not so easy because there are many factors for the evaluation of the scheduler's ability.

The scenario-based simulation feature will be useful for those who customize the scheduler to evaluate their scheduler.

### Goals

Users can simulate their scheduler with some defined scenarios and can evaluate their scheduler with the result.

### Non-Goals

See the result of scenario-based simulation from Web UI. (may be implemented in the future, but out of scope of this proposal.)

## Proposal

### Implementation design details

#### The current simulator and proposal

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
// Scenario is the Schema for the scenarios API
type Scenario struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`

  Spec   ScenarioSpec   `json:"spec,omitempty"`
  Status ScenarioStatus `json:"status,omitempty"`
}

// ScenarioSpec defines the desired state of Scenario
type ScenarioSpec struct {
	// Events field has all operations for a scenario.
	// Also you can add new events during the scenario is running.
	//
	// +patchMergeKey=ID
	// +patchStrategy=merge
	Events []*ScenarioEvent `json:"events"`
}

type ScenarioEvent struct {
	// ID for this event. Normally, the system sets this field for you.
	ID string `json:"id"`
	// Step indicates the step at which the event occurs.
	Step ScenarioStep `json:"step"`
	// Operation describes which operation this event wants to do.
	// Only "Create", "Patch", "Delete", "Done" are valid operations in ScenarioEvent.
	Operation OperationType `json:"operation"`

	// One of the following four fields must be specified.
	// If more than one is specified or if all are empty, the event is invalid and the scenario will fail.

	// CreateOperation is the operation to create new resource.
	// When use CreateOperation, Operation should be "Create".
	//
	// +optional
	CreateOperation *CreateOperation `json:"createOperation,omitempty"`
	// PatchOperation is the operation to patch a resource.
	// When use PatchOperation, Operation should be "Patch".
	//
	// +optional
	PatchOperation *PatchOperation `json:"patchOperation,omitempty"`
	// DeleteOperation indicates the operation to delete a resource.
	// When use DeleteOperation, Operation should be "Delete".
	//
	// +optional
	DeleteOperation *DeleteOperation `json:"deleteOperation,omitempty"`
	// DoneOperation indicates the operation to mark the scenario as DONE.
	// When use DoneOperation, Operation should be "Done".
	//
	// +optional
	DoneOperation *DoneOperation `json:"doneOperation,omitempty"`
}

// OperationType describes Operation.
// Please see the following defined OperationType, all operation types not listed below are invalid.
type OperationType string

const (
	CreateOperationType         OperationType = "Create"
	PatchOperationType          OperationType = "Patch"
	DeleteOperationType         OperationType = "Delete"
	DoneOperationType           OperationType = "Done"
	PodScheduledOperationType   OperationType = "PodScheduled"
	PodUnscheduledOperationType OperationType = "PodUnscheduled"
	PodPreemptedOperationType   OperationType = "PodPreempted"
)

type CreateOperation struct {
	// Object is the Object to be created.
	Object unstructured.Unstructured `json:"object"`

	// +optional
	CreateOptions metav1.CreateOptions `json:"createOptions,omitempty"`
}

type PatchOperation struct {
	TypeMeta   metav1.TypeMeta   `json:"typeMeta"`
	ObjectMeta metav1.ObjectMeta `json:"objectMeta"`
	// Patch is the patch for target.
	Patch string `json:"patch"`

	// +optional
	PatchOptions metav1.PatchOptions `json:"patchOptions,omitempty"`
}

type DeleteOperation struct {
	TypeMeta   metav1.TypeMeta   `json:"typeMeta"`
	ObjectMeta metav1.ObjectMeta `json:"objectMeta"`

	// +optional
	DeleteOptions metav1.DeleteOptions `json:"deleteOptions,omitempty"`
}

type DoneOperation struct {
	Done bool `json:"done"`
}

// ScenarioStep is the step simply represented by numbers and used in the simulation.
// In ScenarioStep, step is moved to next step when it can no longer schedule any more Pods in that step.
// See [TODO: document here] for more information about ScenarioStep.
type ScenarioStep int32

// ScenarioStatus defines the observed state of Scenario
type ScenarioStatus struct {
	// The phase is a simple, high-level summary of where the Scenario is in its lifecycle.
	//
	// +optional
	Phase ScenarioPhase `json:"phase,omitempty"`
	// Current state of scheduler.
	//
	// +optional
	SchedulerStatus SchedulerStatus `json:"schedulerStatus,omitempty"`
	// A human readable message indicating details about why the scenario is in this phase.
	//
	// +optional
	Message *string `json:"message,omitempty"`
	// Step indicates the current step.
	//
	// +optional
	Step ScenarioStep `json:"step,omitempty"`
	// ScenarioResult has the result of the simulation.
	// Just before Step advances, this result is updated based on all occurrences at that step.
	//
	// +optional
	ScenarioResult ScenarioResult `json:"scenarioResult,omitempty"`
}

type SchedulerStatus string

const (
	// SchedulerWillRun indicates the scheduler is expected to start to schedule.
	// In other words, the scheduler is currently stopped,
	// and will start to schedule Pods when the state is SchedulerWillRun.
	SchedulerWillRun SchedulerStatus = "WillRun"
	// SchedulerRunning indicates the scheduler is scheduling Pods.
	SchedulerRunning SchedulerStatus = "Running"
	// SchedulerWillStop indicates the scheduler is expected to stop scheduling.
	// In other words, the scheduler is currently scheduling Pods,
	// and will stop scheduling when the state is SchedulerWillStop.
	SchedulerWillStop SchedulerStatus = "WillStop"
	// SchedulerStoped indicates the scheduler stops scheduling Pods.
	SchedulerStoped SchedulerStatus = "Stoped"
	// SchedulerUnknown indicates the scheduler's status is unknown.
	SchedulerUnknown ScenarioPhase = "Unknown"
)

type ScenarioPhase string

const (
	// ScenarioPending phase indicates the scenario isn't started yet.
	// e.g. waiting for another scenario to finish running.
	ScenarioPending ScenarioPhase = "Pending"
	// ScenarioRunning phase indicates the scenario is running.
	ScenarioRunning ScenarioPhase = "Running"
	// ScenarioPaused phase indicates all ScenarioSpec.Events
	// has been finished but has not been marked as done by ScenarioDone ScenarioEvent.
	ScenarioPaused ScenarioPhase = "Paused"
	// ScenarioSucceeded phase describes Scenario is fully completed
	// by ScenarioDone ScenarioEvent. User
	// can’t add any ScenarioEvent once
	// Scenario reached at the phase.
	ScenarioSucceeded ScenarioPhase = "Succeeded"
	// ScenarioFailed phase indicates something wrong happened during running scenario.
	// For example:
	// - the controller cannot create resource for some reason.
	// - users change the scheduler configuration via simulator API.
	ScenarioFailed  ScenarioPhase = "Failed"
	ScenarioUnknown ScenarioPhase = "Unknown"
)

type ScenarioResult struct {
	// SimulatorVersion represents the version of the simulator that runs this scenario.
	SimulatorVersion string `json:"simulatorVersion"`
	// Timeline is a map of events keyed with ScenarioStep.
	// This may have many of the same events as .spec.events, but has additional PodScheduled and Delete events for Pods
	// to represent a Pod is scheduled or preempted by the scheduler.
	//
	// +patchMergeKey=ID
	// +patchStrategy=merge
	Timeline map[ScenarioStep][]ScenarioTimelineEvent `json:"timeline"`
}

type ScenarioTimelineEvent struct {
	// The ID will be the same as spec.ScenarioEvent.ID if it is from the defined event.
	// Otherwise, it'll be newly generated.
	ID string
	// Step indicates the step at which the event occurs.
	Step ScenarioStep `json:"step"`
	// Operation describes which operation this event wants to do.
	// Only "Create", "Patch", "Delete", "Done", "PodScheduled", "PodUnscheduled", "PodPreempted" are valid operations in ScenarioTimelineEvent.
	Operation OperationType `json:"operation"`

	// Only one of the following fields must be non-empty.

	// Create is the result of ScenarioSpec.Events.CreateOperation.
	// When Create is non nil, Operation should be "Create".
	Create *CreateOperationResult `json:"create"`
	// Patch is the result of ScenarioSpec.Events.PatchOperation.
	// When Patch is non nil, Operation should be "Patch".
	Patch *PatchOperationResult `json:"patch"`
	// Delete is the result of ScenarioSpec.Events.DeleteOperation.
	// When Delete is non nil, Operation should be "Delete".
	Delete *DeleteOperationResult `json:"delete"`
	// Done is the result of ScenarioSpec.Events.DoneOperation.
	// When Done is non nil, Operation should be "Done".
	Done *DoneOperationResult `json:"done"`
	// PodScheduled represents the Pod is scheduled to a Node.
	// When PodScheduled is non nil, Operation should be "PodScheduled".
	PodScheduled *PodResult `json:"podScheduled"`
	// PodUnscheduled represents the scheduler tried to schedule the Pod, but cannot schedule to any Node.
	// When PodUnscheduled is non nil, Operation should be "PodUnscheduled".
	PodUnscheduled *PodResult `json:"podUnscheduled"`
	// PodPreempted represents the scheduler preempted the Pod.
	// When PodPreempted is non nil, Operation should be "PodPreempted".
	PodPreempted *PodResult `json:"podPreempted"`
}

type CreateOperationResult struct {
	// Operation is the operation that was done.
	Operation CreateOperation `json:"operation"`
	// Result is the resource after patch.
	Result unstructured.Unstructured `json:"result"`
}

type PatchOperationResult struct {
	// Operation is the operation that was done.
	Operation PatchOperation `json:"operation"`
	// Result is the resource after patch.
	Result unstructured.Unstructured `json:"result"`
}

type DeleteOperationResult struct {
	// Operation is the operation that was done.
	Operation DeleteOperation `json:"operation"`
}

type DoneOperationResult struct {
	// Operation is the operation that was done.
	Operation DoneOperation `json:"operation"`
}

// PodResult has the results related to the specific Pod.
// Depending on the status of the Pod, some fields may be empty.
type PodResult struct {
	Pod v1.Pod `json:"pod"`
	// BoundTo indicates to which Node the Pod was scheduled.
	BoundTo *string `json:"boundTo"`
	// PreemptedBy indicates which Pod the Pod was deleted for.
	// This field may be nil if this Pod has not been preempted.
	PreemptedBy *string `json:"preemptedBy"`
	// CreatedAt indicates when the Pod was created.
	CreatedAt ScenarioStep `json:"createdAt"`
	// BoundAt indicates when the Pod was scheduled.
	// This field may be nil if this Pod has not been scheduled.
	BoundAt *ScenarioStep `json:"boundAt"`
	// PreemptedAt indicates when the Pod was preempted.
	// This field may be nil if this Pod has not been preempted.
	PreemptedAt *ScenarioStep `json:"preemptedAt"`
	// ScheduleResult has the results of all scheduling for the Pod.
	//
	// If the scheduler working with a simulator isn't worked on scheduling framework,
	// this field will be empty.
	// TODO: add the link to doc when it's empty.
	//
	// +patchStrategy=replace
	// +optional
	ScheduleResult []ScenarioPodScheduleResult `json:"scheduleResult"`
}

type ScenarioPodScheduleResult struct {
	// Step indicates the step scheduling at which the scheduling is performed.
	Step *ScenarioStep `json:"step"`
	// AllCandidateNodes indicates all candidate Nodes before Filter.
	AllCandidateNodes []string `json:"allCandidateNodes"`
	// AllFilteredNodes indicates all candidate Nodes after Filter.
	AllFilteredNodes []string `json:"allFilteredNodes"`
	// PluginResults has each plugin’s result.
	PluginResults ScenarioPluginsResults `json:"pluginResults"`
}

type (
	NodeName   string
	PluginName string
)

type ScenarioPluginsResults struct {
	// Filter has each filter plugin’s result.
	Filter map[NodeName]map[PluginName]string `json:"filter"`
	// Score has each score plugin’s score.
	Score map[NodeName]map[PluginName]ScenarioPluginsScoreResult `json:"score"`
}

type ScenarioPluginsScoreResult struct {
	// RawScore has the score from Score method of Score plugins.
	RawScore int64 `json:"rawScore"`
	// NormalizedScore has the score calculated by NormalizeScore method of Score plugins.
	NormalizedScore int64 `json:"normalizedScore"`
	// FinalScore has score plugin’s final score calculated by normalizing with NormalizedScore and applied Score plugin weight.
	FinalScore int64 `json:"finalScore"`
}

```

#### The concept "ScenarioStep"

ScenarioStep is: 
- simply represented by numbers. like 1, 2, 3…
- moved to next step **when it can no longer schedule any more Pods**.

The following shows what happens at a single step in ScenarioStep:

1. run all operations defined for that step.
2. the scenario controller changes status.SchedulerStatus to SchedulerWillRun.
3. the scheduler starts scheduling and changes status.SchedulerStatus to SchedulerRunning.
4. the scenario controller changes status.SchedulerStatus to SchedulerWillStop, 
when it can no longer schedule any more Pods.
5. the scheduler stop scheduling and changes status.SchedulerStatus to SchedulerStoped.
6. update status.scenarioResult.
7. move to next step.

##### How to detect "when it can no longer schedule any more Pods" at (4)

In the scheduler, the unscheduled Pods are stored in queue.

In a single step, the resources are created/edited/deleted only at (1) in the above description. (except pod deletion by preemption)
So, the amount of k8s resources in the cluster doesn't get increased during (4) rather got decreased by scheduled Pod.

This means that the number of Pods can be scheduled will decrease and eventually no more Pods can be scheduled 
(or all Pods will be scheduled).
Thus, all we have to do in (4) is wait until no more pods are scheduled for a while.

##### Why scheduler needs to restarts/stops scheduling loop?

To ensure that the results of the simulation do not vary significantly from run to run. 

If we don't have "ScenarioStep" concept and when users want to define multiple operations for the same time, users expect them to run concurrently and them to be run at the same time. But, in practice it is difficult to run them at the same time, because the scheduler is constantly attempting to schedule. 
For example, suppose a user has defined a scenario to create 1000 Nodes at the same time. Since it is strictly impossible to create 1000 Nodes at the same time, pending Pods will be scheduled to the Nodes created first. And depending on what order the Nodes were created, the results of the simulation may change.
To prevent this, the scheduler needs to be stopped scheduling until 1000 Nodes are created.

##### How to stop scheduling loop

We can prevent a scheduling queue from poping next Pods by replacing `Scheduler.NextPod` function.
https://github.com/kubernetes/kubernetes/blob/867b5cc31b376c9f5d04cf9278112368b0337104/pkg/scheduler/scheduler.go#L75

We can provide the function to override `Scheduler.NextPod` so that users can use scenario outside of simulator.
The override function basically behave like normal `NextPod`, 
but checks the running Scenario's status.SchedulerStatus and decide to stop/restart scheduling.

##### Adding events to running Scenario 

It is allowed to add events while the Scenario is running.

Note that it does not make sense to add past ScenarioStep events.
The scenario will continue to run until all events in .spec.Events are completed, and when all events are completed, the scenario will be "Paused" phase in the case .spec.Events doesn't have "Done" operation.

So, it is strongly recommended adding events to running Scenario only after Scenario have reached "Paused" phase. 
(since ScenarioStep has stopped moving forward in "Paused" phase as described above.)
Otherwise, you may add the past ScenarioStep events and they are ignored by running Scenario.

##### Configure when to update ScenarioResult

As described in the above, the controller only update status.scenarioResult in Scenario resource when proceeding to the next ScenarioStep.

This is because kube-apiserver will be so busy if the controller update status.scenarioResult everytime it updated,
especially when the size of Scenario is so big.

> etcd is designed to handle small key value pairs typical for metadata. Larger requests will work, but may increase the latency of other requests. By default, the maximum size of any request is 1.5 MiB. This limit is configurable through --max-request-bytes flag for etcd server.
https://etcd.io/docs/v3.4/dev-guide/limit/#request-size-limit

And, sometimes, users may want to reduce this updating request more for some reasons. 

For example, 
- when watching Scenario for dynamic event addition 
- when using Scenario for accurate benchmark testing, users may want to reduce the request to update Scenario for kube-apiserver as much as possible

We can add a new configuration environment variable `UPDATE_SCENARIO_RESULTS_STRATEGY` and define some strategy like:
- `UPDATE_SCENARIO_RESULTS_STRATEGY=AtMovingNextStep`: default value. update status.scenarioResult in Scenario resource when proceeding to the next ScenarioStep.
- `UPDATE_SCENARIO_RESULTS_STRATEGY=OnPause`: update status.scenarioResult in Scenario resource when the Scenario's phase becomes `Paused`, `Succeeded` or `Failed`.
- `UPDATE_SCENARIO_RESULTS_STRATEGY=OnDone`: update status.scenarioResult in Scenario resource when the Scenario's phase becomes `Succeeded` or `Failed`.

#### The result calculation packages

ScenarioResult only has the simple data that represent what was happened during the scenario.

So, we will provide useful functions and data structures to analyze the result. 

For example:
- the function to aggregate changes in allocation rate of the entire cluster.
- the function to aggregate changes in resource utilization for each Node.
- the function to aggregate data by Pod.
- the generic iterator function that users can aggregate custom values.
- (Do you have any other idea? Tell us!)

By putting only the minimum simple information in ScenarioResult and providing functions to change it into a user-friendly structs, many data structures can be supported in the future without any changes to API.

#### The case kube-apiserver have Scenarios when the controller start to run

When the controller is started and finds the Scenario which phase is "Running", the controller just changes the status "Failed" with updating the `.status.message` like "the controller restarted while the Scenario was running".

In the future, it would be nice if we could implement the endpoint in simulator that tells the Scenario controller that the simulator is going to be shuted down. 

#### Prohibitions and Restrictions

When Scenario is created, the scenario is started by the controller. The scenario is run one by one, and multiple scenarios are never run at the same time. 
This means the controller will run the next Scenario after the current running Scenario becomes "Failed" or "Succeeded".

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

## Questions/Answers

### can scenario work with all kind of schedulers?

Respondent: @sanposhiho.

The current simulator has a scheduler of only fixed one version. (v1.22 at the time I write this.)

Also, schedulers can be customized by adding your plugin to scheduler, and custom plugins can be used in simulator/Scenario. (see [docs/how-to-use-custom-plugins](./docs/how-to-use-custom-plugins))

So, to summarize, the current simulator/Scenario only supports:
- non-customized scheduler of fixed version.
- scheduler above + custom plugins.

For who want to use scheduler not supported by simulator, (e.g. scheduler of different versions, patched scheduler, or completely original scheduler)
the simulator will support "outside scheduler", [issue here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/issues/182).

The outside scheduler literally means the scheduler outside of simulator,
communicates with the api-server in simulator and schedule all the Pod.

Given the current simulator only supports limited scheduler described above,
I _guess_ more users will want to use outside scheduler rather than scheduler in simulator.

Get back to the original story - So, which kinds of scheduler can scenario work with.

As described in KEP so far, for scenario, we need to stop scheduler somehow.

If scheduler in simulator is used, it's easy. 
We just need to replace NextPod field in scheduler as described in [# How to stop scheduling loop](#How-to-stop-scheduling-loop).
Users don't need to anything for it.

And for any outside scheduler that schedules Pods one by one in loop like scheduler in [kubernetes/kubernetes repo](https://github.com/kubernetes/kubernetes/tree/master/pkg/scheduler),
it's also easy.

I believe we can provide the function that checks the running Scenario's status.SchedulerStatus and decide to stop/restart scheduling. 
Users only need to add that func in the beginning of process to schedule a Pod.
By doing this, Scenario can be used with almost any scheduler.

So... which kind of scheduler **cannot** scenario work with?

I think that, for example, schedulers like following are that scenario cannot work with:
- scheduler that schedules Pods by the conditions other than k8s resources.
    - For example, the scheduler that schedules Pods by checking metrics server.
    - The scenario controller cannot manage metrics value and the Scenario may yield different results for each run.
- scheduler that creates the new resources.
    - Only binding Pods or preemption are allowed for scheduler.
