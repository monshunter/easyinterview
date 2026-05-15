package review_test

import (
	"testing"

	sharedevents "github.com/monshunter/easyinterview/backend/internal/shared/events"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	reviewstore "github.com/monshunter/easyinterview/backend/internal/store/review"
)

func TestReportGeneratedPayload(t *testing.T) {
	payload, err := reviewstore.BuildReportGeneratedPayload(reviewstore.ReportGeneratedInput{
		ReportID:           "0197d120-0000-7000-8000-000000000200",
		SessionID:          "0197d120-0000-7000-8000-000000000201",
		TargetJobID:        "0197d120-0000-7000-8000-000000000202",
		PreparednessLevel:  sharedtypes.ReadinessTierBasicallyReady,
		QuestionIssueCount: 2,
		PromptVersion:      "v0.1.0",
		RubricVersion:      "v0.1.0",
		ModelID:            "model-profile:report.generate.default",
	})
	if err != nil {
		t.Fatalf("BuildReportGeneratedPayload: %v", err)
	}
	if payload.ReportID == "" || payload.PreparednessLevel != sharedtypes.ReadinessTierBasicallyReady || payload.QuestionIssueCount != 2 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestReportGenerationFailedPayload(t *testing.T) {
	payload, err := reviewstore.BuildReportGenerationFailedPayload(reviewstore.ReportGenerationFailedInput{
		ReportID:  "0197d120-0000-7000-8000-000000000200",
		SessionID: "0197d120-0000-7000-8000-000000000201",
		ErrorCode: "AI_PROVIDER_TIMEOUT",
		Retryable: true,
	})
	if err != nil {
		t.Fatalf("BuildReportGenerationFailedPayload: %v", err)
	}
	if payload.ErrorCode != "AI_PROVIDER_TIMEOUT" || !payload.Retryable {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestReportOutboxPayloadRejectsPII(t *testing.T) {
	_, err := reviewstore.BuildReportGeneratedPayload(reviewstore.ReportGeneratedInput{
		ReportID:           "0197d120-0000-7000-8000-000000000200",
		SessionID:          "0197d120-0000-7000-8000-000000000201",
		TargetJobID:        "0197d120-0000-7000-8000-000000000202",
		PreparednessLevel:  sharedtypes.ReadinessTierNeedsPractice,
		QuestionIssueCount: 1,
		PromptVersion:      "v0.1.0",
		RubricVersion:      "v0.1.0",
		ModelID:            "prompt_body",
	})
	if err == nil {
		t.Fatal("expected PII boundary error")
	}
	if string(sharedevents.EventNameReportGenerated) != "report.generated" {
		t.Fatalf("event name drift: %s", sharedevents.EventNameReportGenerated)
	}
}
