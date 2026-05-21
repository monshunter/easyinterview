package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func TestUpdateMyProfilePatchAndValidation(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())

	// Seed first.
	seedRec := httptest.NewRecorder()
	seedReq := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me", nil), "user-a")
	h.GetMyProfile(seedRec, seedReq)

	headline := "Senior frontend engineer focused on growth-stage SaaS"
	years := int32(5)
	body, _ := json.Marshal(api.UpdateProfileRequest{
		Headline:          &headline,
		YearsOfExperience: &years,
	})
	rec := httptest.NewRecorder()
	req := contextWithUser(httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/me", bytes.NewReader(body)), "user-a")
	h.UpdateMyProfile(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var got api.CandidateProfile
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Headline == nil || *got.Headline != headline {
		t.Fatalf("headline = %v, want %q", got.Headline, headline)
	}
	if got.YearsOfExperience == nil || *got.YearsOfExperience != 5 {
		t.Fatalf("yearsOfExperience = %v, want 5", got.YearsOfExperience)
	}
	// Region / preferredPracticeLanguage stay from seed defaults; currentRole still nil.
	if got.CurrentRole != nil {
		t.Fatalf("currentRole must remain null after patch, got %q", *got.CurrentRole)
	}

	// profile_version bumped: 1 (seed) + 1 (UpsertLite increment on fakeStore).
	store.mu.Lock()
	v1 := store.profiles["user-a"].ProfileVersion
	store.mu.Unlock()
	if v1 < 2 {
		t.Fatalf("profile_version after patch = %d, want >=2", v1)
	}

	// Validation: yearsOfExperience = -1 returns 422 + VALIDATION_FAILED, and
	// profile_version does not change.
	bad := int32(-1)
	body2, _ := json.Marshal(api.UpdateProfileRequest{YearsOfExperience: &bad})
	rec2 := httptest.NewRecorder()
	req2 := contextWithUser(httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/me", bytes.NewReader(body2)), "user-a")
	h.UpdateMyProfile(rec2, req2)
	if rec2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid patch want 422, got %d body=%s", rec2.Code, rec2.Body.String())
	}
	var errBody api.ApiErrorResponse
	if err := json.Unmarshal(rec2.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("unmarshal err body: %v", err)
	}
	if errBody.Error.Code != sharederrors.CodeValidationFailed {
		t.Fatalf("error code = %q, want %q", errBody.Error.Code, sharederrors.CodeValidationFailed)
	}
	store.mu.Lock()
	v2 := store.profiles["user-a"].ProfileVersion
	store.mu.Unlock()
	if v1 != v2 {
		t.Fatalf("profile_version changed after rejected patch: %d -> %d", v1, v2)
	}

	// Empty-string clears the column (D-2 spec: empty string is legal).
	empty := ""
	body3, _ := json.Marshal(api.UpdateProfileRequest{Headline: &empty})
	rec3 := httptest.NewRecorder()
	req3 := contextWithUser(httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/me", bytes.NewReader(body3)), "user-a")
	h.UpdateMyProfile(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("empty-string patch want 200, got %d body=%s", rec3.Code, rec3.Body.String())
	}
	var clear api.CandidateProfile
	if err := json.Unmarshal(rec3.Body.Bytes(), &clear); err != nil {
		t.Fatalf("unmarshal clear: %v", err)
	}
	if clear.Headline == nil || *clear.Headline != "" {
		t.Fatalf("headline after empty-string patch = %v, want \"\"", clear.Headline)
	}
}
