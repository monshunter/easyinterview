package runner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrPrivacyRequestNotFound = errors.New("privacy delete request not found")

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
	var userID string
	err := s.db.QueryRowContext(ctx, `
select user_id
from privacy_requests
where id = $1 and request_type = 'delete'`, privacyRequestID).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrPrivacyRequestNotFound
	}
	if err != nil {
		return "", fmt.Errorf("lookup privacy delete request user: %w", err)
	}
	return userID, nil
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

func (s *SQLStore) MarkDeleteRequestCompleted(ctx context.Context, privacyRequestID string, deletedFileCount int, now time.Time) error {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return s.updateRequest(ctx, `
update privacy_requests
set status = 'completed',
    completed_at = $2,
    error_code = null,
    metadata = metadata || jsonb_build_object('deletedFileCount', $3::int)
where id = $1 and request_type = 'delete'`, privacyRequestID, now, deletedFileCount)
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
