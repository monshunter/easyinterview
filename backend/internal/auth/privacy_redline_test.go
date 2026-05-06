package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestAuthPrivacyObservableSurfacesDoNotLeakSecretsOrPII(t *testing.T) {
	dispatcher := &payloadRecordingDispatcher{}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"})
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:           &recordingChallengeStore{},
		Dispatcher:      dispatcher,
		DeliverySecrets: sink,
		TokenGenerator:  fixedTokenGenerator("raw-magic-token"),
		ChallengePepper: "pepper-secret",
		Now:             func() time.Time { return time.Date(2026, 5, 6, 11, 10, 0, 0, time.UTC) },
		NewID:           fixedIDs("challenge-privacy"),
	})
	if _, err := service.StartEmailChallenge(context.Background(), auth.StartEmailChallengeInput{
		Email:      "candidate@example.com",
		RemoteAddr: "203.0.113.40:5588",
		UserAgent:  "unit-test-agent",
	}); err != nil {
		t.Fatalf("StartEmailChallenge: %v", err)
	}
	assertNoAuthLeak(t, "email dispatch payload", fmt.Sprintf("%+v", dispatcher.payload))
	assertNoAuthLeak(t, "dev sink metadata", fmt.Sprintf("%+v", sink))

	verifyStore := &verifyStore{
		challenge: auth.ChallengeRecord{ID: "challenge-privacy", Email: "candidate@example.com", ExpiresAt: time.Now().Add(auth.ChallengeTTL)},
		user:      auth.UserContext{ID: "user-privacy", Email: "candidate@example.com"},
	}
	verifyService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:                 verifyStore,
		SessionTokenGenerator: fixedTokenGenerator("raw-session-cookie"),
		ChallengePepper:       "pepper-secret",
		SessionCookieSecret:   "session-secret",
		Now:                   func() time.Time { return time.Date(2026, 5, 6, 11, 10, 0, 0, time.UTC) },
		NewID:                 fixedIDs("user-privacy", "session-privacy"),
	})
	handler := auth.NewHandler(auth.HandlerOptions{Passwordless: verifyService})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/email/verify?token=raw-magic-token", nil)
	rec := httptest.NewRecorder()
	handler.VerifyAuthEmailChallenge(rec, req)
	assertNoAuthLeak(t, "verify response body", rec.Body.String())

	failing := auth.NewBackgroundMailDispatcher(auth.BackgroundMailDispatcherOptions{
		Writer: &failingDeliveryWriter{err: fmt.Errorf("failed raw-magic-token candidate@example.com http://api.test/verify?token=raw-magic-token")},
	})
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   "challenge-privacy",
		"templateKey":       "auth_magic_link",
		"locale":            "en",
		"deliverySecretRef": "auth_challenge:challenge-privacy",
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := failing.Enqueue(context.Background(), payload); err != nil {
		t.Fatal(err)
	}
	if err := failing.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
	assertNoAuthLeak(t, "dispatcher failure summaries", strings.Join(failing.ErrorSummaries(), "\n"))
}

func assertNoAuthLeak(t *testing.T, surface string, text string) {
	t.Helper()
	for _, forbidden := range []string{
		"raw-magic-token",
		"raw-session-cookie",
		"candidate@example.com",
		"pepper-secret",
		"session-secret",
		"http://api.test/api/v1/auth/email/verify",
	} {
		if contains(text, forbidden) {
			t.Fatalf("%s leaked %s: %s", surface, forbidden, text)
		}
	}
}
