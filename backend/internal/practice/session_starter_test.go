package practice

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestStartPracticeSessionRunsThreeStepFlowWithAIOutsideTransactions(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	store := &recordingPlanStore{
		reservation: SessionReservation{
			SessionID:          "session-1",
			PlanID:             "plan-1",
			TargetJobID:        "target-1",
			Goal:               sharedtypes.PracticeGoalBaseline,
			Mode:               sharedtypes.PracticeModeAssisted,
			InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
			Language:           "zh-CN",
			HintsEnabled:       true,
			CreatedAt:          now.Add(-time.Hour),
			UpdatedAt:          now.Add(-time.Hour),
		},
	}
	registryClient := &fakePromptResolver{resolution: registry.PromptResolution{
		FeatureKey:          "practice.session.first_question",
		PromptVersion:       "prompt.v1",
		RubricVersion:       "rubric.v1",
		ModelProfileName:    "practice.first_question.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		UserMessageTemplate: "ask the first question",
	}}
	ai := &fakeAIClient{content: `{"question":"请用 STAR 描述你主导设计系统迁移的项目，重点说明跨 12 个团队的协调过程。","intent":"behavioral.leadership.design_system","focus_dimension":"leadership","expected_signals":["scope","tradeoffs"],"time_budget_seconds":180}`, store: store}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: registryClient,
		AI:       ai,
		Now:      func() time.Time { return now },
		NewID:    sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	session, err := service.StartPracticeSession(context.Background(), StartSessionRequest{
		UserID:             "user-1",
		PlanID:             "plan-1",
		HintsEnabled:       true,
		IdempotencyKeyHash: "key-hash",
		RequestFingerprint: "fingerprint",
	})
	if err != nil {
		t.Fatalf("StartPracticeSession returned error: %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve", "ai", "commit"}) {
		t.Fatalf("three-step order = %#v", store.steps)
	}
	if !ai.calledOutsideTransaction {
		t.Fatalf("AI call must happen outside the repository transaction window")
	}
	if session.Status != sharedtypes.SessionStatusRunning || session.CurrentTurn == nil {
		t.Fatalf("unexpected session: %+v", session)
	}
	if session.CurrentTurn.QuestionText != "请用 STAR 描述你主导设计系统迁移的项目，重点说明跨 12 个团队的协调过程。" ||
		session.CurrentTurn.QuestionIntent != "behavioral.leadership.design_system" ||
		session.CurrentTurn.TurnIndex != 1 {
		t.Fatalf("unexpected first turn: %+v", session.CurrentTurn)
	}
	if store.commit.IdempotencyRecordID != "idem-1" ||
		store.commit.UserID != "user-1" ||
		store.commit.TurnID != "turn-1" ||
		store.commit.SessionEventID != "event-1" ||
		store.commit.OutboxEventID != "outbox-1" ||
		store.commit.AuditEventID != "audit-1" {
		t.Fatalf("commit ids not generated: %+v", store.commit)
	}
	if ai.profileName != "practice.first_question.default" {
		t.Fatalf("AI profile = %q", ai.profileName)
	}
	meta := ai.payload.Metadata
	if meta.FeatureKey != "practice.session.first_question" ||
		meta.PromptVersion != "prompt.v1" ||
		meta.RubricVersion != "rubric.v1" ||
		meta.Language != "zh-CN" ||
		meta.FeatureFlag != "none" ||
		meta.DataSourceVersion != "registry.v1" {
		t.Fatalf("AI metadata incomplete: %+v", meta)
	}
	if meta.TaskRun.UserID != "user-1" ||
		meta.TaskRun.Capability != aiclient.AITaskRunTaskQuestionGenerate ||
		meta.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		meta.TaskRun.ResourceID != "target-1" {
		t.Fatalf("AI task run context incomplete: %+v", meta.TaskRun)
	}
}

func TestStartPracticeSessionRejectsMissingFirstQuestionText(t *testing.T) {
	store := &recordingPlanStore{
		reservation: SessionReservation{
			SessionID:          "session-1",
			PlanID:             "plan-1",
			TargetJobID:        "target-1",
			Goal:               sharedtypes.PracticeGoalBaseline,
			Mode:               sharedtypes.PracticeModeAssisted,
			InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
			Language:           "zh-CN",
		},
	}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{FeatureKey: "practice.session.first_question", PromptVersion: "p", RubricVersion: "r", ModelProfileName: "practice.first_question.default", FeatureFlag: "none", DataSourceVersion: "registry.v1"}},
		AI:       &fakeAIClient{content: `{"questionIntent":"missing.text"}`, store: store},
		NewID:    sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	if _, err := service.StartPracticeSession(context.Background(), StartSessionRequest{UserID: "user-1", PlanID: "plan-1", IdempotencyKeyHash: "key-hash", RequestFingerprint: "fingerprint"}); err == nil {
		t.Fatalf("expected invalid first question error")
	}
	if len(store.steps) != 3 || store.steps[0] != "reserve" || store.steps[1] != "ai" || store.steps[2] != "fail" {
		t.Fatalf("invalid first question should persist failed reservation without commit, steps=%v", store.steps)
	}
	if store.fail.ErrorCode != sharederrors.CodeAiOutputInvalid || store.fail.Retryable {
		t.Fatalf("invalid first question failure not recorded correctly: %+v", store.fail)
	}
}

func TestStartPracticeSessionRejectsNonJSONFirstQuestionResponse(t *testing.T) {
	store := &recordingPlanStore{
		reservation: SessionReservation{
			SessionID:          "session-1",
			PlanID:             "plan-1",
			TargetJobID:        "target-1",
			Goal:               sharedtypes.PracticeGoalBaseline,
			Mode:               sharedtypes.PracticeModeAssisted,
			InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
			Language:           "zh-CN",
		},
	}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{FeatureKey: "practice.session.first_question", PromptVersion: "p", RubricVersion: "r", ModelProfileName: "practice.first_question.default", FeatureFlag: "none", DataSourceVersion: "registry.v1"}},
		AI:       &fakeAIClient{content: `Here is a first question without strict JSON.`, store: store},
		NewID:    sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	if _, err := service.StartPracticeSession(context.Background(), StartSessionRequest{UserID: "user-1", PlanID: "plan-1", IdempotencyKeyHash: "key-hash", RequestFingerprint: "fingerprint"}); err == nil {
		t.Fatalf("expected non-JSON first question to be rejected")
	}
	if len(store.steps) != 3 || store.steps[0] != "reserve" || store.steps[1] != "ai" || store.steps[2] != "fail" {
		t.Fatalf("non-JSON first question should persist failed reservation without commit, steps=%v", store.steps)
	}
	if store.fail.ErrorCode != sharederrors.CodeAiOutputInvalid || store.fail.Retryable {
		t.Fatalf("non-JSON first question failure not recorded correctly: %+v", store.fail)
	}
}

type fakePromptResolver struct {
	resolution registry.PromptResolution
	err        error
}

func (r *fakePromptResolver) ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	if r.err != nil {
		return registry.PromptResolution{}, r.err
	}
	r.resolution.FeatureKey = featureKey
	return r.resolution, nil
}

type fakeAIClient struct {
	content                  string
	err                      error
	meta                     aiclient.AICallMeta
	profileName              string
	payload                  aiclient.CompletePayload
	calledOutsideTransaction bool
	store                    *recordingPlanStore
}

func (c *fakeAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.profileName = profileName
	c.payload = payload
	c.calledOutsideTransaction = c.store == nil || !c.store.inTx
	if c.store != nil {
		c.store.steps = append(c.store.steps, "ai")
	}
	if c.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, c.err
	}
	return aiclient.CompleteResponse{Content: c.content}, c.meta, nil
}

func (c *fakeAIClient) Transcribe(ctx context.Context, input string, payload aiclient.TranscriptionInput) (aiclient.TranscriptionResponse, aiclient.AICallMeta, error) {
	return aiclient.TranscriptionResponse{}, aiclient.AICallMeta{}, nil
}

func (c *fakeAIClient) Stream(ctx context.Context, profileName string, payload aiclient.CompletePayload) (<-chan aiclient.AIStreamEvent, error) {
	return nil, nil
}

func (c *fakeAIClient) Synthesize(ctx context.Context, profileName string, input aiclient.SynthesisInput) (aiclient.SynthesisResponse, aiclient.AICallMeta, error) {
	return aiclient.SynthesisResponse{}, aiclient.AICallMeta{}, nil
}

func firstQuestionJSON(t *testing.T, text, intent string) string {
	t.Helper()
	raw, err := json.Marshal(map[string]string{"questionText": text, "questionIntent": intent})
	if err != nil {
		t.Fatalf("marshal question: %v", err)
	}
	return string(raw)
}
