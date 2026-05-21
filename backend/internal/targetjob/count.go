package targetjob

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// CountTargetJobsForUser is the cross-owner internal API consumed by
// backend-jobs-recommendations/001 BuildJobMatchProfile (spec D-18 sources
// aggregation). It returns the count of non-deleted target_jobs rows owned
// by the supplied userID. Read-only; cross-user isolated.
func CountTargetJobsForUser(ctx context.Context, db *sql.DB, userID string) (int, error) {
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
		"SELECT COUNT(*) FROM target_jobs WHERE user_id = $1 AND deleted_at IS NULL",
		uid,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("targetjob: count for user: %w", err)
	}
	return n, nil
}

var (
	ErrCounterDBRequired     = errors.New("targetjob: counter requires a non-nil *sql.DB")
	ErrCounterUserIDRequired = errors.New("targetjob: counter requires a non-empty userID")
)
