package aiclient

import "testing"

func TestAITaskRunCapabilityIncludesPracticeChat(t *testing.T) {
	ctx := AITaskRunContext{
		UserID:       "01918fa0-0010-7a00-8a00-000000000001",
		Capability:   AITaskRunTaskPracticeChat,
		ResourceType: AITaskRunResourceTargetJob,
		ResourceID:   "01918fa0-0020-7a00-8a00-000000000002",
	}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("Validate should accept practice_chat: %v", err)
	}
}

func TestAITaskRunCapabilityIncludesReportGenerate(t *testing.T) {
	ctx := AITaskRunContext{
		UserID:       "01918fa0-0010-7a00-8a00-000000000001",
		Capability:   AITaskRunTaskReportGenerate,
		ResourceType: AITaskRunResourceFeedbackReport,
		ResourceID:   "01918fa0-0020-7a00-8a00-000000000002",
	}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("Validate should accept report_generate: %v", err)
	}
	if _, ok := allowedAITaskRunCapabilities[AITaskRunTaskReportGenerate]; !ok {
		t.Fatalf("allowedAITaskRunCapabilities missing %q", AITaskRunTaskReportGenerate)
	}
}

func TestRemovedStructuredTaskCapabilitiesAreRejected(t *testing.T) {
	ctx := AITaskRunContext{
		UserID:       "01918fa0-0010-7a00-8a00-000000000001",
		Capability:   AITaskRunCapability("question_generate"),
		ResourceType: AITaskRunResourceFeedbackReport,
		ResourceID:   "01918fa0-0020-7a00-8a00-000000000002",
	}
	if err := ctx.Validate(); err == nil {
		t.Fatal("Validate should reject removed question task capability")
	}
}
