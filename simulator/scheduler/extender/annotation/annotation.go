package annotation

const (
	// ExtenderFilterResultAnnotationKey has the filtering result of extender.
	ExtenderFilterResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/extender-filter-result"
	// ExtenderPrioritizeResultAnnotationKey has the prioritizing result of extender.
	ExtenderPrioritizeResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/extender-prioritize-result"
	// ExtenderPreemptResultAnnotationKey has the preemption result of extender.
	ExtenderPreemptResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/extender-preempt-result"
	// ExtenderBindResultAnnotationKey has the binding result of extender.
	ExtenderBindResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/extender-bind-result"
)
