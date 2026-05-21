package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/store"
)

type fakeSavedStore struct {
	list   []store.SavedSearchRecord
	listErr error
	create  store.SavedSearchRecord
	createErr error
}

func (f *fakeSavedStore) ListSavedSearchesByUser(ctx context.Context, userID string) ([]store.SavedSearchRecord, error) {
	return f.list, f.listErr
}
func (f *fakeSavedStore) CreateSavedSearch(ctx context.Context, in store.CreateSavedSearchInput) (store.SavedSearchRecord, error) {
	return f.create, f.createErr
}

type fakeRunStore struct {
	out store.SearchRunRecord
	err error
}

func (f *fakeRunStore) CreateSearchRun(ctx context.Context, in store.CreateSearchRunInput) (store.SearchRunRecord, error) {
	return f.out, f.err
}

type fakeSearchAI struct {
	res handler.SearchAIResult
	err error
}

func (f *fakeSearchAI) Search(ctx context.Context, userID, query string, filters json.RawMessage) (handler.SearchAIResult, error) {
	return f.res, f.err
}

func TestListSavedSearchesHappyPath(t *testing.T) {
	now := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	fake := &fakeSavedStore{list: []store.SavedSearchRecord{
		{ID: "s-1", Label: "Frontend remote", Query: "frontend", CreatedAt: now, UpdatedAt: now},
	}}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetSearch(fake, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/saved-searches", nil)
	w := httptest.NewRecorder()
	h.ListSavedSearches(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestCreateSavedSearchHappyPath(t *testing.T) {
	now := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	fake := &fakeSavedStore{create: store.SavedSearchRecord{ID: "s-1", Label: "Frontend", Query: "frontend roles", CreatedAt: now, UpdatedAt: now}}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetSearch(fake, nil, nil)
	h.SetWatchlist(&fakeWatchlistStore{}, func() string { return "s-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/saved-searches", strings.NewReader(`{"label":"Frontend","query":"frontend roles"}`))
	w := httptest.NewRecorder()
	h.CreateSavedSearch(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
}

func TestCreateSavedSearchValidation(t *testing.T) {
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetSearch(&fakeSavedStore{}, nil, nil)
	h.SetWatchlist(&fakeWatchlistStore{}, func() string { return "s-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/saved-searches", strings.NewReader(`{"label":"","query":""}`))
	w := httptest.NewRecorder()
	h.CreateSavedSearch(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestSearchJobsHappyPath(t *testing.T) {
	saved := &fakeSavedStore{}
	runs := &fakeRunStore{}
	ai := &fakeSearchAI{res: handler.SearchAIResult{MatchedJobMatchIDs: []string{"rec-1"}, PromptVersion: "p1", ModelProfileName: "jd_match.search.default", DataSourceVersion: "jd_match.v1"}}
	rec := &fakeRecStore{get: jdmatch.RecommendationRecord{ID: "rec-1", Title: "T", Company: "Acme", Location: "Shanghai", Score: 92, FitMust: 4, FitTotal: 5}}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetSearch(saved, runs, ai)
	h.SetRecommendations(rec, rec)
	h.SetWatchlist(&fakeWatchlistStore{}, func() string { return "sr-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/search", strings.NewReader(`{"query":"frontend platform"}`))
	w := httptest.NewRecorder()
	h.SearchJobs(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	var body struct {
		SearchRunID string `json:"searchRunId"`
		Items       []struct {
			ID string `json:"id"`
		} `json:"items"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.SearchRunID == "" || len(body.Items) != 1 || body.Items[0].ID != "rec-1" {
		t.Fatalf("body = %+v", body)
	}
}

func TestSearchJobsTimeout(t *testing.T) {
	ai := &fakeSearchAI{err: handler.SearchTimeoutErr}
	rec := &fakeRecStore{}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetSearch(&fakeSavedStore{}, &fakeRunStore{}, ai)
	h.SetRecommendations(rec, rec)
	h.SetWatchlist(&fakeWatchlistStore{}, func() string { return "sr-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/search", strings.NewReader(`{"query":"frontend"}`))
	w := httptest.NewRecorder()
	h.SearchJobs(w, req)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", w.Code)
	}
}

func TestSearchJobsValidation(t *testing.T) {
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetSearch(&fakeSavedStore{}, &fakeRunStore{}, &fakeSearchAI{})
	h.SetRecommendations(&fakeRecStore{}, &fakeRecStore{})
	h.SetWatchlist(&fakeWatchlistStore{}, func() string { return "sr-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/search", strings.NewReader(`{"query":""}`))
	w := httptest.NewRecorder()
	h.SearchJobs(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestSearchHandlers401(t *testing.T) {
	h := handler.New(handler.Options{Session: stubSession("", false)})
	h.SetSearch(&fakeSavedStore{}, &fakeRunStore{}, &fakeSearchAI{})
	h.SetRecommendations(&fakeRecStore{}, &fakeRecStore{})
	h.SetWatchlist(&fakeWatchlistStore{}, func() string { return "x" })
	for _, fn := range []func(http.ResponseWriter, *http.Request){h.ListSavedSearches, h.CreateSavedSearch, h.SearchJobs} {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/x", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		fn(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
	}
	_ = errors.New // keep errors import alive
}
