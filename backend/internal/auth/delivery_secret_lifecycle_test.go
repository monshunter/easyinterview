package auth_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/auth"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type sharedDeliverySecretRedis struct {
	mu     sync.Mutex
	values map[string]string
	keys   []string
	ttls   []time.Duration
}

type sharedDeliverySecretRedisClient struct{ backend *sharedDeliverySecretRedis }

func (c sharedDeliverySecretRedisClient) Set(_ context.Context, key, value string, ttl time.Duration) error {
	c.backend.mu.Lock()
	defer c.backend.mu.Unlock()
	if c.backend.values == nil {
		c.backend.values = map[string]string{}
	}
	c.backend.values[key] = value
	c.backend.keys = append(c.backend.keys, key)
	c.backend.ttls = append(c.backend.ttls, ttl)
	return nil
}

func (c sharedDeliverySecretRedisClient) Get(_ context.Context, key string) (string, bool, error) {
	c.backend.mu.Lock()
	defer c.backend.mu.Unlock()
	value, ok := c.backend.values[key]
	return value, ok, nil
}

func (c sharedDeliverySecretRedisClient) Del(_ context.Context, key string) error {
	c.backend.mu.Lock()
	defer c.backend.mu.Unlock()
	delete(c.backend.values, key)
	return nil
}

type lifecycleSecretStore struct {
	putErr  error
	getErr  error
	delErr  error
	secrets map[string]string
	deleted []string
}

func (s *lifecycleSecretStore) PutDeliverySecret(_ context.Context, ref, token string, _ time.Duration) error {
	if s.putErr != nil {
		return s.putErr
	}
	if s.secrets == nil {
		s.secrets = map[string]string{}
	}
	s.secrets[ref] = token
	return nil
}

func (s *lifecycleSecretStore) GetDeliverySecret(_ context.Context, ref string) (string, bool, error) {
	if s.getErr != nil {
		return "", false, s.getErr
	}
	token, ok := s.secrets[ref]
	return token, ok, nil
}

func (s *lifecycleSecretStore) DeleteDeliverySecret(_ context.Context, ref string) error {
	s.deleted = append(s.deleted, ref)
	if s.delErr != nil {
		return s.delErr
	}
	delete(s.secrets, ref)
	return nil
}

type failingDispatcher struct {
	called bool
	err    error
}

func (d *failingDispatcher) Enqueue(context.Context, jobs.EmailDispatchPayload) error {
	d.called = true
	return d.err
}

func TestStartEmailChallengeDoesNotEnqueueWhenDeliverySecretStorageFails(t *testing.T) {
	secrets := &lifecycleSecretStore{putErr: errors.New("redis://user:password@private-host:6379 123456")}
	dispatcher := &failingDispatcher{}
	service := newLifecycleEmailCodeService(secrets, dispatcher)

	_, err := service.StartEmailChallenge(context.Background(), auth.StartEmailChallengeInput{Email: "candidate@example.test"})
	if err == nil {
		t.Fatal("expected delivery secret storage failure")
	}
	if dispatcher.called {
		t.Fatal("dispatch must not run after delivery secret storage failure")
	}
	for _, forbidden := range []string{"password", "private-host", "123456"} {
		if contains(err.Error(), forbidden) {
			t.Fatalf("storage error leaked %q: %v", forbidden, err)
		}
	}
}

func TestStartEmailChallengeDeletesDeliverySecretWhenEnqueueFails(t *testing.T) {
	secrets := &lifecycleSecretStore{}
	dispatcher := &failingDispatcher{err: errors.New("database unavailable")}
	service := newLifecycleEmailCodeService(secrets, dispatcher)

	_, err := service.StartEmailChallenge(context.Background(), auth.StartEmailChallengeInput{Email: "candidate@example.test"})
	if err == nil {
		t.Fatal("expected enqueue failure")
	}
	if len(secrets.deleted) != 1 || secrets.deleted[0] != "auth_challenge:challenge-lifecycle" {
		t.Fatalf("deleted refs = %#v", secrets.deleted)
	}
	if _, ok := secrets.secrets["auth_challenge:challenge-lifecycle"]; ok {
		t.Fatal("enqueue failure left a delivery secret behind")
	}
}

func newLifecycleEmailCodeService(secrets auth.DeliverySecretStore, dispatcher auth.MailDispatcher) *auth.EmailCodeService {
	return auth.NewEmailCodeService(auth.EmailCodeServiceOptions{
		Store:               &recordingChallengeStore{},
		Dispatcher:          dispatcher,
		DeliverySecrets:     secrets,
		TokenGenerator:      fixedTokenGenerator("123456"),
		ChallengePepper:     "pepper",
		SessionCookieSecret: "session-secret",
		NewID:               func() string { return "challenge-lifecycle" },
		Now:                 func() time.Time { return time.Unix(0, 0).UTC() },
	})
}

func TestEmailCodeDeliveryWorksAcrossIndependentRedisBackedInstances(t *testing.T) {
	backend := &sharedDeliverySecretRedis{}
	producerSecrets, err := auth.NewRedisDeliverySecretStoreWithClient(sharedDeliverySecretRedisClient{backend}, "shared-pepper")
	if err != nil {
		t.Fatalf("producer store: %v", err)
	}
	consumerSecrets, err := auth.NewRedisDeliverySecretStoreWithClient(sharedDeliverySecretRedisClient{backend}, "shared-pepper")
	if err != nil {
		t.Fatalf("consumer store: %v", err)
	}
	dispatcher := &payloadRecordingDispatcher{}
	service := newLifecycleEmailCodeService(producerSecrets, dispatcher)

	result, err := service.StartEmailChallenge(context.Background(), auth.StartEmailChallengeInput{Email: "candidate@example.test"})
	if err != nil || !result.Accepted {
		t.Fatalf("StartEmailChallenge = %+v, %v", result, err)
	}
	var message string
	writer := auth.NewSMTPDeliveryWriter(auth.SMTPDeliveryWriterOptions{
		SMTPAddr:        "smtp.example.test:587",
		FromAddress:     "noreply@example.test",
		DeliverySecrets: consumerSecrets,
		LookupChallengeEmail: func(string) (string, error) {
			return "candidate@example.test", nil
		},
		Send: func(envelope auth.SMTPEnvelope) error {
			message = string(envelope.Message)
			return nil
		},
	})
	if err := writer.Write(dispatcher.payload); err != nil {
		t.Fatalf("consumer Write: %v", err)
	}
	if !strings.Contains(message, "123456") {
		t.Fatalf("cross-instance message did not contain the generated six-digit code")
	}
	if _, ok, err := producerSecrets.GetDeliverySecret(context.Background(), dispatcher.payload["deliverySecretRef"]); err != nil || ok {
		t.Fatalf("successful cross-instance delivery did not delete secret: ok=%v err=%v", ok, err)
	}
	backend.mu.Lock()
	defer backend.mu.Unlock()
	if len(backend.ttls) != 1 || backend.ttls[0] != auth.ChallengeTTL {
		t.Fatalf("Redis TTLs = %#v", backend.ttls)
	}
	for _, stored := range append(append([]string{}, backend.keys...), mapValues(backend.values)...) {
		for _, forbidden := range []string{"123456", "auth_challenge:challenge-lifecycle"} {
			if strings.Contains(stored, forbidden) {
				t.Fatalf("Redis material leaked %q", forbidden)
			}
		}
	}
	for _, value := range dispatcher.payload {
		if value == "123456" {
			t.Fatal("async payload leaked raw code")
		}
	}
}

func mapValues(values map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}
	return out
}
