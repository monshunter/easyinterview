package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestHandlerImplementsUpdateResumeSurface(t *testing.T) {
	var _ interface {
		UpdateResume(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestUpdateResumeOverwritesEditableFields(t *testing.T) {
	now := time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC).Format(time.RFC3339)
	svc := &fakeUpdateResumeService{out: api.Resume{
		Id:          "resume-1",
		Title:       "Resume",
		DisplayName: "Updated CV",
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

	h.UpdateResume(rec, newUpdateResumeRequest(`{"displayName":"Updated CV","structuredProfile":{"headline":"Staff engineer"}}`), "resume-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.in.UserID != "user-1" || svc.in.ResumeID != "resume-1" || !svc.in.DisplayNameSet || !svc.in.StructuredProfileSet || svc.in.StructuredProfile["headline"] != "Staff engineer" {
		t.Fatalf("update input = %+v", svc.in)
	}
}

func TestUpdateResumeValidationAndErrors(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		err    error
		status int
		code   string
	}{
		{name: "empty body", body: `{}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "non editable field", body: `{"title":"x"}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "null structured profile", body: `{"structuredProfile":null}`, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
		{name: "not found", body: `{"displayName":"x"}`, err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeResourceNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeUpdateResumeService{err: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.UpdateResume(rec, newUpdateResumeRequest(tc.body), "resume-1")

			assertAPIError(t, rec, tc.status, tc.code)
		})
	}
}

func TestUpdateResumeRequiresIdempotencyKey(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeUpdateResumeService{},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/resumes/resume-1", strings.NewReader(`{"displayName":"x"}`))

	h.UpdateResume(rec, req, "resume-1")

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestUpdateResumeFixtureParity(t *testing.T) {
	fixture := loadResumeBodyFixture(t, "updateResume.json")
	entry := fixture.Scenarios["default"]
	var want api.Resume
	if err := json.Unmarshal(entry.Response.Body, &want); err != nil {
		t.Fatalf("decode update fixture body: %v", err)
	}
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeUpdateResumeService{out: want},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.UpdateResume(rec, newUpdateResumeRequest(string(entry.Request.Body)), want.Id)

	if rec.Code != entry.Response.Status {
		t.Fatalf("status = %d want %d body=%s", rec.Code, entry.Response.Status, rec.Body.String())
	}
	var got api.Resume
	decodeResponse(t, rec, &got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fixture parity mismatch\ngot:  %+v\nwant: %+v", got, want)
	}
}

type fakeUpdateResumeService struct {
	fakeRegisterService
	in  resume.UpdateResumeRequest
	out api.Resume
	err error
}

func (s *fakeUpdateResumeService) UpdateResume(_ context.Context, in resume.UpdateResumeRequest) (api.Resume, error) {
	s.in = in
	return s.out, s.err
}

func newUpdateResumeRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/resumes/resume-1", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-update")
	return req
}

type resumeBodyFixtureEntry struct {
	Request struct {
		Body json.RawMessage `json:"body"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

type resumeBodyFixture struct {
	Scenarios map[string]resumeBodyFixtureEntry `json:"scenarios"`
}

func loadResumeBodyFixture(t *testing.T, name string) resumeBodyFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", name))
	if err != nil {
		t.Fatalf("read %s fixture: %v", name, err)
	}
	var fixture resumeBodyFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode %s fixture: %v", name, err)
	}
	return fixture
}
