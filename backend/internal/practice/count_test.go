package practice_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/practice"
)

func TestCountPracticeSessionsForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM practice_sessions WHERE user_id = $1",
	)).WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(8))
	got, err := practice.CountPracticeSessionsForUser(context.Background(), db, "user-A")
	if err != nil {
		t.Fatalf("CountPracticeSessionsForUser: %v", err)
	}
	if got != 8 {
		t.Fatalf("count = %d, want 8", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCountPracticeSessionsForUserCrossUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM practice_sessions WHERE user_id = $1",
	)).WithArgs("user-B").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	got, _ := practice.CountPracticeSessionsForUser(context.Background(), db, "user-B")
	if got != 0 {
		t.Fatalf("count = %d, want 0", got)
	}
}

func TestCountPracticeSessionsForUserRejectsNilDB(t *testing.T) {
	_, err := practice.CountPracticeSessionsForUser(context.Background(), nil, "user-A")
	if !errors.Is(err, practice.ErrCounterDBRequired) {
		t.Fatalf("err = %v, want ErrCounterDBRequired", err)
	}
}

func TestCountPracticeSessionsForUserRejectsEmptyUserID(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	_, err := practice.CountPracticeSessionsForUser(context.Background(), db, "  ")
	if !errors.Is(err, practice.ErrCounterUserIDRequired) {
		t.Fatalf("err = %v, want ErrCounterUserIDRequired", err)
	}
}
