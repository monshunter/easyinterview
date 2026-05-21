package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/profile"
)

func ptr(s string) *string { return &s }

func newHandlerForTest(store profile.Store, settings profile.SettingsReader) *Handler {
	return New(Options{
		Store:    store,
		Settings: settings,
		Session:  func(ctx context.Context) (string, bool) { return userFromCtx(ctx) },
		NewID:    func() string { return "01918fa0-0000-7000-8000-000000003000" },
	})
}

type userCtxKey struct{}

func userFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userCtxKey{}).(string)
	return v, ok && strings.TrimSpace(v) != ""
}

func contextWithUser(req *http.Request, userID string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), userCtxKey{}, userID))
}

func defaultSettings() fakeSettings {
	region := "CN-SH"
	return fakeSettings{defaults: profile.UserSettings{
		PreferredPracticeLanguage: "en",
		UiLanguage:                "zh-CN",
		Region:                    &region,
	}}
}

func TestGetMyProfileSeedAndReuse(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())

	// First call seeds the row.
	rec := httptest.NewRecorder()
	req := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me", nil), "user-a")
	h.GetMyProfile(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first call: want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var first api.CandidateProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &first); err != nil {
		t.Fatalf("unmarshal first: %v", err)
	}
	if first.Headline != nil {
		t.Fatalf("headline must be null for seed, got %q", *first.Headline)
	}
	if first.YearsOfExperience != nil {
		t.Fatalf("yearsOfExperience must be null for seed, got %d", *first.YearsOfExperience)
	}
	if first.PreferredPracticeLanguage != "en" {
		t.Fatalf("preferredPracticeLanguage = %q", first.PreferredPracticeLanguage)
	}
	if first.UiLanguage != "zh-CN" {
		t.Fatalf("uiLanguage = %q", first.UiLanguage)
	}
	if first.Region == nil || *first.Region != "CN-SH" {
		t.Fatalf("region must come from settings: %#v", first.Region)
	}

	// Capture row state before re-read.
	store.mu.Lock()
	if got := len(store.profiles); got != 1 {
		store.mu.Unlock()
		t.Fatalf("store rows after seed = %d, want 1", got)
	}
	store.profiles["user-a"].UpdatedAt = time.Now().UTC()
	store.mu.Unlock()

	// Second call returns the same row without re-seeding.
	rec2 := httptest.NewRecorder()
	req2 := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me", nil), "user-a")
	h.GetMyProfile(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second call: want 200, got %d body=%s", rec2.Code, rec2.Body.String())
	}
	store.mu.Lock()
	rows := len(store.profiles)
	version := store.profiles["user-a"].ProfileVersion
	store.mu.Unlock()
	if rows != 1 {
		t.Fatalf("store rows after second call = %d, want 1", rows)
	}
	if version != 1 {
		t.Fatalf("profile_version after re-read = %d, want 1 (no bump on read)", version)
	}
}

func TestGetMyProfileUnauthorized(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me", nil)
	h.GetMyProfile(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", rec.Code)
	}
}

func TestGetMyProfileCrossUserIsolation(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())

	// Seed for user A.
	recA := httptest.NewRecorder()
	reqA := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me", nil), "user-a")
	h.GetMyProfile(recA, reqA)
	// Patch user A's headline to a known value.
	store.mu.Lock()
	h1 := "Senior frontend engineer"
	store.profiles["user-a"].Headline = &h1
	store.mu.Unlock()

	// Seed for user B; verify separate row, no leak from A.
	recB := httptest.NewRecorder()
	reqB := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me", nil), "user-b")
	h.GetMyProfile(recB, reqB)
	if recB.Code != http.StatusOK {
		t.Fatalf("user B want 200, got %d", recB.Code)
	}
	var bProfile api.CandidateProfile
	if err := json.Unmarshal(recB.Body.Bytes(), &bProfile); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if bProfile.Headline != nil {
		t.Fatalf("user B headline must remain null after user A patch, got %q", *bProfile.Headline)
	}
}

func ptrInt32(v int32) *int32 { return &v }
