package aiclient

import "testing"

func TestAITaskRunCapabilityIncludesHintGenerate(t *testing.T) {
	ctx := AITaskRunContext{
		UserID:       "01918fa0-0010-7a00-8a00-000000000001",
		Capability:   AITaskRunTaskHintGenerate,
		ResourceType: AITaskRunResourceTargetJob,
		ResourceID:   "01918fa0-0020-7a00-8a00-000000000002",
	}
	if err := ctx.Validate(); err != nil {
		t.Fatalf("Validate should accept hint_generate: %v", err)
	}
	if _, ok := allowedAITaskRunCapabilities[AITaskRunTaskHintGenerate]; !ok {
		t.Fatalf("allowedAITaskRunCapabilities missing %q", AITaskRunTaskHintGenerate)
	}
}
