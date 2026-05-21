package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

type fakeRecStore struct {
	list      store.ListRecommendationsResult
	listErr   error
	get       jdmatch.RecommendationRecord
	getErr    error
	dismissed jdmatch.RecommendationRecord
	dismissErr error
}

func (f *fakeRecStore) ListRecommendationsByUser(ctx context.Context, userID string, filter store.ListRecommendationsFilter) (store.ListRecommendationsResult, error) {
	return f.list, f.listErr
}

func (f *fakeRecStore) GetRecommendationByIDForUser(ctx context.Context, userID, id string) (jdmatch.RecommendationRecord, error) {
	return f.get, f.getErr
}

func (f *fakeRecStore) MarkRecommendationDismissed(ctx context.Context, in store.MarkRecommendationDismissedInput) (jdmatch.RecommendationRecord, error) {
	return f.dismissed, f.dismissErr
}

func TestListJobRecommendationsHappyPath(t *testing.T) {
	fake := &fakeRecStore{
		list: store.ListRecommendationsResult{
			Items: []jdmatch.RecommendationRecord{
				{ID: "rec-1", Title: "T1", Company: "Acme", Location: "Shanghai", Score: 92, FitMust: 4, FitTotal: 5, Reasons: []string{"r1"}, Risks: []string{}, Highlights: []string{}},
			},
			PageSize:   20,
			HasMore:    false,
			NextCursor: "",
		},
	}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetRecommendations(fake, fake)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/recommendations?pageSize=20", nil)
	w := httptest.NewRecorder()
	h.ListJobRecommendations(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Items []struct {
			ID    string `json:"id"`
			Score int    `json:"score"`
		} `json:"items"`
		PageInfo struct {
			HasMore bool `json:"hasMore"`
		} `json:"pageInfo"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 || body.Items[0].ID != "rec-1" || body.Items[0].Score != 92 {
		t.Fatalf("items=%+v", body.Items)
	}
}

func TestGetJobRecommendationCrossUser404(t *testing.T) {
	fake := &fakeRecStore{getErr: jdmatch.ErrNotFound}
	h := handler.New(handler.Options{Session: stubSession("user-B", true)})
	h.SetRecommendations(fake, fake)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/recommendations/rec-1", nil)
	w := httptest.NewRecorder()
	h.GetJobRecommendation(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestMarkJobNotRelevantHappyPath(t *testing.T) {
	dismissed := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	fake := &fakeRecStore{
		dismissed: jdmatch.RecommendationRecord{
			ID:          "rec-1",
			DismissedAt: &dismissed,
		},
	}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetRecommendations(fake, fake)
	body := strings.NewReader(`{"reason":"wrong_level","freeNote":"too senior"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/recommendations/rec-1/dismiss", body)
	w := httptest.NewRecorder()
	h.MarkJobNotRelevant(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var resp struct {
		JobMatchID  string `json:"jobMatchId"`
		DismissedAt string `json:"dismissedAt"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.JobMatchID != "rec-1" || resp.DismissedAt == "" {
		t.Fatalf("resp = %+v", resp)
	}
}

func TestMarkJobNotRelevantAlreadyDismissed(t *testing.T) {
	fake := &fakeRecStore{dismissErr: jdmatch.ErrAlreadyDismissed}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetRecommendations(fake, fake)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/recommendations/rec-1/dismiss", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	h.MarkJobNotRelevant(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestRecommendationsHandler401WhenUnauth(t *testing.T) {
	fake := &fakeRecStore{}
	h := handler.New(handler.Options{Session: stubSession("", false)})
	h.SetRecommendations(fake, fake)
	for _, path := range []string{"/api/v1/jd-match/recommendations", "/api/v1/jd-match/recommendations/rec-1", "/api/v1/jd-match/recommendations/rec-1/dismiss"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		if strings.HasSuffix(path, "/recommendations") {
			h.ListJobRecommendations(w, req)
		} else if strings.HasSuffix(path, "/dismiss") {
			req = httptest.NewRequest(http.MethodPost, path, nil)
			w = httptest.NewRecorder()
			h.MarkJobNotRelevant(w, req)
		} else {
			h.GetJobRecommendation(w, req)
		}
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("path=%s status=%d, want 401", path, w.Code)
		}
	}
}
