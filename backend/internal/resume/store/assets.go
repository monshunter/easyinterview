package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWithParseJob(ctx context.Context, in CreateAssetInput) (CreateAssetResult, error) {
	if r == nil || r.db == nil {
		return CreateAssetResult{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	parseStatus := in.ParseStatus
	if parseStatus == "" {
		parseStatus = sharedtypes.TargetJobParseStatusQueued
	}
	jobStatus := in.JobStatus
	if jobStatus == "" {
		jobStatus = sharedtypes.JobStatusQueued
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CreateAssetResult{}, fmt.Errorf("begin resume asset create: %w", err)
	}
	defer tx.Rollback()

	if in.DedupeKey != "" {
		existing, hit, err := lookupActiveRegisterDedupe(ctx, tx, in.DedupeKey)
		if err != nil {
			return CreateAssetResult{}, err
		}
		if hit {
			if err := tx.Commit(); err != nil {
				return CreateAssetResult{}, fmt.Errorf("commit resume register replay: %w", err)
			}
			existing.Existing = true
			return existing, nil
		}
	}

	guidedAnswers, err := nullableJSON(in.GuidedAnswers)
	if err != nil {
		return CreateAssetResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
insert into resume_assets (
  id, user_id, file_object_id, title, language, parse_status,
  source_type, original_text, guided_answers, latest_parse_job_id, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		in.AssetID,
		in.UserID,
		nullableStringPtr(in.FileObjectID),
		in.Title,
		in.Language,
		string(parseStatus),
		in.SourceType,
		nullableString(in.RawText),
		guidedAnswers,
		in.JobID,
		now,
		now,
	); err != nil {
		return CreateAssetResult{}, fmt.Errorf("insert resume asset: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"resumeAssetId": in.AssetID,
		"userId":        in.UserID,
		"sourceType":    in.SourceType,
		"request":       in.RequestPayload,
	})
	if err != nil {
		return CreateAssetResult{}, fmt.Errorf("marshal resume parse job payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status,
  payload, available_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		in.JobID,
		string(jobs.JobTypeResumeParse),
		string(api.ResourceTypeResumeAsset),
		in.AssetID,
		nullableString(in.DedupeKey),
		string(jobStatus),
		payload,
		now,
		now,
		now,
	); err != nil {
		return CreateAssetResult{}, fmt.Errorf("insert resume parse job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return CreateAssetResult{}, fmt.Errorf("commit resume asset create: %w", err)
	}
	return CreateAssetResult{
		AssetID:      in.AssetID,
		JobID:        in.JobID,
		JobStatus:    jobStatus,
		JobCreatedAt: now,
		JobUpdatedAt: now,
	}, nil
}

func lookupActiveRegisterDedupe(ctx context.Context, tx *sql.Tx, dedupeKey string) (CreateAssetResult, bool, error) {
	var out CreateAssetResult
	var status string
	err := tx.QueryRowContext(ctx, `
select resource_id, id, status, created_at, updated_at from async_jobs
where job_type = $1 and dedupe_key = $2 and status in ('queued','running')
order by created_at desc
limit 1
for update`,
		string(jobs.JobTypeResumeParse),
		dedupeKey,
	).Scan(&out.AssetID, &out.JobID, &status, &out.JobCreatedAt, &out.JobUpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return CreateAssetResult{}, false, nil
	}
	if err != nil {
		return CreateAssetResult{}, false, fmt.Errorf("lookup resume register dedupe: %w", err)
	}
	out.JobStatus = sharedtypes.JobStatus(status)
	return out, true, nil
}

func nullableString(in string) any {
	if in == "" {
		return nil
	}
	return in
}

func nullableStringPtr(in *string) any {
	if in == nil || *in == "" {
		return nil
	}
	return *in
}

func nullableJSON(in map[string]any) (any, error) {
	if len(in) == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("marshal guided answers: %w", err)
	}
	return raw, nil
}
