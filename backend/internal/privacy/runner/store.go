package runner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrPrivacyRequestNotFound        = errors.New("privacy delete request not found")
	ErrPrivacyDeleteAlreadyCompleted = errors.New("privacy delete request already completed")
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) *SQLStore {
	return &SQLStore{db: db}
}

func (s *SQLStore) LookupDeleteRequestUser(ctx context.Context, privacyRequestID string) (string, error) {
	if err := s.checkDB(); err != nil {
		return "", err
	}
	var userID sql.NullString
	var status string
	err := s.db.QueryRowContext(ctx, `
	select user_id, status
	from privacy_requests
	where id = $1 and request_type = 'delete'`, privacyRequestID).Scan(&userID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrPrivacyRequestNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup privacy delete request user: %w", err)
	}
	if !userID.Valid {
		if status == "completed" {
			return "", ErrPrivacyDeleteAlreadyCompleted
		}
		return "", nil
	}
	return userID.String, nil
}

func (s *SQLStore) MarkDeleteRequestProcessing(ctx context.Context, privacyRequestID string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return s.updateRequest(ctx, `
update privacy_requests
set status = 'processing',
    metadata = metadata || jsonb_build_object('processingAt', to_jsonb($2::text))
where id = $1 and request_type = 'delete'`, privacyRequestID, now.Format(time.RFC3339Nano))
}

func (s *SQLStore) MarkDeleteRequestCompleted(ctx context.Context, privacyRequestID string, userID string, deletedFileCount int, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if err := s.checkDB(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin privacy delete completion: %w", err)
	}
	defer tx.Rollback()

	var email sql.NullString
	if err := tx.QueryRowContext(ctx, `select email from users where id = $1`, userID).Scan(&email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrPrivacyRequestNotFound
		}
		return fmt.Errorf("lookup privacy delete user identity: %w", err)
	}

	res, err := tx.ExecContext(ctx, `
update privacy_requests
set status = 'completed',
    completed_at = $2,
    error_code = null,
    user_id = null,
    metadata = metadata || jsonb_build_object('deletedFileCount', $3::int)
where id = $1 and request_type = 'delete' and user_id = $4`, privacyRequestID, now, deletedFileCount, userID)
	if err != nil {
		return fmt.Errorf("complete privacy delete request: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete privacy delete request rows affected: %w", err)
	}
	if rows == 0 {
		return ErrPrivacyRequestNotFound
	}
	if err := hardDeleteAccountIdentity(ctx, tx, userID, email.String); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit privacy delete completion: %w", err)
	}
	return nil
}

func (s *SQLStore) MarkDeleteRequestFailed(ctx context.Context, privacyRequestID string, errorCode string, errorMessage string, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return s.updateRequest(ctx, `
update privacy_requests
set status = 'failed',
    completed_at = $2,
    error_code = $3,
    metadata = metadata || jsonb_build_object('errorMessage', $4)
where id = $1 and request_type = 'delete'`, privacyRequestID, now, errorCode, errorMessage)
}

func (s *SQLStore) updateRequest(ctx context.Context, query string, args ...any) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update privacy delete request: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("privacy delete request rows affected: %w", err)
	}
	if rows == 0 {
		return ErrPrivacyRequestNotFound
	}
	return nil
}

func (s *SQLStore) checkDB() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("privacy runner SQL store is not configured")
	}
	return nil
}

type accountIdentityDeleteExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func hardDeleteAccountIdentity(ctx context.Context, exec accountIdentityDeleteExecutor, userID string, email string) error {
	if _, err := exec.ExecContext(ctx, `
delete from resume_version_suggestions
where resume_version_id in (select id from resume_versions where user_id = $1)`, userID); err != nil {
		return fmt.Errorf("delete resume version suggestions for privacy user: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `
update resume_versions
set parent_version_id = null
where user_id = $1 and parent_version_id is not null`, userID); err != nil {
		return fmt.Errorf("clear resume version parent links for privacy user: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `delete from resume_versions where user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete resume versions for privacy user: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `delete from auth_challenges where user_id = $1 or email = $2`, userID, email); err != nil {
		return fmt.Errorf("delete auth challenges for privacy user: %w", err)
	}
	res, err := exec.ExecContext(ctx, `delete from users where id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete privacy user identity: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete privacy user identity rows affected: %w", err)
	}
	if rows == 0 {
		return ErrPrivacyRequestNotFound
	}
	return nil
}
