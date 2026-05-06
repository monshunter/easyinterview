package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestStartAuthEmailChallengeCreatesHashedChallengeAndDispatchesDevLink(t *testing.T) {
	store := &recordingChallengeStore{}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{
		VerifyBaseURL: "http://api.test/api/v1/auth/email/verify",
	})
	dispatcher := auth.NewImmediateMailDispatcher(sink)
	now := time.Date(2026, 5, 6, 10, 0, 0, 0, time.UTC)
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:           store,
		Dispatcher:      dispatcher,
		DeliverySecrets: sink,
		TokenGenerator:  fixedTokenGenerator("raw-token-for-test"),
		ChallengePepper: "pepper",
		Now:             func() time.Time { return now },
		NewID:           fixedIDs("018f2a40-0000-7000-9000-000000000010"),
	})
	handler := auth.NewHandler(auth.HandlerOptions{Passwordless: service})

	body := bytes.NewBufferString(`{"email":"Candidate@Example.COM","returnTo":"/practice?planId=plan_1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/email/start", body)
	req.RemoteAddr = "203.0.113.20:5588"
	req.Header.Set("User-Agent", "unit-test-agent")
	rec := httptest.NewRecorder()

	handler.StartAuthEmailChallenge(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.challenge.Email != "candidate@example.com" {
		t.Fatalf("email normalized = %q", store.challenge.Email)
	}
	if store.challenge.TokenHash == "" || store.challenge.TokenHash == "raw-token-for-test" {
		t.Fatalf("challenge token must be stored as non-empty hash, got %q", store.challenge.TokenHash)
	}
	if store.challenge.IPHash == "" || store.challenge.IPHash == "203.0.113.20" {
		t.Fatalf("IP must be stored as hash, got %q", store.challenge.IPHash)
	}
	if store.challenge.UserAgentHash == "" || store.challenge.UserAgentHash == "unit-test-agent" {
		t.Fatalf("UA must be stored as hash, got %q", store.challenge.UserAgentHash)
	}
	if !store.challenge.ExpiresAt.Equal(now.Add(auth.ChallengeTTL)) {
		t.Fatalf("expiresAt = %s", store.challenge.ExpiresAt)
	}

	link, ok := sink.MagicLinkForChallenge("018f2a40-0000-7000-9000-000000000010")
	if !ok {
		t.Fatal("dev mail sink did not expose retrieval link")
	}
	if !contains(link, "token=raw-token-for-test") {
		t.Fatalf("retrieval link missing token: %s", link)
	}
	if sink.ContainsStoredSecret("raw-token-for-test") {
		t.Fatal("dev sink stored raw token instead of transient retrieval secret")
	}
	if sink.ContainsStoredSecret("http://api.test/api/v1/auth/email/verify") {
		t.Fatal("dev sink stored full magic-link URL")
	}
	if sink.ContainsStoredSecret("Candidate@Example.COM") || sink.ContainsStoredSecret("candidate@example.com") {
		t.Fatal("dev sink stored recipient email")
	}
}

type recordingChallengeStore struct {
	challenge auth.ChallengeRecord
}

func (s *recordingChallengeStore) CountRecentChallenges(context.Context, string, string, time.Time) (int, error) {
	return 0, nil
}

func (s *recordingChallengeStore) CreateChallenge(_ context.Context, rec auth.ChallengeRecord) error {
	s.challenge = rec
	return nil
}

func (s *recordingChallengeStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	panic("not used")
}

func (s *recordingChallengeStore) FindOrCreateUserByEmail(context.Context, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *recordingChallengeStore) CreateSession(context.Context, auth.SessionRecord) error {
	panic("not used")
}

func (s *recordingChallengeStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	panic("not used")
}

func (s *recordingChallengeStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *recordingChallengeStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	panic("not used")
}

func (s *recordingChallengeStore) RevokeSession(context.Context, string, time.Time) error {
	panic("not used")
}

func (s *recordingChallengeStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}

type fixedTokenGenerator string

func (g fixedTokenGenerator) GenerateToken() (string, error) {
	return string(g), nil
}

func fixedIDs(ids ...string) func() string {
	i := 0
	return func() string {
		if i >= len(ids) {
			return ids[len(ids)-1]
		}
		id := ids[i]
		i++
		return id
	}
}

func TestStartAuthEmailChallengeRejectsMalformedJSON(t *testing.T) {
	handler := auth.NewHandler(auth.HandlerOptions{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/email/start", bytes.NewBufferString(`{`))
	rec := httptest.NewRecorder()

	handler.StartAuthEmailChallenge(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("bad error JSON: %v", err)
	}
	if _, ok := payload["error"]; !ok {
		t.Fatalf("missing B1 error envelope: %s", rec.Body.String())
	}
}
