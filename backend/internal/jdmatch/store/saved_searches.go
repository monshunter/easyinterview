package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
)

// SavedSearchRecord is the saved_searches row projection.
type SavedSearchRecord struct {
	ID           string
	UserID       string
	Label        string
	Query        string
	Filters      json.RawMessage
	NewJobsCount *int
	LastRunAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// SearchRunRecord is the jd_match_search_runs row projection. Used by
// the searchJobs handler + privacy delete cascade.
type SearchRunRecord struct {
	ID                string
	UserID            string
	SearchRunID       string
	Query             string
	Filters           json.RawMessage
	ResultCount       int
	PromptVersion     *string
	RubricVersion     *string
	ModelID           *string
	DataSourceVersion string
	CreatedAt         time.Time
}

// ListSavedSearchesByUser returns user-owned saved_searches ordered by
// created_at DESC, id DESC.
func (r *Repository) ListSavedSearchesByUser(ctx context.Context, userID string) ([]SavedSearchRecord, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return nil, jdmatch.ErrUserIDRequired
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, label, query, filters, new_jobs_count, last_run_at, created_at, updated_at
		FROM saved_searches WHERE user_id = $1 ORDER BY created_at ASC, id ASC`,
		uid,
	)
	if err != nil {
		return nil, fmt.Errorf("jdmatch store: list saved_searches: %w", err)
	}
	defer rows.Close()
	out := make([]SavedSearchRecord, 0)
	for rows.Next() {
		rec, err := scanSavedSearchRow(rows)
		if err != nil {
			return nil, fmt.Errorf("jdmatch store: scan saved_searches: %w", err)
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// CreateSavedSearchInput captures the createSavedSearch parameters.
type CreateSavedSearchInput struct {
	ID      string
	UserID  string
	Label   string
	Query   string
	Filters json.RawMessage
}

// CreateSavedSearch persists a new saved_searches row. label / query
// land in the row but never appear in log / audit / outbox per D-7.
func (r *Repository) CreateSavedSearch(ctx context.Context, in CreateSavedSearchInput) (SavedSearchRecord, error) {
	if r == nil || r.db == nil {
		return SavedSearchRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.Label) == "" || strings.TrimSpace(in.Query) == "" {
		return SavedSearchRecord{}, jdmatch.ErrValidationFailed
	}
	now := r.now()
	filters := in.Filters
	if len(filters) == 0 {
		filters = json.RawMessage(`{}`)
	}
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO saved_searches (id, user_id, label, query, filters, new_jobs_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $6)
		RETURNING id, user_id, label, query, filters, new_jobs_count, last_run_at, created_at, updated_at`,
		in.ID, in.UserID, in.Label, in.Query, filters, now,
	)
	rec, err := scanSavedSearchRow(row)
	if err != nil {
		return SavedSearchRecord{}, fmt.Errorf("jdmatch store: insert saved_searches: %w", err)
	}
	return rec, nil
}

// DeleteSavedSearchesForUser removes every saved_searches row for the
// supplied user (privacy delete cascade).
func (r *Repository) DeleteSavedSearchesForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM saved_searches WHERE user_id = $1`, uid)
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: delete saved_searches: %w", err)
	}
	return res.RowsAffected()
}

// CreateSearchRunInput captures the jd_match_search_runs INSERT
// payload. query / filters never appear in log / audit / outbox.
type CreateSearchRunInput struct {
	ID                string
	UserID            string
	SearchRunID       string
	Query             string
	Filters           json.RawMessage
	ResultCount       int
	PromptVersion     string
	RubricVersion     string
	ModelID           string
	DataSourceVersion string
}

// CreateSearchRun persists the search run audit row.
func (r *Repository) CreateSearchRun(ctx context.Context, in CreateSearchRunInput) (SearchRunRecord, error) {
	if r == nil || r.db == nil {
		return SearchRunRecord{}, fmt.Errorf("jdmatch store: db is nil")
	}
	if strings.TrimSpace(in.UserID) == "" || strings.TrimSpace(in.SearchRunID) == "" {
		return SearchRunRecord{}, jdmatch.ErrUserIDRequired
	}
	now := r.now()
	filters := in.Filters
	if len(filters) == 0 {
		filters = json.RawMessage(`{}`)
	}
	row := r.db.QueryRowContext(ctx,
		`INSERT INTO jd_match_search_runs (id, user_id, search_run_id, query, filters, result_count, prompt_version, rubric_version, model_id, data_source_version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, user_id, search_run_id, query, filters, result_count, prompt_version, rubric_version, model_id, data_source_version, created_at`,
		in.ID, in.UserID, in.SearchRunID, in.Query, filters, in.ResultCount,
		nullableString(in.PromptVersion), nullableString(in.RubricVersion), nullableString(in.ModelID), in.DataSourceVersion, now,
	)
	var (
		rec       SearchRunRecord
		promptVer sql.NullString
		rubricVer sql.NullString
		modelID   sql.NullString
	)
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.SearchRunID, &rec.Query, &rec.Filters, &rec.ResultCount,
		&promptVer, &rubricVer, &modelID, &rec.DataSourceVersion, &rec.CreatedAt); err != nil {
		return SearchRunRecord{}, fmt.Errorf("jdmatch store: insert search_run: %w", err)
	}
	if promptVer.Valid {
		v := promptVer.String
		rec.PromptVersion = &v
	}
	if rubricVer.Valid {
		v := rubricVer.String
		rec.RubricVersion = &v
	}
	if modelID.Valid {
		v := modelID.String
		rec.ModelID = &v
	}
	return rec, nil
}

// DeleteSearchRunsForUser removes every jd_match_search_runs row for
// the supplied user.
func (r *Repository) DeleteSearchRunsForUser(ctx context.Context, userID string) (int64, error) {
	if r == nil || r.db == nil {
		return 0, fmt.Errorf("jdmatch store: db is nil")
	}
	uid := strings.TrimSpace(userID)
	if uid == "" {
		return 0, jdmatch.ErrUserIDRequired
	}
	res, err := r.db.ExecContext(ctx, `DELETE FROM jd_match_search_runs WHERE user_id = $1`, uid)
	if err != nil {
		return 0, fmt.Errorf("jdmatch store: delete search_runs: %w", err)
	}
	return res.RowsAffected()
}

func scanSavedSearchRow(row interface{ Scan(dest ...any) error }) (SavedSearchRecord, error) {
	var (
		rec          SavedSearchRecord
		newJobsCount sql.NullInt32
		lastRunAt    sql.NullTime
	)
	if err := row.Scan(&rec.ID, &rec.UserID, &rec.Label, &rec.Query, &rec.Filters, &newJobsCount, &lastRunAt, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return SavedSearchRecord{}, err
	}
	if newJobsCount.Valid {
		v := int(newJobsCount.Int32)
		rec.NewJobsCount = &v
	}
	if lastRunAt.Valid {
		t := lastRunAt.Time
		rec.LastRunAt = &t
	}
	return rec, nil
}

func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// silence unused import when no errors helper required
var _ = errors.New
