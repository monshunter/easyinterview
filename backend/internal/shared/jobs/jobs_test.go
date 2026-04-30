package jobs

import "testing"

func TestBuildEmailDispatchPayload(t *testing.T) {
	t.Run("rejects redacted fields", func(t *testing.T) {
		for _, field := range []string{"rawMagicLinkToken", "magicLinkUrl", "recipientEmail", "emailBody"} {
			_, err := BuildEmailDispatchPayload(map[string]string{
				"authChallengeId": "challenge-1",
				field:             "secret",
			})
			if err == nil {
				t.Fatalf("BuildEmailDispatchPayload allowed redacted field %s", field)
			}
		}
	})

	t.Run("accepts allowed fields", func(t *testing.T) {
		payload, err := BuildEmailDispatchPayload(map[string]string{
			"authChallengeId":   "challenge-1",
			"userId":            "user-1",
			"templateKey":       "login",
			"locale":            "en-US",
			"deliverySecretRef": "secret-ref-1",
			"dedupeKey":         "dedupe-1",
		})
		if err != nil {
			t.Fatalf("BuildEmailDispatchPayload: %v", err)
		}
		if payload["deliverySecretRef"] != "secret-ref-1" {
			t.Fatalf("payload did not preserve allowed field: %#v", payload)
		}
	})
}
