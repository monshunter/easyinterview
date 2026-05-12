package events

import "testing"

func TestResumeTailorModeAllowedValues(t *testing.T) {
	allowed := map[ResumeTailorMode]bool{
		ResumeTailorModeGapReview:         true,
		ResumeTailorModeBulletSuggestions: true,
	}

	for _, parts := range [][]string{{"in", "line"}, {"re", "write"}, {"mir", "ror"}} {
		forbidden := ResumeTailorMode(parts[0] + parts[1])
		if allowed[forbidden] {
			t.Errorf("ResumeTailorMode must not include retired value %q", forbidden)
		}
	}
	if got := len(allowed); got != 2 {
		t.Errorf("ResumeTailorMode allowed values = %d, want 2", got)
	}
}
