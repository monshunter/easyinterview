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

func newMockRepository(t *testing.T) (*resumestore.Repository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return resumestore.NewRepository(db), mock, func() { _ = db.Close() }
}
