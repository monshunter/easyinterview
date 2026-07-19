package auth_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

func TestUserContextHasNoDisplayPreferencesOrAnalyticsProjection(t *testing.T) {
	typeOfUser := reflect.TypeOf(auth.UserContext{})
	for _, field := range []string{"UILanguage", "PreferredPracticeLanguage", "AnalyticsOptIn"} {
		if _, ok := typeOfUser.FieldByName(field); ok {
			t.Fatalf("auth.UserContext must not expose obsolete/current-user-external field %s", field)
		}
	}
}

func TestUserContextOwnsAccountThemeProjection(t *testing.T) {
	typeOfUser := reflect.TypeOf(auth.UserContext{})
	if _, ok := typeOfUser.FieldByName("DisplayPreferences"); !ok {
		t.Fatal("auth.UserContext must expose account-owned DisplayPreferences")
	}
}

func TestSQLStoreUpdateUserContextCommitsProfileAndThemeTogether(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := auth.NewSQLStore(db)
	now := time.Date(2026, 7, 19, 3, 0, 0, 0, time.UTC)
	displayName := "Alice Candidate"
	acceptedTerms := true
	prefs := auth.AccountDisplayPreferences{Theme: auth.AccountThemePlum}

	mock.ExpectBegin()
	mock.ExpectExec("update users").
		WithArgs("user-1", sql.NullString{String: displayName, Valid: true}, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("update user_settings").
		WithArgs("user-1", "plum", nil, nil, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("from users u").
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "display_name", "profile_completed_at", "terms_accepted_at",
			"theme", "custom_accent_hue", "custom_accent_chroma",
		}).AddRow("user-1", "alice@example.com", displayName, now, now, "plum", nil, nil))
	mock.ExpectCommit()

	got, err := store.UpdateUserContext(context.Background(), "user-1", auth.UpdateUserContextInput{
		DisplayName:        &displayName,
		AcceptedTerms:      &acceptedTerms,
		DisplayPreferences: &prefs,
	}, now)
	if err != nil {
		t.Fatalf("UpdateUserContext: %v", err)
	}
	if got.DisplayName != displayName || got.DisplayPreferences.Theme != auth.AccountThemePlum {
		t.Fatalf("updated context = %+v", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStoreUpdateUserContextRollsBackProfileWhenThemeWriteFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := auth.NewSQLStore(db)
	now := time.Date(2026, 7, 19, 3, 0, 0, 0, time.UTC)
	displayName := "Alice Candidate"
	acceptedTerms := true
	prefs := auth.AccountDisplayPreferences{Theme: auth.AccountThemePlum}

	mock.ExpectBegin()
	mock.ExpectExec("update users").
		WithArgs("user-1", sql.NullString{String: displayName, Valid: true}, now).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("update user_settings").
		WithArgs("user-1", "plum", nil, nil, now).
		WillReturnError(driver.ErrBadConn)
	mock.ExpectRollback()

	if _, err := store.UpdateUserContext(context.Background(), "user-1", auth.UpdateUserContextInput{
		DisplayName:        &displayName,
		AcceptedTerms:      &acceptedTerms,
		DisplayPreferences: &prefs,
	}, now); err == nil {
		t.Fatal("UpdateUserContext error = nil, want rollback on theme write failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
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
			sqlmock.AnyArg(),
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

	mock.ExpectQuery("from users u").
		WithArgs("018f2a40-0000-7000-9000-000000000003").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"email",
			"display_name",
			"profile_completed_at",
			"terms_accepted_at",
			"theme",
			"custom_accent_hue",
			"custom_accent_chroma",
		}).AddRow(
			"018f2a40-0000-7000-9000-000000000003",
			"candidate@example.com",
			"Candidate",
			now,
			now,
			"ocean",
			nil,
			nil,
		))

	if _, err := store.GetUserContext(context.Background(), "018f2a40-0000-7000-9000-000000000003"); err != nil {
		t.Fatalf("GetUserContext: %v", err)
	}

	mock.ExpectQuery("select analytics_opt_in").
		WithArgs("018f2a40-0000-7000-9000-000000000003").
		WillReturnRows(sqlmock.NewRows([]string{"analytics_opt_in"}).AddRow(true))
	if optIn, err := store.GetAnalyticsOptIn(context.Background(), "018f2a40-0000-7000-9000-000000000003"); err != nil || !optIn {
		t.Fatalf("GetAnalyticsOptIn = %t, %v", optIn, err)
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

func TestSQLStorePrivacyDeleteHandoffSoftDeletesUserAndRevokesSessions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := auth.NewSQLStore(db)
	now := time.Date(2026, 5, 26, 9, 30, 0, 0, time.UTC)
	userID := "018f2a40-0000-7000-9000-000000000101"
	privacyRequestID := "018f2a40-0000-7000-9000-000000000201"
	jobID := "018f2a40-0000-7000-9000-000000000301"
	dedupe := &notRawDedupeKey{raw: "delete-key"}

	mock.ExpectQuery("from async_jobs").
		WithArgs(dedupe, "privacy_delete").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectBegin()
	mock.ExpectExec("update users").
		WithArgs(now, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("update sessions").
		WithArgs(now, userID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec("insert into privacy_requests").
		WithArgs(privacyRequestID, userID, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("insert into async_jobs").
		WithArgs(jobID, "privacy_delete", privacyRequestID, dedupe, now).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if _, err := store.CreatePrivacyDeleteHandoff(context.Background(), userID, "delete-key", privacyRequestID, jobID, now); err != nil {
		t.Fatalf("handoff: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStorePrivacyDeleteDedupeKeyIsScopedByUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := auth.NewSQLStore(db)
	now := time.Date(2026, 5, 6, 19, 30, 0, 0, time.UTC)
	rawKey := "v1.1777777777.018f2a40-0000-7000-9000-000000000001"

	expectPrivacyHandoffInsert := func(userID, privacyRequestID, jobID string, dedupeMatcher sqlmock.Argument) {
		mock.ExpectQuery("from async_jobs").
			WithArgs(dedupeMatcher, "privacy_delete").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectBegin()
		mock.ExpectExec("update users").
			WithArgs(now, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("update sessions").
			WithArgs(now, userID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("insert into privacy_requests").
			WithArgs(privacyRequestID, userID, now).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("insert into async_jobs").
			WithArgs(jobID, "privacy_delete", privacyRequestID, dedupeMatcher, now).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
	}

	userOneScoped := &notRawDedupeKey{raw: rawKey}
	userTwoScoped := &notRawDedupeKey{raw: rawKey}
	expectPrivacyHandoffInsert("018f2a40-0000-7000-9000-000000000101", "018f2a40-0000-7000-9000-000000000201", "018f2a40-0000-7000-9000-000000000301", userOneScoped)
	expectPrivacyHandoffInsert("018f2a40-0000-7000-9000-000000000102", "018f2a40-0000-7000-9000-000000000202", "018f2a40-0000-7000-9000-000000000302", userTwoScoped)

	if _, err := store.CreatePrivacyDeleteHandoff(context.Background(), "018f2a40-0000-7000-9000-000000000101", rawKey, "018f2a40-0000-7000-9000-000000000201", "018f2a40-0000-7000-9000-000000000301", now); err != nil {
		t.Fatalf("first user handoff: %v", err)
	}
	if _, err := store.CreatePrivacyDeleteHandoff(context.Background(), "018f2a40-0000-7000-9000-000000000102", rawKey, "018f2a40-0000-7000-9000-000000000202", "018f2a40-0000-7000-9000-000000000302", now); err != nil {
		t.Fatalf("second user handoff: %v", err)
	}
	if userOneScoped.value == "" || userTwoScoped.value == "" {
		t.Fatalf("scoped dedupe keys were not observed: one=%q two=%q", userOneScoped.value, userTwoScoped.value)
	}
	if userOneScoped.value == userTwoScoped.value {
		t.Fatalf("different users must not share privacy_delete dedupe key: %q", userOneScoped.value)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSQLStoreCountRecentChallengesCountsAllRecentAttempts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	store := auth.NewSQLStore(db)
	since := time.Date(2026, 5, 6, 20, 10, 0, 0, time.UTC)
	mock.ExpectQuery(`select count\(\*\)\s+from auth_challenges\s+where created_at >= \$1\s+and \(email = \$2 or ip_hash = \$3\)`).
		WithArgs(since, "candidate@example.com", "ip-hash").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	got, err := store.CountRecentChallenges(context.Background(), "candidate@example.com", "ip-hash", since)
	if err != nil {
		t.Fatalf("CountRecentChallenges: %v", err)
	}
	if got != 2 {
		t.Fatalf("recent challenge count = %d, want 2", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

type notRawDedupeKey struct {
	raw   string
	value string
}

func (m *notRawDedupeKey) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	m.value = s
	return s != "" && s != m.raw
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
