package annotation

const (
	// PreFilterStatusResultAnnotationKey has the prefilter result(framework.Status).
	PreFilterStatusResultAnnotationKey = "scheduler-simulator/prefilter-result-status"
	// PreFilterResultAnnotationKey has the prefilter result(framework.PreFilterResult).
	PreFilterResultAnnotationKey = "scheduler-simulator/prefilter-result"
	// FilterResultAnnotationKey has the filtering result.
	FilterResultAnnotationKey = "scheduler-simulator/filter-result"
	// PostFilterResultAnnotationKey has the post filter result.
	PostFilterResultAnnotationKey = "scheduler-simulator/postfilter-result"
	// PreScoreResultAnnotationKey has the prescore result.
	PreScoreResultAnnotationKey = "scheduler-simulator/prescore-result"
	// ScoreResultAnnotationKey has the scoring result.
	ScoreResultAnnotationKey = "scheduler-simulator/score-result"
	// FinalScoreResultAnnotationKey has the final score(= normalized and applied score plugin weight).
	FinalScoreResultAnnotationKey = "scheduler-simulator/finalscore-result"
	// ReserveResultAnnotationKey has the reserve result.
	ReserveResultAnnotationKey = "scheduler-simulator/reserve-result"
	// PermitStatusResultAnnotationKey has the permit result.
	PermitStatusResultAnnotationKey = "scheduler-simulator/permit-result"
	// PermitTimeoutResultAnnotationKey has the permit result.
	PermitTimeoutResultAnnotationKey = "scheduler-simulator/permit-result-timeout"
	// PreBindResultAnnotationKey has the prebind result.
	PreBindResultAnnotationKey = "scheduler-simulator/prebind-result"
	// BindResultAnnotationKey has the prebind result.
	BindResultAnnotationKey = "scheduler-simulator/bind-result"
	// SelectedNodeAnnotationKey has the selected node name. It's filled when a Pod go through the Reserve phase.
	SelectedNodeAnnotationKey = "scheduler-simulator/selected-node"
)
