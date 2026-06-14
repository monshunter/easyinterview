package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestArchiveResumeReturns202AndScopesUser(t *testing.T) {
	now := time.Date(2026, 6, 14, 8, 50, 0, 0, time.UTC).Format(time.RFC3339)
	svc := &fakeArchiveResumeService{out: api.Resume{
		Id:          "resume-1",
		Title:       "Resume",
		DisplayName: "Archived CV",
		Language:    "en",
		ParseStatus: sharedtypes.TargetJobParseStatusReady,
		Status:      ptrString("archived"),
		DeletedAt:   &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.ArchiveResume(rec, newArchiveResumeRequest(), "resume-1")

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.userID != "user-1" || svc.resumeID != "resume-1" {
		t.Fatalf("archive scope user=%q resume=%q", svc.userID, svc.resumeID)
	}
	if rec.Header().Get("X-Idempotency-Resource-Type") != "resume" || rec.Header().Get("X-Idempotency-Resource-ID") != "resume-1" {
		t.Fatalf("response resource headers type=%q id=%q", rec.Header().Get("X-Idempotency-Resource-Type"), rec.Header().Get("X-Idempotency-Resource-ID"))
	}
}

func TestArchiveResumeAlreadyArchivedReturnsConflict(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeArchiveResumeService{err: resume.ErrAlreadyArchived},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.ArchiveResume(rec, newArchiveResumeRequest(), "resume-1")

	assertAPIError(t, rec, http.StatusConflict, sharederrors.CodeTargetInvalidStateTransition)
}

func TestArchiveResumeRequiresIdempotencyKey(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeArchiveResumeService{},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/resume-1/archive", nil)

	h.ArchiveResume(rec, req, "resume-1")

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

type fakeArchiveResumeService struct {
	fakeRegisterService
	userID   string
	resumeID string
	out      api.Resume
	err      error
}

func (s *fakeArchiveResumeService) ArchiveResume(_ context.Context, userID string, resumeID string) (api.Resume, error) {
	s.userID = userID
	s.resumeID = resumeID
	return s.out, s.err
}

func newArchiveResumeRequest() *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes/resume-1/archive", strings.NewReader(`{}`))
	req.Header.Set(idempotency.HeaderName, "idem-archive")
	return req
}

func ptrString(in string) *string {
	return &in
}
