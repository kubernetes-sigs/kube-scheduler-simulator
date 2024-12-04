package annotation

const (
	// PreFilterStatusResultAnnotationKey has the prefilter result(framework.Status).
	PreFilterStatusResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/prefilter-result-status"
	// PreFilterResultAnnotationKey has the prefilter result(framework.PreFilterResult).
	PreFilterResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/prefilter-result"
	// FilterResultAnnotationKey has the filtering result.
	FilterResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/filter-result"
	// PostFilterResultAnnotationKey has the post filter result.
	PostFilterResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/postfilter-result"
	// PreScoreResultAnnotationKey has the prescore result.
	PreScoreResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/prescore-result"
	// ScoreResultAnnotationKey has the scoring result.
	ScoreResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/score-result"
	// FinalScoreResultAnnotationKey has the final score(= normalized and applied score plugin weight).
	FinalScoreResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/finalscore-result"
	// ReserveResultAnnotationKey has the reserve result.
	ReserveResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/reserve-result"
	// PermitStatusResultAnnotationKey has the permit result.
	PermitStatusResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/permit-result"
	// PermitTimeoutResultAnnotationKey has the permit result.
	PermitTimeoutResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/permit-result-timeout"
	// PreBindResultAnnotationKey has the prebind result.
	PreBindResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/prebind-result"
	// BindResultAnnotationKey has the prebind result.
	BindResultAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/bind-result"
	// SelectedNodeAnnotationKey has the selected node name. It's filled when a Pod go through the Reserve phase.
	SelectedNodeAnnotationKey = "kube-scheduler-simulator.sigs.k8s.io/selected-node"
)
