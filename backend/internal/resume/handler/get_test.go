package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestHandlerImplementsGetResumeSurface(t *testing.T) {
	var _ interface {
		GetResume(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestGetResume(t *testing.T) {
	now := time.Date(2026, 5, 13, 4, 0, 0, 0, time.UTC).Format(time.RFC3339)
	svc := &fakeGetService{out: api.ResumeAsset{
		Id:          "asset-1",
		Title:       "Resume",
		Language:    "en",
		ParseStatus: sharedtypes.TargetJobParseStatusQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) {
			return "user-1", true
		},
	})
	rec := httptest.NewRecorder()

	h.GetResume(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/asset-1", nil), "asset-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.userID != "user-1" || svc.assetID != "asset-1" {
		t.Fatalf("service scope user=%q asset=%q", svc.userID, svc.assetID)
	}
}

func TestGetResumeNotFoundAndCrossUserReturns404(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeGetService{err: resume.ErrNotFound},
		Session: func(context.Context) (string, bool) {
			return "user-2", true
		},
	})
	rec := httptest.NewRecorder()

	h.GetResume(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/asset-owned-by-user-1", nil), "asset-owned-by-user-1")

	assertAPIError(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
}

func TestGetResumeFixtureParity(t *testing.T) {
	fixture := loadGetFixture(t)
	t.Run("default", func(t *testing.T) {
		want := fixture.Scenarios["default"].Response.Body
		svc := &fakeGetService{out: want.Resume}
		h := resumehandler.New(resumehandler.Options{
			Service: svc,
			Session: func(context.Context) (string, bool) {
				return "user-1", true
			},
		})
		rec := httptest.NewRecorder()

		h.GetResume(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/"+want.Resume.Id, nil), want.Resume.Id)

		if rec.Code != fixture.Scenarios["default"].Response.Status {
			t.Fatalf("status = %d want %d body=%s", rec.Code, fixture.Scenarios["default"].Response.Status, rec.Body.String())
		}
		var got api.ResumeAsset
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !reflect.DeepEqual(got, want.Resume) {
			t.Fatalf("fixture parity mismatch\ngot:  %+v\nwant: %+v", got, want.Resume)
		}
	})
	t.Run("not-found", func(t *testing.T) {
		want := fixture.Scenarios["not-found"].Response.Body.Error
		h := resumehandler.New(resumehandler.Options{
			Service: &fakeGetService{err: resume.ErrNotFound},
			Session: func(context.Context) (string, bool) {
				return "user-1", true
			},
		})
		rec := httptest.NewRecorder()

		h.GetResume(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/missing", nil), "missing")

		if rec.Code != fixture.Scenarios["not-found"].Response.Status {
			t.Fatalf("status = %d want %d body=%s", rec.Code, fixture.Scenarios["not-found"].Response.Status, rec.Body.String())
		}
		var got api.ApiErrorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		got.Error.RequestID = want.Error.RequestID
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("fixture parity mismatch\ngot:  %+v\nwant: %+v", got, want)
		}
	})
}

type fakeGetService struct {
	userID  string
	assetID string
	out     api.ResumeAsset
	err     error
}

func (s *fakeGetService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, nil
}

func (s *fakeGetService) GetResume(_ context.Context, userID string, assetID string) (api.ResumeAsset, error) {
	s.userID = userID
	s.assetID = assetID
	return s.out, s.err
}

type getFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int `json:"status"`
			Body   struct {
				Resume api.ResumeAsset
				Error  api.ApiErrorResponse
			}
		} `json:"response"`
	} `json:"scenarios"`
}

func loadGetFixture(t *testing.T) getFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "getResume.json"))
	if err != nil {
		t.Fatalf("read get fixture: %v", err)
	}
	var wire struct {
		Scenarios map[string]struct {
			Response struct {
				Status int             `json:"status"`
				Body   json.RawMessage `json:"body"`
			} `json:"response"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatalf("decode get fixture: %v", err)
	}
	out := getFixture{Scenarios: make(map[string]struct {
		Response struct {
			Status int `json:"status"`
			Body   struct {
				Resume api.ResumeAsset
				Error  api.ApiErrorResponse
			}
		} `json:"response"`
	}, len(wire.Scenarios))}
	for name, scenario := range wire.Scenarios {
		entry := out.Scenarios[name]
		entry.Response.Status = scenario.Response.Status
		if scenario.Response.Status >= 400 {
			if err := json.Unmarshal(scenario.Response.Body, &entry.Response.Body.Error); err != nil {
				t.Fatalf("decode get fixture error scenario %q: %v", name, err)
			}
		} else {
			if err := json.Unmarshal(scenario.Response.Body, &entry.Response.Body.Resume); err != nil {
				t.Fatalf("decode get fixture success scenario %q: %v", name, err)
			}
		}
		out.Scenarios[name] = entry
	}
	for _, scenario := range []string{"default", "not-found"} {
		if _, ok := out.Scenarios[scenario]; !ok {
			t.Fatalf("getResume fixture missing scenario %q", scenario)
		}
	}
	return out
}
