package practice

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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
			RoleTitle:          "Staff Frontend Architect",
			Seniority:          "staff",
			TopSkills:          []string{"React", "design systems", "cross-team migration"},
			RubricDimensions:   []string{"practice_depth", "language_consistency"},
			ResumeProfile:      "Candidate led GraphQL platform migration and owned frontend reliability for a high-traffic SaaS product.",
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
		OutputSchema:        practiceOutputSchema(`{"type":"object","required":["questionText","questionIntent"],"properties":{"questionText":{"type":"string"},"questionIntent":{"type":"string"}}}`),
		UserMessageTemplate: "Respond in {{language}}. Role: {{role_title}} ({{seniority}}). Top required skills: {{top_skills}}. Resume profile: {{resume_profile}}. Rubric dimensions: {{rubric_dimensions}}. Practice goal: {{practice_goal}}.",
	}}
	ai := &fakeAIClient{content: `{"questionText":"请用 STAR 描述你主导设计系统迁移的项目，重点说明跨 12 个团队的协调过程。","questionIntent":"behavioral.leadership.design_system","focusDimension":"leadership","expectedSignals":["scope","tradeoffs"],"timeBudgetSeconds":180}`, store: store}
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
	if len(meta.OutputSchema) == 0 {
		t.Fatalf("AI metadata OutputSchema must be populated")
	}
	if meta.TaskRun.UserID != "user-1" ||
		meta.TaskRun.Capability != aiclient.AITaskRunTaskQuestionGenerate ||
		meta.TaskRun.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		meta.TaskRun.ResourceID != "target-1" {
		t.Fatalf("AI task run context incomplete: %+v", meta.TaskRun)
	}
	userPrompt := ai.payload.Messages[len(ai.payload.Messages)-1].Content
	for _, forbidden := range []string{"{{language}}", "{{role_title}}", "{{top_skills}}", "{{resume_profile}}", "{{practice_goal}}", "{{rubric_dimensions}}"} {
		if strings.Contains(userPrompt, forbidden) {
			t.Fatalf("first-question prompt still contains raw template marker %q: %s", forbidden, userPrompt)
		}
	}
	for _, required := range []string{"zh-CN", "Staff Frontend Architect", "React, design systems, cross-team migration", "Candidate led GraphQL platform migration", "practice_depth, language_consistency", "baseline"} {
		if !strings.Contains(userPrompt, required) {
			t.Fatalf("first-question prompt missing %q: %s", required, userPrompt)
		}
	}
}

func TestStartPracticeSessionRepairsWrongLanguageExactlyOnce(t *testing.T) {
	now := time.Date(2026, 4, 28, 13, 45, 12, 0, time.UTC)
	store := &recordingPlanStore{reservation: SessionReservation{
		IdempotencyRecordID: "idem-1",
		SessionID:           "session-1",
		PlanID:              "plan-1",
		TargetJobID:         "target-1",
		Goal:                sharedtypes.PracticeGoalBaseline,
		Mode:                sharedtypes.PracticeModeAssisted,
		InterviewerPersona:  sharedtypes.InterviewerRoleHiringManager,
		Language:            "zh-CN",
		RoleTitle:           "后端工程师",
		CreatedAt:           now,
	}}
	ai := &fakeAIClient{contents: []string{
		firstQuestionJSON(t, "Tell me about a production incident.", "incident"),
		firstQuestionJSON(t, "请介绍一次你处理线上故障的经历。", "incident"),
	}, store: store}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			FeatureKey:          firstQuestionFeatureKey,
			PromptVersion:       "prompt.v1",
			RubricVersion:       "rubric.v1",
			ModelProfileName:    "practice.first_question.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			OutputSchema:        practiceOutputSchema(`{"type":"object","required":["questionText","questionIntent"],"properties":{"questionText":{"type":"string"},"questionIntent":{"type":"string"}}}`),
			UserMessageTemplate: "Respond in {{language}} for {{role_title}}.",
		}},
		AI:    ai,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	session, err := service.StartPracticeSession(context.Background(), StartSessionRequest{
		UserID:             "user-1",
		PlanID:             "plan-1",
		IdempotencyKeyHash: "key-hash",
		RequestFingerprint: "fingerprint",
	})
	if err != nil {
		t.Fatalf("StartPracticeSession: %v", err)
	}
	if session.CurrentTurn == nil || session.CurrentTurn.QuestionText != "请介绍一次你处理线上故障的经历。" {
		t.Fatalf("first question did not use repaired output: %+v", session.CurrentTurn)
	}
	if len(ai.payloads) != 2 || !reflect.DeepEqual(store.steps, []string{"reserve", "ai", "ai", "commit"}) {
		t.Fatalf("first-question repair count/order drift: calls=%d steps=%v", len(ai.payloads), store.steps)
	}
}

func TestStartPracticeSessionFailsReservationWhenPromptResolutionFails(t *testing.T) {
	store := &recordingPlanStore{
		reservation: SessionReservation{
			IdempotencyRecordID: "idem-1",
			SessionID:           "session-1",
			PlanID:              "plan-1",
			TargetJobID:         "target-1",
			Goal:                sharedtypes.PracticeGoalBaseline,
			Mode:                sharedtypes.PracticeModeAssisted,
			InterviewerPersona:  sharedtypes.InterviewerRoleHiringManager,
			Language:            "zh-CN",
		},
	}
	ai := &fakeAIClient{content: firstQuestionJSON(t, "Question?", "behavioral"), store: store}
	service := NewService(ServiceOptions{
		Store:    store,
		Registry: &fakePromptResolver{err: registry.ErrPromptUnsupported},
		AI:       ai,
		NewID:    sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	_, err := service.StartPracticeSession(context.Background(), StartSessionRequest{
		UserID:             "user-1",
		PlanID:             "plan-1",
		IdempotencyKeyHash: "key-hash",
		RequestFingerprint: "fingerprint",
	})
	if err == nil {
		t.Fatalf("expected prompt resolution failure")
	}
	var svcErr *ServiceError
	if !errors.As(err, &svcErr) || svcErr.Code != sharederrors.CodeAiProviderConfigInvalid {
		t.Fatalf("expected AI_PROVIDER_CONFIG_INVALID service error, got %v", err)
	}
	if !reflect.DeepEqual(store.steps, []string{"reserve", "fail"}) {
		t.Fatalf("prompt resolution failure must fail reserved session without AI/commit, steps=%v", store.steps)
	}
	if store.fail.ErrorCode != sharederrors.CodeAiProviderConfigInvalid || store.fail.Retryable {
		t.Fatalf("prompt resolution failure not recorded as terminal config failure: %+v", store.fail)
	}
	if ai.profileName != "" {
		t.Fatalf("AI client must not be called after prompt resolution failure")
	}
}

func TestParseFirstQuestionUsesCanonicalOutputKeys(t *testing.T) {
	t.Run("rejects alias-only output", func(t *testing.T) {
		_, err := parseFirstQuestion(`{"question":"Alias question?","intent":"alias.intent"}`)
		if err == nil {
			t.Fatal("expected alias-only first question output to fail")
		}
		var apiErr *sharederrors.APIError
		if !errors.As(err, &apiErr) || apiErr.Code != sharederrors.CodeAiOutputInvalid {
			t.Fatalf("expected AI_OUTPUT_INVALID, got %T %v", err, err)
		}
	})

	t.Run("canonical keys win over unknown aliases", func(t *testing.T) {
		got, err := parseFirstQuestion(`{"questionText":"Canonical question?","questionIntent":"canonical.intent","question":"Alias question?","intent":"alias.intent"}`)
		if err != nil {
			t.Fatalf("parseFirstQuestion returned error: %v", err)
		}
		if got.Text != "Canonical question?" || got.Intent != "canonical.intent" {
			t.Fatalf("unexpected first question: %+v", got)
		}
	})

	t.Run("rejects missing canonical intent", func(t *testing.T) {
		_, err := parseFirstQuestion(`{"questionText":"Canonical question?"}`)
		if err == nil {
			t.Fatal("expected missing questionIntent to fail")
		}
		if code, ok := aiErrorCode(err); !ok || code != sharederrors.CodeAiOutputInvalid {
			t.Fatalf("expected AI_OUTPUT_INVALID, got %T %v", err, err)
		}
	})
}

func TestStartPracticeSessionRejectsInvalidFirstQuestionOutput(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "missing question text", content: `{"questionIntent":"missing.text"}`},
		{name: "non-JSON response", content: `Here is a first question without strict JSON.`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
				AI:       &fakeAIClient{content: tc.content, store: store},
				NewID:    sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
			})

			if _, err := service.StartPracticeSession(context.Background(), StartSessionRequest{UserID: "user-1", PlanID: "plan-1", IdempotencyKeyHash: "key-hash", RequestFingerprint: "fingerprint"}); err == nil {
				t.Fatal("expected invalid first question output to be rejected")
			}
			if !reflect.DeepEqual(store.steps, []string{"reserve", "ai", "ai", "fail"}) {
				t.Fatalf("invalid first question should fail the reservation without commit, steps=%v", store.steps)
			}
			if store.fail.ErrorCode != sharederrors.CodeAiOutputInvalid || store.fail.Retryable {
				t.Fatalf("invalid first question failure not recorded correctly: %+v", store.fail)
			}
		})
	}
}

type fakePromptResolver struct {
	resolution  registry.PromptResolution
	resolutions map[string]registry.PromptResolution
	err         error
	errs        map[string]error
}

func (r *fakePromptResolver) ResolveActive(ctx context.Context, featureKey, language string) (registry.PromptResolution, error) {
	if r.errs != nil {
		if err, ok := r.errs[featureKey]; ok {
			return registry.PromptResolution{}, err
		}
	}
	if r.err != nil {
		return registry.PromptResolution{}, r.err
	}
	if r.resolutions != nil {
		if resolution, ok := r.resolutions[featureKey]; ok {
			resolution.FeatureKey = featureKey
			return resolution, nil
		}
		return registry.PromptResolution{}, registry.ErrPromptUnsupported
	}
	r.resolution.FeatureKey = featureKey
	return r.resolution, nil
}

type fakeAIClient struct {
	content                  string
	contents                 []string
	err                      error
	errs                     []error
	meta                     aiclient.AICallMeta
	profileName              string
	profileNames             []string
	payload                  aiclient.CompletePayload
	payloads                 []aiclient.CompletePayload
	calledOutsideTransaction bool
	store                    *recordingPlanStore
}

func (c *fakeAIClient) Complete(ctx context.Context, profileName string, payload aiclient.CompletePayload) (aiclient.CompleteResponse, aiclient.AICallMeta, error) {
	c.profileName = profileName
	c.profileNames = append(c.profileNames, profileName)
	c.payload = payload
	c.payloads = append(c.payloads, payload)
	c.calledOutsideTransaction = c.store == nil || !c.store.inTx
	if c.store != nil {
		c.store.steps = append(c.store.steps, "ai")
	}
	if len(c.errs) > 0 {
		err := c.errs[0]
		c.errs = c.errs[1:]
		if err != nil {
			return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, err
		}
	}
	if c.err != nil {
		return aiclient.CompleteResponse{}, aiclient.AICallMeta{}, c.err
	}
	content := c.content
	if len(c.contents) > 0 {
		content = c.contents[0]
		c.contents = c.contents[1:]
	}
	return aiclient.CompleteResponse{Content: content}, c.meta, nil
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

func practiceOutputSchema(s string) *json.RawMessage {
	raw := json.RawMessage(s)
	return &raw
}
