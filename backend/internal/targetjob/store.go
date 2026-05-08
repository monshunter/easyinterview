package targetjob

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

// SourceType mirrors B4 baseline check list for target_jobs.source_type and
// target_job_sources.source_type (see migrations/000001_create_baseline.up.sql).
type SourceType string

const (
	SourceTypeManualText SourceType = "manual_text"
	SourceTypeURL        SourceType = "url"
	SourceTypeFile       SourceType = "file"
	SourceTypeManualForm SourceType = "manual_form"
)

// RequirementKind mirrors B4 baseline check list for target_job_requirements.kind.
type RequirementKind string

const (
	RequirementMustHave       RequirementKind = "must_have"
	RequirementNiceToHave     RequirementKind = "nice_to_have"
	RequirementHiddenSignal   RequirementKind = "hidden_signal"
	RequirementInterviewFocus RequirementKind = "interview_focus"
)

// EvidenceLevel mirrors B4 baseline check list for target_job_requirements.evidence_level.
type EvidenceLevel string

const (
	EvidenceExplicit EvidenceLevel = "explicit"
	EvidenceInferred EvidenceLevel = "inferred"
)

// FreshnessStatus mirrors B4 baseline check list for target_job_sources.freshness_status.
type FreshnessStatus string

const (
	FreshnessFresh   FreshnessStatus = "fresh"
	FreshnessStale   FreshnessStatus = "stale"
	FreshnessExpired FreshnessStatus = "expired"
)

// ListMaxPageSize bounds the per-page row count regardless of caller request.
const ListMaxPageSize = 100

// ErrTargetJobNotFound is returned by reads / writes that would touch a row
// the caller is not permitted to see (cross-user) or that has been soft
// deleted. Per spec §3.1 D-9 handlers map this to HTTP 404 + B1
// TARGET_JOB_NOT_FOUND, never to FORBIDDEN.
var ErrTargetJobNotFound = errors.New("target job not found")

// TargetJobRecord is the in-process representation of a row in target_jobs.
type TargetJobRecord struct {
	ID                     string
	UserID                 string
	ProfileID              string
	Status                 sharedtypes.TargetJobStatus
	AnalysisStatus         sharedtypes.TargetJobParseStatus
	Title                  string
	CompanyName            string
	LocationText           string
	EmploymentType         string
	SeniorityLevel         string
	TargetLanguage         string
	SourceType             SourceType
	SourceURL              string
	SourceFileObjectID     string
	RawJDText              string
	Summary                json.RawMessage
	FitSummary             json.RawMessage
	Notes                  string
	LatestReportID         string
	OpenQuestionIssueCount int32
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              *time.Time
}

// RequirementRecord represents a row in target_job_requirements.
type RequirementRecord struct {
	ID            string
	TargetJobID   string
	Kind          RequirementKind
	Label         string
	Description   string
	EvidenceLevel EvidenceLevel
	DisplayOrder  int32
	CreatedAt     time.Time
}

// SourceRecord represents a row in target_job_sources.
type SourceRecord struct {
	ID              string
	TargetJobID     string
	SourceType      SourceType
	URL             string
	FileObjectID    string
	SnapshotText    string
	FetchedAt       *time.Time
	FreshnessStatus FreshnessStatus
	CreatedAt       time.Time
}

// ListFilter is the read-side filter passed to ListTargetJobsForUser. The
// store enforces user_id scope and deleted_at filtering on top of these.
type ListFilter struct {
	Status         *sharedtypes.TargetJobStatus
	AnalysisStatus *sharedtypes.TargetJobParseStatus
	SearchQuery    string
	Cursor         string
	PageSize       int32
}

// ListResult captures one page of TargetJob rows plus a cursor pointing at
// the next page. The cursor is opaque base64url(updated_at + id).
type ListResult struct {
	Items      []TargetJobRecord
	NextCursor string
	HasMore    bool
}

// UpdateLifecycleFields is the optional-field update payload for
// UpdateTargetJobLifecycle. Nil pointers leave the field untouched.
type UpdateLifecycleFields struct {
	Status          *sharedtypes.TargetJobStatus
	LocationText    *string
	Notes           *string
	TitleHint       *string
	CompanyNameHint *string
	DedupeKey       string
	DedupeMarkerID  string
}

// ApplyParseResultInput is the parse-pipeline output the store will merge in
// a single transaction with the target_jobs analysis-status / summary fields.
type ApplyParseResultInput struct {
	TargetJobID      string
	AnalysisStatus   sharedtypes.TargetJobParseStatus
	Summary          json.RawMessage
	FitSummary       json.RawMessage
	LatestParseJobID string
	Requirements     []RequirementRecord
	Now              time.Time
}

// ImportTargetJobInput is the input to the compound ImportTargetJob store
// method. The service layer pre-generates IDs, chooses the right initial
// status and runner-vs-synchronous path, and hands the assembled input here.
//
// Source-type specific contracts:
//   - url / manual_text / file: InitialAnalysisStatus must be queued, the
//     RunnerJob fields (JobID / OutboxEventID / Payloads) must be filled,
//     and DraftRequirements must be empty.
//   - manual_form: InitialAnalysisStatus must be ready, RunnerJob fields
//     must be empty, and DraftRequirements must contain at least one entry.
type ImportTargetJobInput struct {
	UserID    string
	DedupeKey string

	TargetJobID            string
	Title                  string
	CompanyName            string
	LocationText           string
	EmploymentType         string
	SeniorityLevel         string
	TargetLanguage         string
	APISourceType          SourceType
	SourceURL              string
	SourceFileObjectID     string
	RawJDText              string
	InitialLifecycleStatus sharedtypes.TargetJobStatus
	InitialAnalysisStatus  sharedtypes.TargetJobParseStatus

	// SourceRecord fields. For manual_form, leave SourceID empty to skip the
	// target_job_sources insert (per plan §3.1).
	SourceID           string
	SourceSnapshotText string
	SourceFetchedAt    *time.Time

	// Runner-bound async path (url / manual_text / file). All four must be
	// set together; for manual_form they must all be empty.
	JobID              string
	OutboxEventID      string
	OutboxEventPayload []byte
	JobPayload         []byte

	// Manual form synchronous-ready path. Empty for runner-bound source
	// variants.
	DraftRequirements []RequirementRecord

	Now time.Time
}

// ImportTargetJobResult is what the service layer returns to the handler so
// it can shape the generated TargetJobWithJob response.
type ImportTargetJobResult struct {
	TargetJobID  string
	JobID        string
	JobStatus    sharedtypes.JobStatus
	JobCreatedAt time.Time
	JobUpdatedAt time.Time
	Existing     bool
}

// ClaimedJob is the structured handoff between the store-level claim
// query and the drainer's per-job handler. The drainer treats async_jobs
// rows as the source of truth for retry / completion bookkeeping.
type ClaimedJob struct {
	JobID        string
	JobType      string
	ResourceType string
	ResourceID   string
	Payload      []byte
	Attempts     int32
	MaxAttempts  int32
	AvailableAt  time.Time
}

// JobOutcome captures the parse-pipeline result so the drainer can update
// async_jobs.status / error_code / completed_at consistently.
type JobOutcome struct {
	Succeeded    bool
	ErrorCode    string
	ErrorMessage string
	Retryable    bool
}

// FileAttachmentRecord is the minimal projection of file_objects needed by
// the import flow to confirm a referenced upload belongs to the caller and
// has the expected purpose.
type FileAttachmentRecord struct {
	ID      string
	UserID  string
	Purpose string
}

// Store is the persistence boundary for the target_jobs / target_job_requirements
// / target_job_sources tables. Every read / write filters by user_id; cross-user
// or soft-deleted rows surface as ErrTargetJobNotFound (handler maps to 404 +
// TARGET_JOB_NOT_FOUND).
type Store interface {
	ImportTargetJob(ctx context.Context, in ImportTargetJobInput) (ImportTargetJobResult, error)
	InsertTargetJob(ctx context.Context, rec TargetJobRecord) error
	InsertTargetJobSource(ctx context.Context, rec SourceRecord) error
	GetTargetJobByUser(ctx context.Context, userID string, targetJobID string) (TargetJobRecord, []RequirementRecord, []SourceRecord, error)
	ListTargetJobsForUser(ctx context.Context, userID string, filter ListFilter) (ListResult, error)
	LookupUpdateDedupe(ctx context.Context, userID string, dedupeKey string) (TargetJobRecord, []RequirementRecord, bool, error)
	UpdateTargetJobLifecycle(ctx context.Context, userID string, targetJobID string, fields UpdateLifecycleFields, now time.Time) (TargetJobRecord, error)
	ApplyParseResult(ctx context.Context, in ApplyParseResultInput) error
	UpdateSourceFreshness(ctx context.Context, targetJobID string, freshness FreshnessStatus, now time.Time) error
	UpdateSourceSnapshot(ctx context.Context, sourceID string, sanitizedURL string, snapshotText string, fetchedAt time.Time, now time.Time) error
	LookupFileAttachmentForUser(ctx context.Context, userID string, fileObjectID string) (FileAttachmentRecord, error)
	ClaimNextAsyncJob(ctx context.Context, jobTypes []string, now time.Time) (ClaimedJob, bool, error)
	FinalizeAsyncJob(ctx context.Context, jobID string, outcome JobOutcome, now time.Time) error
	EnqueueSourceRefresh(ctx context.Context, jobID string, targetJobID string, now time.Time) error
	WriteParseFailedOutbox(ctx context.Context, eventID string, targetJobID string, payload []byte, now time.Time) error
	WriteTargetParsedOutbox(ctx context.Context, eventID string, targetJobID string, payload []byte, now time.Time) error
	GetTargetJobForParse(ctx context.Context, targetJobID string) (TargetJobRecord, []SourceRecord, error)
	UpdateTargetJobAnalysisFailure(ctx context.Context, targetJobID string, now time.Time) error
}

// SQLStore is the default Postgres-backed Store implementation.
type SQLStore struct {
	db *sql.DB
}

// NewSQLStore wires a SQLStore against the given *sql.DB.
func NewSQLStore(db *sql.DB) *SQLStore { return &SQLStore{db: db} }

func (s *SQLStore) checkDB() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("targetjob store db is nil")
	}
	return nil
}

func (s *SQLStore) InsertTargetJob(ctx context.Context, rec TargetJobRecord) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	summary := nullJSON(rec.Summary)
	fitSummary := nullJSON(rec.FitSummary)
	_, err := s.db.ExecContext(ctx, `
insert into target_jobs (
  id,
  user_id,
  profile_id,
  status,
  analysis_status,
  title,
  company_name,
  location_text,
  employment_type,
  seniority_level,
  target_language,
  source_type,
  source_url,
  source_file_object_id,
  raw_jd_text,
  summary,
  fit_summary,
  notes,
  open_question_issue_count,
  created_at,
  updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		rec.ID,
		nullableUUID(rec.UserID),
		nullableUUID(rec.ProfileID),
		string(rec.Status),
		string(rec.AnalysisStatus),
		nullableString(rec.Title),
		nullableString(rec.CompanyName),
		nullableString(rec.LocationText),
		nullableString(rec.EmploymentType),
		nullableString(rec.SeniorityLevel),
		rec.TargetLanguage,
		string(rec.SourceType),
		nullableString(rec.SourceURL),
		nullableUUID(rec.SourceFileObjectID),
		nullableString(rec.RawJDText),
		summary,
		fitSummary,
		nullableString(rec.Notes),
		rec.OpenQuestionIssueCount,
		rec.CreatedAt,
		rec.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert target_jobs: %w", err)
	}
	return nil
}

func (s *SQLStore) InsertTargetJobSource(ctx context.Context, rec SourceRecord) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	freshness := rec.FreshnessStatus
	if freshness == "" {
		freshness = FreshnessFresh
	}
	var fetchedAt any
	if rec.FetchedAt != nil {
		fetchedAt = *rec.FetchedAt
	}
	_, err := s.db.ExecContext(ctx, `
insert into target_job_sources (
  id,
  target_job_id,
  source_type,
  url,
  file_object_id,
  snapshot_text,
  fetched_at,
  freshness_status,
  created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		rec.ID,
		rec.TargetJobID,
		string(rec.SourceType),
		nullableString(rec.URL),
		nullableUUID(rec.FileObjectID),
		nullableString(rec.SnapshotText),
		fetchedAt,
		string(freshness),
		rec.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert target_job_sources: %w", err)
	}
	return nil
}

func (s *SQLStore) GetTargetJobByUser(ctx context.Context, userID string, targetJobID string) (TargetJobRecord, []RequirementRecord, []SourceRecord, error) {
	if err := s.checkDB(); err != nil {
		return TargetJobRecord{}, nil, nil, err
	}
	rec := TargetJobRecord{ID: targetJobID, UserID: userID}
	var (
		profileID          sql.NullString
		title              sql.NullString
		companyName        sql.NullString
		locationText       sql.NullString
		employmentType     sql.NullString
		seniorityLevel     sql.NullString
		sourceURL          sql.NullString
		sourceFileObjectID sql.NullString
		rawJDText          sql.NullString
		notes              sql.NullString
		latestReportID     sql.NullString
		summary            []byte
		fitSummary         []byte
		status             string
		analysisStatus     string
		sourceType         string
	)
	err := s.db.QueryRowContext(ctx, `
select id, user_id, profile_id, status, analysis_status, title, company_name, location_text,
       employment_type, seniority_level, target_language, source_type, source_url, source_file_object_id,
       raw_jd_text, summary, fit_summary, notes, latest_report_id, open_question_issue_count,
       created_at, updated_at
from target_jobs
where id = $1 and user_id = $2 and deleted_at is null`,
		targetJobID,
		userID,
	).Scan(
		&rec.ID,
		&rec.UserID,
		&profileID,
		&status,
		&analysisStatus,
		&title,
		&companyName,
		&locationText,
		&employmentType,
		&seniorityLevel,
		&rec.TargetLanguage,
		&sourceType,
		&sourceURL,
		&sourceFileObjectID,
		&rawJDText,
		&summary,
		&fitSummary,
		&notes,
		&latestReportID,
		&rec.OpenQuestionIssueCount,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return TargetJobRecord{}, nil, nil, ErrTargetJobNotFound
	}
	if err != nil {
		return TargetJobRecord{}, nil, nil, fmt.Errorf("select target_jobs: %w", err)
	}
	rec.ProfileID = profileID.String
	rec.Status = sharedtypes.TargetJobStatus(status)
	rec.AnalysisStatus = sharedtypes.TargetJobParseStatus(analysisStatus)
	rec.Title = title.String
	rec.CompanyName = companyName.String
	rec.LocationText = locationText.String
	rec.EmploymentType = employmentType.String
	rec.SeniorityLevel = seniorityLevel.String
	rec.SourceType = SourceType(sourceType)
	rec.SourceURL = sourceURL.String
	rec.SourceFileObjectID = sourceFileObjectID.String
	rec.RawJDText = rawJDText.String
	rec.Notes = notes.String
	rec.LatestReportID = latestReportID.String
	if len(summary) > 0 {
		rec.Summary = append(json.RawMessage{}, summary...)
	}
	if len(fitSummary) > 0 {
		rec.FitSummary = append(json.RawMessage{}, fitSummary...)
	}

	requirements, err := s.listRequirementsForJob(ctx, s.db, targetJobID)
	if err != nil {
		return TargetJobRecord{}, nil, nil, err
	}
	sources, err := s.listSourcesForJob(ctx, s.db, targetJobID)
	if err != nil {
		return TargetJobRecord{}, nil, nil, err
	}
	return rec, requirements, sources, nil
}

func (s *SQLStore) ListTargetJobsForUser(ctx context.Context, userID string, filter ListFilter) (ListResult, error) {
	if err := s.checkDB(); err != nil {
		return ListResult{}, err
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > ListMaxPageSize {
		pageSize = ListMaxPageSize
	}
	var (
		args    []any
		clauses []string
		nextArg = func(v any) string {
			args = append(args, v)
			return fmt.Sprintf("$%d", len(args))
		}
	)
	args = append(args, userID)
	clauses = append(clauses, "user_id = $1", "deleted_at is null")
	if filter.Status != nil {
		clauses = append(clauses, "status = "+nextArg(string(*filter.Status)))
	}
	if filter.AnalysisStatus != nil {
		clauses = append(clauses, "analysis_status = "+nextArg(string(*filter.AnalysisStatus)))
	}
	if q := strings.TrimSpace(filter.SearchQuery); q != "" {
		clauses = append(clauses,
			"to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(company_name,'')) @@ plainto_tsquery('simple', "+nextArg(q)+")")
	}
	if filter.Cursor != "" {
		updatedAt, id, err := decodeCursor(filter.Cursor)
		if err != nil {
			return ListResult{}, fmt.Errorf("decode cursor: %w", err)
		}
		uPlaceholder := nextArg(updatedAt)
		idPlaceholder := nextArg(id)
		clauses = append(clauses, fmt.Sprintf("(updated_at, id) < (%s, %s)", uPlaceholder, idPlaceholder))
	}

	limitArg := nextArg(int(pageSize) + 1)
	query := `
select id, user_id, profile_id, status, analysis_status, title, company_name, location_text,
       employment_type, seniority_level, target_language, source_type, source_url, source_file_object_id,
       raw_jd_text, summary, fit_summary, notes, latest_report_id, open_question_issue_count,
       created_at, updated_at
from target_jobs
where ` + strings.Join(clauses, " and ") + `
order by updated_at desc, id desc
limit ` + limitArg

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return ListResult{}, fmt.Errorf("list target_jobs: %w", err)
	}
	defer rows.Close()

	out := ListResult{}
	for rows.Next() {
		rec, err := scanTargetJobRow(rows)
		if err != nil {
			return ListResult{}, err
		}
		out.Items = append(out.Items, rec)
	}
	if err := rows.Err(); err != nil {
		return ListResult{}, fmt.Errorf("list target_jobs rows: %w", err)
	}

	if int32(len(out.Items)) > pageSize {
		last := out.Items[pageSize-1]
		out.Items = out.Items[:pageSize]
		out.HasMore = true
		out.NextCursor = encodeCursor(last.UpdatedAt, last.ID)
	}
	return out, nil
}

func (s *SQLStore) UpdateTargetJobLifecycle(ctx context.Context, userID string, targetJobID string, fields UpdateLifecycleFields, now time.Time) (TargetJobRecord, error) {
	if err := s.checkDB(); err != nil {
		return TargetJobRecord{}, err
	}
	if fields.DedupeKey != "" {
		return s.updateTargetJobLifecycleIdempotent(ctx, userID, targetJobID, fields, now)
	}
	return updateTargetJobLifecycleRow(ctx, s.db, userID, targetJobID, fields, now)
}

// LookupUpdateDedupe checks whether updateTargetJob has already completed for
// this user-scoped dedupe key. It is a preflight convenience for the service;
// the transactional update path still rechecks under an advisory lock to close
// races.
func (s *SQLStore) LookupUpdateDedupe(ctx context.Context, userID string, dedupeKey string) (TargetJobRecord, []RequirementRecord, bool, error) {
	if err := s.checkDB(); err != nil {
		return TargetJobRecord{}, nil, false, err
	}
	targetID, hit, err := lookupExistingUpdateDedupe(ctx, s.db, dedupeKey)
	if err != nil {
		return TargetJobRecord{}, nil, false, err
	}
	if !hit {
		return TargetJobRecord{}, nil, false, nil
	}
	rec, reqs, _, err := s.GetTargetJobByUser(ctx, userID, targetID)
	if err != nil {
		return TargetJobRecord{}, nil, false, err
	}
	return rec, reqs, true, nil
}

func (s *SQLStore) updateTargetJobLifecycleIdempotent(ctx context.Context, userID string, targetJobID string, fields UpdateLifecycleFields, now time.Time) (TargetJobRecord, error) {
	if fields.DedupeMarkerID == "" {
		return TargetJobRecord{}, fmt.Errorf("update target job requires DedupeMarkerID")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return TargetJobRecord{}, fmt.Errorf("begin update target job lifecycle: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `select pg_advisory_xact_lock(hashtext($1))`, fields.DedupeKey); err != nil {
		return TargetJobRecord{}, fmt.Errorf("lock update dedupe key: %w", err)
	}
	existingTargetID, hit, err := lookupExistingUpdateDedupe(ctx, tx, fields.DedupeKey)
	if err != nil {
		return TargetJobRecord{}, err
	}
	if hit {
		rec, err := selectTargetJobRecordByUser(ctx, tx, userID, existingTargetID)
		if err != nil {
			return TargetJobRecord{}, err
		}
		if err := tx.Commit(); err != nil {
			return TargetJobRecord{}, fmt.Errorf("commit update target job dedupe hit: %w", err)
		}
		return rec, nil
	}

	updated, err := updateTargetJobLifecycleRow(ctx, tx, userID, targetJobID, fields, now)
	if err != nil {
		return TargetJobRecord{}, err
	}
	if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, payload, result,
  available_at, completed_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9,$9,$9)`,
		fields.DedupeMarkerID,
		string(jobs.JobTypeTargetImport),
		"target_job_update",
		targetJobID,
		fields.DedupeKey,
		string(sharedtypes.JobStatusSucceeded),
		[]byte(`{}`),
		[]byte(`{}`),
		now,
	); err != nil {
		return TargetJobRecord{}, fmt.Errorf("insert update dedupe marker: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return TargetJobRecord{}, fmt.Errorf("commit update target job lifecycle: %w", err)
	}
	return updated, nil
}

func updateTargetJobLifecycleRow(ctx context.Context, q rowQueryer, userID string, targetJobID string, fields UpdateLifecycleFields, now time.Time) (TargetJobRecord, error) {
	var (
		sets    []string
		args    []any
		nextArg = func(v any) string {
			args = append(args, v)
			return fmt.Sprintf("$%d", len(args))
		}
	)
	if fields.Status != nil {
		sets = append(sets, "status = "+nextArg(string(*fields.Status)))
	}
	if fields.LocationText != nil {
		sets = append(sets, "location_text = "+nextArg(*fields.LocationText))
	}
	if fields.Notes != nil {
		sets = append(sets, "notes = "+nextArg(*fields.Notes))
	}
	if fields.TitleHint != nil {
		sets = append(sets, "title = coalesce(title, "+nextArg(*fields.TitleHint)+")")
	}
	if fields.CompanyNameHint != nil {
		sets = append(sets, "company_name = coalesce(company_name, "+nextArg(*fields.CompanyNameHint)+")")
	}
	sets = append(sets, "updated_at = "+nextArg(now))

	args = append(args, targetJobID, userID)
	query := fmt.Sprintf(`
update target_jobs
set %s
where id = $%d and user_id = $%d and deleted_at is null
returning id, user_id, profile_id, status, analysis_status, title, company_name, location_text,
          employment_type, seniority_level, target_language, source_type, source_url, source_file_object_id,
          raw_jd_text, summary, fit_summary, notes, latest_report_id, open_question_issue_count,
          created_at, updated_at`,
		strings.Join(sets, ", "),
		len(args)-1,
		len(args),
	)
	rows := q.QueryRowContext(ctx, query, args...)
	rec, err := scanTargetJobRow(rows)
	if errors.Is(err, sql.ErrNoRows) {
		return TargetJobRecord{}, ErrTargetJobNotFound
	}
	if err != nil {
		return TargetJobRecord{}, fmt.Errorf("update target_jobs lifecycle: %w", err)
	}
	return rec, nil
}

func lookupExistingUpdateDedupe(ctx context.Context, q rowQueryer, dedupeKey string) (string, bool, error) {
	var resourceID string
	err := q.QueryRowContext(ctx, `
select resource_id
from async_jobs
where job_type = $1 and dedupe_key = $2
order by created_at desc
limit 1`,
		string(jobs.JobTypeTargetImport),
		dedupeKey,
	).Scan(&resourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("lookup update dedupe marker: %w", err)
	}
	return resourceID, true, nil
}

func selectTargetJobRecordByUser(ctx context.Context, q rowQueryer, userID string, targetJobID string) (TargetJobRecord, error) {
	rec, err := scanTargetJobRow(q.QueryRowContext(ctx, `
select id, user_id, profile_id, status, analysis_status, title, company_name, location_text,
       employment_type, seniority_level, target_language, source_type, source_url, source_file_object_id,
       raw_jd_text, summary, fit_summary, notes, latest_report_id, open_question_issue_count,
       created_at, updated_at
from target_jobs
where id = $1 and user_id = $2 and deleted_at is null`,
		targetJobID,
		userID,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return TargetJobRecord{}, ErrTargetJobNotFound
	}
	if err != nil {
		return TargetJobRecord{}, fmt.Errorf("select target_jobs by update dedupe: %w", err)
	}
	return rec, nil
}

func (s *SQLStore) ApplyParseResult(ctx context.Context, in ApplyParseResultInput) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin apply parse result: %w", err)
	}
	defer tx.Rollback()

	existing, err := s.listRequirementsForJobTx(ctx, tx, in.TargetJobID)
	if err != nil {
		return err
	}
	existingByKey := map[string]struct{}{}
	maxOrder := int32(0)
	for _, r := range existing {
		key := string(r.Kind) + "\x00" + r.Label
		existingByKey[key] = struct{}{}
		if r.DisplayOrder > maxOrder {
			maxOrder = r.DisplayOrder
		}
	}

	nextOrder := maxOrder + 1
	for _, req := range in.Requirements {
		key := string(req.Kind) + "\x00" + req.Label
		if _, dup := existingByKey[key]; dup {
			continue
		}
		evidence := req.EvidenceLevel
		if evidence == "" {
			evidence = EvidenceExplicit
		}
		_, err := tx.ExecContext(ctx, `
insert into target_job_requirements (
  id, target_job_id, kind, label, description, evidence_level, display_order, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
			req.ID,
			in.TargetJobID,
			string(req.Kind),
			req.Label,
			nullableString(req.Description),
			string(evidence),
			nextOrder,
			in.Now,
		)
		if err != nil {
			return fmt.Errorf("insert target_job_requirements: %w", err)
		}
		existingByKey[key] = struct{}{}
		nextOrder++
	}

	res, err := tx.ExecContext(ctx, `
update target_jobs
set analysis_status = $1,
    summary = $2,
    fit_summary = $3,
    latest_parse_job_id = $4,
    updated_at = $5
where id = $6 and deleted_at is null`,
		string(in.AnalysisStatus),
		nullJSON(in.Summary),
		nullJSON(in.FitSummary),
		nullableUUID(in.LatestParseJobID),
		in.Now,
		in.TargetJobID,
	)
	if err != nil {
		return fmt.Errorf("update target_jobs analysis_status: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("apply parse result rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTargetJobNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit apply parse result: %w", err)
	}
	return nil
}

// ImportTargetJob writes target_jobs (+ optional target_job_sources +
// target_job_requirements + outbox_events + async_jobs) atomically. It
// dedupes by (user_id, idempotency_key) hashed into in.DedupeKey: if an
// existing async_jobs row with the same dedupe_key and job_type
// jobs.JobTypeTargetImport is found, the result wraps that row instead of inserting
// a new TargetJob (see spec C-12 / D-6).
//
// Source variant contract:
//   - Runner-bound (url / manual_text / file): all of JobID, OutboxEventID,
//     OutboxEventPayload, JobPayload must be set; DraftRequirements must be
//     empty; InitialAnalysisStatus must be queued.
//   - manual_form: JobID and outbox / job payload fields must be empty;
//     DraftRequirements must contain at least one entry; InitialAnalysisStatus
//     must be ready.
func (s *SQLStore) ImportTargetJob(ctx context.Context, in ImportTargetJobInput) (ImportTargetJobResult, error) {
	if err := s.checkDB(); err != nil {
		return ImportTargetJobResult{}, err
	}
	if in.UserID == "" || in.TargetJobID == "" || in.DedupeKey == "" {
		return ImportTargetJobResult{}, fmt.Errorf("import target job requires userId, targetJobId, dedupeKey")
	}

	if existing, hit, err := s.lookupExistingImport(ctx, in.DedupeKey); err != nil {
		return ImportTargetJobResult{}, err
	} else if hit {
		return existing, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ImportTargetJobResult{}, fmt.Errorf("begin import target job: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
insert into target_jobs (
  id, user_id, status, analysis_status, title, company_name, location_text,
  employment_type, seniority_level, target_language, source_type, source_url,
  source_file_object_id, raw_jd_text, summary, fit_summary,
  open_question_issue_count, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		in.TargetJobID,
		in.UserID,
		string(in.InitialLifecycleStatus),
		string(in.InitialAnalysisStatus),
		nullableString(in.Title),
		nullableString(in.CompanyName),
		nullableString(in.LocationText),
		nullableString(in.EmploymentType),
		nullableString(in.SeniorityLevel),
		in.TargetLanguage,
		string(in.APISourceType),
		nullableString(in.SourceURL),
		nullableUUID(in.SourceFileObjectID),
		nullableString(in.RawJDText),
		[]byte(`{}`),
		[]byte(`{}`),
		int32(0),
		in.Now,
		in.Now,
	); err != nil {
		return ImportTargetJobResult{}, fmt.Errorf("insert target_jobs: %w", err)
	}

	if in.SourceID != "" {
		var fetchedAt any
		if in.SourceFetchedAt != nil {
			fetchedAt = *in.SourceFetchedAt
		}
		if _, err := tx.ExecContext(ctx, `
insert into target_job_sources (
  id, target_job_id, source_type, url, file_object_id, snapshot_text,
  fetched_at, freshness_status, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			in.SourceID,
			in.TargetJobID,
			string(in.APISourceType),
			nullableString(in.SourceURL),
			nullableUUID(in.SourceFileObjectID),
			nullableString(in.SourceSnapshotText),
			fetchedAt,
			string(FreshnessFresh),
			in.Now,
		); err != nil {
			return ImportTargetJobResult{}, fmt.Errorf("insert target_job_sources: %w", err)
		}
	}

	for _, req := range in.DraftRequirements {
		evidence := req.EvidenceLevel
		if evidence == "" {
			evidence = EvidenceExplicit
		}
		if _, err := tx.ExecContext(ctx, `
insert into target_job_requirements (
  id, target_job_id, kind, label, description, evidence_level, display_order, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
			req.ID,
			in.TargetJobID,
			string(req.Kind),
			req.Label,
			nullableString(req.Description),
			string(evidence),
			req.DisplayOrder,
			in.Now,
		); err != nil {
			return ImportTargetJobResult{}, fmt.Errorf("insert draft requirement: %w", err)
		}
	}

	if in.JobID == "" {
		return ImportTargetJobResult{}, fmt.Errorf("import target job requires JobID")
	}

	jobStatus := sharedtypes.JobStatusQueued
	runnerBound := in.OutboxEventID != ""
	if !runnerBound {
		jobStatus = sharedtypes.JobStatusSucceeded
	}

	if runnerBound {
		if _, err := tx.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1,$2,$3,$4,$5,$6,$7,$8)`,
			in.OutboxEventID,
			string(events.EventNameTargetImportRequested),
			1,
			"target_job",
			in.TargetJobID,
			in.OutboxEventPayload,
			"pending",
			in.Now,
		); err != nil {
			return ImportTargetJobResult{}, fmt.Errorf("insert outbox_events: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, payload, available_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)`,
			in.JobID,
			string(jobs.JobTypeTargetImport),
			"target_job",
			in.TargetJobID,
			in.DedupeKey,
			string(sharedtypes.JobStatusQueued),
			in.JobPayload,
			in.Now,
			in.Now,
		); err != nil {
			return ImportTargetJobResult{}, fmt.Errorf("insert async_jobs: %w", err)
		}
	} else {
		// manual_form: a terminal succeeded async_jobs row records the
		// import for SELECT-based dedupe and supplies a real Job.id for the
		// response. The unique-active partial index does not cover
		// succeeded rows, so dedupe is best-effort within the synchronous
		// race window — acceptable since manual_form completes inside a
		// single request lifecycle.
		if _, err := tx.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, payload,
  available_at, completed_at, created_at, updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$8,$8,$8)`,
			in.JobID,
			string(jobs.JobTypeTargetImport),
			"target_job",
			in.TargetJobID,
			in.DedupeKey,
			string(sharedtypes.JobStatusSucceeded),
			[]byte(`{}`),
			in.Now,
		); err != nil {
			return ImportTargetJobResult{}, fmt.Errorf("insert manual_form async_jobs marker: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return ImportTargetJobResult{}, fmt.Errorf("commit import target job: %w", err)
	}

	return ImportTargetJobResult{
		TargetJobID:  in.TargetJobID,
		JobID:        in.JobID,
		JobStatus:    jobStatus,
		JobCreatedAt: in.Now,
		JobUpdatedAt: in.Now,
		Existing:     false,
	}, nil
}

func (s *SQLStore) lookupExistingImport(ctx context.Context, dedupeKey string) (ImportTargetJobResult, bool, error) {
	var (
		jobID      string
		resourceID string
		statusStr  string
		createdAt  time.Time
		updatedAt  time.Time
	)
	err := s.db.QueryRowContext(ctx, `
select id, resource_id, status, created_at, updated_at
from async_jobs
where job_type = $1 and dedupe_key = $2
order by created_at desc
limit 1`,
		string(jobs.JobTypeTargetImport),
		dedupeKey,
	).Scan(&jobID, &resourceID, &statusStr, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ImportTargetJobResult{}, false, nil
	}
	if err != nil {
		return ImportTargetJobResult{}, false, fmt.Errorf("lookup existing %s: %w", jobs.JobTypeTargetImport, err)
	}
	return ImportTargetJobResult{
		TargetJobID:  resourceID,
		JobID:        jobID,
		JobStatus:    sharedtypes.JobStatus(statusStr),
		JobCreatedAt: createdAt,
		JobUpdatedAt: updatedAt,
		Existing:     true,
	}, true, nil
}

// ClaimNextAsyncJob atomically picks the oldest queued row whose job_type
// matches one of the provided values and whose available_at <= now. The
// claim flips status to running, increments attempts, and stamps locked_at,
// then returns the claimed row to the drainer. The (false, nil) tuple
// means there is currently nothing to do.
func (s *SQLStore) ClaimNextAsyncJob(ctx context.Context, jobTypes []string, now time.Time) (ClaimedJob, bool, error) {
	if err := s.checkDB(); err != nil {
		return ClaimedJob{}, false, err
	}
	if len(jobTypes) == 0 {
		return ClaimedJob{}, false, fmt.Errorf("ClaimNextAsyncJob requires at least one job type")
	}
	// Render the IN (...) list with positional placeholders.
	placeholders := make([]string, 0, len(jobTypes))
	args := make([]any, 0, len(jobTypes)+1)
	for i, jt := range jobTypes {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, jt)
	}
	args = append(args, now)
	query := fmt.Sprintf(`
update async_jobs
set status = 'running',
    attempts = attempts + 1,
    locked_at = $%[1]d,
    updated_at = $%[1]d
where id = (
  select id from async_jobs
  where status = 'queued' and available_at <= $%[1]d and job_type in (%s)
  order by available_at asc, created_at asc
  for update skip locked
  limit 1
)
returning id, job_type, resource_type, resource_id, payload, attempts, max_attempts, available_at`,
		len(args),
		strings.Join(placeholders, ","),
	)
	var (
		claimed ClaimedJob
		payload []byte
	)
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&claimed.JobID,
		&claimed.JobType,
		&claimed.ResourceType,
		&claimed.ResourceID,
		&payload,
		&claimed.Attempts,
		&claimed.MaxAttempts,
		&claimed.AvailableAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ClaimedJob{}, false, nil
	}
	if err != nil {
		return ClaimedJob{}, false, fmt.Errorf("claim async_jobs: %w", err)
	}
	if len(payload) > 0 {
		claimed.Payload = append([]byte{}, payload...)
	}
	return claimed, true, nil
}

// FinalizeAsyncJob applies the outcome to async_jobs. Failed + retryable
// requeues the row by clearing locked_at and bumping available_at by a
// modest backoff; failed + non-retryable terminates with status='failed'.
func (s *SQLStore) FinalizeAsyncJob(ctx context.Context, jobID string, outcome JobOutcome, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	if jobID == "" {
		return fmt.Errorf("FinalizeAsyncJob requires jobID")
	}
	if outcome.Succeeded {
		_, err := s.db.ExecContext(ctx, `
update async_jobs
set status = 'succeeded',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = null,
    error_message = null
where id = $2`, now, jobID)
		if err != nil {
			return fmt.Errorf("finalize async_jobs succeeded: %w", err)
		}
		return nil
	}
	if outcome.Retryable {
		_, err := s.db.ExecContext(ctx, `
update async_jobs
set status = case when attempts >= max_attempts then 'dead' else 'queued' end,
    available_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = $2,
    error_message = $3
where id = $4`,
			now.Add(15*time.Second),
			outcome.ErrorCode,
			nullableString(outcome.ErrorMessage),
			jobID,
		)
		if err != nil {
			return fmt.Errorf("finalize async_jobs retryable: %w", err)
		}
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
update async_jobs
set status = 'failed',
    completed_at = $1,
    updated_at = $1,
    locked_at = null,
    error_code = $2,
    error_message = $3
where id = $4`,
		now,
		outcome.ErrorCode,
		nullableString(outcome.ErrorMessage),
		jobID,
	)
	if err != nil {
		return fmt.Errorf("finalize async_jobs failed: %w", err)
	}
	return nil
}

// EnqueueSourceRefresh writes the placeholder source_refresh async_jobs row
// (D-3 / plan 4.5). Payload is intentionally empty: the row exists only as
// a downstream trigger for a future refresh implementation.
func (s *SQLStore) EnqueueSourceRefresh(ctx context.Context, jobID string, targetJobID string, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
insert into async_jobs (
  id, job_type, resource_type, resource_id, dedupe_key, status, payload,
  available_at, created_at, updated_at
) values ($1, $2, 'target_job', $3, null, 'queued', '{}'::jsonb, $4, $4, $4)`,
		jobID,
		string(jobs.JobTypeSourceRefresh),
		targetJobID,
		now,
	)
	if err != nil {
		return fmt.Errorf("enqueue source_refresh: %w", err)
	}
	return nil
}

// WriteTargetParsedOutbox inserts an events.EventNameTargetParsed event row.
func (s *SQLStore) WriteTargetParsedOutbox(ctx context.Context, eventID string, targetJobID string, payload []byte, now time.Time) error {
	return s.writeOutbox(ctx, eventID, string(events.EventNameTargetParsed), targetJobID, payload, now)
}

// WriteParseFailedOutbox inserts an events.EventNameTargetAnalysisFailed event row.
func (s *SQLStore) WriteParseFailedOutbox(ctx context.Context, eventID string, targetJobID string, payload []byte, now time.Time) error {
	return s.writeOutbox(ctx, eventID, string(events.EventNameTargetAnalysisFailed), targetJobID, payload, now)
}

func (s *SQLStore) writeOutbox(ctx context.Context, eventID, eventName, targetJobID string, payload []byte, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
insert into outbox_events (
  id, event_name, event_version, aggregate_type, aggregate_id, payload, publish_status, created_at
) values ($1, $2, 1, 'target_job', $3, $4, 'pending', $5)`,
		eventID,
		eventName,
		targetJobID,
		payload,
		now,
	)
	if err != nil {
		return fmt.Errorf("insert outbox %s: %w", eventName, err)
	}
	return nil
}

// GetTargetJobForParse returns the bare row needed by the parse executor
// without joining cross-tenant filters (the drainer is internal). Sources
// are returned alongside so the executor can fetch URL bodies.
func (s *SQLStore) GetTargetJobForParse(ctx context.Context, targetJobID string) (TargetJobRecord, []SourceRecord, error) {
	if err := s.checkDB(); err != nil {
		return TargetJobRecord{}, nil, err
	}
	var rec TargetJobRecord
	var (
		profileID          sql.NullString
		title              sql.NullString
		companyName        sql.NullString
		locationText       sql.NullString
		employmentType     sql.NullString
		seniorityLevel     sql.NullString
		sourceURL          sql.NullString
		sourceFileObjectID sql.NullString
		rawJDText          sql.NullString
		notes              sql.NullString
		latestReportID     sql.NullString
		summary            []byte
		fitSummary         []byte
		status             string
		analysisStatus     string
		sourceType         string
	)
	err := s.db.QueryRowContext(ctx, `
select id, user_id, profile_id, status, analysis_status, title, company_name, location_text,
       employment_type, seniority_level, target_language, source_type, source_url, source_file_object_id,
       raw_jd_text, summary, fit_summary, notes, latest_report_id, open_question_issue_count,
       created_at, updated_at
from target_jobs
where id = $1 and deleted_at is null`,
		targetJobID,
	).Scan(
		&rec.ID, &rec.UserID, &profileID, &status, &analysisStatus,
		&title, &companyName, &locationText, &employmentType, &seniorityLevel,
		&rec.TargetLanguage, &sourceType, &sourceURL, &sourceFileObjectID,
		&rawJDText, &summary, &fitSummary, &notes, &latestReportID,
		&rec.OpenQuestionIssueCount, &rec.CreatedAt, &rec.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return TargetJobRecord{}, nil, ErrTargetJobNotFound
	}
	if err != nil {
		return TargetJobRecord{}, nil, fmt.Errorf("select target_jobs for parse: %w", err)
	}
	rec.ProfileID = profileID.String
	rec.Status = sharedtypes.TargetJobStatus(status)
	rec.AnalysisStatus = sharedtypes.TargetJobParseStatus(analysisStatus)
	rec.Title = title.String
	rec.CompanyName = companyName.String
	rec.LocationText = locationText.String
	rec.EmploymentType = employmentType.String
	rec.SeniorityLevel = seniorityLevel.String
	rec.SourceType = SourceType(sourceType)
	rec.SourceURL = sourceURL.String
	rec.SourceFileObjectID = sourceFileObjectID.String
	rec.RawJDText = rawJDText.String
	rec.Notes = notes.String
	rec.LatestReportID = latestReportID.String
	if len(summary) > 0 {
		rec.Summary = append(json.RawMessage{}, summary...)
	}
	if len(fitSummary) > 0 {
		rec.FitSummary = append(json.RawMessage{}, fitSummary...)
	}

	sources, err := s.listSourcesForJob(ctx, s.db, targetJobID)
	if err != nil {
		return TargetJobRecord{}, nil, err
	}
	return rec, sources, nil
}

// UpdateTargetJobAnalysisFailure flips analysis_status to 'failed' on the
// target_jobs row without touching summary / fit_summary.
func (s *SQLStore) UpdateTargetJobAnalysisFailure(ctx context.Context, targetJobID string, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
update target_jobs
set analysis_status = 'failed', updated_at = $1
where id = $2 and deleted_at is null`,
		now,
		targetJobID,
	)
	if err != nil {
		return fmt.Errorf("update target_jobs analysis_status=failed: %w", err)
	}
	return nil
}

// LookupFileAttachmentForUser confirms the referenced file_object belongs
// to userID, is not soft-deleted, and returns its purpose so callers can
// reject mismatched uploads (e.g., a resume blob being passed in a target
// import). ErrTargetJobNotFound is returned for cross-user or soft-deleted
// rows so handlers can render HTTP 404 + TARGET_JOB_NOT_FOUND without
// leaking existence (spec D-9).
func (s *SQLStore) LookupFileAttachmentForUser(ctx context.Context, userID string, fileObjectID string) (FileAttachmentRecord, error) {
	if err := s.checkDB(); err != nil {
		return FileAttachmentRecord{}, err
	}
	var (
		rec     FileAttachmentRecord
		purpose string
	)
	err := s.db.QueryRowContext(ctx, `
select id, user_id, purpose
from file_objects
where id = $1 and user_id = $2 and deleted_at is null`,
		fileObjectID,
		userID,
	).Scan(&rec.ID, &rec.UserID, &purpose)
	if errors.Is(err, sql.ErrNoRows) {
		return FileAttachmentRecord{}, ErrTargetJobNotFound
	}
	if err != nil {
		return FileAttachmentRecord{}, fmt.Errorf("lookup file_objects: %w", err)
	}
	rec.Purpose = purpose
	return rec, nil
}

func (s *SQLStore) UpdateSourceFreshness(ctx context.Context, targetJobID string, freshness FreshnessStatus, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
update target_job_sources
set freshness_status = $1,
    fetched_at = $2
where target_job_id = $3`,
		string(freshness),
		now,
		targetJobID,
	)
	if err != nil {
		return fmt.Errorf("update target_job_sources freshness: %w", err)
	}
	if _, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("update target_job_sources rows affected: %w", err)
	}
	return nil
}

// UpdateSourceSnapshot persists the sanitized URL and fetched JD body for a URL
// source row after the SSRF guard has accepted the upstream response.
func (s *SQLStore) UpdateSourceSnapshot(ctx context.Context, sourceID string, sanitizedURL string, snapshotText string, fetchedAt time.Time, now time.Time) error {
	if err := s.checkDB(); err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx, `
update target_job_sources
set url = $1,
    snapshot_text = $2,
    fetched_at = $3,
    freshness_status = $4
where id = $5`,
		nullableString(sanitizedURL),
		nullableString(snapshotText),
		fetchedAt,
		string(FreshnessFresh),
		sourceID,
	)
	if err != nil {
		return fmt.Errorf("update target_job_sources snapshot: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update target_job_sources snapshot rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTargetJobNotFound
	}
	return nil
}

// ----- helpers -----

type rowScanner interface {
	Scan(dest ...any) error
}

type rowQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func scanTargetJobRow(scanner rowScanner) (TargetJobRecord, error) {
	var (
		rec                TargetJobRecord
		profileID          sql.NullString
		title              sql.NullString
		companyName        sql.NullString
		locationText       sql.NullString
		employmentType     sql.NullString
		seniorityLevel     sql.NullString
		sourceURL          sql.NullString
		sourceFileObjectID sql.NullString
		rawJDText          sql.NullString
		notes              sql.NullString
		latestReportID     sql.NullString
		summary            []byte
		fitSummary         []byte
		status             string
		analysisStatus     string
		sourceType         string
	)
	err := scanner.Scan(
		&rec.ID,
		&rec.UserID,
		&profileID,
		&status,
		&analysisStatus,
		&title,
		&companyName,
		&locationText,
		&employmentType,
		&seniorityLevel,
		&rec.TargetLanguage,
		&sourceType,
		&sourceURL,
		&sourceFileObjectID,
		&rawJDText,
		&summary,
		&fitSummary,
		&notes,
		&latestReportID,
		&rec.OpenQuestionIssueCount,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	)
	if err != nil {
		return TargetJobRecord{}, err
	}
	rec.ProfileID = profileID.String
	rec.Status = sharedtypes.TargetJobStatus(status)
	rec.AnalysisStatus = sharedtypes.TargetJobParseStatus(analysisStatus)
	rec.Title = title.String
	rec.CompanyName = companyName.String
	rec.LocationText = locationText.String
	rec.EmploymentType = employmentType.String
	rec.SeniorityLevel = seniorityLevel.String
	rec.SourceType = SourceType(sourceType)
	rec.SourceURL = sourceURL.String
	rec.SourceFileObjectID = sourceFileObjectID.String
	rec.RawJDText = rawJDText.String
	rec.Notes = notes.String
	rec.LatestReportID = latestReportID.String
	if len(summary) > 0 {
		rec.Summary = append(json.RawMessage{}, summary...)
	}
	if len(fitSummary) > 0 {
		rec.FitSummary = append(json.RawMessage{}, fitSummary...)
	}
	return rec, nil
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func (s *SQLStore) listRequirementsForJob(ctx context.Context, q queryer, targetJobID string) ([]RequirementRecord, error) {
	rows, err := q.QueryContext(ctx, `
select id, target_job_id, kind, label, description, evidence_level, display_order, created_at
from target_job_requirements
where target_job_id = $1
order by display_order asc, created_at asc`,
		targetJobID,
	)
	if err != nil {
		return nil, fmt.Errorf("list target_job_requirements: %w", err)
	}
	defer rows.Close()
	var out []RequirementRecord
	for rows.Next() {
		var (
			rec         RequirementRecord
			description sql.NullString
			kind        string
			evidence    string
		)
		if err := rows.Scan(
			&rec.ID,
			&rec.TargetJobID,
			&kind,
			&rec.Label,
			&description,
			&evidence,
			&rec.DisplayOrder,
			&rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan target_job_requirements: %w", err)
		}
		rec.Kind = RequirementKind(kind)
		rec.Description = description.String
		rec.EvidenceLevel = EvidenceLevel(evidence)
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows target_job_requirements: %w", err)
	}
	return out, nil
}

func (s *SQLStore) listRequirementsForJobTx(ctx context.Context, tx *sql.Tx, targetJobID string) ([]RequirementRecord, error) {
	return s.listRequirementsForJob(ctx, txQueryer{tx: tx}, targetJobID)
}

type txQueryer struct{ tx *sql.Tx }

func (t txQueryer) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (s *SQLStore) listSourcesForJob(ctx context.Context, q queryer, targetJobID string) ([]SourceRecord, error) {
	rows, err := q.QueryContext(ctx, `
select id, target_job_id, source_type, url, file_object_id, snapshot_text, fetched_at, freshness_status, created_at
from target_job_sources
where target_job_id = $1
order by created_at desc`,
		targetJobID,
	)
	if err != nil {
		return nil, fmt.Errorf("list target_job_sources: %w", err)
	}
	defer rows.Close()
	var out []SourceRecord
	for rows.Next() {
		var (
			rec          SourceRecord
			urlStr       sql.NullString
			fileObjectID sql.NullString
			snapshot     sql.NullString
			fetchedAt    sql.NullTime
			sourceType   string
			freshness    string
		)
		if err := rows.Scan(
			&rec.ID,
			&rec.TargetJobID,
			&sourceType,
			&urlStr,
			&fileObjectID,
			&snapshot,
			&fetchedAt,
			&freshness,
			&rec.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan target_job_sources: %w", err)
		}
		rec.SourceType = SourceType(sourceType)
		rec.URL = urlStr.String
		rec.FileObjectID = fileObjectID.String
		rec.SnapshotText = snapshot.String
		if fetchedAt.Valid {
			t := fetchedAt.Time
			rec.FetchedAt = &t
		}
		rec.FreshnessStatus = FreshnessStatus(freshness)
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows target_job_sources: %w", err)
	}
	return out, nil
}

// nullableString returns nil for empty strings so Postgres stores NULL rather
// than an empty literal in nullable columns.
func nullableString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// nullableUUID is the same as nullableString but documents that the column
// expects a UUID.
func nullableUUID(v string) any {
	if v == "" {
		return nil
	}
	return v
}

func nullJSON(raw json.RawMessage) any {
	if len(raw) == 0 {
		return []byte(`{}`)
	}
	return []byte(raw)
}

// encodeCursor and decodeCursor pack (updated_at, id) so list pagination
// remains stable when updated_at ties exist.
func encodeCursor(updatedAt time.Time, id string) string {
	payload := updatedAt.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

func decodeCursor(cursor string) (time.Time, string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("malformed cursor")
	}
	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return t, parts[1], nil
}

var _ Store = (*SQLStore)(nil)
