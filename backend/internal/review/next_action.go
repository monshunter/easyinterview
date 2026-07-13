package review

type NextActionType string

const (
	NextActionRetryCurrentRound NextActionType = "retry_current_round"
	NextActionNextRound         NextActionType = "next_round"
	NextActionReviewEvidence    NextActionType = "review_evidence"
)
