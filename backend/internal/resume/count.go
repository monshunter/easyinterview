package resume

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// CountResumesForUser is the cross-owner internal API consumed by
// backend-jobs-recommendations/001 BuildJobMatchProfile (spec D-18 sources
// aggregation). It returns the count of non-deleted resume_assets rows
// owned by the supplied userID. Read-only; does not mutate any state,
// does not write audit_events, does not call resume parse jobs, and is
// fully isolated per-user (cross-user reads must pass a different userID).
//
// Errors:
//   - ErrCounterDBRequired if db is nil
//   - ErrCounterUserIDRequired if userID is empty/blank
//   - underlying *sql error on query failure
func CountResumesForUser(ctx context.Context, db *sql.DB, userID string) (int, error) {
	if db == nil {
		return 0, ErrCounterDBRequired
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, ErrCounterUserIDRequired
	}
	var n int
	err := db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM resume_assets WHERE user_id = $1 AND deleted_at IS NULL",
		uid,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("resume: count for user: %w", err)
	}
	return n, nil
}

var (
	// ErrCounterDBRequired is returned when CountResumesForUser is called
	// with a nil *sql.DB.
	ErrCounterDBRequired = errors.New("resume: counter requires a non-nil *sql.DB")
	// ErrCounterUserIDRequired is returned when CountResumesForUser is
	// called with an empty userID.
	ErrCounterUserIDRequired = errors.New("resume: counter requires a non-empty userID")
)
