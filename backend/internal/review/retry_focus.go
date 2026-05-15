package review

import (
	"sort"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func selectRetryFocusTurnIDs(assessments []QuestionAssessmentDraft) []string {
	indexes := make([]int, 0, len(assessments))
	for i := range assessments {
		if assessments[i].OverallStatus == sharedtypes.DimensionStatusNeedsWork ||
			assessments[i].ReviewStatus == sharedtypes.QuestionReviewStatusQueuedForRetry {
			indexes = append(indexes, i)
		}
	}
	sort.SliceStable(indexes, func(i, j int) bool {
		return assessments[indexes[i]].TurnIndex < assessments[indexes[j]].TurnIndex
	})
	if len(indexes) > 5 {
		indexes = indexes[:5]
	}
	out := make([]string, 0, len(indexes))
	for _, idx := range indexes {
		assessments[idx].IncludedInRetryPlan = true
		out = append(out, assessments[idx].TurnID)
	}
	return out
}
