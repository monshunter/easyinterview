package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

// async_jobs.resource_type values owned by this subject. D-20 keeps tailor run
// state on the async job row; parse jobs keep the API-facing resume_asset
// resource type while the physical table is now resumes.
type resourceType string

const (
	resourceTypeResume          resourceType = "resume_asset"
	resourceTypeResumeTailorRun resourceType = "resume_tailor_run"
)

// tailorJobPayloadColumns decode the requestResumeTailor payload back out of
// the async_jobs row so getResumeTailorRun can reconstruct the run.
type tailorJobPayload struct {
	ResumeID    string `json:"resumeId"`
	TargetJobID string `json:"targetJobId,omitempty"`
	Mode        string `json:"mode"`
}

// tailorJobResult is the ephemeral resume.tailor output persisted into
// async_jobs.result by the tailor job.
type tailorJobResult struct {
	MatchSummary json.RawMessage          `json:"matchSummary,omitempty"`
	Suggestions  []tailorResultSuggestion `json:"suggestions"`
	Provenance   tailorResultProvenance   `json:"provenance"`
}

type tailorResultSuggestion struct {
	OriginalBullet  string `json:"originalBullet"`
	SuggestedBullet string `json:"suggestedBullet"`
	Reason          string `json:"reason"`
}

type tailorResultProvenance struct {
	PromptVersion     string `json:"promptVersion,omitempty"`
	RubricVersion     string `json:"rubricVersion,omitempty"`
	ModelID           string `json:"modelId,omitempty"`
	Provider          string `json:"provider,omitempty"`
	Language          string `json:"language,omitempty"`
	FeatureFlag       string `json:"featureFlag,omitempty"`
	DataSourceVersion string `json:"dataSourceVersion,omitempty"`
}

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

	var resumeExists int
	if err := tx.QueryRowContext(ctx, `
select 1 from resumes
where id = $1 and user_id = $2 and deleted_at is null`,
		in.ResumeID,
		in.UserID,
	).Scan(&resumeExists); errors.Is(err, sql.ErrNoRows) {
		return CreateTailorRunResult{}, ErrAssetNotFound
	} else if err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("check resume ownership for tailor run: %w", err)
	}
	targetJobID := strings.TrimSpace(in.TargetJobID)
	if targetJobID != "" {
		var targetExists int
		if err := tx.QueryRowContext(ctx, `
select 1 from target_jobs
where id = $1 and user_id = $2 and deleted_at is null`,
			targetJobID,
			in.UserID,
		).Scan(&targetExists); errors.Is(err, sql.ErrNoRows) {
			return CreateTailorRunResult{}, ErrAssetNotFound
		} else if err != nil {
			return CreateTailorRunResult{}, fmt.Errorf("check target job ownership for tailor run: %w", err)
		}
	}

	payloadMap := map[string]any{
		"resumeId": in.ResumeID,
		"mode":     in.Mode,
	}
	if targetJobID != "" {
		payloadMap["targetJobId"] = targetJobID
	}
	payload, err := json.Marshal(payloadMap)
	if err != nil {
		return CreateTailorRunResult{}, fmt.Errorf("encode resume tailor payload: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status,
  payload, available_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$8,$8)`,
		in.JobID,
		string(jobs.JobTypeResumeTailor),
		string(resourceTypeResumeTailorRun),
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

// GetTailorRun reconstructs a tailor run from the async_jobs row keyed by
// tailorRunId. User scope is enforced by joining resumes on the payload
// resumeId; cross-user reads return ErrTailorRunNotFound.
func (r *Repository) GetTailorRun(ctx context.Context, userID string, tailorRunID string) (TailorRunRecord, error) {
	if r == nil || r.db == nil {
		return TailorRunRecord{}, fmt.Errorf("resume store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select aj.resource_id, rs.user_id, aj.status, aj.payload, aj.result, aj.error_code,
       aj.created_at, aj.updated_at
from async_jobs aj
join resumes rs on rs.id = (aj.payload->>'resumeId')::uuid and rs.user_id = $2 and rs.deleted_at is null
where aj.resource_type = $3 and aj.resource_id = $1`,
		tailorRunID,
		userID,
		string(resourceTypeResumeTailorRun),
	)
	rec, err := scanTailorRun(row)
	if errors.Is(err, sql.ErrNoRows) {
		return TailorRunRecord{}, ErrTailorRunNotFound
	}
	if err != nil {
		return TailorRunRecord{}, err
	}
	return rec, nil
}

// GetForTailor loads the resume + target job context the tailor job feeds to
// the model. It joins resumes directly (D-20: no versions).
func (r *Repository) GetForTailor(ctx context.Context, tailorRunID string) (TailorJobContext, error) {
	if r == nil || r.db == nil {
		return TailorJobContext{}, fmt.Errorf("resume store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select aj.resource_id, rs.user_id, rs.id, coalesce(aj.payload->>'targetJobId', ''),
       coalesce(aj.payload->>'mode', ''),
       coalesce(rs.language, 'en'), coalesce(rs.parsed_summary, '{}'::jsonb),
       coalesce(rs.structured_profile, '{}'::jsonb),
       coalesce(tj.summary, '{}'::jsonb), coalesce(tj.title, ''),
       coalesce(tj.company_name, ''), coalesce(tj.seniority_level, ''),
       coalesce(tj.raw_jd_text, '')
from async_jobs aj
join resumes rs on rs.id = (aj.payload->>'resumeId')::uuid and rs.deleted_at is null
left join target_jobs tj on tj.id = nullif(aj.payload->>'targetJobId', '')::uuid
  and tj.user_id = rs.user_id and tj.deleted_at is null
where aj.resource_type = $2 and aj.resource_id = $1`,
		tailorRunID,
		string(resourceTypeResumeTailorRun),
	)
	var rec TailorJobContext
	var resumeSummary, structuredProfile, targetSummary []byte
	if err := row.Scan(
		&rec.TailorRunID,
		&rec.UserID,
		&rec.ResumeID,
		&rec.TargetJobID,
		&rec.Mode,
		&rec.Language,
		&resumeSummary,
		&structuredProfile,
		&targetSummary,
		&rec.TargetTitle,
		&rec.TargetCompany,
		&rec.TargetSeniority,
		&rec.RawJDText,
	); errors.Is(err, sql.ErrNoRows) {
		return TailorJobContext{}, ErrTailorRunNotFound
	} else if err != nil {
		return TailorJobContext{}, err
	}
	rec.ResumeSummary = append(json.RawMessage(nil), resumeSummary...)
	rec.StructuredProfile = append(json.RawMessage(nil), structuredProfile...)
	rec.TargetSummary = append(json.RawMessage(nil), targetSummary...)
	rec.OriginalBullet = firstStructuredProfileBullet(rec.StructuredProfile)
	return rec, nil
}

// CompleteTailorRunSuccess persists the ephemeral tailor result onto the
// driving async_jobs row and emits resume.tailor.completed (aggregate_type
// resume, aggregate_id resumeId). The runner kernel finalizes async_jobs.status
// separately; getResumeTailorRun derives 'ready' from a present result.
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
	resultRaw, err := marshalTailorResult(matchSummary, in.Suggestions, in.Provenance)
	if err != nil {
		return err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin resume tailor success: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
update async_jobs
set result = $1, updated_at = $2
where resource_type = $3 and resource_id = $4 and status = 'running'`,
		resultRaw,
		now,
		string(resourceTypeResumeTailorRun),
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
	if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload,
  publish_status, next_attempt_at, created_at
) values ($1,$2,1,$3,$4,$5,'pending',$6,$6)`,
		in.OutboxEventID,
		string(events.EventNameResumeTailorCompleted),
		resumeAggregateType,
		in.ResumeID,
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

func marshalTailorResult(matchSummary json.RawMessage, suggestions []TailorSuggestionInput, provenance VersionProvenance) (json.RawMessage, error) {
	out := tailorJobResult{
		MatchSummary: matchSummary,
		Suggestions:  make([]tailorResultSuggestion, 0, len(suggestions)),
		Provenance: tailorResultProvenance{
			PromptVersion:     provenance.PromptVersion,
			RubricVersion:     provenance.RubricVersion,
			ModelID:           provenance.ModelID,
			Provider:          provenance.Provider,
			Language:          provenance.Language,
			FeatureFlag:       provenance.FeatureFlag,
			DataSourceVersion: provenance.DataSourceVersion,
		},
	}
	for _, suggestion := range suggestions {
		out.Suggestions = append(out.Suggestions, tailorResultSuggestion{
			OriginalBullet:  suggestion.OriginalBullet,
			SuggestedBullet: suggestion.SuggestedBullet,
			Reason:          suggestion.Reason,
		})
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal resume tailor result: %w", err)
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

// scanTailorRun maps an async_jobs row (resource_id, owner user_id, status,
// payload, result, error_code, timestamps) into a TailorRunRecord.
func scanTailorRun(row rowScanner) (TailorRunRecord, error) {
	var (
		rec       TailorRunRecord
		jobStatus string
		payload   []byte
		result    []byte
		errorCode sql.NullString
	)
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&jobStatus,
		&payload,
		&result,
		&errorCode,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return TailorRunRecord{}, err
	}
	var meta tailorJobPayload
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &meta)
	}
	rec.ResumeID = strings.TrimSpace(meta.ResumeID)
	rec.TargetJobID = strings.TrimSpace(meta.TargetJobID)
	rec.Mode = strings.TrimSpace(meta.Mode)

	hasResult := len(result) > 0 && string(result) != "{}" && string(result) != "null"
	var parsed tailorJobResult
	if hasResult {
		if err := json.Unmarshal(result, &parsed); err != nil {
			hasResult = false
		}
	}
	rec.Status = mapTailorRunStatus(jobStatus, hasResult)
	if hasResult {
		rec.MatchSummary = append(json.RawMessage(nil), parsed.MatchSummary...)
		rec.Suggestions = suggestionsToJSON(parsed.Suggestions)
		rec.Provenance = VersionProvenance{
			PromptVersion:     parsed.Provenance.PromptVersion,
			RubricVersion:     parsed.Provenance.RubricVersion,
			ModelID:           parsed.Provenance.ModelID,
			Provider:          parsed.Provenance.Provider,
			Language:          parsed.Provenance.Language,
			FeatureFlag:       parsed.Provenance.FeatureFlag,
			DataSourceVersion: parsed.Provenance.DataSourceVersion,
		}
	}
	if errorCode.Valid {
		rec.ErrorCode = &errorCode.String
	}
	return rec, nil
}

// mapTailorRunStatus collapses async_jobs lifecycle status onto the
// ResumeTailorRun status enum. A present result means the tailor output was
// already persisted, so the run is ready even if the kernel has not yet flipped
// the async_jobs row to 'succeeded'.
func mapTailorRunStatus(jobStatus string, hasResult bool) string {
	if hasResult {
		return "ready"
	}
	switch jobStatus {
	case "queued":
		return "queued"
	case "running":
		return "generating"
	case "succeeded":
		return "ready"
	case "failed", "dead", "cancelled":
		return "failed"
	default:
		return "queued"
	}
}

func suggestionsToJSON(in []tailorResultSuggestion) json.RawMessage {
	if len(in) == 0 {
		return json.RawMessage(`[]`)
	}
	raw, err := json.Marshal(in)
	if err != nil {
		return json.RawMessage(`[]`)
	}
	return raw
}
