package review

import (
	"testing"

	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestDecideNextAction(t *testing.T) {
	for _, tc := range []struct {
		readiness sharedtypes.ReadinessTier
		count     int
		want      NextActionType
	}{
		{readiness: sharedtypes.ReadinessTierNotReady, count: 0, want: NextActionReviewEvidence},
		{readiness: sharedtypes.ReadinessTierNotReady, count: 1, want: NextActionRetryCurrentRound},
		{readiness: sharedtypes.ReadinessTierNeedsPractice, count: 2, want: NextActionRetryCurrentRound},
		{readiness: sharedtypes.ReadinessTierBasicallyReady, count: 0, want: NextActionNextRound},
		{readiness: sharedtypes.ReadinessTierBasicallyReady, count: 2, want: NextActionNextRound},
		{readiness: sharedtypes.ReadinessTierBasicallyReady, count: 3, want: NextActionReviewEvidence},
		{readiness: sharedtypes.ReadinessTierWellPrepared, count: 0, want: NextActionNextRound},
		{readiness: sharedtypes.ReadinessTierWellPrepared, count: 5, want: NextActionReviewEvidence},
	} {
		if got := decideNextAction(tc.readiness, tc.count); got != tc.want {
			t.Fatalf("%s/%d = %s, want %s", tc.readiness, tc.count, got, tc.want)
		}
	}
}
