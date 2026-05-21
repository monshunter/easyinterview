package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
)

// Repository exposes the agent_scans / jd_match_recommendations /
// watchlist_items / saved_searches / jd_match_search_runs repositories.
// Construct via NewRepository.
type Repository struct {
	db  *sql.DB
	now func() time.Time
}

// NewRepository wires the JD-Match store layer. now is overridable so
// tests can pin timestamps without touching the wall clock.
func NewRepository(db *sql.DB, now func() time.Time) *Repository {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Repository{db: db, now: now}
}

// GetLatestAgentScanForUser returns the most recent agent_scans row owned
// by the supplied userID (ordered by created_at DESC). When no row
// exists (the user has never triggered or scheduled a scan), the
// repository returns jdmatch.ErrNotFound so the caller can render the
// D-3 lazy-idle baseline.
func (r *Repository) GetLatestAgentScanForUser(ctx context.Context, userID string) (jdmatch.AgentScanRecord, error) {
	if r == nil || r.db == nil {
		return jdmatch.AgentScanRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrUserIDRequired
	}
	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, status, started_at, finished_at, last_scan_at, next_scan_at, error_message, recommendation_count, created_at, updated_at
		FROM agent_scans
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1`,
		uid,
	)
	rec, err := scanAgentScanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrNotFound
	}
	if err != nil {
		return jdmatch.AgentScanRecord{}, fmt.Errorf("jdmatch store: scan agent_scan: %w", err)
	}
	return rec, nil
}

// CreateAgentScanInput captures the fields the service supplies when
// kicking off a new scan row (typically with status='idle' for a lazy
// trigger or status='scanning' for an in-flight job).
type CreateAgentScanInput struct {
	ID         string
	UserID     string
	Status     jdmatch.AgentScanStatus
	StartedAt  *time.Time
	LastScanAt *time.Time
	NextScanAt *time.Time
}

// CreateAgentScan persists a new agent_scans row. Returns the freshly
// inserted record.
func (r *Repository) CreateAgentScan(ctx context.Context, in CreateAgentScanInput) (jdmatch.AgentScanRecord, error) {
	if r == nil || r.db == nil {
		return jdmatch.AgentScanRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.UserID) == "" {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrUserIDRequired
	}
	if !isValidAgentScanStatus(in.Status) {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrInvalidStatus
	}
	now := r.now()
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO agent_scans (id, user_id, status, started_at, last_scan_at, next_scan_at, recommendation_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $7)`,
		in.ID, in.UserID, string(in.Status), in.StartedAt, in.LastScanAt, in.NextScanAt, now,
	)
	if err != nil {
		return jdmatch.AgentScanRecord{}, fmt.Errorf("jdmatch store: insert agent_scan: %w", err)
	}
	return jdmatch.AgentScanRecord{
		ID:         in.ID,
		UserID:     in.UserID,
		Status:     in.Status,
		StartedAt:  in.StartedAt,
		LastScanAt: in.LastScanAt,
		NextScanAt: in.NextScanAt,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// UpdateAgentScanStatusInput drives an in-place transition: callers may
// supply any subset of the optional fields to advance the row (e.g. set
// status=scanning + started_at; later set status=idle + finished_at +
// last_scan_at + next_scan_at; on failure set status=error +
// error_message + finished_at).
type UpdateAgentScanStatusInput struct {
	ID                  string
	UserID              string
	Status              jdmatch.AgentScanStatus
	StartedAt           *time.Time
	FinishedAt          *time.Time
	LastScanAt          *time.Time
	NextScanAt          *time.Time
	ErrorMessage        *string
	RecommendationCount *int
}

// UpdateAgentScanStatus advances an existing agent_scans row. The
// update is scoped to (id, user_id) so cross-user mutation is impossible
// at the SQL layer. Returns ErrNotFound when the (id, user_id) tuple
// does not match a row.
func (r *Repository) UpdateAgentScanStatus(ctx context.Context, in UpdateAgentScanStatusInput) (jdmatch.AgentScanRecord, error) {
	if r == nil || r.db == nil {
		return jdmatch.AgentScanRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.UserID) == "" {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrUserIDRequired
	}
	if !isValidAgentScanStatus(in.Status) {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrInvalidStatus
	}
	now := r.now()
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE agent_scans
		SET status = $3,
		    started_at = COALESCE($4, started_at),
		    finished_at = COALESCE($5, finished_at),
		    last_scan_at = COALESCE($6, last_scan_at),
		    next_scan_at = COALESCE($7, next_scan_at),
		    error_message = $8,
		    recommendation_count = COALESCE($9, recommendation_count),
		    updated_at = $10
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, status, started_at, finished_at, last_scan_at, next_scan_at, error_message, recommendation_count, created_at, updated_at`,
		in.ID, in.UserID, string(in.Status), in.StartedAt, in.FinishedAt, in.LastScanAt, in.NextScanAt, in.ErrorMessage, in.RecommendationCount, now,
	)
	rec, err := scanAgentScanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return jdmatch.AgentScanRecord{}, jdmatch.ErrNotFound
	}
	if err != nil {
		return jdmatch.AgentScanRecord{}, fmt.Errorf("jdmatch store: update agent_scan: %w", err)
	}
	return rec, nil
}

// DeleteAgentScansForUser removes every agent_scans row owned by the
// supplied user (cascading the per-user privacy delete chain in spec
// §3.1.2). Returns the number of rows removed.
func (r *Repository) DeleteAgentScansForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM agent_scans WHERE user_id = $1`, uid)
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: delete agent_scans: %w", err)
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: rows affected: %w", err)
	}
	return count, nil
}

func scanAgentScanRow(row interface{ Scan(dest ...any) error }) (jdmatch.AgentScanRecord, error) {
	var (
		rec        jdmatch.AgentScanRecord
		status     string
		errMessage sql.NullString
	)
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&status,
		&rec.StartedAt,
		&rec.FinishedAt,
		&rec.LastScanAt,
		&rec.NextScanAt,
		&errMessage,
		&rec.RecommendationCount,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return jdmatch.AgentScanRecord{}, err
	}
	rec.Status = jdmatch.AgentScanStatus(status)
	if errMessage.Valid {
		v := errMessage.String
		rec.ErrorMessage = &v
	}
	return rec, nil
}

func isValidAgentScanStatus(s jdmatch.AgentScanStatus) bool {
	for _, v := range jdmatch.AllAgentScanStatuses {
		if v == s {
			return true
		}
	}
	return false
}
