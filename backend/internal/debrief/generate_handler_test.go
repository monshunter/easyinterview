package debrief

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	"github.com/monshunter/easyinterview/backend/internal/shared/featurekeys"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestGenerateHandler_HappyResolution(t *testing.T) {
	now := time.Date(2026, 5, 16, 14, 0, 0, 0, time.UTC)
	store := &recordingGenerateStore{context: validGenerateContext()}
	reg := &recordingPromptResolver{resolution: validGenerateResolution()}
	ai := &recordingAIClient{
		response: aiclient.CompleteResponse{Content: `{"questions":[{"questionText":"Tell me about the migration.","myAnswerSummary":"I led the rollout.","interviewerReaction":"Asked for metrics.","aiAnalysis":"Add concrete adoption metrics next time."}],"riskItems":[{"label":"Metrics missing","severity":"medium"}]}`},
		meta:     validGenerateMeta(),
	}
	taskRuns := &recordingTaskRunWriter{}
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:      store,
		Registry:   reg,
		AI:         ai,
		AITaskRuns: taskRuns,
		Now:        func() time.Time { return now },
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000f001", "01918fa0-0000-7000-8000-00000000f002"),
	})

	outcome := handler.Handle(context.Background(), validClaimedDebriefJob())

	if !outcome.Succeeded || outcome.ErrorCode != "" {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if store.updateCalls != 1 || store.update.DebriefID != "01918fa0-0000-7000-8000-00000000d010" {
		t.Fatalf("update drifted: calls=%d update=%+v", store.updateCalls, store.update)
	}
	if len(store.update.Questions) != 1 || store.update.Questions[0].AIAnalysis != "Add concrete adoption metrics next time." {
		t.Fatalf("questions not enriched: %+v", store.update.Questions)
	}
	if len(store.update.RiskItems) != 1 || store.update.RiskItems[0].Label != "Metrics missing" {
		t.Fatalf("risk items not persisted: %+v", store.update.RiskItems)
	}
	if store.update.Provenance.PromptVersion != "v0.1.0" || store.update.Provenance.ModelID != "stub-model" {
		t.Fatalf("provenance drifted: %+v", store.update.Provenance)
	}
	if len(taskRuns.rows) != 1 || taskRuns.rows[0].Capability != aiclient.AITaskRunTaskDebriefGenerate || taskRuns.rows[0].Status != aiclient.AITaskRunStatusSuccess {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestGenerateHandler_PromptContextAssembled(t *testing.T) {
	store := &recordingGenerateStore{context: validGenerateContext()}
	ai := &recordingAIClient{
		response: aiclient.CompleteResponse{Content: `{"questions":[{"questionText":"Tell me about the migration.","myAnswerSummary":"I led the rollout.","aiAnalysis":"OK"}],"riskItems":[]}`},
		meta:     validGenerateMeta(),
	}
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:    store,
		Registry: &recordingPromptResolver{resolution: validGenerateResolution()},
		AI:       ai,
		NewID:    sequenceID("01918fa0-0000-7000-8000-00000000f101", "01918fa0-0000-7000-8000-00000000f102"),
	})

	outcome := handler.Handle(context.Background(), validClaimedDebriefJob())

	if !outcome.Succeeded {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	userPrompt := ""
	for _, message := range ai.payload.Messages {
		if message.Role == "user" {
			userPrompt += message.Content
		}
	}
	for _, want := range []string{"Staff Frontend Engineer", "Tell me about the migration.", "I led the rollout.", "zh-CN"} {
		if !strings.Contains(userPrompt, want) {
			t.Fatalf("prompt missing %q: %s", want, userPrompt)
		}
	}
}

func TestGenerateHandler_F3ResolveFailed(t *testing.T) {
	store := &recordingGenerateStore{context: validGenerateContext()}
	taskRuns := &recordingTaskRunWriter{}
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:      store,
		Registry:   &recordingPromptResolver{err: errors.New("registry unavailable")},
		AI:         &recordingAIClient{},
		AITaskRuns: taskRuns,
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000f201"),
	})

	outcome := handler.Handle(context.Background(), validClaimedDebriefJob())

	if outcome.Succeeded || !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if store.updateCalls != 0 {
		t.Fatalf("F3 failure must not complete debrief, updateCalls=%d", store.updateCalls)
	}
	if len(taskRuns.rows) != 1 ||
		taskRuns.rows[0].Status != aiclient.AITaskRunStatusFailed ||
		taskRuns.rows[0].FeatureKey != featurekeys.DebriefGenerate.String() ||
		taskRuns.rows[0].ErrorCode != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestGenerateHandler_A3Timeout(t *testing.T) {
	store := &recordingGenerateStore{context: validGenerateContext()}
	taskRuns := &recordingTaskRunWriter{}
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:      store,
		Registry:   &recordingPromptResolver{resolution: validGenerateResolution()},
		AI:         &recordingAIClient{meta: validGenerateMeta(), err: context.DeadlineExceeded},
		AITaskRuns: taskRuns,
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000f301"),
	})

	outcome := handler.Handle(context.Background(), validClaimedDebriefJob())

	if outcome.Succeeded || !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if store.updateCalls != 0 {
		t.Fatalf("A3 timeout must not complete debrief, updateCalls=%d", store.updateCalls)
	}
	if len(taskRuns.rows) != 1 ||
		taskRuns.rows[0].Status != aiclient.AITaskRunStatusTimeout ||
		taskRuns.rows[0].ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestGenerateHandler_ParseEmpty(t *testing.T) {
	store := &recordingGenerateStore{context: validGenerateContext()}
	taskRuns := &recordingTaskRunWriter{}
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:      store,
		Registry:   &recordingPromptResolver{resolution: validGenerateResolution()},
		AI:         &recordingAIClient{response: aiclient.CompleteResponse{Content: `{"questions":[],"riskItems":[]}`}, meta: validGenerateMeta()},
		AITaskRuns: taskRuns,
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000f401"),
	})

	outcome := handler.Handle(context.Background(), validClaimedDebriefJob())

	if outcome.Succeeded || !outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if store.updateCalls != 0 {
		t.Fatalf("parse failure must not complete debrief, updateCalls=%d", store.updateCalls)
	}
	if len(taskRuns.rows) != 1 ||
		taskRuns.rows[0].Status != aiclient.AITaskRunStatusFailed ||
		taskRuns.rows[0].ValidationStatus != aiclient.ValidationStatusInvalid ||
		taskRuns.rows[0].ErrorCode != sharederrors.CodeAiOutputInvalid {
		t.Fatalf("task run row drifted: %+v", taskRuns.rows)
	}
}

func TestGenerateHandler_PermanentFailAt5Attempts(t *testing.T) {
	store := &recordingGenerateStore{context: validGenerateContext()}
	job := validClaimedDebriefJob()
	job.Attempts = 5
	job.MaxAttempts = 5
	handler := NewGenerateHandler(GenerateHandlerOptions{
		Store:      store,
		Registry:   &recordingPromptResolver{resolution: validGenerateResolution()},
		AI:         &recordingAIClient{meta: validGenerateMeta(), err: context.DeadlineExceeded},
		AITaskRuns: &recordingTaskRunWriter{},
		NewID:      sequenceID("01918fa0-0000-7000-8000-00000000f501"),
	})

	outcome := handler.Handle(context.Background(), job)

	if outcome.Succeeded || outcome.Retryable || outcome.ErrorCode != sharederrors.CodeAiProviderTimeout {
		t.Fatalf("attempts=max must return permanent failure outcome, got %+v", outcome)
	}
	if store.updateCalls != 0 {
		t.Fatalf("permanent failure must keep debrief draft, updateCalls=%d", store.updateCalls)
	}
}

func validClaimedDebriefJob() targetjob.ClaimedJob {
	return targetjob.ClaimedJob{
		JobID:        "01918fa0-0000-7000-8000-00000000d011",
		JobType:      "debrief_generate",
		ResourceType: "debrief",
		ResourceID:   "01918fa0-0000-7000-8000-00000000d010",
		Payload:      []byte(`{"debriefId":"01918fa0-0000-7000-8000-00000000d010","targetJobId":"01918fa0-0000-7000-8000-00000000c001","language":"zh-CN","questionCount":1}`),
		Attempts:     0,
		MaxAttempts:  5,
	}
}

func validGenerateContext() GenerateContext {
	return GenerateContext{
		UserID:        "01918fa0-0000-7000-8000-000000000001",
		DebriefID:     "01918fa0-0000-7000-8000-00000000d010",
		TargetJobID:   "01918fa0-0000-7000-8000-00000000c001",
		Language:      "zh-CN",
		TargetTitle:   "Staff Frontend Engineer",
		TargetSummary: "Design systems and cross-functional leadership.",
		Questions: []QuestionInput{{
			QuestionText:        "Tell me about the migration.",
			MyAnswerSummary:     "I led the rollout.",
			InterviewerReaction: "Asked for metrics.",
		}},
	}
}

func validGenerateResolution() registry.PromptResolution {
	return registry.PromptResolution{
		FeatureKey:          featurekeys.DebriefGenerate.String(),
		PromptVersion:       "v0.1.0",
		RubricVersion:       "v0.1.0",
		ModelProfileName:    "debrief.generate.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "debrief/01918fa0-0000-7000-8000-00000000d010@v1",
		SystemMessage:       "Analyze a real interview debrief.",
		UserMessageTemplate: "Role {{targetTitle}}\nQuestions {{questions}}\nLanguage {{language}}",
	}
}

func validGenerateMeta() aiclient.AICallMeta {
	return aiclient.AICallMeta{
		Provider:          "stub",
		ModelFamily:       "stub-family",
		ModelID:           "stub-model",
		PromptVersion:     "v0.1.0",
		RubricVersion:     "v0.1.0",
		ModelProfileName:  "debrief.generate.default",
		FeatureKey:        featurekeys.DebriefGenerate.String(),
		FeatureFlag:       "none",
		DataSourceVersion: "debrief/01918fa0-0000-7000-8000-00000000d010@v1",
		Language:          "zh-CN",
		ValidationStatus:  aiclient.ValidationStatusOK,
	}
}

type recordingGenerateStore struct {
	context     GenerateContext
	update      UpdateDebriefCompletedInput
	updateCalls int
}

func (s *recordingGenerateStore) LoadGenerateContext(context.Context, GenerateJobPayload) (GenerateContext, error) {
	return s.context, nil
}

func (s *recordingGenerateStore) UpdateDebriefCompleted(_ context.Context, in UpdateDebriefCompletedInput) (DebriefRecord, error) {
	s.updateCalls++
	s.update = in
	return DebriefRecord{ID: in.DebriefID}, nil
}
