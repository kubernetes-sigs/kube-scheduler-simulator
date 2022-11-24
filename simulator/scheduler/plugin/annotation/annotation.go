package annotation

const (
	// FilterResultAnnotationKey has the filtering result.
	FilterResultAnnotationKey = "scheduler-simulator/filter-result"

	PostFilterResultAnnotationKey = "scheduler-simulator/postFilter-result"
	// ScoreResultAnnotationKey has the scoring result.
	ScoreResultAnnotationKey = "scheduler-simulator/score-result"
	// FinalScoreResultAnnotationKey has the final score(= normalized and applied score plugin weight).
	FinalScoreResultAnnotationKey = "scheduler-simulator/finalscore-result"
)
