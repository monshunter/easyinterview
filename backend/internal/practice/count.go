package practice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// CountPracticeSessionsForUser is the cross-owner internal API consumed by
// backend-jobs-recommendations/001 BuildJobMatchProfile (spec D-18 sources
// aggregation). Read-only; cross-user isolated.
func CountPracticeSessionsForUser(ctx context.Context, db *sql.DB, userID string) (int, error) {
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
		"SELECT COUNT(*) FROM practice_sessions WHERE user_id = $1",
		uid,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("practice: count for user: %w", err)
	}
	return n, nil
}

var (
	ErrCounterDBRequired     = errors.New("practice: counter requires a non-nil *sql.DB")
	ErrCounterUserIDRequired = errors.New("practice: counter requires a non-empty userID")
)
