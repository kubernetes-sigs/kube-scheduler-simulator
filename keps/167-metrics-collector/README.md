# KEP-167: Metrics Collector

## Summary
Introduce Custom Resource **MetricsCollector** that collects metrics from an endpoint(such as the scheduler) which is specified by the user, and publishes them as time-series data.
That resource(referred to as `MetricsCollector`) could target for collection an endpoint whose metrics are exposed in plain text format that can be read by Prometheus itself or Prometheus scrapers.
These collected data will be pulished via kubernetes resources or on the API endpoint.
We would define this new resource to collect metrics during scheduler benchmarking as time-series data.

## Motivation
The kube-scheduler-simulator can simulate large-scale scheduling thanks to the CRD `Scenario` feature. And now, the next feature we want is the benchmarking tool. By being able to achieve it, the kube-scheduler-simulator would be somewhat complete as a simulator.

As you know, there is a scheduler benchmark tool, `scheduler_perf, implemented on the kubernetes/kubernetes repository. The Scheduler_perf allows us to investigate scheduler metrics, throughput, etc.. when performing some operation for a Pod, Node, and so on that was preplanned by you. About this operation, in our simulator, you can test a more flexible and wider range of conditions than the scheduler_perf by using the `Scenario` feature.

We plan to propose a new Custon Resource Definition `Benchmarker` for benchmarking using `Scenario`. However, we may face a problem. It is how to collect the metrics.
We initially considered implementing it on Benchmarker and having it take on the task of collecting metrics as well, although decided against it for two reasons. 
The first is that the `Benchmarker` works closely with the `Scenario` feature. The `Scenario` is a unique feature of the kube-scheduler-simulator, then the `Benchmarker` is also necessarily unique to the simulator. However, the responsibility to "collect metrics" is a more general feature, and you may want to use it elsewhere in the future.
Such tools should be designed to work standalone as much as possible.
The second is that, in terms of responsibilities, benchmark execution and metrics collection should be separated as another tasks.
For these reasons, the `Benchmarker` is responsible for benchmarking by using the simulator's feature `Scenario` and the general feature `MetricsCollector`, while the `MetricsCollector` is responsible only for collecting metrics.

We also would like to aim to make it as easy as possible to set up the collection and extract metrics for a specific period, to provide the `MetricsCollector` as a general-purpose tool.


### Goals 
- Allow users to obtain scheduler metrics as time-series data.

### Non-Goals
- Execute benchmarking.
- See the result from Web UI. (maybe implemented in the future, but out of the scope of this proposal.)

## User Stories
### Story #1
A user wants to tune the scheduler and would use the simulator's Scenario feature to simulate a production environment. By starting MetricsCollector to collect metrics before the Scenario is run, the user would be able to obtain the scheduler metrics for the period of the Scenario in chronological order after the Scenario is finished.

### Story #2
In the future, we plan to ship a benchmarking tool (`Benchmarker`) that utilizes the `Scenario`. This `Benchmarker` would allow users to more easily perform a variety of benchmarks against the scheduler.
The `MetricsCollector` is the metrics collection tool that works best with the `Benchmarker`. Users can freely specify the type of metrics to be acquired, the time period, and even the frequency of acquisition, allowing them to benchmark according to their own needs.

## Proposal
The MetricsCollector has a simple feature. The ability to periodically collect metrics and publish it as time-series data.
This is because to make the metrics available to tools other than Benchmarker by implementing only simple functionality. This allows it to be use for purposes other than benchmarking.
The following describes the expected operation of each feature and how to use it.

### Feature 1: Collection
This feature continuously collects metrics that are published on the endpoints at the specified destination. Once a MetricsCollector resource is created and ready, it continues to collect metrics, accessing the endpoints at regular intervals. Those Metrics are stored in chronological order in `.status.MetricFamilies`.

Since `.status.MetricFamilies` is an array field, the collected metrics would have array indexes in chronological order. This index represents the order in which the endpoints were invoked.
For example, suppose the MetricsCollector collects metrics at on-second intervals, and the result of the first invoke somes back later than the second. Even in that case, the result of the first is tied to the **first index**.
If the first does not return for some reason, it is recordde there as a blank value with the index. (It is difficult to decide whether a blank value should be nil or a specific value, but if a value close to the values that the metrics can take is adopted, it will be difficult to determine whether there was an error or a normal value, so it is better not to adopt a zero value, for example).

In other words, since the indexes represent a time series, even if the acquisition of data for a metric in the middle fails, the acquisition time of all metrics can be determined without problems from the start time and collection period.

#### How to collect metrics
MetricsCollector collects metrics from the target endpoints uning HTTP, a common API communication protocol. The endpoint requires that the metrics be provided in plain text format that can be read by Prometheus. This is different from how `scheduler_perf` does it on the kubernetes/kubernetes repository.
This is because our project was also intended to support a scheduler running externally, and for this reason we needed a more usually way to reach the metrics.
As mentioned above, this feature expects to collect metrics by invoke the `/metrics/*` API exposed by the scheduler(target) endpoint. **Users should expose the port to the scheduler endpoint in advance.**

The current scheduler in our simulator does not have the ability to collect metrics and expose endpoints. With the implementation of this feature, these modifications will be required.


### Feature 2: Expose
MetricsCollector exposes collected metrics arranged in time series. Those metrics are exposed in two ways: referened as kubernetes resources or provided via APIs in plain text format like other endpoints. Although these ways are different, both provide the same data. The `Benchmarker` works with this `MetricsCollector` using the first way. The second means is provided to make it accessible via HTTP, even from tools that not related to kubernetes. For example, to see the benchmark results, we may provide a WebUI in the future, and this API may be useful in that case as well.

As an option, it is also possible to obtain data only for a specified portion of the period. By asking the user to specify an index of the metrics, MetricsCollector returns the metrics for the period of the specified index. 
Therefore, MetricsCollector would provide an API to return the latest index of the metrics that MetricsCollector have. Since `.status.Count` represents the latest index, it is also ok to refer this instead of the API. This feature allow us to access to the metrics for the period of time you want to capture, so to speak, like a **bookmark**.

### API Examples
- `/currentCount`: Returns the current `.status.Count` value.
- `/metrics/result`: Returns all time-series metrics currently being retrieved.
- `/metrics/result?start=5`: Returns all metrics after index 5 of the time-series metrics currently being retrieved.
`/metrics/result?start=5&end=100: Returns metrics from index 5 onwards to index 100 of the time series metrics currently retrieved.


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
  //
  // In the future, recommended values could be researched and specified by us.
  ThroughputSampleFrequencySeconds int64
}

// MetricsSet manages the name of the metrics to be collected.
// We can be set in the same way as the scheduling-framework plugin
type MetricsSet struct {
  // Stores the names of metrics not to be collected.
  // "*" means that all metrics will not be collected.
  Disabled []string
  
  // Stores the names of metrics to be collected.
  // Empty means that all default metrics are collected.
  Enabled []string
}

type MetricsCollectorStatus struct {
  // Phase is the current phase of this resource.
  Phase CollectorPhase

  // Count is the current number of metrics retrieved.
  // MetricsCollector users can refer to and remember this value to enable retrieval or timing of arbitrary timing metrics.
  // This keeps it loosely coupled with `Scenario`.
  // e.g.) len(MetricFamilies[i].Metric) == 3 → count == 3
  Count int

  // StartAt represents the time when the MetricsCollector resource was created and started collecting. 
  // The use can later calculate when the metrics were collected by using this value,
  // `thegroupputSampleFrequencySeconds` and the index of each metric in `MetricFamilies`.
  StartAt time.Time

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
    disabled:
      - hoge-metrics
```

### Filtering of Metrics
Among the metrics, users can specify only those items they wish to collect.
Users can specify enabled and disabled like the scheduler-framework plugins. It should also be possible to do the same for configuration.
The default will be to collect all metrics published by the scheduler.

### Calculating throughput of scheduler
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
Since this is an important point, we would explain it again.
MetricsCollector periodically collects metrics from the scheduler's api endpoints at the time interval specified by the `.spec.theholdputSampleFrequencySeconds` value.
This means that metrics collection will begin when the MetricsCollector is applied and the phase is running.

Therefore, if the user wants to start collecting metrics at an arbitrary time, you should use `.status.Count`.
By user refer to and record the .status.Count values at the start and end of the benchmark, the user can obtain only the indicators during the test run.
This method is also useful if you want to see metrics for any period during the test.

However, this does not allow users to start collecting metrics atarbitrary times. To achieve this , the `.status.Count` (index) is used. Get the value of count at the start of the benchmark. Then, when finished, retrieve the metrics by specifying the index after the value of count as described earlier. With just this, users get only the metrics they want. (e.g., during benchmarking). Of course, both the start and end of the index can be specified. For the most part, user would want to specify the start and end for use.

By adopting the above approach, we can avoid establishing a close relationship with the `scenario` to collect metrics. 
This is why MetricsCollector is a more generally available resource(tool).

### Metrics exposing and results calculating 
MetricsCollector exposes time-series metrics. This means that users only need to query the MetricsCollector once at the end of the period for which users need to retrieve metrics to get all the data. **There is no need to access the MetricsCollector sequentially to check the values of the metrics.**

Also, the metrics are only the values obtained would be exposed without modification or calculation. For example, MetricsCollector does not calculate averages, etc., but leaves this to the user. Collection is the only purpose of the MetricsCollector. 

If it is useful to have such a calculation function in the future, we will consider introducing it, but it should be limited to cases where the calculation requires the use of internal values such as `thegroupputSampleFrequencySeconds`. In other words, we would not introduce calculations that are useful only for specific tools.




### How to use 
#### With `Benchmarker`
1. Create a MetricsCollector resource by Benchmarker.
2. MetricsCollector's `Phase` becomes running and starts collecting.
3. Get MetricsCollector's `.status.Count` at the start of the benchmark testing by Benchmarker.
4. (Optional) Get MetricsCollector's `.status.Count` at any time you want by Benchmarker.
5. Finish the Benchmarker's execution.
6. Get MetricsCollector's `.status.MetricFamilies` (get benchmark results)  by Benchmarker.
7. Calculate the benchmark results from the metrics (metrics for any period can be checked using the value of 4).

#### With an external tool
1. Create a MetricsCollector resource and wait until the phase is running
2. (Optional)Get MetricsCollector's `.status.Count` (via API).
3. Execute Processes, instructions, load tests, etc., where scheduling occurs by the tool.
4. (Optional) Get MetricsCollector's `.status.Count` at any time if you need it (via API).
5. Complete all tests or operations in 3.
6. Get MetricsCollector's `.status.MetricFamilies` (via API). 
7. Calculate the benchmark results from the metrics (metrics for any period can be checked using the value of 4).

## Compare the other tools
There are several tools that collect metrics.

First is `Prometheus`, a well-known monitoring system that hs appeared frequently in the context of kubernetes and can be used in conjunction with various exporters to monitor specific software and more.
For example, Prometheus is used in `vertial-pod-autoscalor` to monitor multiple targets by deploying Prometheus in a cluster.
The reason we did not adopt Prometheus instead of MetricsCollector, is that Prometheus is only a monitoring tool. We don't need the feature to raise alerts and don't have multiple monitoring targets. There are many unnecessary features. 
In contrast, Prometheus does not have string or log collection capabilities. We needed to be able to overlay the execution history of the Benchmarker, such as Node, Pod creation, etc., on the metrics, but we can't do that.

Another perspective is the integration aspect with Benchmarker, as MetricsCollector keeps the collected metrics as a k8s resource, making it easy to integrate with Benchmarker. This is one of the main reasons to define MetricsCollector anew.

Then there is metrics-server. This one collects metrics from kubelet so that they can be retrieved from kubernetes api-server, and has a slightly different usage. The uses for this tool are slightly different.
It should also be noted that it must be deployed and used in a cluster, so the user have to prepare a cluster.

I also found a tool called kube-state-metrics, but it is not a replacement for MetricsCollector since it only monitors kubernetes objects.

Considering the integration aspect with the Benchmarker, the required features, and the possibility of general use in various scenarios, we propose the implementation of MetricsCollector.

### Issues
The following is a description of the problems that are currently being considered. We cannot be determined until they are tried.

- Metrics misalignment due to scheduler endpoint access delays
- Load of parsing and formatting of acquired metrics
- High throughputSampleFrequency causes high-frequency access to the scheduler endpoint, load, and network congestion
