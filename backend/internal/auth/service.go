package auth

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type PasswordlessServiceOptions struct {
	Store           Store
	Dispatcher      MailDispatcher
	DeliverySecrets DeliverySecretStore
	TokenGenerator  TokenGenerator
	ChallengePepper string
	Now             func() time.Time
	NewID           func() string
}

type PasswordlessService struct {
	store           Store
	dispatcher      MailDispatcher
	deliverySecrets DeliverySecretStore
	tokenGenerator  TokenGenerator
	challengePepper string
	now             func() time.Time
	newID           func() string
}

type StartEmailChallengeInput struct {
	Email        string
	ReturnTo     string
	RemoteAddr   string
	UserAgent    string
	AcceptLocale string
}

type StartEmailChallengeResult struct {
	ChallengeID string
	Accepted    bool
	RateLimited bool
}

func NewPasswordlessService(opts PasswordlessServiceOptions) *PasswordlessService {
	now := opts.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	generator := opts.TokenGenerator
	if generator == nil {
		generator = SecureTokenGenerator{}
	}
	newID := opts.NewID
	if newID == nil {
		newID = NewID
	}
	return &PasswordlessService{
		store:           opts.Store,
		dispatcher:      opts.Dispatcher,
		deliverySecrets: opts.DeliverySecrets,
		tokenGenerator:  generator,
		challengePepper: opts.ChallengePepper,
		now:             now,
		newID:           newID,
	}
}

func (s *PasswordlessService) StartEmailChallenge(ctx context.Context, in StartEmailChallengeInput) (StartEmailChallengeResult, error) {
	if s == nil || s.store == nil {
		return StartEmailChallengeResult{}, fmt.Errorf("passwordless service store is nil")
	}
	if s.dispatcher == nil {
		return StartEmailChallengeResult{}, fmt.Errorf("passwordless service dispatcher is nil")
	}
	if s.deliverySecrets == nil {
		return StartEmailChallengeResult{}, fmt.Errorf("passwordless service delivery secrets are nil")
	}
	email := normalizeEmail(in.Email)
	if email == "" {
		return StartEmailChallengeResult{}, fmt.Errorf("email is required")
	}
	now := s.now().UTC()
	challengeID := s.newID()
	ipHash := hashWithPepper(s.challengePepper, clientIP(in.RemoteAddr))
	uaHash := hashWithPepper(s.challengePepper, strings.TrimSpace(in.UserAgent))
	recent, err := s.store.CountRecentChallenges(ctx, email, ipHash, now.Add(-RateLimitWindow))
	if err != nil {
		return StartEmailChallengeResult{}, err
	}
	if recent >= RateLimitThreshold-1 {
		return StartEmailChallengeResult{ChallengeID: challengeID, Accepted: true, RateLimited: true}, nil
	}
	token, err := s.tokenGenerator.GenerateToken()
	if err != nil {
		return StartEmailChallengeResult{}, err
	}
	tokenHash := hashWithPepper(s.challengePepper, token)
	if err := s.store.CreateChallenge(ctx, ChallengeRecord{
		ID:            challengeID,
		Email:         email,
		TokenHash:     tokenHash,
		Purpose:       ChallengePurposeLogin,
		IPHash:        ipHash,
		UserAgentHash: uaHash,
		ExpiresAt:     now.Add(ChallengeTTL),
		CreatedAt:     now,
	}); err != nil {
		return StartEmailChallengeResult{}, err
	}

	deliverySecretRef := "auth_challenge:" + challengeID
	s.deliverySecrets.PutDeliverySecret(deliverySecretRef, token)
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   challengeID,
		"templateKey":       "auth_magic_link",
		"locale":            localeOrDefault(in.AcceptLocale),
		"deliverySecretRef": deliverySecretRef,
		"dedupeKey":         hashWithPepper(s.challengePepper, "email:"+email),
	})
	if err != nil {
		return StartEmailChallengeResult{}, err
	}
	if err := s.dispatcher.Enqueue(ctx, payload); err != nil {
		return StartEmailChallengeResult{}, err
	}
	return StartEmailChallengeResult{ChallengeID: challengeID, Accepted: true}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func clientIP(remoteAddr string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(remoteAddr)
}

func localeOrDefault(locale string) string {
	if locale = strings.TrimSpace(locale); locale != "" {
		return locale
	}
	return "en"
}
