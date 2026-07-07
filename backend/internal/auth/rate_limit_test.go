package auth_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestStartEmailChallengeDedupesThirdRecentEmailOrIPRequest(t *testing.T) {
	store := &rateLimitStore{recentCount: 2}
	dispatcher := &recordingDispatcher{}
	now := time.Date(2026, 5, 6, 10, 5, 0, 0, time.UTC)
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:           store,
		Dispatcher:      dispatcher,
		DeliverySecrets: auth.NewDevMailSink(auth.DevMailSinkOptions{}),
		TokenGenerator:  fixedTokenGenerator("raw-token-for-rate-limit"),
		ChallengePepper: "pepper",
		Now:             func() time.Time { return now },
		NewID:           fixedIDs("018f2a40-0000-7000-9000-000000000011"),
	})
	handler := auth.NewHandler(auth.HandlerOptions{EmailCode: service})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/email/start", bytes.NewBufferString(`{"email":"Candidate@Example.COM"}`))
	req.RemoteAddr = "203.0.113.21:5588"
	rec := httptest.NewRecorder()

	handler.StartAuthEmailChallenge(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if store.created {
		t.Fatal("third recent request must not create another challenge")
	}
	if dispatcher.enqueued {
		t.Fatal("third recent request must not enqueue another email dispatch")
	}
	if store.countEmail != "candidate@example.com" {
		t.Fatalf("count email = %q", store.countEmail)
	}
	if store.countIPHash == "" || store.countIPHash == "203.0.113.21" {
		t.Fatalf("count IP must use hash, got %q", store.countIPHash)
	}
	if !store.countSince.Equal(now.Add(-auth.RateLimitWindow)) {
		t.Fatalf("count since = %s", store.countSince)
	}
}

type rateLimitStore struct {
	recentCount int
	created     bool
	countEmail  string
	countIPHash string
	countSince  time.Time
}

func (s *rateLimitStore) CountRecentChallenges(_ context.Context, email string, ipHash string, since time.Time) (int, error) {
	s.countEmail = email
	s.countIPHash = ipHash
	s.countSince = since
	return s.recentCount, nil
}

func (s *rateLimitStore) CreateChallenge(context.Context, auth.ChallengeRecord) error {
	s.created = true
	return nil
}

func (s *rateLimitStore) ConsumeChallenge(context.Context, string, time.Time) (auth.ChallengeRecord, error) {
	panic("not used")
}

func (s *rateLimitStore) CreateUserByEmail(context.Context, string, string, string, time.Time) (auth.UserContext, error) {
	panic("not used")
}

func (s *rateLimitStore) FindUserByEmail(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *rateLimitStore) CreateSession(context.Context, auth.SessionRecord) error {
	panic("not used")
}

func (s *rateLimitStore) GetSessionByHash(context.Context, string, time.Time) (auth.SessionRecord, error) {
	panic("not used")
}

func (s *rateLimitStore) GetUserContext(context.Context, string) (auth.UserContext, error) {
	panic("not used")
}

func (s *rateLimitStore) TouchSession(context.Context, string, time.Time, time.Time) error {
	panic("not used")
}

func (s *rateLimitStore) RevokeSession(context.Context, string, time.Time) error {
	panic("not used")
}

func (s *rateLimitStore) CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (auth.PrivacyDeleteHandoff, error) {
	panic("not used")
}

type recordingDispatcher struct {
	enqueued bool
}

func (d *recordingDispatcher) Enqueue(context.Context, jobs.EmailDispatchPayload) error {
	d.enqueued = true
	return nil
}
