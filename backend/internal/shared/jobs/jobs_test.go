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

func TestGeneratedJobMappings(t *testing.T) {
	if AsynqTaskTargetImport != "target.import" {
		t.Fatalf("AsynqTaskTargetImport = %q", AsynqTaskTargetImport)
	}
	if AsynqTaskPrivacyDelete != "privacy.delete" {
		t.Fatalf("AsynqTaskPrivacyDelete = %q", AsynqTaskPrivacyDelete)
	}
	if AsynqTaskEmailDispatch != "email.dispatch" {
		t.Fatalf("AsynqTaskEmailDispatch = %q", AsynqTaskEmailDispatch)
	}
	if JobTriggerEventSemanticSourceEventOnly != "source_event_only" {
		t.Fatalf("JobTriggerEventSemanticSourceEventOnly = %q", JobTriggerEventSemanticSourceEventOnly)
	}
	if JobTriggerEventSemanticTriggerCreatesJob != "trigger_creates_job" {
		t.Fatalf("JobTriggerEventSemanticTriggerCreatesJob = %q", JobTriggerEventSemanticTriggerCreatesJob)
	}
	if JobTriggerEventSemantics[JobTypeReportGenerate] != JobTriggerEventSemanticSourceEventOnly {
		t.Fatalf("report_generate semantic = %q", JobTriggerEventSemantics[JobTypeReportGenerate])
	}
	if !IsSourceEventOnly(JobTypeReportGenerate) {
		t.Fatalf("report_generate must be source_event_only")
	}
	if IsSourceEventOnly(JobTypeTargetImport) {
		t.Fatalf("target_import must remain trigger_creates_job")
	}
	for _, internalOnly := range []JobType{JobTypeSourceRefresh, JobTypeEmailDispatch} {
		if containsJobType(APIFacingJobTypes, internalOnly) {
			t.Fatalf("%s must stay out of APIFacingJobTypes", internalOnly)
		}
	}
}

func containsJobType(types []JobType, want JobType) bool {
	for _, typ := range types {
		if typ == want {
			return true
		}
	}
	return false
}
