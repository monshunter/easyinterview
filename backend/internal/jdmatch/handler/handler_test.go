package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/handler"
	"github.com/monshunter/easyinterview/backend/internal/jdmatch/service"
)

func stubSession(userID string, ok bool) handler.SessionResolver {
	return func(ctx context.Context) (string, bool) { return userID, ok }
}

type fakeAgentScans struct {
	record jdmatch.AgentScanRecord
	err    error
}

func (f *fakeAgentScans) GetLatestAgentScanForUser(ctx context.Context, userID string) (jdmatch.AgentScanRecord, error) {
	if f.err != nil {
		return jdmatch.AgentScanRecord{}, f.err
	}
	return f.record, nil
}

func TestGetJobMatchProfileHappyPath(t *testing.T) {
	years := int32(6)
	headline := "Senior frontend engineer"
	h := handler.New(handler.Options{
		Session: stubSession("user-A", true),
		AgentScans: &fakeAgentScans{
			record: jdmatch.AgentScanRecord{Status: jdmatch.AgentScanStatusIdle},
		},
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.JobMatchProfileResult{
				Profile: api.JobMatchProfile{
					DisplayName:       "Alice Example",
					Headline:          &headline,
					YearsOfExperience: &years,
					Skills:            []string{},
					Sources: api.JobMatchProfileSourceCounts{
						Resumes:  3,
						Jds:      5,
						Mocks:    8,
						Debriefs: 2,
					},
				},
			}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/profile", nil)
	w := httptest.NewRecorder()
	h.GetJobMatchProfile(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var got api.JobMatchProfile
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.DisplayName != "Alice Example" {
		t.Fatalf("displayName = %q", got.DisplayName)
	}
	if got.Skills == nil || len(got.Skills) != 0 {
		t.Fatalf("skills should marshal as []: %#v", got.Skills)
	}
	if got.Sources.Resumes != 3 || got.Sources.Jds != 5 || got.Sources.Mocks != 8 || got.Sources.Debriefs != 2 {
		t.Fatalf("sources = %#v", got.Sources)
	}
	if got.AvatarUrl != nil || got.LocationText != nil || got.CompensationText != nil {
		t.Fatalf("avatarUrl / locationText / compensationText must be nil at P0 baseline: %+v", got)
	}
}

func TestGetJobMatchProfile401WhenSessionMissing(t *testing.T) {
	h := handler.New(handler.Options{
		Session: stubSession("", false),
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			t.Fatalf("ProfileBuilder must not be called on 401 path")
			return service.JobMatchProfileResult{}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/profile", nil)
	w := httptest.NewRecorder()
	h.GetJobMatchProfile(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestGetJobMatchProfile500WhenBuilderFails(t *testing.T) {
	h := handler.New(handler.Options{
		Session: stubSession("user-A", true),
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.JobMatchProfileResult{}, errors.New("upstream failure")
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/profile", nil)
	w := httptest.NewRecorder()
	h.GetJobMatchProfile(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
}

func TestGetAgentScanStatusFirstTimeIdle(t *testing.T) {
	h := handler.New(handler.Options{
		Session:    stubSession("user-A", true),
		AgentScans: &fakeAgentScans{err: jdmatch.ErrNotFound},
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.JobMatchProfileResult{}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/agent-status", nil)
	w := httptest.NewRecorder()
	h.GetAgentScanStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", w.Code, w.Body.String())
	}
	var got api.AgentScanStatus
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Status != api.JobMatchAgentStatusIdle {
		t.Fatalf("status = %q, want idle", got.Status)
	}
	if got.LastScanAt != nil || got.NextScanAt != nil || got.Message != nil {
		t.Fatalf("first-time idle must have null timestamps + message, got %+v", got)
	}
}

func TestGetAgentScanStatusExistingRow(t *testing.T) {
	last := time.Date(2026, 5, 21, 5, 0, 0, 0, time.UTC)
	next := last.Add(4 * time.Hour)
	h := handler.New(handler.Options{
		Session: stubSession("user-A", true),
		AgentScans: &fakeAgentScans{
			record: jdmatch.AgentScanRecord{
				Status:     jdmatch.AgentScanStatusIdle,
				LastScanAt: &last,
				NextScanAt: &next,
			},
		},
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.JobMatchProfileResult{}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/agent-status", nil)
	w := httptest.NewRecorder()
	h.GetAgentScanStatus(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got api.AgentScanStatus
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.Status != api.JobMatchAgentStatusIdle {
		t.Fatalf("status = %q", got.Status)
	}
	if got.LastScanAt == nil || *got.LastScanAt != "2026-05-21T05:00:00Z" {
		t.Fatalf("lastScanAt = %v", got.LastScanAt)
	}
	if got.NextScanAt == nil || *got.NextScanAt != "2026-05-21T09:00:00Z" {
		t.Fatalf("nextScanAt = %v", got.NextScanAt)
	}
}

func TestGetAgentScanStatus401WhenSessionMissing(t *testing.T) {
	h := handler.New(handler.Options{
		Session:    stubSession("", false),
		AgentScans: &fakeAgentScans{},
		ProfileBuilder: func(ctx context.Context, userID string) (service.JobMatchProfileResult, error) {
			return service.JobMatchProfileResult{}, nil
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/jd-match/agent-status", nil)
	w := httptest.NewRecorder()
	h.GetAgentScanStatus(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}
