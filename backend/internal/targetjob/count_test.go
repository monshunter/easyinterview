package targetjob_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

func TestCountTargetJobsForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM target_jobs WHERE user_id = $1 AND deleted_at IS NULL",
	)).WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	got, err := targetjob.CountTargetJobsForUser(context.Background(), db, "user-A")
	if err != nil {
		t.Fatalf("CountTargetJobsForUser: %v", err)
	}
	if got != 5 {
		t.Fatalf("count = %d, want 5", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCountTargetJobsForUserCrossUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM target_jobs WHERE user_id = $1 AND deleted_at IS NULL",
	)).WithArgs("user-B").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	got, _ := targetjob.CountTargetJobsForUser(context.Background(), db, "user-B")
	if got != 0 {
		t.Fatalf("count = %d, want 0", got)
	}
}

func TestCountTargetJobsForUserRejectsNilDB(t *testing.T) {
	_, err := targetjob.CountTargetJobsForUser(context.Background(), nil, "user-A")
	if !errors.Is(err, targetjob.ErrCounterDBRequired) {
		t.Fatalf("err = %v, want ErrCounterDBRequired", err)
	}
}

func TestCountTargetJobsForUserRejectsEmptyUserID(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	_, err := targetjob.CountTargetJobsForUser(context.Background(), db, "")
	if !errors.Is(err, targetjob.ErrCounterUserIDRequired) {
		t.Fatalf("err = %v, want ErrCounterUserIDRequired", err)
	}
}
