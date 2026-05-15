package review

import (
	"fmt"
	"testing"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestSelectRetryFocusTurns(t *testing.T) {
	t.Run("mixed_and_ordered", func(t *testing.T) {
		assessments := []QuestionAssessmentDraft{
			{TurnID: "turn-3", TurnIndex: 3, OverallStatus: sharedtypes.DimensionStatusStrong},
			{TurnID: "turn-1", TurnIndex: 1, OverallStatus: sharedtypes.DimensionStatusNeedsWork},
			{TurnID: "turn-2", TurnIndex: 2, ReviewStatus: sharedtypes.QuestionReviewStatusQueuedForRetry},
		}
		got := selectRetryFocusTurnIDs(assessments)
		if fmt.Sprint(got) != "[turn-1 turn-2]" {
			t.Fatalf("got %v", got)
		}
		if !assessments[1].IncludedInRetryPlan || !assessments[2].IncludedInRetryPlan || assessments[0].IncludedInRetryPlan {
			t.Fatalf("included flags = %+v", assessments)
		}
	})
	t.Run("all_strong_empty", func(t *testing.T) {
		got := selectRetryFocusTurnIDs([]QuestionAssessmentDraft{{TurnID: "turn-1", TurnIndex: 1, OverallStatus: sharedtypes.DimensionStatusStrong}})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
	t.Run("max_five", func(t *testing.T) {
		assessments := make([]QuestionAssessmentDraft, 0, 7)
		for i := 7; i >= 1; i-- {
			assessments = append(assessments, QuestionAssessmentDraft{
				TurnID:        fmt.Sprintf("turn-%d", i),
				TurnIndex:     i,
				OverallStatus: sharedtypes.DimensionStatusNeedsWork,
			})
		}
		got := selectRetryFocusTurnIDs(assessments)
		if fmt.Sprint(got) != "[turn-1 turn-2 turn-3 turn-4 turn-5]" {
			t.Fatalf("got %v", got)
		}
	})
}
