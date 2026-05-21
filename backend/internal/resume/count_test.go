package resume_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/resume"
)

func TestCountResumesForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM resume_assets WHERE user_id = $1 AND deleted_at IS NULL",
	)).
		WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	got, err := resume.CountResumesForUser(context.Background(), db, "user-A")
	if err != nil {
		t.Fatalf("CountResumesForUser: %v", err)
	}
	if got != 3 {
		t.Fatalf("count = %d, want 3", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCountResumesForUserCrossUserIsolation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	// User B has zero resumes; the query is constrained to user_id = $1.
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM resume_assets WHERE user_id = $1 AND deleted_at IS NULL",
	)).
		WithArgs("user-B").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	got, err := resume.CountResumesForUser(context.Background(), db, "user-B")
	if err != nil {
		t.Fatalf("CountResumesForUser: %v", err)
	}
	if got != 0 {
		t.Fatalf("count = %d, want 0", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCountResumesForUserRejectsNilDB(t *testing.T) {
	_, err := resume.CountResumesForUser(context.Background(), nil, "user-A")
	if !errors.Is(err, resume.ErrCounterDBRequired) {
		t.Fatalf("err = %v, want ErrCounterDBRequired", err)
	}
}

func TestCountResumesForUserRejectsEmptyUserID(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	_, err = resume.CountResumesForUser(context.Background(), db, "   ")
	if !errors.Is(err, resume.ErrCounterUserIDRequired) {
		t.Fatalf("err = %v, want ErrCounterUserIDRequired", err)
	}
}
