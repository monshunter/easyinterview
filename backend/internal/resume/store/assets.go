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

	"github.com/monshunter/easyinterview/backend/internal/shared/events"
	"github.com/monshunter/easyinterview/backend/internal/shared/jobs"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

const resumeAggregateType = "resume"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetForParse(ctx context.Context, resumeID string) (ParseAssetRecord, error) {
	if r == nil || r.db == nil {
		return ParseAssetRecord{}, fmt.Errorf("resume store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select rs.id, rs.user_id, rs.language, rs.parse_status, coalesce(rs.source_type, ''),
       coalesce(rs.original_text, ''),
       coalesce(rs.file_object_id::text, ''), coalesce(fo.object_key, '')
from resumes rs
left join file_objects fo on fo.id = rs.file_object_id and fo.deleted_at is null
where rs.id = $1 and rs.deleted_at is null`,
		resumeID,
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

const resumeSelectColumns = `id, user_id, file_object_id, title, display_name, language, parse_status,
       parsed_summary, original_text, structured_profile, parsed_text_snapshot,
       source_type, error_code, latest_parse_job_id, created_at, updated_at, deleted_at`

func (r *Repository) Get(ctx context.Context, userID string, resumeID string) (ResumeRecord, error) {
	if r == nil || r.db == nil {
		return ResumeRecord{}, fmt.Errorf("resume store db is nil")
	}
	row := r.db.QueryRowContext(ctx, `
select `+resumeSelectColumns+`
from resumes
where id = $1 and user_id = $2 and deleted_at is null`,
		resumeID,
		userID,
	)
	rec, err := scanResume(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ResumeRecord{}, ErrAssetNotFound
	}
	if err != nil {
		return ResumeRecord{}, err
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
select ` + resumeSelectColumns + `
from resumes
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
	items := make([]ResumeRecord, 0, pageSize)
	for rows.Next() {
		rec, err := scanResume(rows)
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

// UpdateResume overwrites only the editable fields present in the request on an
// existing resume (D-20 C-17: save accepted rewrites by overwrite).
func (r *Repository) UpdateResume(ctx context.Context, in UpdateResumeInput) (ResumeRecord, error) {
	if r == nil || r.db == nil {
		return ResumeRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	row := r.db.QueryRowContext(ctx, `
update resumes
set structured_profile = case when $1 then $2 else structured_profile end,
    display_name = coalesce($3, display_name),
    updated_at = $4
where id = $5 and user_id = $6 and deleted_at is null
returning `+resumeSelectColumns,
		in.StructuredProfileSet,
		in.StructuredProfile,
		nullableStringPtr(in.DisplayName),
		now,
		in.ResumeID,
		in.UserID,
	)
	rec, err := scanResume(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ResumeRecord{}, ErrAssetNotFound
	}
	if err != nil {
		return ResumeRecord{}, err
	}
	return rec, nil
}

// DuplicateResume copies the source resume's read-only snapshot into a new
// resume row, applying the supplied structured_profile / display_name and a new
// id (D-20 C-18: save accepted rewrites as a new resume).
func (r *Repository) DuplicateResume(ctx context.Context, in DuplicateResumeInput) (ResumeRecord, error) {
	if r == nil || r.db == nil {
		return ResumeRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("begin resume duplicate: %w", err)
	}
	defer tx.Rollback()

	source, err := scanResume(tx.QueryRowContext(ctx, `
select `+resumeSelectColumns+`
from resumes
where id = $1 and user_id = $2 and deleted_at is null`,
		in.SourceResumeID,
		in.UserID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return ResumeRecord{}, ErrAssetNotFound
	}
	if err != nil {
		return ResumeRecord{}, err
	}

	structuredProfile := in.StructuredProfile
	if !in.StructuredProfileSet {
		structuredProfile = source.StructuredProfile
	}
	if len(structuredProfile) == 0 {
		structuredProfile = json.RawMessage(`{}`)
	}
	displayName := source.DisplayName
	if in.DisplayName != nil {
		displayName = in.DisplayName
	}
	if _, err := tx.ExecContext(ctx, `
insert into resumes (
  id, user_id, file_object_id, title, display_name, language, parse_status,
  source_type, original_text, parsed_summary, parsed_text_snapshot,
  structured_profile, latest_parse_job_id, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,null,$13,$13)`,
		in.NewResumeID,
		in.UserID,
		nullableStringPtr(source.FileObjectID),
		source.Title,
		nullableStringPtr(displayName),
		source.Language,
		string(source.ParseStatus),
		nullableStringPtr(source.SourceType),
		nullableStringPtr(source.OriginalText),
		nonEmptyJSON(source.ParsedSummary),
		nullableStringPtr(source.ParsedTextSnapshot),
		structuredProfile,
		now,
	); err != nil {
		return ResumeRecord{}, fmt.Errorf("insert duplicated resume: %w", err)
	}
	rec, err := scanResume(tx.QueryRowContext(ctx, `
select `+resumeSelectColumns+`
from resumes
where id = $1 and user_id = $2`,
		in.NewResumeID,
		in.UserID,
	))
	if err != nil {
		return ResumeRecord{}, fmt.Errorf("read duplicated resume: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return ResumeRecord{}, fmt.Errorf("commit resume duplicate: %w", err)
	}
	return rec, nil
}

func (r *Repository) ArchiveResume(ctx context.Context, in ArchiveResumeInput) (ResumeRecord, error) {
	if r == nil || r.db == nil {
		return ResumeRecord{}, fmt.Errorf("resume store db is nil")
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	rec, err := scanResume(r.db.QueryRowContext(ctx, `
update resumes
set deleted_at = $1, updated_at = $1
where id = $2 and user_id = $3 and deleted_at is null
returning `+resumeSelectColumns,
		now,
		in.ResumeID,
		in.UserID,
	))
	if err == nil {
		return rec, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return ResumeRecord{}, err
	}
	var deletedAt sql.NullTime
	err = r.db.QueryRowContext(ctx, `
select deleted_at
from resumes
where id = $1 and user_id = $2`,
		in.ResumeID,
		in.UserID,
	).Scan(&deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ResumeRecord{}, ErrAssetNotFound
	}
	if err != nil {
		return ResumeRecord{}, err
	}
	if deletedAt.Valid {
		return ResumeRecord{}, ErrAlreadyArchived
	}
	return ResumeRecord{}, ErrAssetNotFound
}

func (r *Repository) MarkParsing(ctx context.Context, in StatusUpdateInput) error {
	return r.updateStatus(ctx, in.UserID, in.AssetID, "", sharedtypes.TargetJobParseStatusProcessing, nil, nil, nil, "", in.Now)
}

func (r *Repository) MarkReady(ctx context.Context, in MarkReadyInput) error {
	return r.updateStatus(ctx, in.UserID, in.AssetID, sharedtypes.TargetJobParseStatusProcessing, sharedtypes.TargetJobParseStatusReady, in.ParsedSummary, in.StructuredProfile, &in.ParsedTextSnapshot, "", in.Now)
}

func (r *Repository) MarkFailed(ctx context.Context, in MarkFailedInput) error {
	return r.updateStatus(ctx, in.UserID, in.AssetID, "", sharedtypes.TargetJobParseStatusFailed, nil, nil, nil, in.ErrorCode, in.Now)
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
	structuredProfile := in.StructuredProfile
	if len(structuredProfile) == 0 {
		structuredProfile = []byte(`{}`)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin resume parse success: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
update resumes
set parse_status = $1, parsed_summary = $2, structured_profile = $3, parsed_text_snapshot = $4,
    display_name = coalesce($5, display_name), error_code = null, updated_at = $6
where id = $7 and user_id = $8 and parse_status = $9 and deleted_at is null`,
		string(sharedtypes.TargetJobParseStatusReady),
		in.ParsedSummary,
		structuredProfile,
		in.ParsedTextSnapshot,
		nullableStringPtr(in.DisplayName),
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
		resumeAggregateType,
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
update resumes
set deleted_at = coalesce(deleted_at, $2), updated_at = $2
where user_id = $1 and deleted_at is null`,
		userID,
		now,
	)
	if err != nil {
		return fmt.Errorf("delete resumes for user: %w", err)
	}
	return nil
}

func (r *Repository) updateStatus(ctx context.Context, userID string, resumeID string, from sharedtypes.TargetJobParseStatus, to sharedtypes.TargetJobParseStatus, parsedSummary json.RawMessage, structuredProfile json.RawMessage, parsedTextSnapshot *string, errorCode string, now time.Time) error {
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
update resumes
set parse_status = $1, error_code = null, updated_at = $2
where id = $3 and user_id = $4 and parse_status in ('queued','failed') and deleted_at is null`,
			string(to), now, resumeID, userID,
		)
	case to == sharedtypes.TargetJobParseStatusReady:
		if len(parsedSummary) == 0 {
			parsedSummary = []byte(`{}`)
		}
		if len(structuredProfile) == 0 {
			structuredProfile = []byte(`{}`)
		}
		res, err = r.db.ExecContext(ctx, `
update resumes
set parse_status = $1, parsed_summary = $2, structured_profile = $3, parsed_text_snapshot = $4, error_code = null, updated_at = $5
where id = $6 and user_id = $7 and parse_status = $8 and deleted_at is null`,
			string(to), parsedSummary, structuredProfile, nullableStringPtr(parsedTextSnapshot), now, resumeID, userID, string(from),
		)
	case to == sharedtypes.TargetJobParseStatusFailed:
		res, err = r.db.ExecContext(ctx, `
update resumes
set parse_status = $1, error_code = $2, updated_at = $3
where id = $4 and user_id = $5 and parse_status in ('queued','processing') and deleted_at is null`,
			string(to), nullableString(errorCode), now, resumeID, userID,
		)
	default:
		res, err = r.db.ExecContext(ctx, `
update resumes
set parse_status = $1, updated_at = $2
where id = $3 and user_id = $4 and parse_status = $5 and deleted_at is null`,
			string(to), now, resumeID, userID, string(from),
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
		return CreateAssetResult{}, fmt.Errorf("begin resume create: %w", err)
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

	if _, err := tx.ExecContext(ctx, `
	insert into resumes (
	  id, user_id, file_object_id, title, display_name, language, parse_status,
	  source_type, original_text, latest_parse_job_id, created_at, updated_at
	) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		in.AssetID,
		in.UserID,
		nullableStringPtr(in.FileObjectID),
		in.Title,
		nil,
		in.Language,
		string(parseStatus),
		in.SourceType,
		nullableString(in.RawText),
		in.JobID,
		now,
		now,
	); err != nil {
		return CreateAssetResult{}, fmt.Errorf("insert resume: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"resumeId":   in.AssetID,
		"userId":     in.UserID,
		"sourceType": in.SourceType,
		"request":    in.RequestPayload,
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
		string(resourceTypeResume),
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
		return CreateAssetResult{}, fmt.Errorf("commit resume create: %w", err)
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

func nonEmptyJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return []byte(`{}`)
	}
	return []byte(raw)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanResume(row rowScanner) (ResumeRecord, error) {
	var rec ResumeRecord
	var fileObjectID, displayName, originalText, parsedTextSnapshot, sourceType, errorCode, latestParseJobID sql.NullString
	var parsedSummary, structuredProfile []byte
	var deletedAt sql.NullTime
	var parseStatus string
	if err := row.Scan(
		&rec.ID,
		&rec.UserID,
		&fileObjectID,
		&rec.Title,
		&displayName,
		&rec.Language,
		&parseStatus,
		&parsedSummary,
		&originalText,
		&structuredProfile,
		&parsedTextSnapshot,
		&sourceType,
		&errorCode,
		&latestParseJobID,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&deletedAt,
	); err != nil {
		return ResumeRecord{}, err
	}
	rec.ParseStatus = sharedtypes.TargetJobParseStatus(parseStatus)
	rec.FileObjectID = stringPtrFromNull(fileObjectID)
	rec.DisplayName = stringPtrFromNull(displayName)
	rec.OriginalText = stringPtrFromNull(originalText)
	rec.ParsedTextSnapshot = stringPtrFromNull(parsedTextSnapshot)
	rec.SourceType = stringPtrFromNull(sourceType)
	rec.ErrorCode = stringPtrFromNull(errorCode)
	rec.LatestParseJobID = stringPtrFromNull(latestParseJobID)
	if len(parsedSummary) == 0 {
		parsedSummary = []byte(`{}`)
	}
	rec.ParsedSummary = append(json.RawMessage(nil), parsedSummary...)
	if len(structuredProfile) == 0 {
		structuredProfile = []byte(`{}`)
	}
	rec.StructuredProfile = append(json.RawMessage(nil), structuredProfile...)
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
	ErrAssetNotFound          = errors.New("resume not found")
	ErrAlreadyArchived        = errors.New("resume already archived")
	ErrTailorRunNotFound      = errors.New("resume tailor run not found")
	ErrInvalidStateTransition = errors.New("invalid resume parse status transition")
	ErrInvalidCursor          = errors.New("invalid resume list cursor")
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
