package practice

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/monshunter/easyinterview/backend/internal/ai/registry"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestStartPracticeSessionMapsAIErrorsToPracticeServiceErrors(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCode   string
		wantPrompt string
	}{
		{
			name:       "timeout",
			err:        sharederrors.Wrap(sharederrors.CodeAiProviderTimeout, "provider timed out with prompt body and response body", true),
			wantCode:   sharederrors.CodeAiProviderTimeout,
			wantPrompt: "prompt body",
		},
		{
			name:       "invalid-output",
			err:        sharederrors.Wrap(sharederrors.CodeAiOutputInvalid, "raw response body had invalid schema", false),
			wantCode:   sharederrors.CodeAiOutputInvalid,
			wantPrompt: "response body",
		},
		{
			name:       "secret-missing",
			err:        errors.New("AI_PROVIDER_SECRET_MISSING provider secret missing: sk-test"),
			wantCode:   sharederrors.CodeAiProviderSecretMissing,
			wantPrompt: "sk-test",
		},
		{
			name:       "fallback-exhausted",
			err:        errors.New("AI_FALLBACK_EXHAUSTED prompt body exhausted every route"),
			wantCode:   sharederrors.CodeAiFallbackExhausted,
			wantPrompt: "prompt body",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &recordingPlanStore{reservation: validSessionReservation()}
			service := NewService(ServiceOptions{
				Store:    store,
				Registry: &fakePromptResolver{resolution: validPromptResolution()},
				AI:       &fakeAIClient{err: tc.err, store: store},
				NewID:    sequenceIDs("idem-1", "session-1", "turn-1", "event-1", "outbox-1"),
			})

			_, err := service.StartPracticeSession(context.Background(), StartSessionRequest{
				UserID:             "user-1",
				PlanID:             "plan-1",
				IdempotencyKeyHash: "key-hash",
				RequestFingerprint: "fingerprint",
			})
			var svcErr *ServiceError
			if !errors.As(err, &svcErr) {
				t.Fatalf("expected ServiceError, got %T: %v", err, err)
			}
			if svcErr.Code != tc.wantCode {
				t.Fatalf("code = %q, want %q", svcErr.Code, tc.wantCode)
			}
			if strings.Contains(svcErr.Message, tc.wantPrompt) || strings.Contains(svcErr.Error(), tc.wantPrompt) {
				t.Fatalf("service error leaked forbidden AI evidence %q: %+v", tc.wantPrompt, svcErr)
			}
			if len(store.steps) != 3 || store.steps[0] != "reserve" || store.steps[1] != "ai" || store.steps[2] != "fail" {
				t.Fatalf("AI error should reserve, call AI, then persist failed reservation; steps=%v", store.steps)
			}
			if store.fail.ErrorCode != tc.wantCode || store.fail.SessionID != "session-1" || store.fail.UserID != "user-1" {
				t.Fatalf("reservation failure was not persisted correctly: %+v", store.fail)
			}
		})
	}
}

func validSessionReservation() SessionReservation {
	return SessionReservation{
		SessionID:          "session-1",
		PlanID:             "plan-1",
		TargetJobID:        "target-1",
		Goal:               sharedtypes.PracticeGoalBaseline,
		Mode:               sharedtypes.PracticeModeAssisted,
		InterviewerPersona: sharedtypes.InterviewerRoleHiringManager,
		Language:           "zh-CN",
	}
}

func validPromptResolution() registry.PromptResolution {
	return registry.PromptResolution{
		FeatureKey:          "practice.session.first_question",
		PromptVersion:       "prompt.v1",
		RubricVersion:       "rubric.v1",
		ModelProfileName:    "practice.first_question.default",
		FeatureFlag:         "none",
		DataSourceVersion:   "registry.v1",
		UserMessageTemplate: "ask the first question",
	}
}
