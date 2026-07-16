package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type DeliverySecretStore interface {
	PutDeliverySecret(context.Context, string, string, time.Duration) error
	GetDeliverySecret(context.Context, string) (string, bool, error)
	DeleteDeliverySecret(context.Context, string) error
}

type MailDispatcher interface {
	Enqueue(context.Context, jobs.EmailDispatchPayload) error
}

type DeliveryWriter interface {
	Write(jobs.EmailDispatchPayload) error
}

type DevMailSinkOptions struct {
	VerifyBaseURL string
}

type DevMailSink struct {
	mu            sync.Mutex
	verifyBaseURL string
	secrets       map[string]string
	deliveries    map[string]DevMailDelivery
}

type DevMailDelivery struct {
	ChallengeID       string
	TemplateKey       string
	Locale            string
	DeliverySecretRef string
	DedupeKey         string
	CreatedAt         time.Time
}

func NewDevMailSink(opts DevMailSinkOptions) *DevMailSink {
	return &DevMailSink{
		verifyBaseURL: opts.VerifyBaseURL,
		secrets:       map[string]string{},
		deliveries:    map[string]DevMailDelivery{},
	}
}

func (s *DevMailSink) PutDeliverySecret(_ context.Context, ref string, token string, _ time.Duration) error {
	if s == nil {
		return fmt.Errorf("dev mail sink is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[ref] = token
	return nil
}

func (s *DevMailSink) GetDeliverySecret(_ context.Context, ref string) (string, bool, error) {
	if s == nil {
		return "", false, fmt.Errorf("dev mail sink is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.secrets[ref]
	return token, ok, nil
}

func (s *DevMailSink) DeleteDeliverySecret(_ context.Context, ref string) error {
	if s == nil {
		return fmt.Errorf("dev mail sink is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.secrets, ref)
	return nil
}

func (s *DevMailSink) Write(payload jobs.EmailDispatchPayload) error {
	if s == nil {
		return fmt.Errorf("dev mail sink is nil")
	}
	challengeID := payload["authChallengeId"]
	if challengeID == "" {
		return fmt.Errorf("%s payload missing authChallengeId", jobs.JobTypeEmailDispatch)
	}
	secretRef := payload["deliverySecretRef"]
	if secretRef == "" {
		return fmt.Errorf("%s payload missing deliverySecretRef", jobs.JobTypeEmailDispatch)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deliveries[challengeID] = DevMailDelivery{
		ChallengeID:       challengeID,
		TemplateKey:       payload["templateKey"],
		Locale:            payload["locale"],
		DeliverySecretRef: secretRef,
		DedupeKey:         payload["dedupeKey"],
		CreatedAt:         time.Now().UTC(),
	}
	return nil
}

func (s *DevMailSink) CodeForChallenge(challengeID string) (string, bool) {
	if s == nil {
		return "", false
	}
	s.mu.Lock()
	delivery, ok := s.deliveries[challengeID]
	if !ok {
		s.mu.Unlock()
		return "", false
	}
	token, ok := s.secrets[delivery.DeliverySecretRef]
	s.mu.Unlock()
	if !ok || token == "" {
		return "", false
	}
	return token, true
}

// ContainsStoredSecret scans only persisted sink delivery metadata. The
// transient secret map is the retrieval boundary and is not part of queued or
// sink delivery payload evidence.
func (s *DevMailSink) ContainsStoredSecret(value string) bool {
	if s == nil || value == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, delivery := range s.deliveries {
		blob := strings.Join([]string{
			delivery.ChallengeID,
			delivery.TemplateKey,
			delivery.Locale,
			delivery.DeliverySecretRef,
			delivery.DedupeKey,
		}, "\n")
		if strings.Contains(blob, value) {
			return true
		}
	}
	return false
}

func (s *DevMailSink) String() string {
	if s == nil {
		return "DevMailSink<nil>"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("DevMailSink{deliveries:%d secrets:redacted verifyBaseURL:redacted}", len(s.deliveries))
}

func (s *DevMailSink) GoString() string {
	return s.String()
}

type ImmediateMailDispatcher struct {
	sink *DevMailSink
}

func NewImmediateMailDispatcher(sink *DevMailSink) *ImmediateMailDispatcher {
	return &ImmediateMailDispatcher{sink: sink}
}

func (d *ImmediateMailDispatcher) Enqueue(_ context.Context, payload jobs.EmailDispatchPayload) error {
	if d == nil || d.sink == nil {
		return fmt.Errorf("mail dispatcher sink is nil")
	}
	return d.sink.Write(payload)
}

// Email delivery now flows through async_jobs(job_type=email_dispatch) via
// EmailDispatchEnqueuer (producer) and EmailDispatchHandler (kernel handler),
// per backend-async-runner spec D-10.
