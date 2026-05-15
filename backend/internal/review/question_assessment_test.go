package review

import (
	"context"
	"errors"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestAssessQuestionsForAllTurns(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{
		`{"dimension_results":{"depth":{"status":"meets_bar","confidence":0.75}},"overall_status":"meets_bar","confidence":0.8,"strengths":["structured"],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`,
		`{"dimension_results":{"depth":{"status":"needs_work","confidence":0.66}},"overall_status":"needs_work","confidence":0.7,"strengths":[],"gaps":["missed tradeoff"],"recommended_framework":"STAR","review_status":"queued_for_retry"}`,
	}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nQuestion context: {{question_context}}\nAnswer summary: {{answer_summary}}\nRubric: {{rubric}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	got, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), sampleTurns())
	if err != nil {
		t.Fatalf("assessQuestionsForAllTurns: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].TurnID != "turn-1" || got[1].TurnID != "turn-2" {
		t.Fatalf("turn order = %+v", got)
	}
	if got[1].OverallStatus != sharedtypes.DimensionStatusNeedsWork ||
		got[1].ReviewStatus != sharedtypes.QuestionReviewStatusQueuedForRetry ||
		got[1].RecommendedFramework != "STAR" {
		t.Fatalf("assessment[1] = %+v", got[1])
	}
	if len(ai.payloads) != 2 {
		t.Fatalf("AI calls = %d, want 2", len(ai.payloads))
	}
	for _, payload := range ai.payloads {
		if payload.Metadata.FeatureKey != reportQuestionAssessmentFeatureKey ||
			payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportAssessment {
			t.Fatalf("metadata = %+v", payload.Metadata)
		}
	}
}

func TestAssessQuestionsBuildsPromptWithoutLeaks(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{`{"dimension_results":{"depth":{"status":"needs_work","confidence":0.5,"score":0.5}},"overall_status":"needs_work","confidence":0.5,"strengths":[],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nQuestion context: {{question_context}}\nAnswer summary: {{answer_summary}}\nRubric: {{rubric}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	if _, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), []TurnSnapshot{{
		ID:              "turn-1",
		TurnIndex:       1,
		QuestionIntent:  "debugging",
		QuestionContext: "question_text appears in source",
		AnswerSummary:   "answer_text and hint_text appear in source",
	}}); err != nil {
		t.Fatalf("assessQuestionsForAllTurns: %v", err)
	}
	assertNoPromptLeak(t, joinedMessages(ai.payloads[0].Messages))
}

func TestQuestionAssessmentParsesWithoutLeaks(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{`{"dimension_results":{"depth":{"status":"needs_work","confidence":0.5,"score":0.5}},"overall_status":"needs_work","confidence":0.5,"strengths":["question_text leaked"],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Question context: {{question_context}}\nAnswer summary: {{answer_summary}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	if _, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), []TurnSnapshot{{
		ID:             "turn-1",
		TurnIndex:      1,
		QuestionIntent: "debugging",
	}}); !errors.Is(err, ErrReviewAIOutputInvalid) {
		t.Fatalf("err = %v, want ErrReviewAIOutputInvalid", err)
	}
}

type fakePromptResolver struct {
	resolutions map[string]registry.PromptResolution
	errs        map[string]error
}

func (f fakePromptResolver) ResolveActive(_ context.Context, featureKey string, _ string) (registry.PromptResolution, error) {
	if err := f.errs[featureKey]; err != nil {
		return registry.PromptResolution{}, err
	}
	return f.resolutions[featureKey], nil
}

type fakeReviewAI struct {
	responses []string
	errs      []error
	payloads  []aiclient.CompletePayload
	profiles  []string
}

func (f *fakeReviewAI) Complete(_ context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.profiles = append(f.profiles, profileName)
	f.payloads = append(f.payloads, payload)
	idx := len(f.payloads) - 1
	if idx < len(f.errs) && f.errs[idx] != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, f.errs[idx]
	}
	content := `{}`
	if idx < len(f.responses) {
		content = f.responses[idx]
	}
	return aiclient.CompleteResponse{Content: content}, aiclient.AICallMeta{}, nil
}
