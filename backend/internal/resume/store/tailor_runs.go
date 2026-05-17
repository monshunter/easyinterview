package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func (r *Repository) CreateTailorRun(ctx context.Context, in CreateTailorRunInput) (CreateTailorRunResult, error) {
	if r == nil || r.db == nil {
		return CreateTailorRunResult{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("begin resume tailor run create: %w", err)
	}
	defer tx.Rollback()

	var assetExists int
	if err := tx.QueryRowContext(ctx, `
select 1 from resume_assets
where id = $1 and user_id = $2 and deleted_at is null`,
		in.ResumeAssetID,
		in.UserID,
	).Scan(&assetExists); errors.Is(err, sql.ErrNoRows) {
		return CreateTailorRunResult{}, ErrAssetNotFound
	} else if err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("check resume asset ownership for tailor run: %w", err)
	}
	var targetExists int
	if err := tx.QueryRowContext(ctx, `
select 1 from target_jobs
where id = $1 and user_id = $2 and deleted_at is null`,
		in.TargetJobID,
		in.UserID,
	).Scan(&targetExists); errors.Is(err, sql.ErrNoRows) {
		return CreateTailorRunResult{}, ErrAssetNotFound
	} else if err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("check target job ownership for tailor run: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
insert into resume_tailor_runs (
  id, user_id, target_job_id, resume_asset_id, mode, status,
  created_at, updated_at
) values ($1,$2,$3,$4,$5,'queued',$6,$6)`,
		in.TailorRunID,
		in.UserID,
		in.TargetJobID,
		in.ResumeAssetID,
		in.Mode,
		now,
	); err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("insert resume tailor run: %w", err)
	}
	payload, err := json.Marshal(map[string]any{
		"tailorRunId":   in.TailorRunID,
		"resumeAssetId": in.ResumeAssetID,
		"targetJobId":   in.TargetJobID,
		"mode":          in.Mode,
	})
	if err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("encode resume tailor payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status,
  payload, available_at, created_at, updated_at
) values ($1,$2,'resume_tailor_run',$3,$4,$5,$6,$7,$7,$7)`,
		in.JobID,
		string(jobs.JobTypeResumeTailor),
		in.TailorRunID,
		nullableString(in.DedupeKey),
		string(sharedtypes.JobStatusQueued),
		payload,
		now,
	); err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("insert resume tailor async job: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("commit resume tailor run create: %w", err)
	}
	return CreateTailorRunResult{
		TailorRunID:  in.TailorRunID,
		JobID:        in.JobID,
		JobStatus:    sharedtypes.JobStatusQueued,
		JobCreatedAt: now,
		JobUpdatedAt: now,
	}, nil
}

func (r *Repository) GetTailorRun(ctx context.Context, userID string, tailorRunID string) (TailorRunRecord, error) {
	if r == nil || r.db == nil {
		return TailorRunRecord{}, fmt.Errorf("resume store db is nil")
	}
	rec, err := scanTailorRun(r.db.QueryRowContext(ctx, tailorRunSelectSQL()+`
where id = $1 and user_id = $2`,
		tailorRunID,
		userID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return TailorRunRecord{}, ErrTailorRunNotFound
	}
	if err != nil {
		return TailorRunRecord{}, err
	}
	return rec, nil
}

func (r *Repository) MarkTailorRunGenerating(ctx context.Context, in TailorRunStatusInput) (TailorRunRecord, error) {
	if r == nil || r.db == nil {
		return TailorRunRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	rec, err := scanTailorRun(r.db.QueryRowContext(ctx, `
update resume_tailor_runs
set status = 'generating',
    updated_at = $1,
    error_code = null
where id = $2 and status in ('queued','failed')
returning id, user_id, target_job_id, resume_asset_id, mode, status,
          match_summary, suggestions, prompt_version, rubric_version, model_id,
          provider, error_code, created_at, updated_at`,
		now,
		in.TailorRunID,
	))
	return mapTailorRunMutationError(ctx, r.db, in.TailorRunID, rec, err)
}

func (r *Repository) MarkTailorRunReady(ctx context.Context, in TailorRunReadyInput) (TailorRunRecord, error) {
	if r == nil || r.db == nil {
		return TailorRunRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	matchSummary := in.MatchSummary
	if len(matchSummary) == 0 {
		matchSummary = json.RawMessage(`{}`)
	}
	suggestions := in.Suggestions
	if len(suggestions) == 0 {
		suggestions = json.RawMessage(`[]`)
	}
	rec, err := scanTailorRun(r.db.QueryRowContext(ctx, `
update resume_tailor_runs
set status = 'ready',
    match_summary = $1,
    suggestions = $2,
    prompt_version = $3,
    rubric_version = $4,
    model_id = $5,
    provider = $6,
    error_code = null,
    generated_at = $7,
    updated_at = $7
where id = $8 and status = 'generating'
returning id, user_id, target_job_id, resume_asset_id, mode, status,
          match_summary, suggestions, prompt_version, rubric_version, model_id,
          provider, error_code, created_at, updated_at`,
		matchSummary,
		suggestions,
		nullableString(in.Provenance.PromptVersion),
		nullableString(in.Provenance.RubricVersion),
		nullableString(in.Provenance.ModelID),
		nullableString(in.Provenance.Provider),
		now,
		in.TailorRunID,
	))
	if err == nil {
		fillTailorRunProvenance(&rec, in.Provenance)
	}
	return mapTailorRunMutationError(ctx, r.db, in.TailorRunID, rec, err)
}

func (r *Repository) MarkTailorRunFailed(ctx context.Context, in TailorRunFailureInput) (TailorRunRecord, error) {
	if r == nil || r.db == nil {
		return TailorRunRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	rec, err := scanTailorRun(r.db.QueryRowContext(ctx, `
update resume_tailor_runs
set status = 'failed',
    error_code = $1,
    updated_at = $2
where id = $3 and status = 'generating'
returning id, user_id, target_job_id, resume_asset_id, mode, status,
          match_summary, suggestions, prompt_version, rubric_version, model_id,
          provider, error_code, created_at, updated_at`,
		nullableString(in.ErrorCode),
		now,
		in.TailorRunID,
	))
	return mapTailorRunMutationError(ctx, r.db, in.TailorRunID, rec, err)
}

func tailorRunSelectSQL() string {
	return `select id, user_id, target_job_id, resume_asset_id, mode, status,
       match_summary, suggestions, prompt_version, rubric_version, model_id,
       provider, error_code, created_at, updated_at
from resume_tailor_runs
`
}

func scanTailorRun(row interface{ Scan(dest ...any) error }) (TailorRunRecord, error) {
	var rec TailorRunRecord
	var matchSummary []byte
	var suggestions []byte
	var promptVersion, rubricVersion, modelID, provider, errorCode sql.NullString
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.TargetJobID,
		&rec.ResumeAssetID,
		&rec.Mode,
		&rec.Status,
		&matchSummary,
		&suggestions,
		&promptVersion,
		&rubricVersion,
		&modelID,
		&provider,
		&errorCode,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return TailorRunRecord{}, err
	}
	rec.MatchSummary = append(json.RawMessage(nil), matchSummary...)
	rec.Suggestions = append(json.RawMessage(nil), suggestions...)
	rec.Provenance = VersionProvenance{
		PromptVersion: stringFromNull(promptVersion),
		RubricVersion: stringFromNull(rubricVersion),
		ModelID:       stringFromNull(modelID),
		Provider:      stringFromNull(provider),
	}
	if errorCode.Valid {
		rec.ErrorCode = &errorCode.String
	}
	return rec, nil
}

func fillTailorRunProvenance(rec *TailorRunRecord, provenance VersionProvenance) {
	if rec == nil {
		return
	}
	if rec.Provenance.PromptVersion == "" {
		rec.Provenance.PromptVersion = provenance.PromptVersion
	}
	if rec.Provenance.RubricVersion == "" {
		rec.Provenance.RubricVersion = provenance.RubricVersion
	}
	if rec.Provenance.ModelID == "" {
		rec.Provenance.ModelID = provenance.ModelID
	}
	if rec.Provenance.Provider == "" {
		rec.Provenance.Provider = provenance.Provider
	}
	if rec.Provenance.Language == "" {
		rec.Provenance.Language = provenance.Language
	}
	if rec.Provenance.FeatureFlag == "" {
		rec.Provenance.FeatureFlag = provenance.FeatureFlag
	}
	if rec.Provenance.DataSourceVersion == "" {
		rec.Provenance.DataSourceVersion = provenance.DataSourceVersion
	}
}

func stringFromNull(in sql.NullString) string {
	if !in.Valid {
		return ""
	}
	return in.String
}

func mapTailorRunMutationError(ctx context.Context, db *sql.DB, id string, rec TailorRunRecord, err error) (TailorRunRecord, error) {
	if !errors.Is(err, sql.ErrNoRows) {
		return rec, err
	}
	var exists int
	if checkErr := db.QueryRowContext(ctx, `select 1 from resume_tailor_runs where id = $1`, id).Scan(&exists); errors.Is(checkErr, sql.ErrNoRows) {
		return TailorRunRecord{}, ErrTailorRunNotFound
	} else if checkErr != nil {
		return TailorRunRecord{}, checkErr
	}
	return TailorRunRecord{}, ErrInvalidStateTransition
}
