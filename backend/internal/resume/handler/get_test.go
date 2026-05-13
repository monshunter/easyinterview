package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
