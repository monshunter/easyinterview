package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/profile"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
)

func seedCard(store *fakeStore, id string, userID string, updatedAt time.Time, source string) {
	store.mu.Lock()
	defer store.mu.Unlock()
	rec := &profile.ExperienceCardRecord{
		ID:          id,
		UserID:      userID,
		ProfileID:   "fake-profile-" + userID,
		Title:       "card-" + id,
		CompanyName: "Acme",
		Situation:   "s",
		Task:        "t",
		Action:      "a",
		Result:      "r",
		Skills:      []string{"go"},
		Language:    "en",
		SourceType:  source,
		Confidence:  "medium",
		UpdatedAt:   updatedAt,
		CreatedAt:   updatedAt,
	}
	store.cards[id] = rec
	store.order = append(store.order, id)
}

func TestListExperienceCardsPagination(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())

	base := time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 25; i++ {
		id := makeULID(i + 1)
		seedCard(store, id, "user-a", base.Add(time.Duration(i)*time.Second), profile.SourceTypeManual)
	}
	// Add a cross-user card that must not leak.
	seedCard(store, makeULID(99), "user-b", base.Add(time.Hour), profile.SourceTypeManual)

	// First page: pageSize=20.
	rec := httptest.NewRecorder()
	req := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me/experience-cards?pageSize=20", nil), "user-a")
	h.ListExperienceCards(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("first page want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var page api.PaginatedExperienceCard
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(page.Items) != 20 {
		t.Fatalf("first page items = %d, want 20", len(page.Items))
	}
	if !page.PageInfo.HasMore {
		t.Fatalf("first page hasMore = false, want true")
	}
	if page.PageInfo.NextCursor == nil || *page.PageInfo.NextCursor == "" {
		t.Fatalf("first page nextCursor empty")
	}
	// updated_at desc ordering: newest first.
	prev := page.Items[0].UpdatedAt
	for i := 1; i < len(page.Items); i++ {
		if page.Items[i].UpdatedAt > prev {
			t.Fatalf("page not sorted desc at index %d", i)
		}
		prev = page.Items[i].UpdatedAt
	}

	// Second page via cursor.
	rec2 := httptest.NewRecorder()
	req2 := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me/experience-cards?pageSize=20&cursor="+*page.PageInfo.NextCursor, nil), "user-a")
	h.ListExperienceCards(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second page want 200, got %d body=%s", rec2.Code, rec2.Body.String())
	}
	var page2 api.PaginatedExperienceCard
	if err := json.Unmarshal(rec2.Body.Bytes(), &page2); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(page2.Items) != 5 {
		t.Fatalf("second page items = %d, want 5", len(page2.Items))
	}
	if page2.PageInfo.HasMore {
		t.Fatalf("second page hasMore = true, want false")
	}
	// generated ExperienceCard does not surface user_id; verify via store.
	store.mu.Lock()
	for _, item := range page2.Items {
		if rec, ok := store.cards[item.Id]; ok && rec.UserID == "user-b" {
			store.mu.Unlock()
			t.Fatalf("user-b card leaked: %+v", item)
		}
	}
	store.mu.Unlock()
}

func TestListExperienceCardsInvalidCursor(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	rec := httptest.NewRecorder()
	req := contextWithUser(httptest.NewRequest(http.MethodGet, "/api/v1/profiles/me/experience-cards?cursor=!!notbase64!!", nil), "user-a")
	h.ListExperienceCards(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid cursor want 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateExperienceCardManualForce(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())

	body, _ := json.Marshal(api.CreateExperienceCardRequest{
		Title:       "drove migration",
		CompanyName: "Acme",
		Situation:   "fragmented system",
		Task:        "unify",
		Action:      "did the work",
		Result:      "win",
		Language:    "zh-CN",
		Skills:      []string{"leadership"},
	})
	req := contextWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/profiles/me/experience-cards", bytes.NewReader(body)), "user-a")
	req.Header.Set("Idempotency-Key", "01918fa0-0000-7000-8000-00000000ik01")
	rec := httptest.NewRecorder()
	h.CreateExperienceCard(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create want 201, got %d body=%s", rec.Code, rec.Body.String())
	}
	var card api.ExperienceCard
	if err := json.Unmarshal(rec.Body.Bytes(), &card); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	store.mu.Lock()
	if got := store.cards[card.Id].SourceType; got != profile.SourceTypeManual {
		t.Fatalf("source_type = %q, want manual", got)
	}
	if got := store.cards[card.Id].Confidence; got != profile.ConfidenceDefaultMedium {
		t.Fatalf("confidence = %q, want medium", got)
	}
	store.mu.Unlock()
}

func TestCreateExperienceCardMissingIdempotencyKey(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	body, _ := json.Marshal(api.CreateExperienceCardRequest{
		Title: "x", CompanyName: "y", Situation: "z", Task: "t",
		Action: "a", Result: "r", Language: "en",
	})
	req := contextWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/profiles/me/experience-cards", bytes.NewReader(body)), "user-a")
	rec := httptest.NewRecorder()
	h.CreateExperienceCard(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing IK want 422, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestCreateExperienceCardValidation(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	body, _ := json.Marshal(api.CreateExperienceCardRequest{Title: "", CompanyName: "y", Situation: "z", Task: "t", Action: "a", Result: "r", Language: "en"})
	req := contextWithUser(httptest.NewRequest(http.MethodPost, "/api/v1/profiles/me/experience-cards", bytes.NewReader(body)), "user-a")
	req.Header.Set("Idempotency-Key", "01918fa0-0000-7000-8000-00000000ik02")
	rec := httptest.NewRecorder()
	h.CreateExperienceCard(rec, req)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing title want 422, got %d", rec.Code)
	}
}

func TestUpdateExperienceCardPatch(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	id := makeULID(7)
	seedCard(store, id, "user-a", time.Now().UTC(), profile.SourceTypeManual)

	newResult := "updated result"
	body, _ := json.Marshal(api.UpdateExperienceCardRequest{Result: &newResult})
	req := contextWithUser(httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/me/experience-cards/"+id, bytes.NewReader(body)), "user-a")
	req.Header.Set("Idempotency-Key", "01918fa0-0000-7000-8000-00000000ik03")
	rec := httptest.NewRecorder()
	h.UpdateExperienceCard(rec, req, id)
	if rec.Code != http.StatusOK {
		t.Fatalf("patch want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	var updated api.ExperienceCard
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if updated.Result != newResult {
		t.Fatalf("result = %q, want %q", updated.Result, newResult)
	}
	// Other fields preserved.
	if updated.Title != "card-"+id {
		t.Fatalf("title overwritten: %q", updated.Title)
	}
}

func TestUpdateExperienceCardCrossUser404(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	id := makeULID(11)
	seedCard(store, id, "user-a", time.Now().UTC(), profile.SourceTypeManual)

	newResult := "x"
	body, _ := json.Marshal(api.UpdateExperienceCardRequest{Result: &newResult})
	req := contextWithUser(httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/me/experience-cards/"+id, bytes.NewReader(body)), "user-b")
	req.Header.Set("Idempotency-Key", "01918fa0-0000-7000-8000-00000000ik04")
	rec := httptest.NewRecorder()
	h.UpdateExperienceCard(rec, req, id)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cross-user want 404, got %d body=%s", rec.Code, rec.Body.String())
	}
	var errBody api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if errBody.Error.Code != sharederrors.CodeResourceNotFound {
		t.Fatalf("error code = %q, want RESOURCE_NOT_FOUND", errBody.Error.Code)
	}
}

func TestUpdateExperienceCardMissingCardReturns404(t *testing.T) {
	store := newFakeStore()
	h := newHandlerForTest(store, defaultSettings())
	body, _ := json.Marshal(api.UpdateExperienceCardRequest{})
	id := makeULID(404)
	req := contextWithUser(httptest.NewRequest(http.MethodPatch, "/api/v1/profiles/me/experience-cards/"+id, bytes.NewReader(body)), "user-a")
	req.Header.Set("Idempotency-Key", "01918fa0-0000-7000-8000-00000000ik05")
	rec := httptest.NewRecorder()
	h.UpdateExperienceCard(rec, req, id)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing card want 404, got %d", rec.Code)
	}
}

// Helpers ---------------------------------------------------------------------

func makeULID(seed int) string {
	// Deterministic, RFC 4122-shape uuid so DB-side uuid columns accept it
	// even if a future test wires the real SQL store.
	hex := "0123456789abcdef"
	out := []byte("01918fa0-0000-7000-8000-000000000000")
	v := seed
	for i := len(out) - 1; i >= 0 && v > 0; i-- {
		if out[i] == '-' {
			continue
		}
		out[i] = hex[v%16]
		v /= 16
	}
	return string(out)
}

// noop_ctx silences linter for context import in case it's unused after edits.
var _ = context.Background
