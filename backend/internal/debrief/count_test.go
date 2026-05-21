package debrief_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/debrief"
)

func TestCountDebriefsForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM debriefs WHERE user_id = $1",
	)).WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	got, err := debrief.CountDebriefsForUser(context.Background(), db, "user-A")
	if err != nil {
		t.Fatalf("CountDebriefsForUser: %v", err)
	}
	if got != 2 {
		t.Fatalf("count = %d, want 2", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestCountDebriefsForUserCrossUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COUNT(*) FROM debriefs WHERE user_id = $1",
	)).WithArgs("user-B").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	got, _ := debrief.CountDebriefsForUser(context.Background(), db, "user-B")
	if got != 0 {
		t.Fatalf("count = %d, want 0", got)
	}
}

func TestCountDebriefsForUserRejectsNilDB(t *testing.T) {
	_, err := debrief.CountDebriefsForUser(context.Background(), nil, "user-A")
	if !errors.Is(err, debrief.ErrCounterDBRequired) {
		t.Fatalf("err = %v, want ErrCounterDBRequired", err)
	}
}

func TestCountDebriefsForUserRejectsEmptyUserID(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	_, err := debrief.CountDebriefsForUser(context.Background(), db, "")
	if !errors.Is(err, debrief.ErrCounterUserIDRequired) {
		t.Fatalf("err = %v, want ErrCounterUserIDRequired", err)
	}
}
