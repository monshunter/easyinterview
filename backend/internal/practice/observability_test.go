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
	"github.com/monshunter/easyinterview/backend/internal/ai/aiclient/observability"
	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestPracticeMetricLabelBoundaryUsesA3Allowlist(t *testing.T) {
	forbidden := map[string]struct{}{
		"feature_key":    {},
		"prompt_version": {},
		"rubric_version": {},
	}
	for _, label := range append(append([]string{}, observability.StandardLabelKeys...), observability.FallbackLabelKeys...) {
		if _, ok := forbidden[label]; ok {
			t.Fatalf("practice AI metrics must not use high-cardinality provenance label %q", label)
		}
	}

	expectedStandard := []string{
		"provider",
		"model_family",
		"model_profile_name",
		"route",
		"capability",
		"language",
		"result",
	}
	if !reflect.DeepEqual(observability.StandardLabelKeys, expectedStandard) {
		t.Fatalf("standard AI metric labels drifted: got %v want %v", observability.StandardLabelKeys, expectedStandard)
	}

	expectedFallback := append(append([]string{}, expectedStandard...), "from_provider", "from_model_family", "to_provider", "to_model_family")
	if !reflect.DeepEqual(observability.FallbackLabelKeys, expectedFallback) {
		t.Fatalf("fallback AI metric labels drifted: got %v want %v", observability.FallbackLabelKeys, expectedFallback)
	}
}

func TestStartPracticeSessionObservedAIClientWritesTaskRunColumns(t *testing.T) {
	now := time.Date(2026, 5, 9, 13, 0, 0, 0, time.UTC)
	userID := "01918fa0-0000-7000-8000-000000000001"
	targetJobID := "01918fa0-0000-7000-8000-000000002000"
	store := &recordingPlanStore{
		reservation: SessionReservation{
			SessionID:          "session-1",
			PlanID:             "plan-1",
			TargetJobID:        targetJobID,
			Goal:               sharedtypes.PracticeGoalBaseline,
			Mode:               sharedtypes.PracticeModeAssisted,
			InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
			Language:           "zh-CN",
			HintsEnabled:       true,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}
	runWriter := &recordingAITaskRunWriter{}
	ai, err := observability.New(
		&fakeAIClient{
			content: firstQuestionJSON(t, "请描述一次跨团队设计系统迁移。", "behavioral.leadership.design_system"),
			meta: aiclient.AICallMeta{
				Provider:         "stub",
				ModelFamily:      "stub",
				ModelID:          "stub-chat-1",
				FallbackChain:    []string{"stub/stub-chat-1"},
				ValidationStatus: aiclient.ValidationStatusOK,
			},
			store: store,
		},
		observability.WithRegisterer(observability.NewInMemoryRegistry()),
		observability.WithLogger(observability.NewMemoryLogger()),
		observability.WithAITaskRunWriter(runWriter),
		observability.WithAuditEventWriter(discardAuditEventWriter{}),
		observability.WithProfileResolver(practiceObservedProfileResolver{
			"practice.first_question.default": {
				Name:       "practice.first_question.default",
				Capability: aiclient.CapabilityChat,
				Status:     aiclient.ProfileStatusActive,
				Default: aiclient.ProviderConfig{
					ProviderRef: "stub",
					Model:       "stub-chat-1",
				},
				Route:     "practice.session.first_question",
				TimeoutMs: 5000,
				Version:   "1.0.0",
			},
		}),
		observability.WithNow(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			PromptVersion:       "prompt.v1",
			RubricVersion:       "rubric.v1",
			ModelProfileName:    "practice.first_question.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			UserMessageTemplate: "ask the first question",
		}},
		AI:    ai,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	if _, err := service.StartPracticeSession(context.Background(), StartSessionRequest{
		UserID:             userID,
		PlanID:             "plan-1",
		HintsEnabled:       true,
		IdempotencyKeyHash: "key-hash",
		RequestFingerprint: "fingerprint",
	}); err != nil {
		t.Fatalf("StartPracticeSession returned error: %v", err)
	}

	rows := runWriter.rows
	if len(rows) != 1 {
		t.Fatalf("expected one ai_task_runs row, got %+v", rows)
	}
	row := rows[0]
	if row.FeatureKey != firstQuestionFeatureKey ||
		row.ModelProfileName != "practice.first_question.default" ||
		row.ModelFamily != "stub" ||
		len(row.FallbackChain) != 1 ||
		row.ValidationStatus != aiclient.ValidationStatusOK ||
		row.Route != "practice.session.first_question" ||
		row.FeatureFlag != "none" ||
		row.DataSourceVersion != "registry.v1" {
		t.Fatalf("ai_task_runs row lost A3/F3 metadata: %+v", row)
	}
	if row.Capability != aiclient.AITaskRunTaskQuestionGenerate ||
		row.UserID != userID ||
		row.ResourceType != aiclient.AITaskRunResourceTargetJob ||
		row.ResourceID != targetJobID {
		t.Fatalf("ai_task_runs row lost practice task context: %+v", row)
	}
}

func TestPracticeObservedAIRedactsPromptResponseFromLogsMetricsAndAudit(t *testing.T) {
	now := time.Date(2026, 5, 9, 13, 15, 0, 0, time.UTC)
	userID := "01918fa0-0000-7000-8000-000000000001"
	targetJobID := "01918fa0-0000-7000-8000-000000002000"
	store := &recordingPlanStore{
		reservation: SessionReservation{
			SessionID:          "session-1",
			PlanID:             "plan-1",
			TargetJobID:        targetJobID,
			Goal:               sharedtypes.PracticeGoalBaseline,
			Mode:               sharedtypes.PracticeModeAssisted,
			InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
			Language:           "zh-CN",
			HintsEnabled:       true,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}
	metricsRegistry := observability.NewInMemoryRegistry()
	logger := observability.NewMemoryLogger()
	runWriter := &recordingAITaskRunWriter{}
	auditWriter := &recordingAuditEventWriter{}
	ai, err := observability.New(
		&fakeAIClient{
			content: firstQuestionJSON(t, "请说明候选人在项目中如何识别风险、验证边界、记录取舍、衡量结果并制定回滚方案，并进一步说明团队背景、目标约束、关键决策、验证方法、失败信号、复盘结论和后续改进，同时包含 response body provider secret sk-test answer_text hint_text 作为隐私测试标记。", "prompt body response body"),
			meta: aiclient.AICallMeta{
				Provider:         "stub",
				ModelFamily:      "stub",
				ModelID:          "stub-chat-1",
				FallbackChain:    []string{"stub/stub-chat-1"},
				ValidationStatus: aiclient.ValidationStatusOK,
			},
			store: store,
		},
		observability.WithRegisterer(metricsRegistry),
		observability.WithLogger(logger),
		observability.WithAITaskRunWriter(runWriter),
		observability.WithAuditEventWriter(auditWriter),
		observability.WithProfileResolver(practiceObservedProfileResolver{
			"practice.first_question.default": {
				Name:       "practice.first_question.default",
				Capability: aiclient.CapabilityChat,
				Status:     aiclient.ProfileStatusActive,
				Default: aiclient.ProviderConfig{
					ProviderRef: "stub",
					Model:       "stub-chat-1",
				},
				Route:     "practice.session.first_question",
				TimeoutMs: 5000,
				Version:   "1.0.0",
			},
		}),
		observability.WithNow(func() time.Time { return now }),
	)
	if err != nil {
		t.Fatalf("observability.New: %v", err)
	}
	service := NewService(ServiceOptions{
		Store: store,
		Registry: &fakePromptResolver{resolution: registry.PromptResolution{
			PromptVersion:       "prompt.v1",
			RubricVersion:       "rubric.v1",
			ModelProfileName:    "practice.first_question.default",
			FeatureFlag:         "none",
			DataSourceVersion:   "registry.v1",
			UserMessageTemplate: "prompt body provider secret sk-test",
		}},
		AI:    ai,
		Now:   func() time.Time { return now },
		NewID: sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1", "audit-1"),
	})

	if _, err := service.StartPracticeSession(context.Background(), StartSessionRequest{
		UserID:             userID,
		PlanID:             "plan-1",
		HintsEnabled:       true,
		IdempotencyKeyHash: "key-hash",
		RequestFingerprint: "fingerprint",
	}); err != nil {
		t.Fatalf("StartPracticeSession returned error: %v", err)
	}

	raw := mustMarshalString(t, map[string]any{
		"logs":           logger.Entries(),
		"run_rows":       runWriter.rows,
		"audit_rows":     auditWriter.rows,
		"runs_labels":    metricsRegistry.CounterLabelValues(observability.MetricRunsTotal),
		"fallbackLabels": metricsRegistry.CounterLabelValues(observability.MetricFallbackTotal),
	})
	for _, forbidden := range []string{"prompt body", "response body", "provider secret", "sk-test", "answer_text", "hint_text"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("observability surface leaked forbidden evidence %q: %s", forbidden, raw)
		}
	}
}

type recordingAITaskRunWriter struct {
	rows []aiclient.AITaskRunRow
}

func (w *recordingAITaskRunWriter) WriteAITaskRun(_ context.Context, row aiclient.AITaskRunRow) error {
	w.rows = append(w.rows, row)
	return nil
}

type discardAuditEventWriter struct{}

func (discardAuditEventWriter) WriteAuditEvent(context.Context, aiclient.AuditEventRow) error {
	return nil
}

type recordingAuditEventWriter struct {
	rows []aiclient.AuditEventRow
}

func (w *recordingAuditEventWriter) WriteAuditEvent(_ context.Context, row aiclient.AuditEventRow) error {
	w.rows = append(w.rows, row)
	return nil
}

type practiceObservedProfileResolver map[string]*aiclient.ModelProfile

func (r practiceObservedProfileResolver) Resolve(name string) (*aiclient.ModelProfile, error) {
	profile, ok := r[name]
	if !ok {
		return nil, errors.New("missing profile: " + name)
	}
	return profile, nil
}

func mustMarshalString(t *testing.T, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return string(raw)
}
