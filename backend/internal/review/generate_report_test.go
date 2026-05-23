package review

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
)

func TestGenerateReportContentSuccess(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{`{
		"summary":"mostly ready",
		"dimension_scores":[{"name":"communication","score":0.82,"reasoning":"structured","supporting_observations":["clear opening"]}],
		"highlights":[{"dimension":"communication","evidence":"clear opening","confidence":0.8}],
		"issues":[{"dimension":"depth","evidence":"missed tradeoff","confidence":0.7}],
		"next_actions":[{"type":"retry_current_round","label":"Retry architecture tradeoff"}],
		"retry_focus_turn_ids":["turn-2"]
	}`}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportGenerateFeatureKey: reportResolution(reportGenerateFeatureKey, "report.generate.default", "Generate report\nSession metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nRubric: {{rubric_dimensions}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	draft, err := svc.generateReportContent(context.Background(), sampleSession(), samplePlan(), sampleTurns())
	if err != nil {
		t.Fatalf("generateReportContent: %v", err)
	}
	if draft.Summary != "mostly ready" || len(draft.Highlights) != 1 || len(draft.Issues) != 1 || len(draft.NextActions) != 1 {
		t.Fatalf("draft = %+v", draft)
	}
	if len(draft.DimensionScores) != 1 || draft.DimensionScores[0].Name != "communication" {
		t.Fatalf("dimension scores = %+v", draft.DimensionScores)
	}
	if ai.profiles[0] != "report.generate.default" {
		t.Fatalf("profile = %q", ai.profiles[0])
	}
	if ai.payloads[0].Metadata.FeatureKey != reportGenerateFeatureKey ||
		ai.payloads[0].Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportGenerate ||
		ai.payloads[0].Metadata.TaskRun.ResourceType != aiclient.AITaskRunResourceFeedbackReport {
		t.Fatalf("metadata = %+v", ai.payloads[0].Metadata)
	}
	if len(ai.payloads[0].Metadata.OutputSchema) == 0 {
		t.Fatalf("metadata OutputSchema must be populated")
	}
}

func TestGenerateReportContentBuildsPromptWithoutLeaks(t *testing.T) {
	ai := &fakeReviewAI{responses: []string{`{"summary":"ok","highlights":[],"issues":[],"next_actions":[]}`}}
	svc := NewService(ServiceOptions{
		Registry: fakePromptResolver{resolutions: map[string]registry.PromptResolution{
			reportGenerateFeatureKey: reportResolution(reportGenerateFeatureKey, "report.generate.default", "Session metadata: {{session_metadata}}\nTurn summaries: {{turn_summaries}}\nRubric: {{rubric_dimensions}}\nReturn strict JSON."),
		}},
		AI: ai,
	})

	if _, err := svc.generateReportContent(context.Background(), sampleSession(), samplePlan(), []TurnSnapshot{{
		ID:              "turn-raw",
		TurnIndex:       1,
		QuestionIntent:  "system-design",
		QuestionContext: "question_text should be redacted",
		AnswerSummary:   "answer_text and hint_text should be redacted",
	}}); err != nil {
		t.Fatalf("generateReportContent: %v", err)
	}
	assertNoPromptLeak(t, joinedMessages(ai.payloads[0].Messages))
}

func sampleSession() SessionSnapshot {
	return SessionSnapshot{
		UserID:      "0197d120-0000-7000-8000-000000000100",
		ReportID:    "0197d120-0000-7000-8000-000000000101",
		SessionID:   "0197d120-0000-7000-8000-000000000102",
		PlanID:      "0197d120-0000-7000-8000-000000000103",
		TargetJobID: "0197d120-0000-7000-8000-000000000104",
		Language:    "en",
	}
}

func samplePlan() PracticePlanSnapshot {
	return PracticePlanSnapshot{ID: "0197d120-0000-7000-8000-000000000103", Goal: "backend-review", Mode: "strict", InterviewerPersona: "staff engineer"}
}

func sampleTurns() []TurnSnapshot {
	return []TurnSnapshot{
		{ID: "turn-2", TurnIndex: 2, QuestionIntent: "tradeoff", QuestionContext: "cache invalidation", AnswerSummary: "named consistency risks"},
		{ID: "turn-1", TurnIndex: 1, QuestionIntent: "architecture", QuestionContext: "service boundaries", AnswerSummary: "explained queue boundaries"},
	}
}

func reportResolution(featureKey string, profile string, template string) registry.PromptResolution {
	return registry.PromptResolution{
		FeatureKey:          featureKey,
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    profile,
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		OutputSchema:        reviewOutputSchema(`{"type":"object","required":["summary"],"properties":{"summary":{"type":"string"}}}`),
		UserMessageTemplate: template,
	}
}

func reviewOutputSchema(s string) *json.RawMessage {
	raw := json.RawMessage(s)
	return &raw
}

func joinedMessages(messages []aiclient.Message) string {
	var b strings.Builder
	for _, msg := range messages {
		b.WriteString(msg.Content)
		b.WriteByte('\n')
	}
	return b.String()
}

func assertNoPromptLeak(t *testing.T, prompt string) {
	t.Helper()
	for _, forbidden := range []string{"question_text", "answer_text", "hint_text", "prompt body", "response body"} {
		if strings.Contains(strings.ToLower(prompt), forbidden) {
			t.Fatalf("prompt leaked forbidden token %q: %s", forbidden, prompt)
		}
	}
}
