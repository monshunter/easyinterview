package store_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestCreateWithParseJobInsertsAssetAndJobAtomically(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 3, 0, 0, 0, time.UTC)
	fileObjectID := "01918fa0-0000-7000-8000-000000000301"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select resource_id, id, status, created_at, updated_at from async_jobs`)).
		WithArgs("resume_parse", "dedupe-1").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(regexp.QuoteMeta(`insert into resume_assets`)).
		WithArgs(
			"asset-1", "user-1", fileObjectID, "Resume", "en", string(sharedtypes.TargetJobParseStatusQueued),
			"upload", nil, nil, "job-1", now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into async_jobs`)).
		WithArgs(
			"job-1", "resume_parse", "resume_asset", "asset-1", "dedupe-1", string(sharedtypes.JobStatusQueued), sqlmock.AnyArg(), now, now, now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	got, err := repo.CreateWithParseJob(context.Background(), resumestore.CreateAssetInput{
		AssetID:      "asset-1",
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
	if got.AssetID != "asset-1" || got.JobID != "job-1" || got.JobStatus != sharedtypes.JobStatusQueued {
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
	mock.ExpectExec(regexp.QuoteMeta(`insert into resume_assets`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into async_jobs`)).
		WillReturnError(errors.New("job insert failed"))
	mock.ExpectRollback()

	_, err := repo.CreateWithParseJob(context.Background(), resumestore.CreateAssetInput{
		AssetID:     "asset-1",
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

func TestRepositoryExposesResumeAssetMethods(t *testing.T) {
	var _ interface {
		CreateWithParseJob(context.Context, resumestore.CreateAssetInput) (resumestore.CreateAssetResult, error)
		Get(context.Context, string, string) (resumestore.AssetRecord, error)
		List(context.Context, string, resumestore.ListFilter) (resumestore.ListResult, error)
		MarkParsing(context.Context, resumestore.StatusUpdateInput) error
		MarkReady(context.Context, resumestore.MarkReadyInput) error
		MarkFailed(context.Context, resumestore.MarkFailedInput) error
		CreateStructuredMasterFromAsset(context.Context, resumestore.CreateStructuredMasterInput) (resumestore.VersionRecord, error)
		DeleteForUser(context.Context, string, time.Time) error
	} = (*resumestore.Repository)(nil)
}

func TestCreateStructuredMasterFromAssetInsertsReadyAssetMaster(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 17, 16, 30, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`select parse_status`)).
		WithArgs("asset-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"parse_status"}).AddRow(string(sharedtypes.TargetJobParseStatusReady)))
	mock.ExpectQuery(regexp.QuoteMeta(`insert into resume_versions`)).
		WithArgs(
			"version-1",
			"user-1",
			"asset-1",
			string(sharedtypes.ResumeVersionTypeStructuredMaster),
			"Structured master",
			[]byte(`{"headline":"Senior engineer"}`),
			"resume_profile.v1",
			"not_applicable",
			"model-1",
			nil,
			now,
		).
		WillReturnRows(versionRows().AddRow(
			"version-1", "user-1", "asset-1", nil, string(sharedtypes.ResumeVersionTypeStructuredMaster), nil,
			"Structured master", nil, nil, []byte(`{"headline":"Senior engineer"}`), nil,
			"resume_profile.v1", "not_applicable", "model-1", nil, now, now, nil,
		))
	mock.ExpectCommit()

	got, err := repo.CreateStructuredMasterFromAsset(context.Background(), resumestore.CreateStructuredMasterInput{
		VersionID:         "version-1",
		UserID:            "user-1",
		ResumeAssetID:     "asset-1",
		DisplayName:       "Structured master",
		StructuredProfile: []byte(`{"headline":"Senior engineer"}`),
		Provenance: resumestore.VersionProvenance{
			PromptVersion: "resume_profile.v1",
			RubricVersion: "not_applicable",
			ModelID:       "model-1",
		},
		Now: now,
	})
	if err != nil {
		t.Fatalf("CreateStructuredMasterFromAsset: %v", err)
	}
	if got.ID != "version-1" || got.VersionType != sharedtypes.ResumeVersionTypeStructuredMaster || got.ParentVersionID != nil || got.TargetJobID != nil || got.SeedStrategy != nil {
		t.Fatalf("version = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCreateStructuredMasterFromAssetValidatesOwnershipReadinessAndUniqueIndex(t *testing.T) {
	tests := []struct {
		name      string
		selectErr error
		status    sharedtypes.TargetJobParseStatus
		insertErr error
		want      error
	}{
		{name: "cross user not found", selectErr: sql.ErrNoRows, want: resumestore.ErrAssetNotFound},
		{name: "parse not ready", status: sharedtypes.TargetJobParseStatusProcessing, want: resumestore.ErrAssetParseNotReady},
		{name: "structured master already exists", status: sharedtypes.TargetJobParseStatusReady, insertErr: &pq.Error{Code: "23505", Constraint: "uq_resume_versions_structured_master_per_asset"}, want: resumestore.ErrStructuredMasterAlreadyExists},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepository(t)
			defer cleanup()
			mock.ExpectBegin()
			selectQuery := mock.ExpectQuery(regexp.QuoteMeta(`select parse_status`)).
				WithArgs("asset-1", "user-1")
			if tc.selectErr != nil {
				selectQuery.WillReturnError(tc.selectErr)
			} else {
				selectQuery.WillReturnRows(sqlmock.NewRows([]string{"parse_status"}).AddRow(string(tc.status)))
			}
			if tc.insertErr != nil {
				mock.ExpectQuery(regexp.QuoteMeta(`insert into resume_versions`)).
					WillReturnError(tc.insertErr)
			}
			mock.ExpectRollback()

			_, err := repo.CreateStructuredMasterFromAsset(context.Background(), resumestore.CreateStructuredMasterInput{
				VersionID:         "version-1",
				UserID:            "user-1",
				ResumeAssetID:     "asset-1",
				DisplayName:       "Structured master",
				StructuredProfile: []byte(`{"headline":"Senior engineer"}`),
			})

			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("sql expectations: %v", err)
			}
		})
	}
}

func TestParseStatusTransition(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 4, 30, 0, 0, time.UTC)

	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusProcessing), now, "asset-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkParsing(context.Background(), resumestore.StatusUpdateInput{UserID: "user-1", AssetID: "asset-1", Now: now}); err != nil {
		t.Fatalf("MarkParsing: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusProcessing), now, "asset-retry", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkParsing(context.Background(), resumestore.StatusUpdateInput{UserID: "user-1", AssetID: "asset-retry", Now: now}); err != nil {
		t.Fatalf("MarkParsing failed retry: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusReady), []byte(`{"summary":"ok"}`), "parsed text", now, "asset-1", "user-1", string(sharedtypes.TargetJobParseStatusProcessing)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkReady(context.Background(), resumestore.MarkReadyInput{
		UserID:             "user-1",
		AssetID:            "asset-1",
		ParsedSummary:      []byte(`{"summary":"ok"}`),
		ParsedTextSnapshot: "parsed text",
		Now:                now,
	}); err != nil {
		t.Fatalf("MarkReady: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusFailed), "AI_OUTPUT_INVALID", now, "asset-2", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.MarkFailed(context.Background(), resumestore.MarkFailedInput{
		UserID:    "user-1",
		AssetID:   "asset-2",
		ErrorCode: "AI_OUTPUT_INVALID",
		Now:       now,
	}); err != nil {
		t.Fatalf("MarkFailed: %v", err)
	}

	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusReady), []byte(`{"summary":"late"}`), "late", now, "asset-ready", "user-1", string(sharedtypes.TargetJobParseStatusProcessing)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err := repo.MarkReady(context.Background(), resumestore.MarkReadyInput{
		UserID:             "user-1",
		AssetID:            "asset-ready",
		ParsedSummary:      []byte(`{"summary":"late"}`),
		ParsedTextSnapshot: "late",
		Now:                now,
	})
	if !errors.Is(err, resumestore.ErrInvalidStateTransition) {
		t.Fatalf("MarkReady invalid transition err = %v, want ErrInvalidStateTransition", err)
	}
}

func TestCompleteParseSuccessWritesReadyStateAndCompletedOutboxAtomically(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 13, 8, 30, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(
			string(sharedtypes.TargetJobParseStatusReady),
			[]byte(`{"basics":{"name":"Ada"}}`),
			"parsed text",
			now,
			"asset-1",
			"user-1",
			string(sharedtypes.TargetJobParseStatusProcessing),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`insert into outbox_events`)).
		WithArgs("event-1", "resume.parse.completed", "resume_asset", "asset-1", []byte(`{"resumeAssetId":"asset-1","userId":"user-1","parseStatus":"ready"}`), now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := repo.CompleteParseSuccess(context.Background(), resumestore.CompleteParseSuccessInput{
		UserID:             "user-1",
		AssetID:            "asset-1",
		ParsedSummary:      []byte(`{"basics":{"name":"Ada"}}`),
		ParsedTextSnapshot: "parsed text",
		OutboxEventID:      "event-1",
		OutboxEventPayload: []byte(`{"resumeAssetId":"asset-1","userId":"user-1","parseStatus":"ready"}`),
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

	mock.ExpectExec(regexp.QuoteMeta(`update resume_assets`)).
		WithArgs(string(sharedtypes.TargetJobParseStatusFailed), "AI_OUTPUT_INVALID", now, "asset-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.CompleteParseFailure(context.Background(), resumestore.CompleteParseFailureInput{
		UserID:    "user-1",
		AssetID:   "asset-1",
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

	firstRows := assetRows()
	for i := 0; i < 21; i++ {
		firstRows.AddRow(
			assetID(i), "user-1", nil, "Resume", "en", string(sharedtypes.TargetJobParseStatusQueued),
			[]byte(`{}`), nil, nil, nil, "paste", nil, "job-1", base.Add(-time.Duration(i)*time.Minute), base.Add(-time.Duration(i)*time.Minute), nil,
		)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, language, parse_status`)).
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

	secondRows := assetRows()
	for i := 20; i < 25; i++ {
		secondRows.AddRow(
			assetID(i), "user-1", nil, "Resume", "en", string(sharedtypes.TargetJobParseStatusQueued),
			[]byte(`{}`), nil, nil, nil, "paste", nil, "job-1", base.Add(-time.Duration(i)*time.Minute), base.Add(-time.Duration(i)*time.Minute), nil,
		)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`select id, user_id, file_object_id, title, language, parse_status`)).
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

func assetRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"file_object_id",
		"title",
		"language",
		"parse_status",
		"parsed_summary",
		"original_text",
		"guided_answers",
		"parsed_text_snapshot",
		"source_type",
		"error_code",
		"latest_parse_job_id",
		"created_at",
		"updated_at",
		"deleted_at",
	})
}

func versionRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"user_id",
		"resume_asset_id",
		"parent_version_id",
		"version_type",
		"target_job_id",
		"display_name",
		"seed_strategy",
		"focus_angle",
		"structured_profile",
		"match_score",
		"prompt_version",
		"rubric_version",
		"model_id",
		"provider",
		"created_at",
		"updated_at",
		"deleted_at",
	})
}

func assetID(i int) string {
	return "01918fa0-0000-7000-8000-00000000a" + string(rune('a'+i/10)) + string(rune('0'+i%10))
}
