package handler_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestHandlerImplementsResumeVersionReadSurface(t *testing.T) {
	var _ interface {
		GetResumeVersion(http.ResponseWriter, *http.Request, string)
		ListResumeVersions(http.ResponseWriter, *http.Request, string)
	} = (*resumehandler.Handler)(nil)
}

func TestGetResumeVersion(t *testing.T) {
	now := time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC)
	version := versionResponse("version-1", "asset-1", now)
	svc := &fakeVersionReadService{getOut: version}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.GetResumeVersion(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resume-versions/version-1", nil), "version-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.getUserID != "user-1" || svc.getVersionID != "version-1" {
		t.Fatalf("get scope user=%q version=%q", svc.getUserID, svc.getVersionID)
	}
	var got api.ResumeVersion
	decodeResponse(t, rec, &got)
	if got.Id != "version-1" || got.ResumeAssetId != "asset-1" {
		t.Fatalf("response = %+v", got)
	}
}

func TestGetResumeVersionErrors(t *testing.T) {
	h := resumehandler.New(resumehandler.Options{
		Service: &fakeVersionReadService{getErr: resume.ErrNotFound},
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.GetResumeVersion(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resume-versions/missing", nil), "missing")

	assertAPIError(t, rec, http.StatusNotFound, sharederrors.CodeTargetJobNotFound)
}

func TestListResumeVersions(t *testing.T) {
	now := time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC)
	svc := &fakeVersionReadService{listOut: api.PaginatedResumeVersion{
		Items: []api.ResumeVersion{
			versionResponse("version-2", "asset-1", now.Add(time.Minute)),
			versionResponse("version-1", "asset-1", now),
		},
		PageInfo: api.PageInfo{PageSize: 2, HasMore: true, NextCursor: ptrString("cursor-2")},
	}}
	h := resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) { return "user-1", true },
	})
	rec := httptest.NewRecorder()

	h.ListResumeVersions(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/asset-1/versions?pageSize=2&cursor=cursor-1", nil), "asset-1")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if svc.listIn.UserID != "user-1" || svc.listIn.ResumeAssetID != "asset-1" || svc.listIn.PageSize != 2 || svc.listIn.Cursor != "cursor-1" {
		t.Fatalf("list input = %+v", svc.listIn)
	}
	var got api.PaginatedResumeVersion
	decodeResponse(t, rec, &got)
	if len(got.Items) != 2 || got.PageInfo.NextCursor == nil || *got.PageInfo.NextCursor != "cursor-2" {
		t.Fatalf("response = %+v", got)
	}
}

func TestListResumeVersionsErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{name: "asset not found", err: resume.ErrNotFound, status: http.StatusNotFound, code: sharederrors.CodeTargetJobNotFound},
		{name: "invalid cursor", err: resume.ErrInvalidCursor, status: http.StatusUnprocessableEntity, code: sharederrors.CodeValidationFailed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := resumehandler.New(resumehandler.Options{
				Service: &fakeVersionReadService{listErr: tc.err},
				Session: func(context.Context) (string, bool) { return "user-1", true },
			})
			rec := httptest.NewRecorder()

			h.ListResumeVersions(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/asset-1/versions", nil), "asset-1")

			assertAPIError(t, rec, tc.status, tc.code)
		})
	}
}

func TestResumeVersionReadFixtureParity(t *testing.T) {
	t.Run("get default", func(t *testing.T) {
		fixture := loadGetVersionFixture(t)
		for _, scenario := range []string{"default", "not-found-404"} {
			t.Run(scenario, func(t *testing.T) {
				want := fixture.Scenarios[scenario].Response
				var version api.ResumeVersion
				if want.Status == http.StatusOK {
					if err := json.Unmarshal(want.Body, &version); err != nil {
						t.Fatalf("decode version body: %v", err)
					}
				}
				svc := &fakeVersionReadService{getOut: version}
				if want.Status == http.StatusNotFound {
					svc.getErr = resume.ErrNotFound
				}
				h := resumehandler.New(resumehandler.Options{
					Service: svc,
					Session: func(context.Context) (string, bool) { return "user-1", true },
				})
				rec := httptest.NewRecorder()

				h.GetResumeVersion(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resume-versions/version-1", nil), "version-1")

				assertRawJSONEqual(t, rec, want.Status, want.Body)
			})
		}
	})

	t.Run("list scenarios", func(t *testing.T) {
		fixture := loadListVersionFixture(t)
		for _, scenario := range []string{"default", "empty", "paginated"} {
			t.Run(scenario, func(t *testing.T) {
				want := fixture.Scenarios[scenario].Response
				var page api.PaginatedResumeVersion
				if err := json.Unmarshal(want.Body, &page); err != nil {
					t.Fatalf("decode list body: %v", err)
				}
				h := resumehandler.New(resumehandler.Options{
					Service: &fakeVersionReadService{listOut: page},
					Session: func(context.Context) (string, bool) { return "user-1", true },
				})
				rec := httptest.NewRecorder()

				h.ListResumeVersions(rec, httptest.NewRequest(http.MethodGet, "/api/v1/resumes/asset-1/versions", nil), "asset-1")

				assertRawJSONEqual(t, rec, want.Status, want.Body)
			})
		}
	})
}

type fakeVersionReadService struct {
	getUserID    string
	getVersionID string
	getOut       api.ResumeVersion
	getErr       error

	listIn  resume.ListVersionRequest
	listOut api.PaginatedResumeVersion
	listErr error
}

func (s *fakeVersionReadService) RegisterResume(context.Context, resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	return api.ResumeAssetWithJob{}, errors.New("not implemented")
}

func (s *fakeVersionReadService) GetResumeVersion(_ context.Context, userID string, versionID string) (api.ResumeVersion, error) {
	s.getUserID = userID
	s.getVersionID = versionID
	return s.getOut, s.getErr
}

func (s *fakeVersionReadService) ListResumeVersions(_ context.Context, in resume.ListVersionRequest) (api.PaginatedResumeVersion, error) {
	s.listIn = in
	return s.listOut, s.listErr
}

func versionResponse(id, assetID string, now time.Time) api.ResumeVersion {
	return api.ResumeVersion{
		Id:            id,
		ResumeAssetId: assetID,
		VersionType:   sharedtypes.ResumeVersionTypeStructuredMaster,
		DisplayName:   "Structured master",
		StructuredProfile: map[string]any{"headline": "Senior engineer", "provenance": map[string]any{
			"promptVersion": "resume_profile.v1", "rubricVersion": "not_applicable", "modelId": "model-1", "language": "en", "featureFlag": "none", "dataSourceVersion": "asset.v1",
		}},
		Provenance: api.GenerationProvenance{
			PromptVersion: "resume_profile.v1", RubricVersion: "not_applicable", ModelId: "model-1", Language: "en", FeatureFlag: "none", DataSourceVersion: "asset.v1",
		},
		Suggestions: []any{},
		CreatedAt:   now.Format(time.RFC3339),
		UpdatedAt:   now.Format(time.RFC3339),
	}
}

func ptrString(in string) *string {
	return &in
}

type getVersionFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int             `json:"status"`
			Body   json.RawMessage `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

type listVersionFixture struct {
	Scenarios map[string]struct {
		Response struct {
			Status int             `json:"status"`
			Body   json.RawMessage `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadGetVersionFixture(t *testing.T) getVersionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "getResumeVersion.json"))
	if err != nil {
		t.Fatalf("read getResumeVersion fixture: %v", err)
	}
	var fixture getVersionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode getResumeVersion fixture: %v", err)
	}
	return fixture
}

func loadListVersionFixture(t *testing.T) listVersionFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "listResumeVersions.json"))
	if err != nil {
		t.Fatalf("read listResumeVersions fixture: %v", err)
	}
	var fixture listVersionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode listResumeVersions fixture: %v", err)
	}
	return fixture
}

func assertRawJSONEqual(t *testing.T, rec *httptest.ResponseRecorder, status int, want json.RawMessage) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d want %d body=%s", rec.Code, status, rec.Body.String())
	}
	var gotBody map[string]any
	var wantBody map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &gotBody); err != nil {
		t.Fatalf("decode got: %v", err)
	}
	if err := json.Unmarshal(want, &wantBody); err != nil {
		t.Fatalf("decode want: %v", err)
	}
	gotNormalized := normalizeFixtureJSON(gotBody)
	wantNormalized := normalizeFixtureJSON(wantBody)
	if !reflect.DeepEqual(gotNormalized, wantNormalized) {
		t.Fatalf("response body mismatch\ngot:  %#v\nwant: %#v", gotBody, wantBody)
	}
}

func normalizeFixtureJSON(in any) any {
	switch v := in.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			if value == nil || key == "requestId" {
				continue
			}
			out[key] = normalizeFixtureJSON(value)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, value := range v {
			out[i] = normalizeFixtureJSON(value)
		}
		return out
	default:
		return in
	}
}
