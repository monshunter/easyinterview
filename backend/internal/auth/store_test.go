package auth_test

import (
	"context"
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/monshunter/easyinterview/backend/internal/auth"
)

func TestStoreSurfaceDoesNotExposeExternalIdentities(t *testing.T) {
	storeType := reflect.TypeOf((*auth.Store)(nil)).Elem()
	for i := 0; i < storeType.NumMethod(); i++ {
		method := storeType.Method(i)
		if contains(method.Name, "ExternalIdentity") || contains(method.Name, "ExternalIdentities") {
			t.Fatalf("P0 auth store must not expose external_identities method: %s", method.Name)
		}
	}
}

func TestSQLStoreAuthTableBoundaries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := auth.NewSQLStore(db)
	now := time.Date(2026, 5, 6, 9, 30, 0, 0, time.UTC)
	expires := now.Add(15 * time.Minute)

	mock.ExpectExec("insert into auth_challenges").
		WithArgs(
			"018f2a40-0000-7000-9000-000000000001",
			sqlmock.AnyArg(),
			"candidate@example.com",
			"challenge-hash",
			"login",
			"ip-hash",
			"ua-hash",
			expires,
			now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.CreateChallenge(context.Background(), auth.ChallengeRecord{
		ID:            "018f2a40-0000-7000-9000-000000000001",
		Email:         "candidate@example.com",
		TokenHash:     "challenge-hash",
		Purpose:       auth.ChallengePurposeLogin,
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		ExpiresAt:     expires,
		CreatedAt:     now,
	}); err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	mock.ExpectExec("insert into sessions").
		WithArgs(
			"018f2a40-0000-7000-9000-000000000002",
			"018f2a40-0000-7000-9000-000000000003",
			"session-hash",
			"active",
			"ip-hash",
			"ua-hash",
			now.Add(30*24*time.Hour),
			now,
			now,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err := store.CreateSession(context.Background(), auth.SessionRecord{
		ID:            "018f2a40-0000-7000-9000-000000000002",
		UserID:        "018f2a40-0000-7000-9000-000000000003",
		SessionHash:   "session-hash",
		Status:        auth.SessionStatusActive,
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		ExpiresAt:     now.Add(30 * 24 * time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	mock.ExpectQuery("from users u join user_settings us").
		WithArgs("018f2a40-0000-7000-9000-000000000003").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"email",
			"display_name",
			"ui_language",
			"preferred_practice_language",
			"analytics_opt_in",
		}).AddRow(
			"018f2a40-0000-7000-9000-000000000003",
			"candidate@example.com",
			"Candidate",
			"zh-CN",
			"en",
			true,
		))

	if _, err := store.GetUserContext(context.Background(), "018f2a40-0000-7000-9000-000000000003"); err != nil {
		t.Fatalf("GetUserContext: %v", err)
	}

	mock.ExpectExec("update sessions set updated_at = \\$1, expires_at = \\$2 where id = \\$3 and status = 'active'").
		WithArgs(now, now.Add(30*24*time.Hour), "018f2a40-0000-7000-9000-000000000002").
		WillReturnResult(sqlmock.NewResult(0, 1))

	if err := store.TouchSession(context.Background(), "018f2a40-0000-7000-9000-000000000002", now, now.Add(30*24*time.Hour)); err != nil {
		t.Fatalf("TouchSession: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func contains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

var _ = (*sql.DB)(nil)
