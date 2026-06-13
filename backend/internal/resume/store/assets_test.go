package store_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCreateWithParseJobInsertsResumeAndJobAtomically(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 3, 0, 0, 0, time.UTC)
	fileObjectID := "01918fa0-0000-7000-8000-000000000301"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select resource_id, id, status, created_at, updated_at from async_jobs`)).
		WithArgs("resume_parse", "dedupe-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(`insert into resumes`)).
		WithArgs(
			"resume-1", "user-1", fileObjectID, "Resume", "en", string(sharedtypes.TargetJobParseStatusQueued),
			"upload", nil, "job-1", now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into async_jobs`)).
		WithArgs(
			"job-1", "resume_parse", "resume", "resume-1", "dedupe-1", string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), now, now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := repo.CreateWithParseJob(context.Background(), resumestore.CreateAssetInput{
		AssetID:      "resume-1",
		UserID:       "user-1",
		JobID:        "job-1",
		DedupeKey:    "dedupe-1",
		SourceType:   "upload",
		FileObjectID: &fileObjectID,
		Title:        "Resume",
		Language:     "en",
		ParseStatus:  sharedtypes.TargetJobParseStatusQueued,
		JobStatus:    sharedtypes.JobStatusQueued,
		Now:          now,
	})
	if err != nil {
		t.Fatalf("CreateWithParseJob: %v", err)
	}
	if got.AssetID != "resume-1" || got.JobID != "job-1" || got.JobStatus != sharedtypes.JobStatusQueued {
		t.Fatalf("result = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateWithParseJobRollsBackWhenJobInsertFails(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 3, 0, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select resource_id, id, status, created_at, updated_at from async_jobs`)).
		WithArgs("resume_parse", "dedupe-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(`insert into resumes`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into async_jobs`)).
		WillReturnError(errors.New("job insert failed"))
	mock.ExpectRollback()

	_, err := repo.CreateWithParseJob(context.Background(), resumestore.CreateAssetInput{
		AssetID:     "resume-1",
		UserID:      "user-1",
		JobID:       "job-1",
		DedupeKey:   "dedupe-1",
		SourceType:  "paste",
		Title:       "Resume",
		Language:    "en",
		RawText:     "resume text",
		ParseStatus: sharedtypes.TargetJobParseStatusQueued,
		JobStatus:   sharedtypes.JobStatusQueued,
		Now:         now,
	})
	if err == nil {
		t.Fatal("expected job insert error")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestRepositoryExposesFlatResumeMethods(t *testing.T) {
	var _ interface {
		CreateWithParseJob(context.Context, resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error)
		Get(context.Context, string, string) (resumestore.ResumeRecord, error)
		List(context.Context, string, resumestore.ListFilter) (resumestore.ListResult, error)
		UpdateResume(context.Context, resumestore.UpdateResumeInput) (resumestore.ResumeRecord, error)
		DuplicateResume(context.Context, resumestore.DuplicateResumeInput) (resumestore.ResumeRecord, error)
		MarkParsing(context.Context, resumestore.StatusUpdateInput) error
		MarkReady(context.Context, resumestore.MarkReadyInput) error
		MarkFailed(context.Context, resumestore.MarkFailedInput) error
		CreateTailorRun(context.Context, resumestore.CreateTailorRunInput) (resumestore.CreateTailorRunResult, error)
		GetTailorRun(context.Context, string, string) (resumestore.TailorRunRecord, error)
		GetForTailor(context.Context, string) (resumestore.TailorJobContext, error)
		CompleteTailorRunSuccess(context.Context, resumestore.CompleteTailorRunSuccessInput) error
		DeleteForUser(context.Context, string, time.Time) error
	} = (*resumestore.Repository)(nil)
}

func TestGetScopesUserAndMapsStructuredProfile(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 6, 13, 18, 0, 0, 0, time.UTC)
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, display_name, language, parse_status`)).
		WithArgs("resume-1", "user-1").
		WillReturnRows(resumeRows().AddRow(
			"resume-1", "user-1", nil, "Resume", "Alice CV", "en", string(sharedtypes.TargetJobParseStatusReady),
			[]byte(`{"headline":"Senior engineer"}`), nil, []byte(`{"basics":{"name":"Alice"}}`), "snapshot",
			"paste", nil, "job-1", now, now, nil,
		))

	got, err := repo.Get(context.Background(), "user-1", "resume-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != "resume-1" || got.DisplayName == nil || *got.DisplayName != "Alice CV" {
		t.Fatalf("resume = %+v", got)
	}
	if string(got.StructuredProfile) != `{"basics":{"name":"Alice"}}` {
		t.Fatalf("structured profile = %s", got.StructuredProfile)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestUpdateResumeOverwritesProfileAndScopesUser(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	createdAt := time.Date(2026, 6, 13, 19, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 6, 13, 19, 30, 0, 0, time.UTC)
	displayName := "Updated CV"

	mock.ExpectQuery(regexp.QuoteMeta(`update resumes`)).
		WithArgs(
			[]byte(`{"headline":"Staff engineer"}`),
			displayName,
			updatedAt,
			"resume-1",
			"user-1",
		).
		WillReturnRows(resumeRows().AddRow(
			"resume-1", "user-1", nil, "Resume", "Updated CV", "en", string(sharedtypes.TargetJobParseStatusReady),
			[]byte(`{}`), nil, []byte(`{"headline":"Staff engineer"}`), nil,
			"paste", nil, "job-1", createdAt, updatedAt, nil,
		))

	got, err := repo.UpdateResume(context.Background(), resumestore.UpdateResumeInput{
		UserID:            "user-1",
		ResumeID:          "resume-1",
		DisplayName:       &displayName,
		StructuredProfile: []byte(`{"headline":"Staff engineer"}`),
		Now:               updatedAt,
	})
	if err != nil {
		t.Fatalf("UpdateResume: %v", err)
	}
	if got.DisplayName == nil || *got.DisplayName != "Updated CV" || string(got.StructuredProfile) != `{"headline":"Staff engineer"}` {
		t.Fatalf("resume = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestUpdateResumeNotFound(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	mock.ExpectQuery(regexp.QuoteMeta(`update resumes`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "resume-1", "user-1").
		WillReturnError(sql.ErrNoRows)

	_, err := repo.UpdateResume(context.Background(), resumestore.UpdateResumeInput{
		UserID:            "user-1",
		ResumeID:          "resume-1",
		StructuredProfile: []byte(`{}`),
		Now:               time.Now(),
	})
	if !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("err = %v, want ErrAssetNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDuplicateResumeCopiesSourceSnapshotAndAppliesProfile(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 6, 13, 20, 0, 0, 0, time.UTC)
	fileObjectID := "01918fa0-0000-7000-8000-000000000301"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, display_name, language, parse_status`)).
		WithArgs("source-1", "user-1").
		WillReturnRows(resumeRows().AddRow(
			"source-1", "user-1", fileObjectID, "Resume", "Source CV", "en", string(sharedtypes.TargetJobParseStatusReady),
			[]byte(`{"headline":"old"}`), "raw text", []byte(`{"headline":"old"}`), "snapshot",
			"upload", nil, "job-1", now, now, nil,
		))
	mock.ExpectExec(regexp.QuoteMeta(`insert into resumes`)).
		WithArgs(
			"resume-new", "user-1", fileObjectID, "Resume", "New CV", "en", string(sharedtypes.TargetJobParseStatusReady),
			"upload", "raw text", sqlmock.AnyArg(), "snapshot", []byte(`{"headline":"new"}`), now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, display_name, language, parse_status`)).
		WithArgs("resume-new", "user-1").
		WillReturnRows(resumeRows().AddRow(
			"resume-new", "user-1", fileObjectID, "Resume", "New CV", "en", string(sharedtypes.TargetJobParseStatusReady),
			[]byte(`{"headline":"old"}`), "raw text", []byte(`{"headline":"new"}`), "snapshot",
			"upload", nil, nil, now, now, nil,
		))
	mock.ExpectCommit()

	displayName := "New CV"
	got, err := repo.DuplicateResume(context.Background(), resumestore.DuplicateResumeInput{
		NewResumeID:       "resume-new",
		UserID:            "user-1",
		SourceResumeID:    "source-1",
		DisplayName:       &displayName,
		StructuredProfile: []byte(`{"headline":"new"}`),
		Now:               now,
	})
	if err != nil {
		t.Fatalf("DuplicateResume: %v", err)
	}
	if got.ID != "resume-new" || got.DisplayName == nil || *got.DisplayName != "New CV" || string(got.StructuredProfile) != `{"headline":"new"}` {
		t.Fatalf("resume = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestDuplicateResumeSourceNotFoundRollsBack(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, display_name, language, parse_status`)).
		WithArgs("source-1", "user-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	_, err := repo.DuplicateResume(context.Background(), resumestore.DuplicateResumeInput{
		NewResumeID:    "resume-new",
		UserID:         "user-1",
		SourceResumeID: "source-1",
		Now:            time.Now(),
	})
	if !errors.Is(err, resumestore.ErrAssetNotFound) {
		t.Fatalf("err = %v, want ErrAssetNotFound", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestParseStatusTransition(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 4, 30, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`update resumes`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusProcessing), now, "resume-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkParsing(context.Background(), resumestore.StatusUpdateInput{UserID: "user-1", AssetID: "resume-1", Now: now}); err != nil {
		t.Fatalf("MarkParsing: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resumes`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusReady), []byte(`{"summary":"ok"}`), []byte(`{"basics":{}}`), "parsed text", now, "resume-1", "user-1", string(sharedtypes.TargetJobParseStatusProcessing)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkReady(context.Background(), resumestore.MarkReadyInput{
		UserID:             "user-1",
		AssetID:            "resume-1",
		ParsedSummary:      []byte(`{"summary":"ok"}`),
		StructuredProfile:  []byte(`{"basics":{}}`),
		ParsedTextSnapshot: "parsed text",
		Now:                now,
	}); err != nil {
		t.Fatalf("MarkReady: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resumes`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusFailed), "AI_OUTPUT_INVALID", now, "resume-2", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkFailed(context.Background(), resumestore.MarkFailedInput{
		UserID:    "user-1",
		AssetID:   "resume-2",
		ErrorCode: "AI_OUTPUT_INVALID",
		Now:       now,
	}); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resumes`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusReady), []byte(`{"summary":"late"}`), []byte(`{}`), "late", now, "resume-ready", "user-1", string(sharedtypes.TargetJobParseStatusProcessing)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.MarkReady(context.Background(), resumestore.MarkReadyInput{
		UserID:             "user-1",
		AssetID:            "resume-ready",
		ParsedSummary:      []byte(`{"summary":"late"}`),
		ParsedTextSnapshot: "late",
		Now:                now,
	})
	if !errors.Is(err, resumestore.ErrInvalidStateTransition) {
		t.Fatalf("MarkReady invalid transition err = %v, want ErrInvalidStateTransition", err)
	}
}

func TestCompleteParseSuccessWritesReadyStateProfileAndCompletedOutboxAtomically(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 8, 30, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`update resumes`)).
		WithArgs(
			string(sharedtypes.TargetJobParseStatusReady),
			[]byte(`{"basics":{"name":"Ada"}}`),
			[]byte(`{"basics":{"name":"Ada"}}`),
			"parsed text",
			now,
			"resume-1",
			"user-1",
			string(sharedtypes.TargetJobParseStatusProcessing),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into outbox_events`)).
		WithArgs("event-1", "resume.parse.completed", "resume", "resume-1", []byte(`{"resumeId":"resume-1","userId":"user-1","parseStatus":"ready"}`), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.CompleteParseSuccess(context.Background(), resumestore.CompleteParseSuccessInput{
		UserID:             "user-1",
		AssetID:            "resume-1",
		ParsedSummary:      []byte(`{"basics":{"name":"Ada"}}`),
		StructuredProfile:  []byte(`{"basics":{"name":"Ada"}}`),
		ParsedTextSnapshot: "parsed text",
		OutboxEventID:      "event-1",
		OutboxEventPayload: []byte(`{"resumeId":"resume-1","userId":"user-1","parseStatus":"ready"}`),
		Now:                now,
	}); err != nil {
		t.Fatalf("CompleteParseSuccess: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCompleteParseFailureMarksFailedWithoutCompletedOutbox(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 8, 45, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`update resumes`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusFailed), "AI_OUTPUT_INVALID", now, "resume-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.CompleteParseFailure(context.Background(), resumestore.CompleteParseFailureInput{
		UserID:    "user-1",
		AssetID:   "resume-1",
		ErrorCode: "AI_OUTPUT_INVALID",
		Now:       now,
	}); err != nil {
		t.Fatalf("CompleteParseFailure: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestListCursorPagination(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	base := time.Date(2026, 5, 13, 5, 0, 0, 0, time.UTC)

	firstRows := resumeRows()
	for i := 0; i < 21; i++ {
		firstRows.AddRow(
			assetID(i), "user-1", nil, "Resume", nil, "en", string(sharedtypes.TargetJobParseStatusQueued),
			[]byte(`{}`), nil, []byte(`{}`), nil, "paste", nil, "job-1", base.Add(-time.Duration(i)*time.Minute), base.Add(-time.Duration(i)*time.Minute), nil,
		)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, display_name, language, parse_status`)).
		WithArgs("user-1", 21).
		WillReturnRows(firstRows)

	first, err := repo.List(context.Background(), "user-1", resumestore.ListFilter{PageSize: 20})
	if err != nil {
		t.Fatalf("List first page: %v", err)
	}
	if len(first.Items) != 20 || !first.HasMore || first.NextCursor == "" {
		t.Fatalf("first page len=%d hasMore=%v cursor=%q", len(first.Items), first.HasMore, first.NextCursor)
	}
	if first.Items[0].ID != assetID(0) || first.Items[19].ID != assetID(19) {
		t.Fatalf("unexpected first page order first=%s last=%s", first.Items[0].ID, first.Items[19].ID)
	}

	secondRows := resumeRows()
	for i := 20; i < 25; i++ {
		secondRows.AddRow(
			assetID(i), "user-1", nil, "Resume", nil, "en", string(sharedtypes.TargetJobParseStatusQueued),
			[]byte(`{}`), nil, []byte(`{}`), nil, "paste", nil, "job-1", base.Add(-time.Duration(i)*time.Minute), base.Add(-time.Duration(i)*time.Minute), nil,
		)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, display_name, language, parse_status`)).
		WithArgs("user-1", sqlmock.AnyArg(), sqlmock.AnyArg(), 21).
		WillReturnRows(secondRows)

	second, err := repo.List(context.Background(), "user-1", resumestore.ListFilter{PageSize: 20, Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("List second page: %v", err)
	}
	if len(second.Items) != 5 || second.HasMore || second.NextCursor != "" {
		t.Fatalf("second page len=%d hasMore=%v cursor=%q", len(second.Items), second.HasMore, second.NextCursor)
	}

	_, err = repo.List(context.Background(), "user-1", resumestore.ListFilter{Cursor: "not-a-valid-cursor"})
	if !errors.Is(err, resumestore.ErrInvalidCursor) {
		t.Fatalf("invalid cursor err = %v, want ErrInvalidCursor", err)
	}
}

func newMockRepository(t *testing.T) (*resumestore.Repository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return resumestore.NewRepository(db), mock, func() { _ = db.Close() }
}

func resumeRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"file_object_id",
		"title",
		"display_name",
		"language",
		"parse_status",
		"parsed_summary",
		"original_text",
		"structured_profile",
		"parsed_text_snapshot",
		"source_type",
		"error_code",
		"latest_parse_job_id",
		"created_at",
		"updated_at",
		"deleted_at",
	})
}

func assetID(i int) string {
	return "01918fa0-0000-7000-8000-00000000a" + string(rune('a'+i/10)) + string(rune('0'+i%10))
}
