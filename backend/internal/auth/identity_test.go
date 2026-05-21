package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestGetUserIdentityForUserSeeded(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT email, display_name FROM users WHERE id = $1 AND deleted_at IS NULL",
	)).WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"email", "display_name"}).
			AddRow("alice@example.com", "Alice Example"))

	got, err := auth.GetUserIdentityForUser(context.Background(), db, "user-A")
	if err != nil {
		t.Fatalf("GetUserIdentityForUser: %v", err)
	}
	if got.DisplayName != "Alice Example" {
		t.Fatalf("displayName = %q, want %q", got.DisplayName, "Alice Example")
	}
	// backend-auth's existing maskEmail keeps the first and last local-part
	// characters; the contract surface is that "***" appears, the raw local
	// part does not survive intact, and the domain is preserved.
	if !strings.Contains(got.EmailMasked, "***") {
		t.Fatalf("emailMasked must contain ***: %q", got.EmailMasked)
	}
	if strings.Contains(got.EmailMasked, "alice") {
		t.Fatalf("emailMasked must not contain raw local-part 'alice': %q", got.EmailMasked)
	}
	if !strings.HasSuffix(got.EmailMasked, "@example.com") {
		t.Fatalf("emailMasked must preserve domain: %q", got.EmailMasked)
	}
	if got.AvatarURL != nil {
		t.Fatalf("avatarUrl must be nil at P0 baseline, got %v", got.AvatarURL)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestGetUserIdentityForUserMissingDisplayName(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT email, display_name FROM users WHERE id = $1 AND deleted_at IS NULL",
	)).WithArgs("user-C").
		WillReturnRows(sqlmock.NewRows([]string{"email", "display_name"}).
			AddRow("c@example.com", nil))

	got, err := auth.GetUserIdentityForUser(context.Background(), db, "user-C")
	if err != nil {
		t.Fatalf("GetUserIdentityForUser: %v", err)
	}
	if got.DisplayName != "Candidate" {
		t.Fatalf("displayName fallback = %q, want %q", got.DisplayName, "Candidate")
	}
	if got.EmailMasked == "" || !strings.Contains(got.EmailMasked, "***") {
		t.Fatalf("emailMasked must mask via ***, got %q", got.EmailMasked)
	}
}

func TestGetUserIdentityForUserNotFound(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT email, display_name FROM users WHERE id = $1 AND deleted_at IS NULL",
	)).WithArgs("user-Z").
		WillReturnError(sql.ErrNoRows)

	_, err := auth.GetUserIdentityForUser(context.Background(), db, "user-Z")
	if !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("err = %v, want ErrUserNotFound", err)
	}
}

func TestGetUserIdentityForUserDoesNotWriteAudit(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	// Only the read-only SELECT is expected. Any INSERT into audit_events
	// or UPDATE on users would fail the ExpectationsWereMet check below
	// because no such expectation is registered.
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT email, display_name FROM users WHERE id = $1 AND deleted_at IS NULL",
	)).WithArgs("user-A").
		WillReturnRows(sqlmock.NewRows([]string{"email", "display_name"}).
			AddRow("alice@example.com", "Alice Example"))

	if _, err := auth.GetUserIdentityForUser(context.Background(), db, "user-A"); err != nil {
		t.Fatalf("GetUserIdentityForUser: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("identity lookup must not perform other queries: %v", err)
	}
}

func TestGetUserIdentityForUserRejectsNilDB(t *testing.T) {
	_, err := auth.GetUserIdentityForUser(context.Background(), nil, "user-A")
	if !errors.Is(err, auth.ErrIdentityDBRequired) {
		t.Fatalf("err = %v, want ErrIdentityDBRequired", err)
	}
}

func TestGetUserIdentityForUserRejectsEmptyUserID(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	_, err := auth.GetUserIdentityForUser(context.Background(), db, "  ")
	if !errors.Is(err, auth.ErrIdentityUserIDRequired) {
		t.Fatalf("err = %v, want ErrIdentityUserIDRequired", err)
	}
}
