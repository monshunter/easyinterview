package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
)

// WatchlistRecord is the watchlist + joined recommendation projection
// used to derive tone (Q-4) in the handler.
type WatchlistRecord struct {
	ID                string
	UserID            string
	LinkedJobMatchID  string
	Label             *string
	ChangeNote        *string
	AddedAt           time.Time
	LinkedTitle       string
	LinkedCompany     string
	LinkedScore       int
	LinkedDismissedAt *time.Time
}

// ListWatchlistByUser returns the user's watchlist joined to
// jd_match_recommendations so the caller can derive tone in process.
func (r *Repository) ListWatchlistByUser(ctx context.Context, userID string) ([]WatchlistRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return nil, jdmatch.ErrUserIDRequired
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT w.id, w.user_id, w.linked_job_match_id, w.label, w.change_note, w.added_at,
		        r.title, r.company, r.score, r.dismissed_at
		FROM watchlist_items w
		JOIN jd_match_recommendations r ON r.id = w.linked_job_match_id AND r.user_id = w.user_id
		WHERE w.user_id = $1
		ORDER BY w.added_at DESC, w.id DESC`,
		uid,
	)
	if err != nil {
		return nil, fmt.Errorf("jdmatch store: list watchlist: %w", err)
	}
	defer rows.Close()
	out := make([]WatchlistRecord, 0)
	for rows.Next() {
		var (
			rec        WatchlistRecord
			label      sql.NullString
			changeNote sql.NullString
			dismissed  sql.NullTime
		)
		if err := rows.Scan(&rec.ID, &rec.UserID, &rec.LinkedJobMatchID, &label, &changeNote, &rec.AddedAt,
			&rec.LinkedTitle, &rec.LinkedCompany, &rec.LinkedScore, &dismissed); err != nil {
			return nil, fmt.Errorf("jdmatch store: scan watchlist: %w", err)
		}
		if label.Valid {
			v := label.String
			rec.Label = &v
		}
		if changeNote.Valid {
			v := changeNote.String
			rec.ChangeNote = &v
		}
		if dismissed.Valid {
			t := dismissed.Time
			rec.LinkedDismissedAt = &t
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// AddWatchlistItemInput captures the addToWatchlist parameters.
type AddWatchlistItemInput struct {
	ID               string
	UserID           string
	LinkedJobMatchID string
	Label            *string
}

// AddWatchlistItem inserts a new watchlist row. On UNIQUE (user_id,
// linked_job_match_id) conflict the existing row is returned so the
// handler can satisfy spec C-6 (duplicate add returns the first row,
// no new row is created).
func (r *Repository) AddWatchlistItem(ctx context.Context, in AddWatchlistItemInput) (WatchlistRecord, error) {
	if r == nil || r.db == nil {
		return WatchlistRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.LinkedJobMatchID) == "" {
		return WatchlistRecord{}, jdmatch.ErrUserIDRequired
	}
	now := r.now()
	// Verify the linked recommendation belongs to this user; cross-
	// user references map to ErrNotFound so the handler returns 404.
	var recOwner string
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id FROM jd_match_recommendations WHERE id = $1 AND deleted_at IS NULL`,
		in.LinkedJobMatchID,
	).Scan(&recOwner)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && recOwner != in.UserID) {
		return WatchlistRecord{}, jdmatch.ErrNotFound
	}
	if err != nil {
		return WatchlistRecord{}, fmt.Errorf("jdmatch store: watchlist ownership probe: %w", err)
	}
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO watchlist_items (id, user_id, linked_job_match_id, label, tone, added_at)
		VALUES ($1, $2, $3, $4, 'ok', $5)
		ON CONFLICT (user_id, linked_job_match_id) DO NOTHING`,
		in.ID, in.UserID, in.LinkedJobMatchID, in.Label, now,
	)
	if err != nil {
		return WatchlistRecord{}, fmt.Errorf("jdmatch store: insert watchlist: %w", err)
	}
	// Re-read joined row so the caller always sees title / company /
	// score from the recommendation regardless of insert-vs-conflict
	// path.
	row := r.db.QueryRowContext(ctx,
		`SELECT w.id, w.user_id, w.linked_job_match_id, w.label, w.change_note, w.added_at,
		        r.title, r.company, r.score, r.dismissed_at
		FROM watchlist_items w
		JOIN jd_match_recommendations r ON r.id = w.linked_job_match_id AND r.user_id = w.user_id
		WHERE w.user_id = $1 AND w.linked_job_match_id = $2`,
		in.UserID, in.LinkedJobMatchID,
	)
	rec, scanErr := scanWatchlistRow(row)
	if errors.Is(scanErr, sql.ErrNoRows) {
		return WatchlistRecord{}, jdmatch.ErrNotFound
	}
	if scanErr != nil {
		return WatchlistRecord{}, fmt.Errorf("jdmatch store: read inserted watchlist row: %w", scanErr)
	}
	return rec, nil
}

// RemoveWatchlistItem deletes the watchlist row for (user_id,
// linked_job_match_id). Returns the count removed (0 / 1).
func (r *Repository) RemoveWatchlistItem(ctx context.Context, userID, linkedJobMatchID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(linkedJobMatchID) == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM watchlist_items WHERE user_id = $1 AND linked_job_match_id = $2`,
		userID, linkedJobMatchID,
	)
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: delete watchlist: %w", err)
	}
	return res.RowsAffected()
}

// DeleteWatchlistForUser removes every watchlist row owned by the
// supplied user (privacy delete cascade).
func (r *Repository) DeleteWatchlistForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM watchlist_items WHERE user_id = $1`, uid)
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: delete watchlist: %w", err)
	}
	return res.RowsAffected()
}

func scanWatchlistRow(row interface{ Scan(dest ...any) error }) (WatchlistRecord, error) {
	var (
		rec        WatchlistRecord
		label      sql.NullString
		changeNote sql.NullString
		dismissed  sql.NullTime
	)
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.LinkedJobMatchID, &label, &changeNote, &rec.AddedAt,
		&rec.LinkedTitle, &rec.LinkedCompany, &rec.LinkedScore, &dismissed); err != nil {
		return WatchlistRecord{}, err
	}
	if label.Valid {
		v := label.String
		rec.Label = &v
	}
	if changeNote.Valid {
		v := changeNote.String
		rec.ChangeNote = &v
	}
	if dismissed.Valid {
		t := dismissed.Time
		rec.LinkedDismissedAt = &t
	}
	return rec, nil
}

// Avoid the unused import warning when pq is not yet wired by other
// methods on this file.
var _ = pq.Array
