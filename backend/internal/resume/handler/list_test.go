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
	resumestore "github.com/monshunter/easyinterview/backend/internal/resume/store"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestHandlerImplementsListResumesSurface(t *testing.T) {
	var _ interface {
		ListResumes(http.ResponseWriter, *http.Request)
	} = (*resumehandler.Handler)(nil)
}

func TestListResumesPassesPaginationAndUserScope(t *testing.T) {
	svc := &fakeListService{out: api.PaginatedResume{
		Items: []api.ResumeSummary{{
			Id:                 "resume-1",
			Title:              "Resume",
			DisplayName:        "Alice CV",
			Language:           "en",
			SourceType:         "paste",
			ParseStatus:        sharedtypes.TargetJobParseStatusReady,
			SummaryHeadline:    strPtr("Senior engineer"),
			HasReadableContent: true,
			UpdatedAt:          "2026-06-13T01:00:00Z",
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
	for _, scenario := range []string{"default", "empty", "paginated", "projection-boundaries"} {
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
			var got api.PaginatedResume
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("fixture parity mismatch\ngot:  %+v\nwant: %+v", got, want)
			}
			assertClosedResumeSummaryKeys(t, rec.Body.Bytes())
		})
	}
}

func TestListResumesInvalidCursorReturnsUnprocessableEntity(t *testing.T) {
	svc := &fakeListService{err: resumestore.ErrInvalidCursor}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/resumes?cursor=not-a-valid-cursor", nil)
	rec := httptest.NewRecorder()

	h.ListResumes(rec, req)

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

type fakeListService struct {
	fakeRegisterService
	in  resume.ListRequest
	out api.PaginatedResume
	err error
}

func (s *fakeListService) ListResumes(_ context.Context, in resume.ListRequest) (api.PaginatedResume, error) {
	s.in = in
	return s.out, s.err
}

func strPtr(v string) *string { return &v }

func assertClosedResumeSummaryKeys(t *testing.T, body []byte) {
	t.Helper()
	var envelope struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("decode summary keys: %v", err)
	}
	allowed := map[string]struct{}{
		"id": {}, "title": {}, "displayName": {}, "language": {}, "sourceType": {},
		"parseStatus": {}, "summaryHeadline": {}, "hasReadableContent": {}, "updatedAt": {},
	}
	for index, item := range envelope.Items {
		if len(item) != len(allowed) {
			t.Fatalf("item[%d] keys = %v, want exact closed summary keys", index, mapKeys(item))
		}
		for key := range item {
			if _, ok := allowed[key]; !ok {
				t.Fatalf("item[%d] contains forbidden detail field %q", index, key)
			}
		}
		for key := range allowed {
			if _, ok := item[key]; !ok {
				t.Fatalf("item[%d] missing required summary field %q", index, key)
			}
		}
	}
}

func mapKeys(in map[string]json.RawMessage) []string {
	out := make([]string, 0, len(in))
	for key := range in {
		out = append(out, key)
	}
	return out
}

type listFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int                 `json:"status"`
			Body   api.PaginatedResume `json:"body"`
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
