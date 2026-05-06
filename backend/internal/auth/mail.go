package auth

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type DeliverySecretStore interface {
	PutDeliverySecret(ref string, token string)
	GetDeliverySecret(ref string) (string, bool)
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

func (s *DevMailSink) PutDeliverySecret(ref string, token string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.secrets[ref] = token
}

func (s *DevMailSink) GetDeliverySecret(ref string) (string, bool) {
	if s == nil {
		return "", false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	token, ok := s.secrets[ref]
	return token, ok
}

func (s *DevMailSink) Write(payload jobs.EmailDispatchPayload) error {
	if s == nil {
		return fmt.Errorf("dev mail sink is nil")
	}
	challengeID := payload["authChallengeId"]
	if challengeID == "" {
		return fmt.Errorf("email_dispatch payload missing authChallengeId")
	}
	secretRef := payload["deliverySecretRef"]
	if secretRef == "" {
		return fmt.Errorf("email_dispatch payload missing deliverySecretRef")
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

func (s *DevMailSink) MagicLinkForChallenge(challengeID string) (string, bool) {
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
	base := s.verifyBaseURL
	s.mu.Unlock()
	if !ok || token == "" || base == "" {
		return "", false
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", false
	}
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	return u.String(), true
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

type BackgroundMailDispatcherOptions struct {
	Writer    DeliveryWriter
	QueueSize int
}

type BackgroundMailDispatcher struct {
	writer DeliveryWriter
	queue  chan jobs.EmailDispatchPayload
	done   chan struct{}

	closeOnce sync.Once
	mu        sync.Mutex
	errors    []string
	closed    bool
}

func NewBackgroundMailDispatcher(opts BackgroundMailDispatcherOptions) *BackgroundMailDispatcher {
	size := opts.QueueSize
	if size <= 0 {
		size = 16
	}
	d := &BackgroundMailDispatcher{
		writer: opts.Writer,
		queue:  make(chan jobs.EmailDispatchPayload, size),
		done:   make(chan struct{}),
	}
	go d.run()
	return d
}

func (d *BackgroundMailDispatcher) Enqueue(ctx context.Context, payload jobs.EmailDispatchPayload) error {
	if d == nil {
		return fmt.Errorf("mail dispatcher is nil")
	}
	d.mu.Lock()
	closed := d.closed
	d.mu.Unlock()
	if closed {
		return fmt.Errorf("mail dispatcher is closed")
	}
	select {
	case d.queue <- payload:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *BackgroundMailDispatcher) Shutdown(ctx context.Context) error {
	if d == nil {
		return nil
	}
	d.closeOnce.Do(func() {
		d.mu.Lock()
		d.closed = true
		d.mu.Unlock()
		close(d.queue)
	})
	select {
	case <-d.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (d *BackgroundMailDispatcher) ErrorSummaries() []string {
	if d == nil {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]string, len(d.errors))
	copy(out, d.errors)
	return out
}

func (d *BackgroundMailDispatcher) run() {
	defer close(d.done)
	for payload := range d.queue {
		if d.writer == nil {
			d.recordError(payload, "email_dispatch failed: writer unavailable")
			continue
		}
		if err := d.writer.Write(payload); err != nil {
			d.recordError(payload, "email_dispatch failed")
		}
	}
}

func (d *BackgroundMailDispatcher) recordError(payload jobs.EmailDispatchPayload, summary string) {
	challengeID := payload["authChallengeId"]
	if challengeID != "" {
		summary += " challenge_id=" + challengeID
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.errors = append(d.errors, summary)
}
