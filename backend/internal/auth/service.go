package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
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
	Metrics               AuthMetrics
	Audit                 AuthAuditRecorder
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
	metrics               AuthMetrics
	audit                 AuthAuditRecorder
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

type CompleteProfileInput struct {
	UserID        string
	DisplayName   string
	AcceptedTerms bool
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
		generator = SixDigitCodeGenerator{}
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
		metrics:               opts.Metrics,
		audit:                 opts.Audit,
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
		s.recordAuthFailure(ctx, "start_challenge", "validation", "", "")
		return StartEmailChallengeResult{}, fmt.Errorf("email is required")
	}
	now := s.now().UTC()
	challengeID := s.newID()
	ipHash := hashWithPepper(s.challengePepper, clientIP(in.RemoteAddr))
	uaHash := hashWithPepper(s.challengePepper, strings.TrimSpace(in.UserAgent))
	recent, err := s.store.CountRecentChallenges(ctx, email, ipHash, now.Add(-RateLimitWindow))
	if err != nil {
		s.recordAuthFailure(ctx, "start_challenge", "store_error", "", challengeID)
		return StartEmailChallengeResult{}, err
	}
	if recent >= RateLimitThreshold-1 {
		s.recordChallengeStarted(ctx, challengeID, "rate_limited")
		return StartEmailChallengeResult{ChallengeID: challengeID, Accepted: true, RateLimited: true}, nil
	}
	token, err := s.tokenGenerator.GenerateToken()
	if err != nil {
		s.recordAuthFailure(ctx, "start_challenge", "token_generation_error", "", challengeID)
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
		s.recordAuthFailure(ctx, "start_challenge", "store_error", "", challengeID)
		return StartEmailChallengeResult{}, err
	}

	deliverySecretRef := "auth_challenge:" + challengeID
	s.deliverySecrets.PutDeliverySecret(deliverySecretRef, token)
	payload, err := jobs.BuildEmailDispatchPayload(map[string]string{
		"authChallengeId":   challengeID,
		"templateKey":       "auth_login_code",
		"locale":            localeOrDefault(in.AcceptLocale),
		"deliverySecretRef": deliverySecretRef,
		"dedupeKey":         hashWithPepper(s.challengePepper, "email:"+email),
	})
	if err != nil {
		s.recordAuthFailure(ctx, "start_challenge", "dispatch_payload_error", "", challengeID)
		return StartEmailChallengeResult{}, err
	}
	if err := s.dispatcher.Enqueue(ctx, payload); err != nil {
		s.recordAuthFailure(ctx, "start_challenge", "dispatch_error", "", challengeID)
		return StartEmailChallengeResult{}, err
	}
	s.recordChallengeStarted(ctx, challengeID, "accepted")
	return StartEmailChallengeResult{ChallengeID: challengeID, Accepted: true}, nil
}

func (s *PasswordlessService) VerifyEmailChallenge(ctx context.Context, in VerifyEmailChallengeInput) (VerifyEmailChallengeResult, error) {
	if s == nil || s.store == nil {
		return VerifyEmailChallengeResult{}, fmt.Errorf("passwordless service store is nil")
	}
	token := strings.TrimSpace(in.Token)
	if !isSixDigitCode(token) {
		s.recordAuthFailure(ctx, "verify_challenge", "invalid", "", "")
		return VerifyEmailChallengeResult{}, ErrChallengeInvalid
	}
	now := s.now().UTC()
	challenge, err := s.store.ConsumeChallenge(ctx, hashWithPepper(s.challengePepper, token), now)
	if err != nil {
		if errors.Is(err, ErrChallengeExpired) || errors.Is(err, ErrChallengeConsumed) || errors.Is(err, ErrChallengeInvalid) {
			s.recordAuthFailure(ctx, "verify_challenge", challengeFailureResult(err), "", "")
			return VerifyEmailChallengeResult{}, err
		}
		s.recordAuthFailure(ctx, "verify_challenge", "store_error", "", "")
		return VerifyEmailChallengeResult{}, err
	}
	user, err := s.store.FindUserByEmail(ctx, challenge.Email)
	if errors.Is(err, ErrUserNotFound) {
		user, err = s.store.CreateUserByEmail(ctx, challenge.Email, "", s.newID(), now)
	}
	if errors.Is(err, ErrEmailRegistered) {
		user, err = s.store.FindUserByEmail(ctx, challenge.Email)
	}
	if err != nil {
		if errors.Is(err, ErrChallengeInvalid) {
			s.recordAuthFailure(ctx, "verify_challenge", challengeFailureResult(err), "", challenge.ID)
			return VerifyEmailChallengeResult{}, err
		}
		s.recordAuthFailure(ctx, "verify_challenge", "store_error", "", challenge.ID)
		return VerifyEmailChallengeResult{}, err
	}
	sessionToken, err := s.sessionTokenGenerator.GenerateToken()
	if err != nil {
		s.recordAuthFailure(ctx, "verify_challenge", "token_generation_error", user.ID, challenge.ID)
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
		s.recordAuthFailure(ctx, "verify_challenge", "store_error", user.ID, challenge.ID)
		return VerifyEmailChallengeResult{}, err
	}
	s.recordSessionMinted(ctx, user.ID, challenge.ID)
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
		if errors.Is(err, sql.ErrNoRows) {
			return CurrentSession{}, ErrSessionRevoked
		}
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

func (s *PasswordlessService) CompleteProfile(ctx context.Context, in CompleteProfileInput) (UserContext, error) {
	if s == nil || s.store == nil {
		return UserContext{}, fmt.Errorf("passwordless service store is nil")
	}
	userID := strings.TrimSpace(in.UserID)
	displayName := normalizeDisplayName(in.DisplayName)
	if userID == "" {
		return UserContext{}, ErrSessionInvalid
	}
	if displayName == "" {
		return UserContext{}, fmt.Errorf("display name is required")
	}
	if !in.AcceptedTerms {
		return UserContext{}, fmt.Errorf("accepted terms is required")
	}
	return s.store.CompleteUserProfile(ctx, userID, displayName, s.now().UTC())
}

func (s *PasswordlessService) Logout(ctx context.Context, current CurrentSession) error {
	if s == nil || s.store == nil {
		return fmt.Errorf("passwordless service store is nil")
	}
	if current.SessionID == "" {
		return nil
	}
	if err := s.store.RevokeSession(ctx, current.SessionID, s.now().UTC()); err != nil {
		s.recordAuthFailure(ctx, "logout", "store_error", current.UserID, "")
		return err
	}
	s.recordLogout(ctx, current.UserID)
	return nil
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
		s.recordAuthFailure(ctx, "delete_handoff", "store_error", current.UserID, "")
		return PrivacyDeleteHandoff{}, err
	}
	s.recordDeleteHandoff(ctx, current.UserID, handoff)
	return handoff, nil
}

func (s *PasswordlessService) RuntimeConfigSessionResolver() func(*http.Request) bool {
	return func(r *http.Request) bool {
		if s == nil || r == nil {
			return false
		}
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil || cookie.Value == "" {
			return false
		}
		current, err := s.ResolveSession(r.Context(), cookie.Value)
		if err != nil {
			return false
		}
		user, err := s.CurrentUser(r.Context(), current.UserID)
		if err != nil {
			return false
		}
		return user.AnalyticsOptIn
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeDisplayName(name string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
}

func isSixDigitCode(token string) bool {
	if len(token) != 6 {
		return false
	}
	for _, ch := range token {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
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

func (s *PasswordlessService) recordChallengeStarted(ctx context.Context, challengeID string, result string) {
	s.metrics.recordChallengeStarted(result)
	s.recordAuthAudit(ctx, AuthAuditEvent{
		Action:      AuthAuditActionChallengeStarted,
		Result:      authMetricResult(result),
		ChallengeID: challengeID,
		TraceID:     AuthTraceIDFromContext(ctx),
	})
}

func (s *PasswordlessService) recordSessionMinted(ctx context.Context, userID string, challengeID string) {
	s.metrics.recordSessionMinted("success")
	s.recordAuthAudit(ctx, AuthAuditEvent{
		Action:      AuthAuditActionSessionMinted,
		Result:      "success",
		UserIDHash:  s.auditUserIDHash(userID),
		ChallengeID: challengeID,
		TraceID:     AuthTraceIDFromContext(ctx),
	})
}

func (s *PasswordlessService) recordLogout(ctx context.Context, userID string) {
	s.metrics.recordLogout("success")
	s.recordAuthAudit(ctx, AuthAuditEvent{
		Action:     AuthAuditActionLogout,
		Result:     "success",
		UserIDHash: s.auditUserIDHash(userID),
		TraceID:    AuthTraceIDFromContext(ctx),
	})
}

func (s *PasswordlessService) recordDeleteHandoff(ctx context.Context, userID string, handoff PrivacyDeleteHandoff) {
	s.metrics.recordDeleteHandoff("success")
	s.recordAuthAudit(ctx, AuthAuditEvent{
		Action:           AuthAuditActionDeleteHandoff,
		Result:           "success",
		UserIDHash:       s.auditUserIDHash(userID),
		PrivacyRequestID: handoff.PrivacyRequestID,
		JobID:            handoff.JobID,
		TraceID:          AuthTraceIDFromContext(ctx),
	})
}

func (s *PasswordlessService) recordAuthFailure(ctx context.Context, operation string, result string, userID string, challengeID string) {
	result = authMetricResult(result)
	operation = authMetricOperation(operation)
	s.metrics.recordFailure(operation, result)
	s.recordAuthAudit(ctx, AuthAuditEvent{
		Action:      AuthAuditActionFailure,
		Operation:   operation,
		Result:      result,
		UserIDHash:  s.auditUserIDHash(userID),
		ChallengeID: challengeID,
		TraceID:     AuthTraceIDFromContext(ctx),
	})
}

func (s *PasswordlessService) recordAuthAudit(ctx context.Context, event AuthAuditEvent) {
	if s == nil || s.audit == nil {
		return
	}
	_ = s.audit.RecordAuthAuditEvent(ctx, event)
}

func (s *PasswordlessService) auditUserIDHash(userID string) string {
	if strings.TrimSpace(userID) == "" {
		return ""
	}
	return hashWithPepper(s.challengePepper, "user:"+userID)
}

func challengeFailureResult(err error) string {
	switch {
	case errors.Is(err, ErrChallengeExpired):
		return "expired"
	case errors.Is(err, ErrChallengeConsumed):
		return "consumed"
	case errors.Is(err, ErrEmailRegistered):
		return "email_registered"
	case errors.Is(err, ErrUserNotFound):
		return "user_not_found"
	case errors.Is(err, ErrChallengeInvalid):
		return "invalid"
	default:
		return "error"
	}
}
