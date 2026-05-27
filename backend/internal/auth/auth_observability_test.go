package auth_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestAuthMetricsAreRegisteredInF1SpecAndUseAllowedLabels(t *testing.T) {
	specPath := filepath.Join("..", "..", "..", "docs", "spec", "observability-stack", "spec.md")
	spec, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read F1 spec: %v", err)
	}
	specText := string(spec)
	expected := map[string][]string{
		auth.MetricAuthChallengeStartedTotal: auth.AuthServiceResultLabelKeys,
		auth.MetricAuthSessionMintedTotal:    auth.AuthServiceResultLabelKeys,
		auth.MetricAuthLogoutTotal:           auth.AuthServiceResultLabelKeys,
		auth.MetricAuthDeleteHandoffTotal:    auth.AuthServiceResultLabelKeys,
		auth.MetricAuthFailureTotal:          auth.AuthOperationResultLabelKeys,
	}
	registry := auth.NewInMemoryAuthMetricRegistry()
	auth.RegisterAuthMetrics(registry, auth.AuthMetricsOptions{Service: "backend"})

	for name, wantLabels := range expected {
		if !contains(specText, "`"+name+"`") {
			t.Fatalf("F1 spec does not register %s", name)
		}
		if !registry.CounterRegistered(name) {
			t.Fatalf("auth metric %s was not registered", name)
		}
		gotLabels := registry.CounterLabelKeys(name)
		if fmt.Sprint(gotLabels) != fmt.Sprint(wantLabels) {
			t.Fatalf("%s labels = %v want %v", name, gotLabels, wantLabels)
		}
		for _, key := range gotLabels {
			if !auth.IsF1AllowedAuthMetricLabel(key) {
				t.Fatalf("%s uses non-F1 label %q", name, key)
			}
		}
	}
}

func TestAuthObservabilityEventsUseF1LabelsAndRedactedAuditFields(t *testing.T) {
	registry := auth.NewInMemoryAuthMetricRegistry()
	metrics := auth.RegisterAuthMetrics(registry, auth.AuthMetricsOptions{Service: "backend"})
	audit := &recordingAuthAudit{}
	ctx := auth.ContextWithAuthTraceID(context.Background(), "4bf92f3577b34da6a3ce929d0e0e4736")

	startService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:           &recordingChallengeStore{},
		Dispatcher:      &recordingDispatcher{},
		DeliverySecrets: auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/api/v1/auth/email/verify"}),
		TokenGenerator:  fixedTokenGenerator("123456"),
		ChallengePepper: "pepper-secret",
		Metrics:         metrics,
		Audit:           audit,
		Now:             func() time.Time { return time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC) },
		NewID:           fixedIDs("challenge-observe"),
	})
	if _, err := startService.StartEmailChallenge(ctx, auth.StartEmailChallengeInput{
		Email:      "candidate@example.com",
		RemoteAddr: "203.0.113.50:5588",
		UserAgent:  "unit-test-agent",
	}); err != nil {
		t.Fatalf("StartEmailChallenge: %v", err)
	}

	verifyService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &verifyStore{
			challenge: auth.ChallengeRecord{ID: "challenge-observe", Email: "candidate@example.com", ExpiresAt: time.Now().Add(auth.ChallengeTTL)},
			user:      auth.UserContext{ID: "user-observe", Email: "candidate@example.com"},
		},
		SessionTokenGenerator: fixedTokenGenerator("raw-session-cookie"),
		ChallengePepper:       "pepper-secret",
		SessionCookieSecret:   "session-secret",
		Metrics:               metrics,
		Audit:                 audit,
		Now:                   func() time.Time { return time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC) },
		NewID:                 fixedIDs("session-observe"),
	})
	if _, err := verifyService.VerifyEmailChallenge(ctx, auth.VerifyEmailChallengeInput{
		Token:      "123456",
		RemoteAddr: "203.0.113.50:5588",
		UserAgent:  "unit-test-agent",
	}); err != nil {
		t.Fatalf("VerifyEmailChallenge: %v", err)
	}

	logoutService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:           &logoutStore{},
		ChallengePepper: "pepper-secret",
		Metrics:         metrics,
		Audit:           audit,
		Now:             func() time.Time { return time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC) },
	})
	if err := logoutService.Logout(ctx, auth.CurrentSession{SessionID: "session-secret-id", UserID: "user-observe"}); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	deleteService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store: &deleteMeStore{handoff: auth.PrivacyDeleteHandoff{
			PrivacyRequestID: "privacy-request-observe",
			JobID:            "job-observe",
			CreatedAt:        time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC),
			UpdatedAt:        time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC),
		}},
		ChallengePepper: "pepper-secret",
		Metrics:         metrics,
		Audit:           audit,
		Now:             func() time.Time { return time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC) },
		NewID:           fixedIDs("privacy-request-observe", "job-observe"),
	})
	if _, err := deleteService.DeleteMe(ctx, auth.CurrentSession{SessionID: "session-secret-id", UserID: "user-observe"}, "delete-key"); err != nil {
		t.Fatalf("DeleteMe: %v", err)
	}

	failureService := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               &verifyStore{consumeErr: auth.ErrChallengeInvalid},
		ChallengePepper:     "pepper-secret",
		SessionCookieSecret: "session-secret",
		Metrics:             metrics,
		Audit:               audit,
		Now:                 func() time.Time { return time.Date(2026, 5, 6, 11, 25, 0, 0, time.UTC) },
	})
	if _, err := failureService.VerifyEmailChallenge(ctx, auth.VerifyEmailChallengeInput{Token: "000000"}); err == nil {
		t.Fatal("invalid token verify unexpectedly succeeded")
	}

	assertCounterValue(t, registry, auth.MetricAuthChallengeStartedTotal, 1, "backend", "accepted")
	assertCounterValue(t, registry, auth.MetricAuthSessionMintedTotal, 1, "backend", "success")
	assertCounterValue(t, registry, auth.MetricAuthLogoutTotal, 1, "backend", "success")
	assertCounterValue(t, registry, auth.MetricAuthDeleteHandoffTotal, 1, "backend", "success")
	assertCounterValue(t, registry, auth.MetricAuthFailureTotal, 1, "backend", "verify_challenge", "invalid")

	if len(audit.events) < 5 {
		t.Fatalf("expected auth audit events, got %+v", audit.events)
	}
	for _, event := range audit.events {
		if event.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
			t.Fatalf("audit event missing trace ID: %+v", event)
		}
		if event.UserIDHash == "user-observe" {
			t.Fatalf("audit event leaked raw user ID: %+v", event)
		}
	}

	observed := fmt.Sprintf("%+v\n%+v", registry.CounterLabelValues(auth.MetricAuthFailureTotal), audit.events)
	assertNoAuthLeak(t, "auth observability", observed)
	for _, forbidden := range []string{"session-secret-id", "http://api.test/api/v1/auth/email/verify"} {
		if contains(observed, forbidden) {
			t.Fatalf("auth observability leaked %q: %s", forbidden, observed)
		}
	}
}

func assertCounterValue(t *testing.T, registry *auth.InMemoryAuthMetricRegistry, name string, want float64, labels ...string) {
	t.Helper()
	if got := registry.CounterValue(name, labels...); got != want {
		t.Fatalf("%s%v = %v want %v", name, labels, got, want)
	}
}

type recordingAuthAudit struct {
	events []auth.AuthAuditEvent
}

func (r *recordingAuthAudit) RecordAuthAuditEvent(_ context.Context, event auth.AuthAuditEvent) error {
	r.events = append(r.events, event)
	return nil
}
