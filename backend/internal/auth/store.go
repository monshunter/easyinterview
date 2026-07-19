package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/idx"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
)

type ChallengePurpose string

const (
	ChallengePurposeLogin  ChallengePurpose = "login"
	ChallengePurposeSignup ChallengePurpose = "signup"
)

type SessionStatus string

const (
	SessionStatusActive  SessionStatus = "active"
	SessionStatusRevoked SessionStatus = "revoked"
	SessionStatusExpired SessionStatus = "expired"
)

var (
	ErrChallengeInvalid  = errors.New("auth challenge invalid")
	ErrChallengeExpired  = errors.New("auth challenge expired")
	ErrChallengeConsumed = errors.New("auth challenge consumed")
	ErrSessionInvalid    = errors.New("auth session invalid")
	ErrSessionExpired    = errors.New("auth session expired")
	ErrSessionRevoked    = errors.New("auth session revoked")
	ErrEmailRegistered   = errors.New("auth email already registered")
)

type ChallengeRecord struct {
	ID            string
	UserID        string
	Email         string
	DisplayName   string
	TokenHash     string
	Purpose       ChallengePurpose
	IPHash        string
	UserAgentHash string
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

type SessionRecord struct {
	ID            string
	UserID        string
	SessionHash   string
	Status        SessionStatus
	IPHash        string
	UserAgentHash string
	ExpiresAt     time.Time
	RevokedAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UserContext struct {
	ID                        string
	Email                     string
	DisplayName               string
	ProfileCompletionRequired bool
	DisplayPreferences        AccountDisplayPreferences
}

type AccountTheme string

const (
	AccountThemeOcean AccountTheme = "ocean"
	AccountThemePlum  AccountTheme = "plum"
)

type CustomAccent struct {
	H float64
	C float64
}

type AccountDisplayPreferences struct {
	Theme        AccountTheme
	CustomAccent *CustomAccent
}

type UpdateUserContextInput struct {
	DisplayName        *string
	AcceptedTerms      *bool
	DisplayPreferences *AccountDisplayPreferences
}

type userContextUpdater interface {
	UpdateUserContext(context.Context, string, UpdateUserContextInput, time.Time) (UserContext, error)
}

type analyticsOptInStore interface {
	GetAnalyticsOptIn(context.Context, string) (bool, error)
}

// Store is the P0 email-code session persistence surface. It intentionally
// omits external_identities methods; that table is only a P1 SSO slot.
type Store interface {
	CountRecentChallenges(context.Context, string, string, time.Time) (int, error)
	CreateChallenge(context.Context, ChallengeRecord) error
	ConsumeChallenge(context.Context, string, time.Time) (ChallengeRecord, error)
	CreateUserByEmail(context.Context, string, string, string, time.Time) (UserContext, error)
	FindUserByEmail(context.Context, string) (UserContext, error)
	CompleteUserProfile(context.Context, string, string, time.Time) (UserContext, error)
	CreateSession(context.Context, SessionRecord) error
	GetSessionByHash(context.Context, string, time.Time) (SessionRecord, error)
	GetUserContext(context.Context, string) (UserContext, error)
	TouchSession(context.Context, string, time.Time, time.Time) error
	RevokeSession(context.Context, string, time.Time) error
	CreatePrivacyDeleteHandoff(context.Context, string, string, string, string, time.Time) (PrivacyDeleteHandoff, error)
}

type PrivacyDeleteHandoff struct {
	PrivacyRequestID string
	JobID            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

type sqlPrivacyDeleteExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func (s *SQLStore) CountRecentChallenges(ctx context.Context, email string, ipHash string, since time.Time) (int, error) {
	if s == nil || s.db == nil {
		return 0, fmt.Errorf("auth store db is nil")
	}
	var count int
	err := s.db.QueryRowContext(ctx, `
	select count(*)
	from auth_challenges
	where created_at >= $1
	  and (email = $2 or ip_hash = $3)`,
		since,
		email,
		ipHash,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count recent auth challenges: %w", err)
	}
	return count, nil
}

func (s *SQLStore) CreateChallenge(ctx context.Context, rec ChallengeRecord) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("auth store db is nil")
	}
	var userID any
	if rec.UserID != "" {
		userID = rec.UserID
	}
	_, err := s.db.ExecContext(ctx, `
insert into auth_challenges (
  id,
  user_id,
  email,
  display_name,
  challenge_token_hash,
  purpose,
  ip_hash,
  user_agent_hash,
  expires_at,
  created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		rec.ID,
		userID,
		rec.Email,
		nullString(rec.DisplayName),
		rec.TokenHash,
		string(rec.Purpose),
		rec.IPHash,
		rec.UserAgentHash,
		rec.ExpiresAt,
		rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert auth challenge: %w", err)
	}
	return nil
}

func (s *SQLStore) ConsumeChallenge(ctx context.Context, tokenHash string, now time.Time) (ChallengeRecord, error) {
	if s == nil || s.db == nil {
		return ChallengeRecord{}, fmt.Errorf("auth store db is nil")
	}
	var rec ChallengeRecord
	var userID sql.NullString
	var displayName sql.NullString
	var ipHash sql.NullString
	var uaHash sql.NullString
	var purpose string
	err := s.db.QueryRowContext(ctx, `
update auth_challenges
set status = 'consumed', consumed_at = $2
where challenge_token_hash = $1
  and status = 'pending'
  and expires_at > $2
returning id, user_id, email, display_name, purpose, ip_hash, user_agent_hash, expires_at, created_at`,
		tokenHash,
		now,
	).Scan(
		&rec.ID,
		&userID,
		&rec.Email,
		&displayName,
		&purpose,
		&ipHash,
		&uaHash,
		&rec.ExpiresAt,
		&rec.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ChallengeRecord{}, ErrChallengeInvalid
	}
	if err != nil {
		return ChallengeRecord{}, fmt.Errorf("consume auth challenge: %w", err)
	}
	rec.UserID = userID.String
	rec.DisplayName = displayName.String
	rec.Purpose = ChallengePurpose(purpose)
	rec.IPHash = ipHash.String
	rec.UserAgentHash = uaHash.String
	return rec, nil
}

func (s *SQLStore) CreateUserByEmail(ctx context.Context, email string, displayName string, userID string, now time.Time) (UserContext, error) {
	if s == nil || s.db == nil {
		return UserContext{}, fmt.Errorf("auth store db is nil")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return UserContext{}, fmt.Errorf("begin find or create user: %w", err)
	}
	defer tx.Rollback()
	var id string
	profileCompletedAt := any(nil)
	termsAcceptedAt := any(nil)
	if strings.TrimSpace(displayName) != "" {
		profileCompletedAt = now
		termsAcceptedAt = now
	}
	err = tx.QueryRowContext(ctx, `
insert into users (id, email, display_name, profile_completed_at, terms_accepted_at, updated_at)
values ($1, $2, $3, $4, $5, $6)
on conflict (email) do nothing
returning id`,
		userID,
		email,
		nullString(displayName),
		profileCompletedAt,
		termsAcceptedAt,
		now,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return UserContext{}, ErrEmailRegistered
	}
	if err != nil {
		return UserContext{}, fmt.Errorf("insert user: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "insert into user_settings (user_id) values ($1) on conflict (user_id) do nothing", id); err != nil {
		return UserContext{}, fmt.Errorf("ensure user settings: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return UserContext{}, fmt.Errorf("commit find or create user: %w", err)
	}
	return s.GetUserContext(ctx, id)
}

func (s *SQLStore) FindUserByEmail(ctx context.Context, email string) (UserContext, error) {
	if s == nil || s.db == nil {
		return UserContext{}, fmt.Errorf("auth store db is nil")
	}
	out, err := s.getUserContext(ctx, "u.email = $1", email)
	if errors.Is(err, sql.ErrNoRows) {
		return UserContext{}, ErrUserNotFound
	}
	if err != nil {
		return UserContext{}, err
	}
	return out, nil
}

func (s *SQLStore) CompleteUserProfile(ctx context.Context, userID string, displayName string, now time.Time) (UserContext, error) {
	if s == nil || s.db == nil {
		return UserContext{}, fmt.Errorf("auth store db is nil")
	}
	result, err := s.db.ExecContext(ctx, `
update users
set display_name = $2,
    terms_accepted_at = coalesce(terms_accepted_at, $3),
    profile_completed_at = coalesce(profile_completed_at, $3),
    updated_at = $3
where id = $1 and deleted_at is null`,
		userID,
		nullString(displayName),
		now,
	)
	if err != nil {
		return UserContext{}, fmt.Errorf("complete user profile: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return UserContext{}, fmt.Errorf("complete user profile rows affected: %w", err)
	}
	if rows == 0 {
		return UserContext{}, ErrUserNotFound
	}
	return s.GetUserContext(ctx, userID)
}

func (s *SQLStore) UpdateUserContext(ctx context.Context, userID string, in UpdateUserContextInput, now time.Time) (UserContext, error) {
	if s == nil || s.db == nil {
		return UserContext{}, fmt.Errorf("auth store db is nil")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return UserContext{}, fmt.Errorf("begin update user context: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if in.DisplayName != nil && in.AcceptedTerms != nil {
		result, execErr := tx.ExecContext(ctx, `
update users
set display_name = $2,
    terms_accepted_at = coalesce(terms_accepted_at, $3),
    profile_completed_at = coalesce(profile_completed_at, $3),
    updated_at = $3
where id = $1 and deleted_at is null`, userID, nullString(*in.DisplayName), now)
		if execErr != nil {
			return UserContext{}, fmt.Errorf("update user profile: %w", execErr)
		}
		rows, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return UserContext{}, fmt.Errorf("update user profile rows affected: %w", rowsErr)
		}
		if rows == 0 {
			return UserContext{}, ErrUserNotFound
		}
	}

	if in.DisplayPreferences != nil {
		var hue any
		var chroma any
		if in.DisplayPreferences.CustomAccent != nil {
			hue = in.DisplayPreferences.CustomAccent.H
			chroma = in.DisplayPreferences.CustomAccent.C
		}
		result, execErr := tx.ExecContext(ctx, `
update user_settings
set theme = $2,
    custom_accent_hue = $3,
    custom_accent_chroma = $4,
    updated_at = $5
where user_id = $1`, userID, string(in.DisplayPreferences.Theme), hue, chroma, now)
		if execErr != nil {
			return UserContext{}, fmt.Errorf("update account theme: %w", execErr)
		}
		rows, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return UserContext{}, fmt.Errorf("update account theme rows affected: %w", rowsErr)
		}
		if rows == 0 {
			return UserContext{}, ErrUserNotFound
		}
	}

	out, err := queryUserContext(ctx, tx, "u.id = $1", userID)
	if err != nil {
		return UserContext{}, fmt.Errorf("select updated user context: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return UserContext{}, fmt.Errorf("commit update user context: %w", err)
	}
	return out, nil
}

func (s *SQLStore) CreateSession(ctx context.Context, rec SessionRecord) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("auth store db is nil")
	}
	_, err := s.db.ExecContext(ctx, `
insert into sessions (
  id,
  user_id,
  session_hash,
  status,
  ip_hash,
  user_agent_hash,
  expires_at,
  created_at,
  updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rec.ID,
		rec.UserID,
		rec.SessionHash,
		string(rec.Status),
		rec.IPHash,
		rec.UserAgentHash,
		rec.ExpiresAt,
		rec.CreatedAt,
		rec.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (s *SQLStore) GetSessionByHash(ctx context.Context, sessionHash string, _ time.Time) (SessionRecord, error) {
	if s == nil || s.db == nil {
		return SessionRecord{}, fmt.Errorf("auth store db is nil")
	}
	var rec SessionRecord
	var status string
	var ipHash sql.NullString
	var uaHash sql.NullString
	var revokedAt sql.NullTime
	err := s.db.QueryRowContext(ctx, `
select id, user_id, session_hash, status, ip_hash, user_agent_hash, expires_at, revoked_at, created_at, updated_at
from sessions
where session_hash = $1`,
		sessionHash,
	).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.SessionHash,
		&status,
		&ipHash,
		&uaHash,
		&rec.ExpiresAt,
		&revokedAt,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return SessionRecord{}, ErrSessionInvalid
	}
	if err != nil {
		return SessionRecord{}, fmt.Errorf("select session: %w", err)
	}
	rec.Status = SessionStatus(status)
	rec.IPHash = ipHash.String
	rec.UserAgentHash = uaHash.String
	if revokedAt.Valid {
		rec.RevokedAt = revokedAt.Time
	}
	return rec, nil
}

func (s *SQLStore) GetUserContext(ctx context.Context, userID string) (UserContext, error) {
	if s == nil || s.db == nil {
		return UserContext{}, fmt.Errorf("auth store db is nil")
	}
	out, err := s.getUserContext(ctx, "u.id = $1", userID)
	if err != nil {
		return UserContext{}, fmt.Errorf("select user context: %w", err)
	}
	return out, nil
}

func (s *SQLStore) getUserContext(ctx context.Context, predicate string, arg any) (UserContext, error) {
	return queryUserContext(ctx, s.db, predicate, arg)
}

type queryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func queryUserContext(ctx context.Context, querier queryRower, predicate string, arg any) (UserContext, error) {
	var out UserContext
	var displayName sql.NullString
	var profileCompletedAt sql.NullTime
	var termsAcceptedAt sql.NullTime
	var theme string
	var customHue sql.NullFloat64
	var customChroma sql.NullFloat64
	err := querier.QueryRowContext(ctx, `
select
  u.id,
  u.email,
  u.display_name,
  u.profile_completed_at,
  u.terms_accepted_at,
  coalesce(us.theme, 'ocean'),
  us.custom_accent_hue,
  us.custom_accent_chroma
from users u
left join user_settings us on us.user_id = u.id
where `+predicate+` and u.deleted_at is null`,
		arg,
	).Scan(
		&out.ID,
		&out.Email,
		&displayName,
		&profileCompletedAt,
		&termsAcceptedAt,
		&theme,
		&customHue,
		&customChroma,
	)
	if err != nil {
		return UserContext{}, err
	}
	out.DisplayName = displayName.String
	out.ProfileCompletionRequired = strings.TrimSpace(displayName.String) == "" || !profileCompletedAt.Valid || !termsAcceptedAt.Valid
	out.DisplayPreferences = AccountDisplayPreferences{Theme: AccountTheme(theme)}
	if customHue.Valid && customChroma.Valid {
		out.DisplayPreferences.CustomAccent = &CustomAccent{H: customHue.Float64, C: customChroma.Float64}
	}
	return out, nil
}

func (s *SQLStore) GetAnalyticsOptIn(ctx context.Context, userID string) (bool, error) {
	if s == nil || s.db == nil {
		return false, fmt.Errorf("auth store db is nil")
	}
	var optIn bool
	if err := s.db.QueryRowContext(ctx, `
select analytics_opt_in
from user_settings
where user_id = $1`, userID).Scan(&optIn); err != nil {
		return false, fmt.Errorf("select analytics opt-in: %w", err)
	}
	return optIn, nil
}

func nullString(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func (s *SQLStore) TouchSession(ctx context.Context, sessionID string, now time.Time, expiresAt time.Time) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("auth store db is nil")
	}
	result, err := s.db.ExecContext(ctx,
		"update sessions set updated_at = $1, expires_at = $2 where id = $3 and status = 'active'",
		now,
		expiresAt,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("touch session rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLStore) RevokeSession(ctx context.Context, sessionID string, now time.Time) error {
	if s == nil || s.db == nil {
		return fmt.Errorf("auth store db is nil")
	}
	_, err := s.db.ExecContext(ctx,
		"update sessions set status = 'revoked', revoked_at = $1, updated_at = $1 where id = $2 and status = 'active'",
		now,
		sessionID,
	)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func (s *SQLStore) CreatePrivacyDeleteHandoff(ctx context.Context, userID string, idempotencyKey string, privacyRequestID string, jobID string, now time.Time) (PrivacyDeleteHandoff, error) {
	if s == nil || s.db == nil {
		return PrivacyDeleteHandoff{}, fmt.Errorf("auth store db is nil")
	}
	if idempotencyKey != "" {
		idempotencyKey = privacyDeleteDedupeKey(userID, idempotencyKey)
		var existing PrivacyDeleteHandoff
		err := s.db.QueryRowContext(ctx, `
	select id, resource_id, created_at, updated_at
	from async_jobs
	where job_type = $2
	  and dedupe_key = $1
	  and status in ('queued', 'running')
	limit 1`,
			idempotencyKey,
			string(jobs.JobTypePrivacyDelete),
		).Scan(&existing.JobID, &existing.PrivacyRequestID, &existing.CreatedAt, &existing.UpdatedAt)
		if err == nil {
			if err := s.softDeleteUserForPrivacy(ctx, userID, now); err != nil {
				return PrivacyDeleteHandoff{}, err
			}
			return existing, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return PrivacyDeleteHandoff{}, fmt.Errorf("select privacy delete handoff: %w", err)
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PrivacyDeleteHandoff{}, fmt.Errorf("begin privacy delete handoff: %w", err)
	}
	defer tx.Rollback()
	if err := softDeleteUserForPrivacy(ctx, tx, userID, now); err != nil {
		return PrivacyDeleteHandoff{}, err
	}
	if _, err := tx.ExecContext(ctx, `
insert into privacy_requests (id, user_id, request_type, status, requested_at)
values ($1, $2, 'delete', 'queued', $3)`,
		privacyRequestID,
		userID,
		now,
	); err != nil {
		return PrivacyDeleteHandoff{}, fmt.Errorf("insert privacy request: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
	insert into async_jobs (id, job_type, resource_type, resource_id, dedupe_key, status, payload, created_at, updated_at)
	values ($1, $2, 'privacy_request', $3, $4, 'queued', '{}'::jsonb, $5, $5)`,
		jobID,
		string(jobs.JobTypePrivacyDelete),
		privacyRequestID,
		idempotencyKey,
		now,
	); err != nil {
		return PrivacyDeleteHandoff{}, fmt.Errorf("insert privacy delete job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return PrivacyDeleteHandoff{}, fmt.Errorf("commit privacy delete handoff: %w", err)
	}
	return PrivacyDeleteHandoff{PrivacyRequestID: privacyRequestID, JobID: jobID, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *SQLStore) softDeleteUserForPrivacy(ctx context.Context, userID string, now time.Time) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin privacy delete user soft-delete: %w", err)
	}
	defer tx.Rollback()
	if err := softDeleteUserForPrivacy(ctx, tx, userID, now); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit privacy delete user soft-delete: %w", err)
	}
	return nil
}

func softDeleteUserForPrivacy(ctx context.Context, exec sqlPrivacyDeleteExecutor, userID string, now time.Time) error {
	res, err := exec.ExecContext(ctx, `
update users
set status = 'deleted',
    deleted_at = coalesce(deleted_at, $1),
    updated_at = $1
where id = $2`, now, userID)
	if err != nil {
		return fmt.Errorf("soft-delete privacy user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("soft-delete privacy user rows affected: %w", err)
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	if _, err := exec.ExecContext(ctx, `
update sessions
set status = 'revoked',
    revoked_at = coalesce(revoked_at, $1),
    updated_at = $1
where user_id = $2 and status = 'active'`, now, userID); err != nil {
		return fmt.Errorf("revoke privacy user sessions: %w", err)
	}
	return nil
}

func privacyDeleteDedupeKey(userID string, idempotencyKey string) string {
	return string(jobs.JobTypePrivacyDelete) + ":" + hashWithPepper(strings.TrimSpace(userID), strings.TrimSpace(idempotencyKey))
}

var _ Store = (*SQLStore)(nil)

func NewID() string {
	return idx.NewID()
}
