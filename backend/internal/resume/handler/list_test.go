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

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestHandlerImplementsListResumesSurface(t *testing.T) {
	var _ interface {
		ListResumes(http.ResponseWriter, *http.Request)
	} = (*resumehandler.Handler)(nil)
}

func TestListResumesPassesPaginationAndUserScope(t *testing.T) {
	svc := &fakeListService{out: api.PaginatedResumeAsset{
		Items: []api.ResumeAsset{{
			Id:          "asset-1",
			Title:       "Resume",
			Language:    "en",
			ParseStatus: sharedtypes.TargetJobParseStatusReady,
			CreatedAt:   "2026-05-13T01:00:00Z",
			UpdatedAt:   "2026-05-13T01:00:00Z",
		}},
		PageInfo: api.PageInfo{PageSize: 5, HasMore: true, NextCursor: strPtr("cursor-2")},
	}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resumes?pageSize=5&cursor=cursor-1", nil)
	rec := httptest.NewRecorder()

	h.ListResumes(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.in.UserID != "user-1" || svc.in.PageSize != 5 || svc.in.Cursor != "cursor-1" {
		t.Fatalf("ListResumes input = %+v", svc.in)
	}
}

func TestListResumesFixtureParity(t *testing.T) {
	fixture := loadListFixture(t)
	for _, scenario := range []string{"default", "empty", "paginated"} {
		t.Run(scenario, func(t *testing.T) {
			want := fixture.Scenarios[scenario].Response.Body
			svc := &fakeListService{out: want}
			h := resumehandler.New(resumehandler.Options{
				Service: svc,
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/resumes?pageSize=20", nil)
			rec := httptest.NewRecorder()

			h.ListResumes(rec, req)

			if rec.Code != fixture.Scenarios[scenario].Response.Status {
				t.Fatalf("status = %d want %d body=%s", rec.Code, fixture.Scenarios[scenario].Response.Status, rec.Body.String())
			}
			var got api.PaginatedResumeAsset
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("fixture parity mismatch\ngot:  %+v\nwant: %+v", got, want)
			}
		})
	}
}

type fakeListService struct {
	fakeRegisterService
	in  resume.ListRequest
	out api.PaginatedResumeAsset
	err error
}

func (s *fakeListService) ListResumes(_ context.Context, in resume.ListRequest) (api.PaginatedResumeAsset, error) {
	s.in = in
	return s.out, s.err
}

func strPtr(v string) *string { return &v }

type listFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int                      `json:"status"`
			Body   api.PaginatedResumeAsset `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadListFixture(t *testing.T) listFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "listResumes.json"))
	if err != nil {
		t.Fatalf("read list fixture: %v", err)
	}
	var fixture listFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode list fixture: %v", err)
	}
	return fixture
}
