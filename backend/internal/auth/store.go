package auth

import (
	"context"
	"database/sql"
	"fmt"
	"time"
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

type ChallengeRecord struct {
	ID            string
	UserID        string
	Email         string
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
	UILanguage                string
	PreferredPracticeLanguage string
	AnalyticsOptIn            bool
}

// Store is the P0 passwordless session persistence surface. It intentionally
// omits external_identities methods; that table is only a P1 SSO slot.
type Store interface {
	CreateChallenge(context.Context, ChallengeRecord) error
	CreateSession(context.Context, SessionRecord) error
	GetUserContext(context.Context, string) (UserContext, error)
	TouchSession(context.Context, string, time.Time, time.Time) error
}

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
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
  challenge_token_hash,
  purpose,
  ip_hash,
  user_agent_hash,
  expires_at,
  created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rec.ID,
		userID,
		rec.Email,
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

func (s *SQLStore) GetUserContext(ctx context.Context, userID string) (UserContext, error) {
	if s == nil || s.db == nil {
		return UserContext{}, fmt.Errorf("auth store db is nil")
	}
	var out UserContext
	var displayName sql.NullString
	err := s.db.QueryRowContext(ctx, `
select
  u.id,
  u.email,
  u.display_name,
  us.ui_language,
  us.preferred_practice_language,
  us.analytics_opt_in
from users u join user_settings us on us.user_id = u.id
where u.id = $1 and u.deleted_at is null`,
		userID,
	).Scan(
		&out.ID,
		&out.Email,
		&displayName,
		&out.UILanguage,
		&out.PreferredPracticeLanguage,
		&out.AnalyticsOptIn,
	)
	if err != nil {
		return UserContext{}, fmt.Errorf("select user context: %w", err)
	}
	out.DisplayName = displayName.String
	return out, nil
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

var _ Store = (*SQLStore)(nil)
