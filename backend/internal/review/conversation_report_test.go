package review

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestGenerateReportUsesOneConversationLevelAICall(t *testing.T) {
	ai := &conversationReportAI{response: `{
		"summary":"mostly ready",
		"dimension_scores":[{"name":"communication","score":4.2,"reasoning":"clear","supporting_observations":["owned the migration","quantified impact"]},{"name":"technical_depth","score":2.4,"reasoning":"thin tradeoffs","supporting_observations":["named caching"]}],
		"highlights":[{"dimension":"communication","evidence":"owned the migration","confidence":0.9}],
		"issues":[{"dimension":"technical_depth","evidence":"missing rollout tradeoffs","confidence":0.7}],
		"next_actions":[{"type":"retry_current_round","label":"Practice rollout tradeoffs"}],
		"retry_focus_competency_codes":["technical_depth"]
	}`}
	repo := &conversationReportRepository{ctx: ReportContext{
		Session: SessionSnapshot{UserID: testUUID(1), ReportID: testUUID(2), SessionID: testUUID(3), PlanID: testUUID(4), TargetJobID: testUUID(5), Language: "en"},
		Plan:    PracticePlanSnapshot{ID: testUUID(4), Goal: "baseline", InterviewerPersona: "hiring_manager"},
		Messages: []MessageSnapshot{
			{Role: "assistant", Content: "What changed?", SeqNo: 3},
			{Role: "assistant", Content: "Tell me about the migration.", SeqNo: 1},
			{Role: "user", Content: "I led it across three teams.", SeqNo: 2},
		},
		Rubric: registry.RubricSchema{Dimensions: []registry.RubricDimension{{Name: "communication", Weight: .5}, {Name: "technical_depth", Weight: .5}}},
	}}
	schema := json.RawMessage(`{"type":"object"}`)
	svc := NewService(ServiceOptions{
		Registry: conversationPromptResolver{resolution: registry.PromptResolution{
			FeatureKey: reportGenerateFeatureKey, PromptVersion: "v0.1.0", RubricVersion: "v0.1.0", ModelProfileName: "report.generate.default",
			FeatureFlag: "none", DataSourceVersion: "registry.v1", OutputSchema: &schema,
			UserMessageTemplate: "Session: {{session_metadata}}\nConversation: {{conversation_messages}}\nRubric: {{rubric_dimensions}}",
		}},
		AI: ai, Repository: repo,
		Now:   func() time.Time { return time.Date(2026, 7, 12, 8, 30, 0, 0, time.UTC) },
		NewID: fixedConversationIDs(testUUID(6), testUUID(7)),
	})

	outcome := svc.GenerateReport(context.Background(), AsyncJob{JobID: testUUID(8), ResourceID: testUUID(2)})
	if !outcome.Succeeded || !outcome.AsyncJobFinalized {
		t.Fatalf("outcome = %+v", outcome)
	}
	if len(ai.payloads) != 1 {
		t.Fatalf("AI calls = %d, want exactly one conversation-level call", len(ai.payloads))
	}
	payload := ai.payloads[0]
	if payload.Metadata.TaskRun.Capability != aiclient.AITaskRunTaskReportGenerate {
		t.Fatalf("capability = %q", payload.Metadata.TaskRun.Capability)
	}
	prompt := joinedConversationMessages(payload.Messages)
	first := strings.Index(prompt, "Tell me about the migration.")
	second := strings.Index(prompt, "I led it across three teams.")
	third := strings.Index(prompt, "What changed?")
	if !(first >= 0 && first < second && second < third) {
		t.Fatalf("conversation not ordered by seqNo: %s", prompt)
	}
	for _, stale := range []string{"turnId", "questionAssessment", "retryFocusTurnIds", "questionBudget", "hint"} {
		if strings.Contains(prompt, stale) {
			t.Fatalf("prompt contains stale structured field %q: %s", stale, prompt)
		}
	}
	if repo.persisted.PreparednessLevel != sharedtypes.ReadinessTierBasicallyReady {
		t.Fatalf("preparedness = %q", repo.persisted.PreparednessLevel)
	}
	if len(repo.persisted.DimensionAssessments) != 2 || repo.persisted.DimensionAssessments[0].Status != sharedtypes.DimensionStatusStrong || repo.persisted.DimensionAssessments[1].Status != sharedtypes.DimensionStatusNeedsWork {
		t.Fatalf("dimension assessments = %+v", repo.persisted.DimensionAssessments)
	}
	if got := strings.Join(repo.persisted.RetryFocusCompetencyCodes, ","); got != "technical_depth" {
		t.Fatalf("retry focus = %q", got)
	}
}

type conversationPromptResolver struct{ resolution registry.PromptResolution }

func (f conversationPromptResolver) ResolveActive(context.Context, string, string) (registry.PromptResolution, error) {
	return f.resolution, nil
}

type conversationReportAI struct {
	response string
	payloads []aiclient.CompletePayload
}

func (f *conversationReportAI) Complete(_ context.Context, _ string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	f.payloads = append(f.payloads, payload)
	return aiclient.CompleteResponse{Content: f.response}, aiclient.AICallMeta{}, nil
}

type conversationReportRepository struct {
	ctx       ReportContext
	persisted ReportResultPersistence
}

func (f *conversationReportRepository) LoadReportContext(context.Context, string) (ReportContext, error) {
	return f.ctx, nil
}
func (f *conversationReportRepository) PersistReportResult(_ context.Context, in ReportResultPersistence) error {
	f.persisted = in
	return nil
}
func (f *conversationReportRepository) PersistReportFailure(context.Context, ReportFailurePersistence) error {
	return nil
}

func fixedConversationIDs(ids ...string) func() string {
	index := 0
	return func() string { value := ids[index]; index++; return value }
}
func joinedConversationMessages(messages []aiclient.Message) string {
	var out strings.Builder
	for _, message := range messages {
		out.WriteString(message.Content)
		out.WriteByte('\n')
	}
	return out.String()
}
func testUUID(suffix int) string { return fmt.Sprintf("01918fa0-0000-7000-8000-%012d", suffix) }
