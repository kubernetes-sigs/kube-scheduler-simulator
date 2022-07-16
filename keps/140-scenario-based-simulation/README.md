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
	// A human-readable message indicating details about why the scenario is in this phase.
	//
	// +optional
	Message *string `json:"message,omitempty"`
	// StepStatus has the status related to step.
	// 
    StepStatus ScenarioStepStatus
	// ScenarioResult has the result of the simulation.
	// Just before Step advances, this result is updated based on all occurrences at that step.
	//
	// +optional
	ScenarioResult ScenarioResult `json:"scenarioResult,omitempty"`
}

type ScenarioStepStatus struct {
    // Step indicates the current step.
    //
    // +optional
    Step ScenarioStep `json:"step,omitempty"`
	// Phase indicates the current phase in single step.
	//
	// Within a single step, the phase proceeds as follows:
	// 1. run all scenario.Spec.Events defined for that step. (OperatingEvents)
    // 2. finish (1) (OperatingEventsFinished)
    // 3. the scheduler starts scheduling. (Scheduling)
    // 4. the scheduler stops scheduling and changes scenario.Status.StepStatus.Phase to SchedulingFinished
    //    when it can no longer schedule any more Pods. (Scheduling -> SchedulingFinished)
    // 5. update status.scenarioResult and move to next step. (StepFinished)
	// +optional
	Phase StepPhase `json:"phase,omitempty"`
}

type StepPhase string

const (
	// OperatingEvents means controller is currently operating event defined for the step.
    OperatingEvents          StepPhase = "OperatingEvents"
    // OperatingEventsFinished means controller have finished operating event defined for the step.
    OperatingEventsFinished  StepPhase = "OperatingEventsFinished"
    // Scheduling means scheduler is scheduling Pods.
    Scheduling               StepPhase = "Scheduling"
    // SchedulingFinished means scheduler is trying to schedule Pods.
	// But, it can no longer schedule any more Pods. 
    SchedulingFinished       StepPhase = "SchedulingFinished"
	// StepFinished means controller is preparing to move to next step.
    StepFinished             StepPhase = "Finished"
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

#### supported scheduler in simulator

The goal of this KEP is to make schedulers that meet the all following criteria work with scenario:

1. schedule Pods by the conditions of k8s resources.
    - In other words, schedulers shouldn't schedule Pods by the conditions other than k8s resources.
    - For example, if the scheduler that schedules Pods by checking some metrics from external server, 
    the scenario controller cannot manage metrics value and the Scenario may yield different results for each run.
2. scheduler that doesn't creates/edits/deletes resource. (except preemption and binding)
    - Basically, no one other than scenario controller should create/edit/delete resources.
3. scheduler included in simulator. Or outside scheduler which is based on scheduler in [kubernetes/kubernetes repo](https://github.com/kubernetes/kubernetes/tree/master/pkg/scheduler)
    - In other words, scheduler needs to have [`NextPod`](https://github.com/kubernetes/kubernetes/blob/867b5cc31b376c9f5d04cf9278112368b0337104/pkg/scheduler/scheduler.go#L75) field and uses it for fetch Pod from queue, 
    because we want to replace it to make scenario work correctly. (see [# Required setup for scheduler](#required-setup-for-scheduler).)

#### Required setup for scheduler

Those who want to use outside scheduler need to do replace [`scheduler.NextPod`](https://github.com/kubernetes/kubernetes/blob/d14ba948ef63769e9767aebd5a08171832a1bbf6/pkg/scheduler/scheduler.go#L73) with a function provided by us.

The function we provide will does:
- restart/stop scheduling (see [# How to stop scheduling loop](#How-to-stop-scheduling-loop))
- check which Pods are scheduled and judge if it can no longer schedule any more Pods. (see [# How scheduler detects "when it can no longer schedule any more Pods" at (4)](#How-scheduler-detects-"when-it-can-no-longer-schedule-any-more-Pods"-at-(4)))

Those who want to use scheduler included in simulator don't need to do anything for setup since we does this set up for users.

#### The concept "ScenarioStep"

ScenarioStep is:
- simply represented by numbers. like 1, 2, 3…
- moved to next step **when it can no longer schedule any more Pods**.

Within a single step, the phase proceeds as follows:
1. run all scenario.Spec.Events defined for that step. (OperatingEvents)
2. finish (1) (OperatingEventsFinished)
3. the scheduler starts scheduling. (Scheduling)
4. the scheduler stops scheduling and changes scenario.Status.StepStatus.Phase to SchedulingFinished
   when it can no longer schedule any more Pods. (Scheduling -> SchedulingFinished)
5. update status.scenarioResult and move to next step. (StepFinished)

##### Why scheduler needs to restarts/stops scheduling loop?

To ensure that the results of the simulation do not vary significantly from run to run.

If we don't have "ScenarioStep" concept and when users want to define multiple operations for the same time, users expect them to run concurrently and them to be run at the same time. But, in practice it is difficult to run them at the same time, because the scheduler is constantly attempting to schedule.
For example, suppose a user has defined a scenario to create 1000 Nodes at the same time. Since it is strictly impossible to create 1000 Nodes at the same time, pending Pods will be scheduled to the Nodes created first. And depending on what order the Nodes were created, the results of the simulation may change.
To prevent this, the scheduler needs to be stopped scheduling until 1000 Nodes are created.

##### How to stop scheduling loop

If the scheduler is based on implementation of the current scheduler in kubernetes/kubernetes, 
we can prevent a scheduling queue from poping next Pods by overwriting [`Scheduler.NextPod`](https://github.com/kubernetes/kubernetes/blob/867b5cc31b376c9f5d04cf9278112368b0337104/pkg/scheduler/scheduler.go#L75) function.

The override function basically behave like normal `NextPod`,
but checks the running Scenario's status.SchedulerStatus and decide to stop/restart scheduling.

##### How scheduler detects "when it can no longer schedule any more Pods" at (4)

The most schedulers try to schedule Pods one by one. 

The idea here is detecting "when it can no longer schedule any more Pods" **by recording Pods scheduler have tried to schedule**.

Basically, scheduler are checking other resources' status and decide the best Node for a Pod.
That means if any resource haven't created/edited/deleted, the scheduling result shouldn't be changed.

Resource creation/changes/deletion should be only run during `OperatingEvents` phase,
and only preemption(deleting Pod) or binding(scheduling Pod to a Node) happens during scheduling.

Thus, we can consider it can no longer schedule any more Pods when the following conditions are met:
1. all unscheduled Pods are tried to be scheduled twice or more.
2. no Pods are preempted and no Pods are scheduled during (1).

For example, there are Pod1 - Pod4 in the cluster and all of them aren't scheduled yet.
1. scheduler try to schedule Pod1 but cannot schedule it. Pod1 is moved back to queue.
2. scheduler try to schedule Pod2 and can schedule it.
3. scheduler try to schedule Pod3 and cannot schedule it. Pod3 is moved back to queue.
4. scheduler try to schedule Pod4 and cannot schedule it. Pod4 is moved back to queue.
5. scheduler try to schedule Pod1 and cannot schedule it. Pod1 is moved back to queue.
6. scheduler try to schedule Pod3 and cannot schedule it. Pod3 is moved back to queue.
7. scheduler try to schedule Pod4 and cannot schedule it. Pod4 is moved back to queue.
8. scheduler try to schedule Pod1 but cannot schedule it. Pod1 is moved back to queue.
9. during (3) - (8), no Pods are preempted, no Pods are scheduled, and Pod1,3,4 are tried to be scheduled twice.
So, we can consider it can no longer schedule any more Pods.

The reason why all unscheduled Pods are tried to be scheduled twice or more is that [binding cycle is run in parallel](https://kubernetes.io/docs/concepts/scheduling-eviction/scheduling-framework/#scheduling-cycle-binding-cycle).

You will understand why it is necessary twice, after seeing the following case:
1. scheduler try to schedule Pod1 but cannot schedule it. Pod1 is moved back to queue.
2. scheduler try to schedule Pod2 and can schedule it.
3. scheduler try to schedule Pod3 and cannot schedule it. Pod3 is moved back to queue.
4. scheduler try to schedule Pod4 and cannot schedule it. Pod4 is moved back to queue.
5. scheduler try to schedule Pod1 and can schedule it. 

Note that we can only record what Pods scheduler are trying to schedule next in `NextPod`. 
That means we cannot see what Pods scheduler moves back to queue when cannot schedule Pods.

During (3) - (5), no Pods are preempted, and Pod1,3,4 are tried to be scheduled once. 
**But, Pod1 is scheduled at (5)**.

In `NextPod`, we may not realize the Pod1 is scheduled at (5), 
since the binding cycle is run in parallel and binding may just not be finished when NextPod is judging if it can no longer schedule any more Pods.

So, to make sure the all unscheduled Pods cannot be scheduled anymore, we need to see the all unscheduled Pods are tried to be scheduled twice or more in `NextPod`.

#### Adding events to running Scenario 

It is allowed to add events while the Scenario is running.

Note that it does not make sense to add past ScenarioStep events.
The scenario will continue to run until all events in .spec.Events are completed, and when all events are completed, the scenario will be "Paused" phase in the case .spec.Events doesn't have "Done" operation.

So, it is strongly recommended adding events to running Scenario only after Scenario have reached "Paused" phase. 
(since ScenarioStep has stopped moving forward in "Paused" phase as described above.)
Otherwise, you may add the past ScenarioStep events and they are ignored by running Scenario.

#### Configure when to update ScenarioResult

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

### Can you elaborate a bit more on which schedulers will be supported?

Respondent: @sanposhiho.

The current simulator has scheduler internally.
(see [simulator/docs/how-it-works.md](simulator/docs/how-it-works.md))

And the scheduler version can be used is only fixed one. (v1.22 at the time I write this.)
That Scheduler can be customized by adding your plugin to scheduler or by scheduler configuration.
(see [docs/how-to-use-custom-plugins](./docs/how-to-use-custom-plugins))

So, that means the current simulator can simulate scheduling with only limited scheduler.

And, for who want to use scheduler not supported by simulator, (e.g. scheduler of different versions, scheduler applied some patches, or completely original scheduler)
we plan to support "outside scheduler" in simulator. ([issue here](https://github.com/kubernetes-sigs/kube-scheduler-simulator/issues/182))

The outside scheduler literally means the scheduler outside of simulator,
communicates with the api-server in simulator and schedule all the Pod.

Given the current simulator only supports limited scheduler described above,
I _guess_ more users will want to use outside scheduler rather than scheduler in simulator.

Any kinds of scheduler can be created in this world.
But, the goal of this KEP is to support only schedulers listed in [# supported scheduler in simulator](#supported-scheduler-in-simulator) in scenario.

However, we believe we can provide a function to make most schedulers work with scenario in the future,
(although we won't go into much detail in this KEP.)

That schedulers only needs to satisfy the condition that it schedule Pods one by one with same process,
and users only need to put the provided function into the top of the process to make scheduler work with scenario like:

```go
// this scheduler trys to schedule Pods one by one.
func (s *scheduler) run() {
	for {
	    scheduleOne(pod)	
    }
}

// scheduleOne schedule a given Pod.
func (s *scheduler) scheduleOne(pod *v1.Pod) {
    // fetch one pod and try to schedule it.
    pod := s.fetchPodFromSomewhere()
	
	// user only need to put this function provided by us so that make this scheduler work with scenario.
	functionProvidedByUs(pod)
	
    // scheduling implementation
}
```

But, even with this function, I think we cannot support the scheduler that doesn't meet the criteria (1)(2) in [# supported scheduler in simulator](#supported-scheduler-in-simulator).
