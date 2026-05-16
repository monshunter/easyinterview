package review

import sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"

type NextActionType string

const (
	NextActionRetryCurrentRound NextActionType = "retry_current_round"
	NextActionNextRound         NextActionType = "next_round"
	NextActionReviewEvidence    NextActionType = "review_evidence"
)

func decideNextAction(readiness sharedtypes.ReadinessTier, retryFocusCount int) NextActionType {
	switch readiness {
	case sharedtypes.ReadinessTierNotReady, sharedtypes.ReadinessTierNeedsPractice:
		if retryFocusCount >= 1 {
			return NextActionRetryCurrentRound
		}
	case sharedtypes.ReadinessTierBasicallyReady, sharedtypes.ReadinessTierWellPrepared:
		if retryFocusCount < 3 {
			return NextActionNextRound
		}
	}
	return NextActionReviewEvidence
}
