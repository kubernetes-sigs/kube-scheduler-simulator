# KEP-167: Metrics Collector

## Summary
Define a new Custom Resource `MetricsCollector` that will periodically collect metrics from the endpoint where metrics are published.

`MetricsCollector` allows users to periodically collect arbitrary metrics exposed to them from a specified server.
This resource is primarily intended to be used in conjunction with the `Scenario` to benchmark simulators, however, it also provides a more general and simple functionality that can be used in other ways. (`Scenario` is the CR to run scenario-based simulation introduced in KEP-140.)

## Motivation
With `Scenario` introduced in KEP-140, we are now able to experiment with scheduling on a somewhat larger scale.
And you will surely want metrics against which you can measure the performance of your scheduler.
This KEP defines a new custom resource `MetricsCollector` to collect and record metrics.

### Goals 
- Allow users to obtain scheduler metrics periodically.
- Combined with `Scenario`, it allows users to record metrics during scenario execution.

### Non-Goals
- Allow users to compare benchmark results from one scheduler to another. (The scope of this proposal is to record(collect) only.)
- See the result from Web UI. (maybe implemented in the future, but out of the scope of this proposal.)

## User Stories
### Story #1
A user has prepared a scheduler to be used for an internal product. 
They want to test the performance of the scheduler before putting it into production. 
This MetricsCollector will help the user to answer that question.

### Story #2
The user wants to tune the scheduler, and while Scenario allows the user to check the scheduler's performance in production to some extent, it does not tell the user how the scheduler is performing. 
By using the `MetricsCollector` together, users will have an indicator of appropriate tuning adjustments.

## Proposal
The job MetricsCollector does is simple and straightforward.
Once the MetricsCollector resource is created, it will continue to periodically retrieve metrics from the target scheduler endpoint.
This is all there is to it, however, it makes it possible to obtain metrics and perform benchmarks from general tools in addition to the `Scenario`.

### MetrocsCollector CRD
The custom resource `MetricsCollector` will be applied to kube-apiserver started in kube-scheduler-simulator.

```go
// MetricsCollector is the Schema for the MetricsCollector API.
type MetricsCollector struct {
    metav1.TypeMeta
    metav1.ObjectMeta
    Spec MetricsCollectorSpec
    Status MetricsCollectorStatus
}

type MetricsCollectorSpec struct {
    // EndpointURL represents the target endpoint.
    EndpointURL

    // MetricsSet represents the name of the metrics to be collected.
    MetricsSet MetricsSet

    // ThroughputSampleFrequencySeconds is the frequency at which metrics are obtained.
    // It should not be extremely small. This is because this value is the frequency at which requests are sent to the scheduler endpoint. However smaller values result in more accurate metrics being collected.
    // Default value is 1 * time.Second.
    ThroughputSampleFrequencySeconds int64
}

// MetricsSet manages the name of the metrics to be collected.
// We can be set in the same way as the scheduling-framework plugin
type MetricsSet struct {

     // Stores the names of metrics not to be collected.
    // "*" denotes all metrics will not be collected.
   Disabled []string
    
    // Stores the names of metrics to be collected.
     Enabled []string
}

type MetricsCollectorStatus struct {

      Phase CollectorPhase

    // Count is the current number of metrics retrieved.
    // MetricsCollector users can refer to and remember this value to enable retrieval or timing of arbitrary timing metrics.
    // This keeps it loosely coupled with `Scenario`.
    // e.g.) len(MetricFamilies[i].Metric) == 3 → count == 3
    Count int

    // MetricFamilies stores the acquired metrics
    // Metrics are converted to the Metric type defined in "github.com/prometheus/client_model/go" and then stored in []*Metric in the appropriate MetricFamily type.
    // .status.Count must be equal to the length of []*Metric.
    // See also:https://github.com/kubernetes/kubernetes/blob/6dbec8e25592d47fc8a8269c86d4b5fa838d320b/vendor/github.com/prometheus/client_model/go/metrics.pb.go#L599
    MetricFamilies []MetricFamily
}

type MetricsCollectorPhase string

Const (
    // MetricsCollectorPending means the MetricsCollector is preparing.
    MetricsCollectorPending MetricsCollectorPhase = "pending"
    // MetricsCollectorRunning means the MetricsCollector is ready and metrics collection has started.
    MetricsCollectorRunning MetricsCollectorPhase = "running"
    // MetricsCollectorFailed the MetricsCollector has stopped or is not ready for some reason.
    MetricsCollectorFailed MetricsCollectorPhase = "failed"
)
```

### How to setting

```yaml
apiVersion: v1
kind: MetricsCollectorConfiguration
spec:
  endpointURL: xxx.scheduler.example.com:10259
  throughputSampleFrequencySeconds: 1000
  metrics:
    enabled:
      - "*"
    disabled:
      - hoge-metrics
```

### Filtering of Metrics
Among the metrics, users can specify only those items they wish to collect.
Users can specify enabled and disabled like the scheduler-framework plugins. It should also be possible to do the same for configuration.
The default will be to collect all metrics published by the scheduler.

### How to collect metrics
This MetricsCollector is intended to collect metrics by calling the `/metrics/*` endpoint exposed by the scheduler.
Users must expose the scheduler endpoint in advance.

A tool similar to this exists within k/k, scheduler_perf, however, it differs from the method of metrics collection employed in this function.
This is because our project was intended to include and support schedulers running externally, and to do so we needed to access metrics in a more generalized way.

On the other hand, the current scheduler in our simulator cannot consciously collect metrics and expose endpoints. Therefore, these modifications will be necessary for the implementation of it.

### Throughput of scheduler
The scheduler_perf has a mechanism of calculate the throughput of the scheduler by sequentially counting the number of pods during benchmark runs.
However, due to the `Scenario`, this may be an inaccurate value, so it is not calculated in this function.
As an extreme example, if `Scenario` created the resources in the following order, the throughput value would take into account the Node time.
```
pod * 10 → Node * 10 → pod * 10 → Node * 10 → pod * 10  →...
```

However, a similar value can be calculated by monitoring the Add events of Pod.
If you want to calculate the throughput value of the scheduler, you need to take action by obtaining the pod_scheduling_duration_seconds metric, but be aware that you can only check throughput in seconds. 
(On the other hand, scheduler_perf calculates throughput well with ticker according to throughputSampleFrequency.)

### Timing of metrics collection
Metrics are automatically collected from the scheduler's api endpoint periodically for each throughputSampleFrequency.
This means that metrics collection will begin when the MetricsCollector is applied and the phase is running.

Therefore, if the user wants to start collecting metrics at an arbitrary time, you should use `.status.Count`.
By user refer to and record the .status.Count values at the start and end of the benchmark, the user can obtain only the indicators during the test run.
This method is also useful if you want to see metrics for any period during the test.

By adopting the above approach, we can avoid establishing a close relationship with the `scenario` to collect metrics at the time of an operation. 
This is why MetricsCollector is a more generally available resource.

### How to use 
#### With `Scenario`
1. Create a MetricsCollector resource and wait until the phase is running
2. Create Scenario resources and confirm the start of the scenario by monitoring the phase.
3. Get MetricsCollector's `.status.Count` at the start of the scenario.
(3.5). (Optional) Get MetricsCollector's `.status.Count` at any scenario timing if you need.
4. The scenario execution is completed.
5. Get MetricsCollector's `.status.MetricFamilies` (get benchmark results)
(5.5) Calculate the benchmark for any period only using the value of `.status.Count` in (3.5)

#### With a tool
1. Create a MetricsCollector resource and wait until the phase is running
2. (Optional)Get MetricsCollector's `.status.Count`.
3. Execute Processes, instructions, load tests, etc., where scheduling occurs.
(3.5). (Optional) Get MetricsCollector's `.status.Count` at any timing if you need.
4. Complete all tests or operations in (3)
5. Get MetricsCollector's `.status.MetricFamilies` (get benchmark results)
(5.5) Calculate the benchmark for any period only using the value of `.status.Count` in (3.5)

### Issues

The following is a description of the problems that are currently being considered. We cannot be determined until they are tried.

- Metrics misalignment due to scheduler endpoint access delays
- Load of parsing and formatting of acquired metrics
- High throughputSampleFrequency causes high-frequency access to the scheduler endpoint, load, and network congestion
