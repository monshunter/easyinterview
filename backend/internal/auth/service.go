package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type PasswordlessServiceOptions struct {
	Store                 Store
	Dispatcher            MailDispatcher
	DeliverySecrets       DeliverySecretStore
	TokenGenerator        TokenGenerator
	SessionTokenGenerator TokenGenerator
	ChallengePepper       string
	SessionCookieSecret   string
	Now                   func() time.Time
	NewID                 func() string
}

type PasswordlessService struct {
	store                 Store
	dispatcher            MailDispatcher
	deliverySecrets       DeliverySecretStore
	tokenGenerator        TokenGenerator
	sessionTokenGenerator TokenGenerator
	challengePepper       string
	sessionCookieSecret   string
	now                   func() time.Time
	newID                 func() string
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

type VerifyEmailChallengeInput struct {
	Token      string
	RemoteAddr string
	UserAgent  string
}

type VerifyEmailChallengeResult struct {
	UserID           string
	SessionToken     string
	SessionExpiresAt time.Time
}

type CurrentSession struct {
	SessionID string
	UserID    string
	ExpiresAt time.Time
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
	sessionGenerator := opts.SessionTokenGenerator
	if sessionGenerator == nil {
		sessionGenerator = SecureTokenGenerator{}
	}
	newID := opts.NewID
	if newID == nil {
		newID = NewID
	}
	return &PasswordlessService{
		store:                 opts.Store,
		dispatcher:            opts.Dispatcher,
		deliverySecrets:       opts.DeliverySecrets,
		tokenGenerator:        generator,
		sessionTokenGenerator: sessionGenerator,
		challengePepper:       opts.ChallengePepper,
		sessionCookieSecret:   opts.SessionCookieSecret,
		now:                   now,
		newID:                 newID,
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

func (s *PasswordlessService) VerifyEmailChallenge(ctx context.Context, in VerifyEmailChallengeInput) (VerifyEmailChallengeResult, error) {
	if s == nil || s.store == nil {
		return VerifyEmailChallengeResult{}, fmt.Errorf("passwordless service store is nil")
	}
	token := strings.TrimSpace(in.Token)
	if token == "" {
		return VerifyEmailChallengeResult{}, ErrChallengeInvalid
	}
	now := s.now().UTC()
	challenge, err := s.store.ConsumeChallenge(ctx, hashWithPepper(s.challengePepper, token), now)
	if err != nil {
		if errors.Is(err, ErrChallengeExpired) || errors.Is(err, ErrChallengeConsumed) || errors.Is(err, ErrChallengeInvalid) {
			return VerifyEmailChallengeResult{}, err
		}
		return VerifyEmailChallengeResult{}, err
	}
	user, err := s.store.FindOrCreateUserByEmail(ctx, challenge.Email, s.newID(), now)
	if err != nil {
		return VerifyEmailChallengeResult{}, err
	}
	sessionToken, err := s.sessionTokenGenerator.GenerateToken()
	if err != nil {
		return VerifyEmailChallengeResult{}, err
	}
	expiresAt := now.Add(SessionTTL)
	sessionID := s.newID()
	sessionHash := hashWithPepper(s.sessionCookieSecret, sessionToken)
	if err := s.store.CreateSession(ctx, SessionRecord{
		ID:            sessionID,
		UserID:        user.ID,
		SessionHash:   sessionHash,
		Status:        SessionStatusActive,
		IPHash:        hashWithPepper(s.challengePepper, clientIP(in.RemoteAddr)),
		UserAgentHash: hashWithPepper(s.challengePepper, strings.TrimSpace(in.UserAgent)),
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		return VerifyEmailChallengeResult{}, err
	}
	return VerifyEmailChallengeResult{
		UserID:           user.ID,
		SessionToken:     sessionToken,
		SessionExpiresAt: expiresAt,
	}, nil
}

func (s *PasswordlessService) ResolveSession(ctx context.Context, rawToken string) (CurrentSession, error) {
	if s == nil || s.store == nil {
		return CurrentSession{}, fmt.Errorf("passwordless service store is nil")
	}
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return CurrentSession{}, ErrSessionInvalid
	}
	now := s.now().UTC()
	rec, err := s.store.GetSessionByHash(ctx, hashWithPepper(s.sessionCookieSecret, token), now)
	if err != nil {
		return CurrentSession{}, err
	}
	switch rec.Status {
	case SessionStatusActive:
	case SessionStatusRevoked:
		return CurrentSession{}, ErrSessionRevoked
	case SessionStatusExpired:
		return CurrentSession{}, ErrSessionExpired
	default:
		return CurrentSession{}, ErrSessionInvalid
	}
	if !rec.ExpiresAt.After(now) {
		return CurrentSession{}, ErrSessionExpired
	}
	nextExpiry := now.Add(SessionTTL)
	if err := s.store.TouchSession(ctx, rec.ID, now, nextExpiry); err != nil {
		return CurrentSession{}, err
	}
	return CurrentSession{
		SessionID: rec.ID,
		UserID:    rec.UserID,
		ExpiresAt: nextExpiry,
	}, nil
}

func (s *PasswordlessService) CurrentUser(ctx context.Context, userID string) (UserContext, error) {
	if s == nil || s.store == nil {
		return UserContext{}, fmt.Errorf("passwordless service store is nil")
	}
	return s.store.GetUserContext(ctx, userID)
}

func (s *PasswordlessService) Logout(ctx context.Context, current CurrentSession) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("passwordless service store is nil")
	}
	if current.SessionID == "" {
		return nil
	}
	return s.store.RevokeSession(ctx, current.SessionID, s.now().UTC())
}

func (s *PasswordlessService) DeleteMe(ctx context.Context, current CurrentSession, idempotencyKey string) (PrivacyDeleteHandoff, error) {
	if s == nil || s.store == nil {
		return PrivacyDeleteHandoff{}, fmt.Errorf("passwordless service store is nil")
	}
	now := s.now().UTC()
	if idempotencyKey == "" {
		idempotencyKey = "privacy_delete:" + current.UserID
	}
	handoff, err := s.store.CreatePrivacyDeleteHandoff(ctx, current.UserID, idempotencyKey, s.newID(), s.newID(), now)
	if err != nil {
		return PrivacyDeleteHandoff{}, err
	}
	if current.SessionID != "" {
		if err := s.store.RevokeSession(ctx, current.SessionID, now); err != nil {
			return PrivacyDeleteHandoff{}, err
		}
	}
	return handoff, nil
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
