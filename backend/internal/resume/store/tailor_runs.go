package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/shared/events"
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
	resumeVersionID := strings.TrimSpace(in.ResumeVersionID)
	if resumeVersionID != "" {
		var versionExists int
		if err := tx.QueryRowContext(ctx, `
select 1 from resume_versions
where id = $1
  and user_id = $2
  and resume_asset_id = $3
  and target_job_id = $4
  and deleted_at is null`,
			resumeVersionID,
			in.UserID,
			in.ResumeAssetID,
			in.TargetJobID,
		).Scan(&versionExists); errors.Is(err, sql.ErrNoRows) {
			return CreateTailorRunResult{}, ErrVersionNotFound
		} else if err != nil {
			return CreateTailorRunResult{}, fmt.Errorf("check resume version ownership for tailor run: %w", err)
		}
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
	payloadMap := map[string]any{
		"tailorRunId":   in.TailorRunID,
		"resumeAssetId": in.ResumeAssetID,
		"targetJobId":   in.TargetJobID,
		"mode":          in.Mode,
	}
	if resumeVersionID != "" {
		payloadMap["resumeVersionId"] = resumeVersionID
	}
	payload, err := json.Marshal(payloadMap)
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

func (r *Repository) GetForTailor(ctx context.Context, tailorRunID string, resumeVersionID string) (TailorJobContext, error) {
	if r == nil || r.db == nil {
		return TailorJobContext{}, fmt.Errorf("resume store db is nil")
	}
	resumeVersionID = strings.TrimSpace(resumeVersionID)
	row := r.db.QueryRowContext(ctx, `
select r.id, r.user_id, r.target_job_id, r.resume_asset_id, r.mode,
       coalesce(ra.language, 'en'), coalesce(ra.parsed_summary, '{}'::jsonb),
       rv.id, rv.structured_profile,
       coalesce(tj.summary, '{}'::jsonb), coalesce(tj.title, ''),
       coalesce(tj.company_name, ''), coalesce(tj.seniority_level, ''),
       coalesce(tj.raw_jd_text, '')
from resume_tailor_runs r
join resume_assets ra on ra.id = r.resume_asset_id and ra.user_id = r.user_id and ra.deleted_at is null
join target_jobs tj on tj.id = r.target_job_id and tj.user_id = r.user_id and tj.deleted_at is null
join lateral (
  select id, structured_profile, updated_at
  from resume_versions rv
  where rv.user_id = r.user_id
    and rv.resume_asset_id = r.resume_asset_id
    and rv.deleted_at is null
    and (
      ($2 <> '' and rv.id = nullif($2, '')::uuid)
      or ($2 = '' and (rv.target_job_id = r.target_job_id or rv.target_job_id is null))
    )
  order by case when $2 <> '' then 0 when rv.target_job_id = r.target_job_id then 0 else 1 end,
           rv.updated_at desc, rv.id desc
  limit 1
) rv on true
where r.id = $1`,
		tailorRunID,
		resumeVersionID,
	)
	var rec TailorJobContext
	var resumeSummary, structuredProfile, targetSummary []byte
	if err := row.Scan(
		&rec.TailorRunID,
		&rec.UserID,
		&rec.TargetJobID,
		&rec.ResumeAssetID,
		&rec.Mode,
		&rec.Language,
		&resumeSummary,
		&rec.ResumeVersionID,
		&structuredProfile,
		&targetSummary,
		&rec.TargetTitle,
		&rec.TargetCompany,
		&rec.TargetSeniority,
		&rec.RawJDText,
	); errors.Is(err, sql.ErrNoRows) {
		var exists int
		if checkErr := r.db.QueryRowContext(ctx, `select 1 from resume_tailor_runs where id = $1`, tailorRunID).Scan(&exists); errors.Is(checkErr, sql.ErrNoRows) {
			return TailorJobContext{}, ErrTailorRunNotFound
		} else if checkErr != nil {
			return TailorJobContext{}, checkErr
		}
		return TailorJobContext{}, ErrVersionNotFound
	} else if err != nil {
		return TailorJobContext{}, err
	}
	rec.ResumeSummary = append(json.RawMessage(nil), resumeSummary...)
	rec.StructuredProfile = append(json.RawMessage(nil), structuredProfile...)
	rec.TargetSummary = append(json.RawMessage(nil), targetSummary...)
	rec.OriginalBullet = firstStructuredProfileBullet(rec.StructuredProfile)
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
          provider, language, feature_flag, data_source_version,
          error_code, created_at, updated_at`,
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
    language = $7,
    feature_flag = $8,
    data_source_version = $9,
    error_code = null,
    generated_at = $10,
    updated_at = $10
where id = $11 and status = 'generating'
returning id, user_id, target_job_id, resume_asset_id, mode, status,
          match_summary, suggestions, prompt_version, rubric_version, model_id,
          provider, language, feature_flag, data_source_version,
          error_code, created_at, updated_at`,
		matchSummary,
		suggestions,
		nullableString(in.Provenance.PromptVersion),
		nullableString(in.Provenance.RubricVersion),
		nullableString(in.Provenance.ModelID),
		nullableString(in.Provenance.Provider),
		provenanceDefault(in.Provenance.Language, "en"),
		provenanceDefault(in.Provenance.FeatureFlag, "none"),
		provenanceDefault(in.Provenance.DataSourceVersion, "not_applicable"),
		now,
		in.TailorRunID,
	))
	if err == nil {
		fillTailorRunProvenance(&rec, in.Provenance)
	}
	return mapTailorRunMutationError(ctx, r.db, in.TailorRunID, rec, err)
}

func (r *Repository) CompleteTailorRunSuccess(ctx context.Context, in CompleteTailorRunSuccessInput) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	matchSummary := in.MatchSummary
	if len(matchSummary) == 0 {
		matchSummary = json.RawMessage(`{}`)
	}
	suggestionsJSON, err := marshalTailorSuggestions(in.Suggestions)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin resume tailor success: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
update resume_tailor_runs
set status = 'ready',
    match_summary = $1,
    suggestions = $2,
    prompt_version = $3,
    rubric_version = $4,
    model_id = $5,
    provider = $6,
    language = $7,
    feature_flag = $8,
    data_source_version = $9,
    error_code = null,
    generated_at = $10,
    updated_at = $10
where id = $11 and status = 'generating'`,
		matchSummary,
		suggestionsJSON,
		nullableString(in.Provenance.PromptVersion),
		nullableString(in.Provenance.RubricVersion),
		nullableString(in.Provenance.ModelID),
		nullableString(in.Provenance.Provider),
		provenanceDefault(in.Provenance.Language, "en"),
		provenanceDefault(in.Provenance.FeatureFlag, "none"),
		provenanceDefault(in.Provenance.DataSourceVersion, "not_applicable"),
		now,
		in.TailorRunID,
	)
	if err != nil {
		return fmt.Errorf("complete resume tailor success: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete resume tailor success rows affected: %w", err)
	}
	if rows == 0 {
		return ErrInvalidStateTransition
	}
	if _, err := tx.ExecContext(ctx, `delete from resume_version_suggestions where tailor_run_id = $1`, in.TailorRunID); err != nil {
		return fmt.Errorf("clear existing resume tailor suggestions: %w", err)
	}
	for _, suggestion := range in.Suggestions {
		if strings.TrimSpace(suggestion.ID) == "" {
			return fmt.Errorf("resume tailor suggestion id is required")
		}
		if _, err := tx.ExecContext(ctx, `
insert into resume_version_suggestions (
  id, resume_version_id, tailor_run_id, original_bullet, suggested_bullet,
  reason, status, created_at
) values ($1,$2,$3,$4,$5,$6,'pending',$7)`,
			suggestion.ID,
			in.ResumeVersionID,
			in.TailorRunID,
			suggestion.OriginalBullet,
			suggestion.SuggestedBullet,
			nullableString(suggestion.Reason),
			now,
		); err != nil {
			return fmt.Errorf("insert resume tailor suggestion: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload,
  publish_status, next_attempt_at, created_at
) values ($1,$2,1,$3,$4,$5,'pending',$6,$6)`,
		in.OutboxEventID,
		string(events.EventNameResumeTailorCompleted),
		string(api.ResourceTypeResumeTailorRun),
		in.TailorRunID,
		in.OutboxEventPayload,
		now,
	); err != nil {
		return fmt.Errorf("insert resume tailor completed outbox: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit resume tailor success: %w", err)
	}
	return nil
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
          provider, language, feature_flag, data_source_version,
          error_code, created_at, updated_at`,
		nullableString(in.ErrorCode),
		now,
		in.TailorRunID,
	))
	return mapTailorRunMutationError(ctx, r.db, in.TailorRunID, rec, err)
}

func marshalTailorSuggestions(in []TailorSuggestionInput) (json.RawMessage, error) {
	out := make([]map[string]string, 0, len(in))
	for _, suggestion := range in {
		out = append(out, map[string]string{
			"originalBullet":  suggestion.OriginalBullet,
			"suggestedBullet": suggestion.SuggestedBullet,
			"reason":          suggestion.Reason,
		})
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal resume tailor suggestions: %w", err)
	}
	return raw, nil
}

func firstStructuredProfileBullet(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}
	return firstBulletValue(value)
}

func firstBulletValue(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"bullets", "items"} {
			if bullet := firstBulletValue(typed[key]); bullet != "" {
				return bullet
			}
		}
		for _, key := range []string{"bullet", "text", "description"} {
			if text, ok := typed[key].(string); ok && strings.TrimSpace(text) != "" {
				return strings.TrimSpace(text)
			}
		}
		for _, nested := range typed {
			if bullet := firstBulletValue(nested); bullet != "" {
				return bullet
			}
		}
	case []any:
		for _, item := range typed {
			if bullet := firstBulletValue(item); bullet != "" {
				return bullet
			}
		}
	case string:
		return strings.TrimSpace(typed)
	}
	return ""
}

func tailorRunSelectSQL() string {
	return `select id, user_id, target_job_id, resume_asset_id, mode, status,
       match_summary, suggestions, prompt_version, rubric_version, model_id,
       provider, language, feature_flag, data_source_version,
       error_code, created_at, updated_at
from resume_tailor_runs
`
}

func scanTailorRun(row interface{ Scan(dest ...any) error }) (TailorRunRecord, error) {
	var rec TailorRunRecord
	var matchSummary []byte
	var suggestions []byte
	var promptVersion, rubricVersion, modelID, provider, language, featureFlag, dataSourceVersion, errorCode sql.NullString
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
		&language,
		&featureFlag,
		&dataSourceVersion,
		&errorCode,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return TailorRunRecord{}, err
	}
	rec.MatchSummary = append(json.RawMessage(nil), matchSummary...)
	rec.Suggestions = append(json.RawMessage(nil), suggestions...)
	rec.Provenance = VersionProvenance{
		PromptVersion:     stringFromNull(promptVersion),
		RubricVersion:     stringFromNull(rubricVersion),
		ModelID:           stringFromNull(modelID),
		Provider:          stringFromNull(provider),
		Language:          stringFromNull(language),
		FeatureFlag:       stringFromNull(featureFlag),
		DataSourceVersion: stringFromNull(dataSourceVersion),
	}
	if errorCode.Valid {
		rec.ErrorCode = &errorCode.String
	}
	return rec, nil
}

func provenanceDefault(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
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
