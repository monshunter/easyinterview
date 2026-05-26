package review

import (
	"context"
	"strings"
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

	got, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), sampleTurns(), sampleRubric())
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
		if len(payload.Metadata.OutputSchema) == 0 {
			t.Fatalf("metadata OutputSchema must be populated")
		}
		if prompt := joinedMessages(payload.Messages); !strings.Contains(prompt, `"name":"depth"`) {
			t.Fatalf("assessment prompt must include rubric dimensions, got %s", prompt)
		}
	}
}

func TestAssessQuestionsMapsScoreLevelToWireStatus(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{`{"dimension_results":{"depth":{"score_level":"proficient","confidence":0.75}},"overall_status":"meets_bar","confidence":0.8,"strengths":["structured"],"gaps":[],"recommended_framework":"STAR","review_status":"open"}`}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nQuestion context: {{question_context}}\nAnswer summary: {{answer_summary}}\nRubric: {{rubric}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	got, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), []TurnSnapshot{{
		ID:             "turn-1",
		TurnIndex:      1,
		QuestionIntent: "architecture",
	}}, sampleRubric())
	if err != nil {
		t.Fatalf("assessQuestionsForAllTurns: %v", err)
	}
	if status := got[0].DimensionResults["depth"].Status; status != sharedtypes.DimensionStatusMeetsBar {
		t.Fatalf("score_level proficient mapped to %q, want meets_bar", status)
	}
}

func TestAssessQuestionsNormalizesRealProviderAssessmentDrift(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{`{"dimension_results":{"depth":{"score_level":"proficient","confidence":0.75}},"confidence":0.8,"strengths":["answer_text label was present"],"gaps":[],"recommended_framework":"STAR","review_status":"ready"}`}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportQuestionAssessmentFeatureKey: reportResolution(reportQuestionAssessmentFeatureKey, "report.assessment.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nQuestion context: {{question_context}}\nAnswer summary: {{answer_summary}}\nRubric: {{rubric}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	got, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), []TurnSnapshot{{
		ID:             "turn-1",
		TurnIndex:      1,
		QuestionIntent: "architecture",
	}}, sampleRubric())
	if err != nil {
		t.Fatalf("assessQuestionsForAllTurns: %v", err)
	}
	if got[0].OverallStatus != sharedtypes.DimensionStatusMeetsBar {
		t.Fatalf("overall_status = %q, want meets_bar", got[0].OverallStatus)
	}
	if got[0].ReviewStatus != sharedtypes.QuestionReviewStatusOpen {
		t.Fatalf("review_status = %q, want open", got[0].ReviewStatus)
	}
	if len(got[0].Strengths) != 1 || got[0].Strengths[0] != "answer summary label was present" {
		t.Fatalf("strengths = %#v", got[0].Strengths)
	}
}

func TestOverallStatusFromDimensionResultsNeedsWorkDominatesMixedStatuses(t *testing.T) {
	results := map[string]DimensionResultDraft{
		"architecture":  {Status: sharedtypes.DimensionStatusMeetsBar},
		"evidence":      {Status: sharedtypes.DimensionStatusNeedsWork},
		"communication": {Status: sharedtypes.DimensionStatusStrong},
	}
	for i := 0; i < 100; i++ {
		if got := overallStatusFromDimensionResults(results); got != sharedtypes.DimensionStatusNeedsWork {
			t.Fatalf("overall status = %q, want needs_work when any dimension needs work", got)
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
	}}, sampleRubric()); err != nil {
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

	got, err := svc.assessQuestionsForAllTurns(context.Background(), sampleSession(), samplePlan(), []TurnSnapshot{{
		ID:             "turn-1",
		TurnIndex:      1,
		QuestionIntent: "debugging",
	}}, sampleRubric())
	if err != nil {
		t.Fatalf("assessQuestionsForAllTurns: %v", err)
	}
	if got[0].Strengths[0] != "question detail leaked" {
		t.Fatalf("strengths = %#v", got[0].Strengths)
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
