package store_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	uploadstore "github.com/monshunter/easyinterview/backend/internal/upload/store"
)

func TestRepositoryCreateInsertsPendingFileObject(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 12, 1, 0, 0, 0, time.UTC)

	mock.ExpectExec(`insert into file_objects`).
		WithArgs(
			"file-1",
			"user-1",
			string(uploadstore.PurposeResume),
			"user-1/resume/file-1.pdf",
			"resume.pdf",
			"application/pdf",
			int64(1024),
			string(uploadstore.RetentionUserOwned),
			string(uploadstore.StatusPending),
			now,
			now,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Create(context.Background(), uploadstore.CreateInput{
		ID:               "file-1",
		UserID:           "user-1",
		Purpose:          uploadstore.PurposeResume,
		ObjectKey:        "user-1/resume/file-1.pdf",
		OriginalFileName: "resume.pdf",
		ContentType:      "application/pdf",
		ByteSize:         1024,
		Now:              now,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryStateTransitions(t *testing.T) {
	for name, setup := range map[string]struct {
		current uploadstore.UploadStatus
		next    uploadstore.UploadStatus
		wantErr bool
	}{
		"pending uploaded":     {current: uploadstore.StatusPending, next: uploadstore.StatusUploaded},
		"pending scan_failed":  {current: uploadstore.StatusPending, next: uploadstore.StatusScanFailed},
		"uploaded scan_failed": {current: uploadstore.StatusUploaded, next: uploadstore.StatusScanFailed},
		"pending deleted":      {current: uploadstore.StatusPending, next: uploadstore.StatusDeleted},
		"uploaded deleted":     {current: uploadstore.StatusUploaded, next: uploadstore.StatusDeleted},
		"scan_failed deleted":  {current: uploadstore.StatusScanFailed, next: uploadstore.StatusDeleted},
		"deleted uploaded":     {current: uploadstore.StatusDeleted, next: uploadstore.StatusUploaded, wantErr: true},
		"scan_failed uploaded": {current: uploadstore.StatusScanFailed, next: uploadstore.StatusUploaded, wantErr: true},
	} {
		t.Run(name, func(t *testing.T) {
			repo, mock, cleanup := newMockRepository(t)
			defer cleanup()
			now := time.Date(2026, 5, 12, 1, 5, 0, 0, time.UTC)

			mock.ExpectBegin()
			mock.ExpectQuery(`select upload_status from file_objects where id = \$1 for update`).
				WithArgs("file-1").
				WillReturnRows(sqlmock.NewRows([]string{"upload_status"}).AddRow(string(setup.current)))
			if !setup.wantErr {
				mock.ExpectExec(`update file_objects set upload_status = \$1, updated_at = \$2`).
					WithArgs(string(setup.next), now, "file-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			} else {
				mock.ExpectRollback()
			}

			err := repo.MarkStatus(context.Background(), "file-1", setup.next, now)
			if setup.wantErr {
				if !errors.Is(err, uploadstore.ErrInvalidStateTransition) {
					t.Fatalf("expected invalid transition, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("MarkStatus: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestRepositoryLockForRegisterScopesUserAndPurpose(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 12, 1, 10, 0, 0, time.UTC)

	mock.ExpectQuery(`from file_objects\s+where id = \$1 and user_id = \$2 and purpose = \$3 and deleted_at is null\s+for update`).
		WithArgs("file-1", "user-1", string(uploadstore.PurposeResume)).
		WillReturnRows(sqlmock.NewRows(fileObjectColumns()).AddRow(
			"file-1", "user-1", string(uploadstore.PurposeResume), "user-1/resume/file-1.pdf", "resume.pdf", "application/pdf", int64(1024), nil, string(uploadstore.RetentionUserOwned), string(uploadstore.StatusPending), now, now, nil,
		))

	rec, err := repo.LockForRegister(context.Background(), "file-1", "user-1", uploadstore.PurposeResume)
	if err != nil {
		t.Fatalf("LockForRegister: %v", err)
	}
	if rec.ID != "file-1" || rec.UserID != "user-1" || rec.Status != uploadstore.StatusPending {
		t.Fatalf("record = %+v", rec)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryLockForRegisterReturnsNotFoundForCrossUser(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	mock.ExpectQuery(`from file_objects\s+where id = \$1 and user_id = \$2 and purpose = \$3 and deleted_at is null\s+for update`).
		WithArgs("file-1", "user-2", string(uploadstore.PurposeResume)).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.LockForRegister(context.Background(), "file-1", "user-2", uploadstore.PurposeResume)
	if !errors.Is(err, uploadstore.ErrFileObjectNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryHardDeleteDeletesRow(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	mock.ExpectExec(`delete from file_objects where id = \$1`).
		WithArgs("file-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := repo.HardDelete(context.Background(), "file-1"); err != nil {
		t.Fatalf("HardDelete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryDeleteFileObjectsForUserHardDeletesScopedRows(t *testing.T) {
	repo, mock, cleanup := newMockRepository(t)
	defer cleanup()
	now := time.Date(2026, 5, 12, 1, 15, 0, 0, time.UTC)

	mock.ExpectBegin()
	mock.ExpectQuery(`from file_objects\s+where user_id = \$1 and deleted_at is null\s+for update`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "object_key", "purpose"}).
			AddRow("file-1", "user-1/resume/file-1.pdf", string(uploadstore.PurposeResume)).
			AddRow("file-2", "user-1/resume/file-2.pdf", string(uploadstore.PurposeResume)))
	mock.ExpectExec(`delete from file_objects where user_id = \$1 and id = any\(\$2\)`).
		WithArgs("user-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	deleted, err := repo.DeleteFileObjectsForUser(context.Background(), "user-1", now)
	if err != nil {
		t.Fatalf("DeleteFileObjectsForUser: %v", err)
	}
	if len(deleted) != 2 || deleted[0].ID != "file-1" || deleted[1].ID != "file-2" {
		t.Fatalf("deleted = %+v", deleted)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func newMockRepository(t *testing.T) (*uploadstore.Repository, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	return uploadstore.NewRepository(db), mock, func() { db.Close() }
}

func fileObjectColumns() []string {
	return []string{"id", "user_id", "purpose", "object_key", "original_file_name", "content_type", "byte_size", "sha256_hex", "retention_policy", "upload_status", "created_at", "updated_at", "deleted_at"}
}
