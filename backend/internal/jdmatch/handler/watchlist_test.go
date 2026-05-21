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

type fakeWatchlistStore struct {
	listOut   []store.WatchlistRecord
	listErr   error
	addOut    store.WatchlistRecord
	addErr    error
	removeOut int64
	removeErr error
}

func (f *fakeWatchlistStore) ListWatchlistByUser(ctx context.Context, userID string) ([]store.WatchlistRecord, error) {
	return f.listOut, f.listErr
}

func (f *fakeWatchlistStore) AddWatchlistItem(ctx context.Context, in store.AddWatchlistItemInput) (store.WatchlistRecord, error) {
	return f.addOut, f.addErr
}

func (f *fakeWatchlistStore) RemoveWatchlistItem(ctx context.Context, userID, linkedJobMatchID string) (int64, error) {
	return f.removeOut, f.removeErr
}

func TestListWatchlistToneDerivation(t *testing.T) {
	added := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	fake := &fakeWatchlistStore{listOut: []store.WatchlistRecord{
		{ID: "w-1", LinkedJobMatchID: "rec-1", LinkedTitle: "T1", LinkedCompany: "Acme", LinkedScore: 92, AddedAt: added},
		{ID: "w-2", LinkedJobMatchID: "rec-2", LinkedTitle: "T2", LinkedCompany: "Lumen", LinkedScore: 78, AddedAt: added},
		{ID: "w-3", LinkedJobMatchID: "rec-3", LinkedTitle: "T3", LinkedCompany: "Globex", LinkedScore: 45, AddedAt: added},
	}}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetWatchlist(fake, func() string { return "w-new" })
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/watchlist", nil)
	w := httptest.NewRecorder()
	h.ListWatchlist(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body struct {
		Items []struct {
			Tone  string `json:"tone"`
			Score int    `json:"score,omitempty"`
		} `json:"items"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 3 {
		t.Fatalf("items=%d, want 3", len(body.Items))
	}
	want := []string{"ok", "warn", "muted"}
	for i, w := range want {
		if body.Items[i].Tone != w {
			t.Fatalf("items[%d].tone = %q, want %q", i, body.Items[i].Tone, w)
		}
	}
}

func TestAddToWatchlistHappyPath(t *testing.T) {
	added := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	fake := &fakeWatchlistStore{addOut: store.WatchlistRecord{
		ID: "w-1", LinkedJobMatchID: "rec-1", LinkedTitle: "T1", LinkedCompany: "Acme", LinkedScore: 92, AddedAt: added,
	}}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetWatchlist(fake, func() string { return "w-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/watchlist", strings.NewReader(`{"jobMatchId":"rec-1"}`))
	w := httptest.NewRecorder()
	h.AddToWatchlist(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
}

func TestAddToWatchlistCrossUser404(t *testing.T) {
	fake := &fakeWatchlistStore{addErr: jdmatch.ErrNotFound}
	h := handler.New(handler.Options{Session: stubSession("user-B", true)})
	h.SetWatchlist(fake, func() string { return "w-new" })
	req := httptest.NewRequest(http.MethodPost, "/api/v1/jd-match/watchlist", strings.NewReader(`{"jobMatchId":"rec-A"}`))
	w := httptest.NewRecorder()
	h.AddToWatchlist(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestRemoveFromWatchlistHappyPath(t *testing.T) {
	fake := &fakeWatchlistStore{removeOut: 1}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetWatchlist(fake, func() string { return "w-new" })
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jd-match/watchlist/rec-1", nil)
	w := httptest.NewRecorder()
	h.RemoveFromWatchlist(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
}

func TestRemoveFromWatchlist404WhenAbsent(t *testing.T) {
	fake := &fakeWatchlistStore{removeOut: 0}
	h := handler.New(handler.Options{Session: stubSession("user-A", true)})
	h.SetWatchlist(fake, func() string { return "w-new" })
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/jd-match/watchlist/rec-1", nil)
	w := httptest.NewRecorder()
	h.RemoveFromWatchlist(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
