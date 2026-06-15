package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	"github.com/monshunter/easyinterview/backend/internal/middleware/idempotency"
	"github.com/monshunter/easyinterview/backend/internal/resume"
	resumehandler "github.com/monshunter/easyinterview/backend/internal/resume/handler"
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
)

func TestHandlerImplementsDuplicateResumeSurface(t *testing.T) {
	var _ interface {
		DuplicateResume(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestDuplicateResumeReturns201(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)
	svc := &fakeDuplicateResumeService{out: api.Resume{
		Id:          "resume-new",
		Title:       "Resume",
		DisplayName: "New CV",
		Language:    "en",
		ParseStatus: sharedtypes.TargetJobParseStatusReady,
		CreatedAt:   now,
		UpdatedAt:   now,
	}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.DuplicateResume(rec, newDuplicateResumeRequest(`{"displayName":"New CV","structuredProfile":{"headline":"new"}}`), "source-1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.in.UserID != "user-1" || svc.in.SourceResumeID != "source-1" || !svc.in.StructuredProfileSet || svc.in.StructuredProfile["headline"] != "new" {
		t.Fatalf("duplicate input = %+v", svc.in)
	}
}

func TestDuplicateResumeAllowsEmptyBody(t *testing.T) {
	svc := &fakeDuplicateResumeService{out: api.Resume{Id: "resume-new"}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.DuplicateResume(rec, newDuplicateResumeRequest(``), "source-1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.in.SourceResumeID != "source-1" || svc.in.StructuredProfile != nil {
		t.Fatalf("duplicate input = %+v", svc.in)
	}
}

func TestDuplicateResumePreservesExplicitEmptyStructuredProfile(t *testing.T) {
	svc := &fakeDuplicateResumeService{out: api.Resume{Id: "resume-new"}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.DuplicateResume(rec, newDuplicateResumeRequest(`{"structuredProfile":{}}`), "source-1")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !svc.in.StructuredProfileSet || len(svc.in.StructuredProfile) != 0 {
		t.Fatalf("duplicate input = %+v", svc.in)
	}
}

func TestDuplicateResumeValidationAndErrors(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		err    error
		status int
		code   string
	}{
		{name: "non editable field", body: `{"title":"x"}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "not found", body: `{}`, err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeResourceNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeDuplicateResumeService{err: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.DuplicateResume(rec, newDuplicateResumeRequest(tc.body), "source-1")

			assertAPIError(t, rec, tc.status, tc.code)
		})
	}
}

func TestDuplicateResumeRequiresIdempotencyKey(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeDuplicateResumeService{},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/source-1/duplicate", strings.NewReader(`{}`))

	h.DuplicateResume(rec, req, "source-1")

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestDuplicateResumeFixtureParity(t *testing.T) {
	fixture := loadResumeBodyFixture(t, "duplicateResume.json")
	entry := fixture.Scenarios["default"]
	var want api.Resume
	if err := json.Unmarshal(entry.Response.Body, &want); err != nil {
		t.Fatalf("decode duplicate fixture body: %v", err)
	}
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeDuplicateResumeService{out: want},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.DuplicateResume(rec, newDuplicateResumeRequest(string(entry.Request.Body)), "source-1")

	if rec.Code != entry.Response.Status {
		t.Fatalf("status = %d want %d body=%s", rec.Code, entry.Response.Status, rec.Body.String())
	}
	var got api.Resume
	decodeResponse(t, rec, &got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fixture parity mismatch\ngot:  %+v\nwant: %+v", got, want)
	}
}

type fakeDuplicateResumeService struct {
	fakeRegisterService
	in  resume.DuplicateResumeRequest
	out api.Resume
	err error
}

func (s *fakeDuplicateResumeService) DuplicateResume(_ context.Context, in resume.DuplicateResumeRequest) (api.Resume, error) {
	s.in = in
	return s.out, s.err
}

func newDuplicateResumeRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/source-1/duplicate", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-duplicate")
	return req
}
