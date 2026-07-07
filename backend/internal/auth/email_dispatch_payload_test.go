package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

func TestStartEmailChallengeUsesGeneratedEmailDispatchPayloadRedaction(t *testing.T) {
	store := &recordingChallengeStore{}
	dispatcher := &payloadRecordingDispatcher{}
	sink := auth.NewDevMailSink(auth.DevMailSinkOptions{})
	service := auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:           store,
		Dispatcher:      dispatcher,
		DeliverySecrets: sink,
		TokenGenerator:  fixedTokenGenerator("raw-token-for-payload-test"),
		ChallengePepper: "pepper",
		Now:             func() time.Time { return time.Date(2026, 5, 6, 10, 10, 0, 0, time.UTC) },
		NewID:           fixedIDs("018f2a40-0000-7000-9000-000000000012"),
	})

	if _, err := service.StartEmailChallenge(context.Background(), auth.StartEmailChallengeInput{
		Email:      "candidate@example.com",
		RemoteAddr: "203.0.113.22:5588",
		UserAgent:  "unit-test-agent",
	}); err != nil {
		t.Fatalf("StartEmailChallenge: %v", err)
	}

	allowed := map[string]struct{}{}
	for _, field := range jobs.EmailDispatchAllowedPayloadFields {
		allowed[field] = struct{}{}
	}
	for field, value := range dispatcher.payload {
		if _, ok := allowed[field]; !ok {
			t.Fatalf("service produced non-helper field %s", field)
		}
		if contains(value, "raw-token-for-payload-test") || contains(value, "candidate@example.com") {
			t.Fatalf("service leaked redacted value in %s=%q", field, value)
		}
	}
	for _, field := range jobs.EmailDispatchRedactedFields {
		if _, err := jobs.BuildEmailDispatchPayload(map[string]string{field: "forbidden"}); err == nil {
			t.Fatalf("BuildEmailDispatchPayload allowed redacted field %s", field)
		}
	}
}

type payloadRecordingDispatcher struct {
	payload jobs.EmailDispatchPayload
}

func (d *payloadRecordingDispatcher) Enqueue(_ context.Context, payload jobs.EmailDispatchPayload) error {
	d.payload = payload
	return nil
}
