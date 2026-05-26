package runner_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/privacy/runner"
)

func TestSQLStoreMarkDeleteRequestCompletedDeletesAccountIdentityAndPreservesRequestTombstone(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := runner.NewSQLStore(db)
	now := time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)
	userID := "018f2a40-0000-7000-9000-000000000101"
	requestID := "018f2a40-0000-7000-9000-000000000201"
	email := "manual-uat-full-funnel@example.test"

	mock.ExpectBegin()
	mock.ExpectQuery("select email from users").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"email"}).AddRow(email))
	mock.ExpectExec("update privacy_requests").
		WithArgs(requestID, now, 1, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("delete from resume_version_suggestions").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("update resume_versions").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("delete from resume_versions").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectExec("delete from auth_challenges").
		WithArgs(userID, email).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("delete from users").
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := store.MarkDeleteRequestCompleted(context.Background(), requestID, userID, 1, now); err != nil {
		t.Fatalf("MarkDeleteRequestCompleted: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStoreLookupDeleteRequestUserTreatsCompletedTombstoneAsIdempotent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := runner.NewSQLStore(db)
	requestID := "018f2a40-0000-7000-9000-000000000201"

	mock.ExpectQuery("select user_id, status from privacy_requests").
		WithArgs(requestID).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "status"}).AddRow(nil, "completed"))

	userID, err := store.LookupDeleteRequestUser(context.Background(), requestID)
	if !errors.Is(err, runner.ErrPrivacyDeleteAlreadyCompleted) {
		t.Fatalf("LookupDeleteRequestUser error = %v, want ErrPrivacyDeleteAlreadyCompleted", err)
	}
	if userID != "" {
		t.Fatalf("userID = %q, want empty", userID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
