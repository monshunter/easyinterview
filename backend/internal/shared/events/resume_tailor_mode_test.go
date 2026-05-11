package events

import "testing"

func TestResumeTailorModeAllowedValues(t *testing.T) {
	allowed := map[ResumeTailorMode]bool{
		ResumeTailorModeGapReview:         true,
		ResumeTailorModeBulletSuggestions: true,
	}

	for _, forbidden := range []ResumeTailorMode{"inline", "rewrite", "mirror"} {
		if allowed[forbidden] {
			t.Errorf("ResumeTailorMode must not include retired value %q", forbidden)
		}
	}
	if got := len(allowed); got != 2 {
		t.Errorf("ResumeTailorMode allowed values = %d, want 2", got)
	}
}
