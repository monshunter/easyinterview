package auth_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/runner"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

var (
	errFakeDelivery  = errors.New("delivery failed")
	errLeakyDelivery = errors.New("provider failed for raw-token candidate@example.com http://api.test/verify?token=raw-token")
)

type failingDeliveryWriter struct {
	err error
}

func (w *failingDeliveryWriter) Write(jobs.EmailDispatchPayload) error { return w.err }

type recordingDeliveryWriter struct {
	delivered []jobs.EmailDispatchPayload
}

func (w *recordingDeliveryWriter) Write(p jobs.EmailDispatchPayload) error {
	w.delivered = append(w.delivered, p)
	return nil
}

func validEmailDispatchPayloadJSON(t *testing.T, challengeID string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]string{
		"authChallengeId":   challengeID,
		"templateKey":       "auth_login_code",
		"locale":            "en",
		"deliverySecretRef": "auth_challenge:" + challengeID,
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestEmailDispatchHandler_DeliversValidPayload(t *testing.T) {
	writer := &recordingDeliveryWriter{}
	h := auth.NewEmailDispatchHandler(writer)
	out := h.Handle(context.Background(), runner.ClaimedJob{
		JobType: string(jobs.JobTypeEmailDispatch),
		Payload: validEmailDispatchPayloadJSON(t, "challenge-1"),
	})
	if !out.Succeeded {
		t.Fatalf("outcome = %+v, want succeeded", out)
	}
	if len(writer.delivered) != 1 || writer.delivered[0]["authChallengeId"] != "challenge-1" {
		t.Fatalf("delivery = %+v", writer.delivered)
	}
}

func TestEmailDispatchHandler_RejectsForbiddenField(t *testing.T) {
	writer := &recordingDeliveryWriter{}
	h := auth.NewEmailDispatchHandler(writer)
	// rawMagicLinkToken is a redacted field; the payload validator must reject it.
	raw, _ := json.Marshal(map[string]string{
		"authChallengeId":   "challenge-1",
		"rawMagicLinkToken": "secret-token",
	})
	out := h.Handle(context.Background(), runner.ClaimedJob{Payload: raw})
	if out.Succeeded || out.Retryable {
		t.Fatalf("outcome = %+v, want permanent (non-retryable) failure", out)
	}
	if len(writer.delivered) != 0 {
		t.Fatalf("must not deliver a forbidden payload")
	}
}

func TestEmailDispatchHandler_RetriesOnWriteError(t *testing.T) {
	h := auth.NewEmailDispatchHandler(&failingDeliveryWriter{err: errFakeDelivery})
	out := h.Handle(context.Background(), runner.ClaimedJob{Payload: validEmailDispatchPayloadJSON(t, "challenge-1")})
	if out.Succeeded || !out.Retryable {
		t.Fatalf("outcome = %+v, want retryable", out)
	}
}

func TestEmailDispatchHandler_PayloadRedaction(t *testing.T) {
	h := auth.NewEmailDispatchHandler(&failingDeliveryWriter{
		err: errLeakyDelivery,
	})
	out := h.Handle(context.Background(), runner.ClaimedJob{Payload: validEmailDispatchPayloadJSON(t, "challenge-1")})
	combined := out.ErrorCode + " " + out.ErrorMessage
	for _, forbidden := range []string{"raw-token", "candidate@example.com", "http://api.test"} {
		if strings.Contains(combined, forbidden) {
			t.Fatalf("handler outcome leaked %q: %s", forbidden, combined)
		}
	}
}

type captureExecer struct {
	lastQuery string
	lastArgs  []any
}

func (e *captureExecer) ExecContext(_ context.Context, query string, args ...any) (sql.Result, error) {
	e.lastQuery = query
	e.lastArgs = args
	return driverResult{}, nil
}

type driverResult struct{}

func (driverResult) LastInsertId() (int64, error) { return 0, nil }
func (driverResult) RowsAffected() (int64, error) { return 1, nil }

func TestEmailDispatchEnqueuer_InsertsEmailDispatchJob(t *testing.T) {
	exec := &captureExecer{}
	enq := auth.NewEmailDispatchEnqueuer(exec, func() string { return "job-1" }, func() time.Time { return time.Unix(0, 0).UTC() })
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   "challenge-1",
		"templateKey":       "auth_login_code",
		"locale":            "en",
		"deliverySecretRef": "auth_challenge:challenge-1",
		"dedupeKey":         "dedupe-hash",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := enq.Enqueue(context.Background(), payload); err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if !strings.Contains(exec.lastQuery, "insert into async_jobs") {
		t.Fatalf("query did not insert an async job: %s", exec.lastQuery)
	}
	if len(exec.lastArgs) < 3 || exec.lastArgs[1] != string(jobs.JobTypeEmailDispatch) || exec.lastArgs[2] != "challenge-1" {
		t.Fatalf("resource_id arg = %v, want challenge-1", exec.lastArgs)
	}
}

func TestStartAuthEmailChallenge_EnqueuesEmailDispatchJob(t *testing.T) {
	exec := &captureExecer{}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{VerifyBaseURL: "http://api.test/verify"})
	enq := auth.NewEmailDispatchEnqueuer(exec, func() string { return "job-1" }, func() time.Time { return time.Unix(0, 0).UTC() })
	service := auth.NewPasswordlessService(auth.PasswordlessServiceOptions{
		Store:               &recordingChallengeStore{},
		Dispatcher:          enq,
		DeliverySecrets:     sink,
		TokenGenerator:      fixedTokenGenerator("123456"),
		ChallengePepper:     "pepper",
		SessionCookieSecret: "session-secret",
		NewID:               func() string { return "challenge-x" },
		Now:                 func() time.Time { return time.Unix(0, 0).UTC() },
	})
	res, err := service.StartEmailChallenge(context.Background(), auth.StartEmailChallengeInput{Email: "candidate@example.com"})
	if err != nil {
		t.Fatalf("StartEmailChallenge: %v", err)
	}
	if !res.Accepted {
		t.Fatalf("challenge not accepted: %+v", res)
	}
	if !strings.Contains(exec.lastQuery, "insert into async_jobs") {
		t.Fatalf("StartEmailChallenge did not enqueue an async job: %s", exec.lastQuery)
	}
	if len(exec.lastArgs) < 3 || exec.lastArgs[1] != string(jobs.JobTypeEmailDispatch) || exec.lastArgs[2] != "challenge-x" {
		t.Fatalf("enqueued resource_id = %v, want challenge-x", exec.lastArgs)
	}
}
