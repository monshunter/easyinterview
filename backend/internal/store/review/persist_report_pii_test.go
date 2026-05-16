package review

import "testing"

func TestPersistReportRedactsRawText(t *testing.T) {
	err := assertNoReviewPersistencePII(map[string]any{
		"highlights": []map[string]any{{
			"dimension":  "depth",
			"evidence":   "question_text leaked",
			"confidence": "high",
		}},
	})
	if err == nil {
		t.Fatal("expected forbidden raw text token to be rejected")
	}
}
