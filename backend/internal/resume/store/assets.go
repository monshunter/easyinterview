package store

import (
	"context"
	"database/sql"
	"encoding/base64"
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

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetForParse(ctx context.Context, assetID string) (ParseAssetRecord, error) {
	if r == nil || r.db == nil {
		return ParseAssetRecord{}, fmt.Errorf("resume store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select ra.id, ra.user_id, ra.language, ra.parse_status, coalesce(ra.source_type, ''),
       coalesce(ra.original_text, ''), coalesce(ra.guided_answers, '{}'::jsonb),
       coalesce(ra.file_object_id::text, ''), coalesce(fo.object_key, '')
from resume_assets ra
left join file_objects fo on fo.id = ra.file_object_id and fo.deleted_at is null
where ra.id = $1 and ra.deleted_at is null`,
		assetID,
	)
	var rec ParseAssetRecord
	var parseStatus string
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&rec.Language,
		&parseStatus,
		&rec.SourceType,
		&rec.OriginalText,
		&rec.GuidedAnswers,
		&rec.FileObjectID,
		&rec.FileObjectKey,
	); errors.Is(err, sql.ErrNoRows) {
		return ParseAssetRecord{}, ErrAssetNotFound
	} else if err != nil {
		return ParseAssetRecord{}, err
	}
	rec.ParseStatus = sharedtypes.TargetJobParseStatus(parseStatus)
	return rec, nil
}

func (r *Repository) Get(ctx context.Context, userID string, assetID string) (AssetRecord, error) {
	if r == nil || r.db == nil {
		return AssetRecord{}, fmt.Errorf("resume store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select id, user_id, file_object_id, title, language, parse_status,
       parsed_summary, original_text, guided_answers, parsed_text_snapshot,
       source_type, error_code, latest_parse_job_id, created_at, updated_at, deleted_at
from resume_assets
where id = $1 and user_id = $2 and deleted_at is null`,
		assetID,
		userID,
	)
	rec, err := scanAsset(row)
	if errors.Is(err, sql.ErrNoRows) {
		return AssetRecord{}, ErrAssetNotFound
	}
	if err != nil {
		return AssetRecord{}, err
	}
	return rec, nil
}

func (r *Repository) List(ctx context.Context, userID string, filter ListFilter) (ListResult, error) {
	if r == nil || r.db == nil {
		return ListResult{}, fmt.Errorf("resume store db is nil")
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = sharedtypes.DefaultPageSize
	}
	if pageSize > sharedtypes.MaxPageSize {
		pageSize = sharedtypes.MaxPageSize
	}
	limit := pageSize + 1
	args := []any{userID, limit}
	query := `
select id, user_id, file_object_id, title, language, parse_status,
       parsed_summary, original_text, guided_answers, parsed_text_snapshot,
       source_type, error_code, latest_parse_job_id, created_at, updated_at, deleted_at
from resume_assets
where user_id = $1 and deleted_at is null`
	if strings.TrimSpace(filter.Cursor) != "" {
		updatedAt, id, err := decodeCursor(filter.Cursor)
		if err != nil {
			return ListResult{}, ErrInvalidCursor
		}
		args = []any{userID, updatedAt, id, limit}
		query += ` and (updated_at, id) < ($2, $3)`
	}
	query += ` order by updated_at desc, id desc limit $` + fmt.Sprint(len(args))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return ListResult{}, err
	}
	defer rows.Close()
	items := make([]AssetRecord, 0, pageSize)
	for rows.Next() {
		rec, err := scanAsset(rows)
		if err != nil {
			return ListResult{}, err
		}
		items = append(items, rec)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, err
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	nextCursor := ""
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = encodeCursor(last.UpdatedAt, last.ID)
	}
	return ListResult{Items: items, NextCursor: nextCursor, HasMore: hasMore, PageSize: pageSize}, nil
}

func (r *Repository) MarkParsing(ctx context.Context, in StatusUpdateInput) error {
	return r.updateStatus(ctx, in.UserID, in.AssetID, "", sharedtypes.TargetJobParseStatusProcessing, nil, nil, "", in.Now)
}

func (r *Repository) MarkReady(ctx context.Context, in MarkReadyInput) error {
	return r.updateStatus(ctx, in.UserID, in.AssetID, sharedtypes.TargetJobParseStatusProcessing, sharedtypes.TargetJobParseStatusReady, in.ParsedSummary, &in.ParsedTextSnapshot, "", in.Now)
}

func (r *Repository) MarkFailed(ctx context.Context, in MarkFailedInput) error {
	return r.updateStatus(ctx, in.UserID, in.AssetID, "", sharedtypes.TargetJobParseStatusFailed, nil, nil, in.ErrorCode, in.Now)
}

func (r *Repository) CompleteParseSuccess(ctx context.Context, in CompleteParseSuccessInput) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if len(in.ParsedSummary) == 0 {
		in.ParsedSummary = []byte(`{}`)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin resume parse success: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
update resume_assets
set parse_status = $1, parsed_summary = $2, parsed_text_snapshot = $3, error_code = null, updated_at = $4
where id = $5 and user_id = $6 and parse_status = $7 and deleted_at is null`,
		string(sharedtypes.TargetJobParseStatusReady),
		in.ParsedSummary,
		in.ParsedTextSnapshot,
		now,
		in.AssetID,
		in.UserID,
		string(sharedtypes.TargetJobParseStatusProcessing),
	)
	if err != nil {
		return fmt.Errorf("complete resume parse success: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("complete resume parse success rows affected: %w", err)
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
		string(events.EventNameResumeParseCompleted),
		string(api.ResourceTypeResumeAsset),
		in.AssetID,
		in.OutboxEventPayload,
		now,
	); err != nil {
		return fmt.Errorf("insert resume parse completed outbox: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit resume parse success: %w", err)
	}
	return nil
}

func (r *Repository) CompleteParseFailure(ctx context.Context, in CompleteParseFailureInput) error {
	return r.MarkFailed(ctx, MarkFailedInput{UserID: in.UserID, AssetID: in.AssetID, ErrorCode: in.ErrorCode, Now: in.Now})
}

func (r *Repository) DeleteForUser(ctx context.Context, userID string, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("resume store db is nil")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
update resume_assets
set deleted_at = coalesce(deleted_at, $2), updated_at = $2
where user_id = $1 and deleted_at is null`,
		userID,
		now,
	)
	if err != nil {
		return fmt.Errorf("delete resume assets for user: %w", err)
	}
	return nil
}

func (r *Repository) updateStatus(ctx context.Context, userID string, assetID string, from sharedtypes.TargetJobParseStatus, to sharedtypes.TargetJobParseStatus, parsedSummary json.RawMessage, parsedTextSnapshot *string, errorCode string, now time.Time) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("resume store db is nil")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	var res sql.Result
	var err error
	switch {
	case to == sharedtypes.TargetJobParseStatusProcessing:
		res, err = r.db.ExecContext(ctx, `
update resume_assets
set parse_status = $1, error_code = null, updated_at = $2
where id = $3 and user_id = $4 and parse_status in ('queued','failed') and deleted_at is null`,
			string(to), now, assetID, userID,
		)
	case to == sharedtypes.TargetJobParseStatusReady:
		if len(parsedSummary) == 0 {
			parsedSummary = []byte(`{}`)
		}
		res, err = r.db.ExecContext(ctx, `
update resume_assets
set parse_status = $1, parsed_summary = $2, parsed_text_snapshot = $3, error_code = null, updated_at = $4
where id = $5 and user_id = $6 and parse_status = $7 and deleted_at is null`,
			string(to), parsedSummary, nullableStringPtr(parsedTextSnapshot), now, assetID, userID, string(from),
		)
	case to == sharedtypes.TargetJobParseStatusFailed:
		res, err = r.db.ExecContext(ctx, `
update resume_assets
set parse_status = $1, error_code = $2, updated_at = $3
where id = $4 and user_id = $5 and parse_status in ('queued','processing') and deleted_at is null`,
			string(to), nullableString(errorCode), now, assetID, userID,
		)
	default:
		res, err = r.db.ExecContext(ctx, `
update resume_assets
set parse_status = $1, updated_at = $2
where id = $3 and user_id = $4 and parse_status = $5 and deleted_at is null`,
			string(to), now, assetID, userID, string(from),
		)
	}
	if err != nil {
		return fmt.Errorf("update resume parse status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update resume parse status rows affected: %w", err)
	}
	if rows == 0 {
		return ErrInvalidStateTransition
	}
	return nil
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

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAsset(row rowScanner) (AssetRecord, error) {
	var rec AssetRecord
	var fileObjectID, originalText, parsedTextSnapshot, sourceType, errorCode, latestParseJobID sql.NullString
	var parsedSummary, guidedAnswers []byte
	var deletedAt sql.NullTime
	var parseStatus string
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&fileObjectID,
		&rec.Title,
		&rec.Language,
		&parseStatus,
		&parsedSummary,
		&originalText,
		&guidedAnswers,
		&parsedTextSnapshot,
		&sourceType,
		&errorCode,
		&latestParseJobID,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&deletedAt,
	); err != nil {
		return AssetRecord{}, err
	}
	rec.ParseStatus = sharedtypes.TargetJobParseStatus(parseStatus)
	rec.FileObjectID = stringPtrFromNull(fileObjectID)
	rec.OriginalText = stringPtrFromNull(originalText)
	rec.ParsedTextSnapshot = stringPtrFromNull(parsedTextSnapshot)
	rec.SourceType = stringPtrFromNull(sourceType)
	rec.ErrorCode = stringPtrFromNull(errorCode)
	rec.LatestParseJobID = stringPtrFromNull(latestParseJobID)
	if len(parsedSummary) == 0 {
		parsedSummary = []byte(`{}`)
	}
	rec.ParsedSummary = append(json.RawMessage(nil), parsedSummary...)
	if len(guidedAnswers) > 0 {
		rec.GuidedAnswers = append(json.RawMessage(nil), guidedAnswers...)
	}
	if deletedAt.Valid {
		rec.DeletedAt = &deletedAt.Time
	}
	return rec, nil
}

func stringPtrFromNull(in sql.NullString) *string {
	if !in.Valid {
		return nil
	}
	return &in.String
}

var (
	ErrAssetNotFound                 = errors.New("resume asset not found")
	ErrAssetParseNotReady            = errors.New("resume asset parse is not ready")
	ErrStructuredMasterAlreadyExists = errors.New("structured master resume version already exists")
	ErrVersionNotFound               = errors.New("resume version not found")
	ErrSuggestionNotFound            = errors.New("resume version suggestion not found")
	ErrSuggestionAlreadyDecided      = errors.New("resume version suggestion already decided")
	ErrTailorRunNotFound             = errors.New("resume tailor run not found")
	ErrInvalidStateTransition        = errors.New("invalid resume parse status transition")
	ErrInvalidCursor                 = errors.New("invalid resume list cursor")
)

func encodeCursor(updatedAt time.Time, id string) string {
	raw := updatedAt.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(cursor))
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
		return time.Time{}, "", fmt.Errorf("cursor missing separator")
	}
	updatedAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return updatedAt, parts[1], nil
}
