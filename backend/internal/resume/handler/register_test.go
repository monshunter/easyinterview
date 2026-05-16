package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
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

func TestHandlerImplementsRegisterResumeSurface(t *testing.T) {
	var _ interface {
		RegisterResume(http.ResponseWriter, *http.Request)
	} = (*resumehandler.Handler)(nil)
}

func TestRegisterSourceType(t *testing.T) {
	job := api.Job{
		Id:           "01918fa0-0000-7000-8000-000000000201",
		JobType:      api.JobTypeResumeParse,
		ResourceType: api.ResourceTypeResumeAsset,
		ResourceId:   "01918fa0-0000-7000-8000-000000000101",
		Status:       sharedtypes.JobStatusQueued,
		CreatedAt:    time.Date(2026, 5, 13, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
		UpdatedAt:    time.Date(2026, 5, 13, 1, 0, 0, 0, time.UTC).Format(time.RFC3339),
	}
	validCases := []struct {
		name string
		body string
		want resume.RegisterInput
	}{
		{
			name: "upload requires fileObjectId",
			body: `{"sourceType":"upload","fileObjectId":"01918fa0-0000-7000-8000-000000000301","title":"Resume","language":"en"}`,
			want: resume.RegisterInput{UserID: "user-1", IdempotencyKey: "idem-1", SourceType: "upload", FileObjectID: "01918fa0-0000-7000-8000-000000000301", Title: "Resume", Language: "en"},
		},
		{
			name: "paste requires rawText",
			body: `{"sourceType":"paste","rawText":"Senior platform resume text","title":"Resume","language":"en"}`,
			want: resume.RegisterInput{UserID: "user-1", IdempotencyKey: "idem-1", SourceType: "paste", RawText: "Senior platform resume text", Title: "Resume", Language: "en"},
		},
		{
			name: "guided requires guidedAnswers",
			body: `{"sourceType":"guided","guidedAnswers":{"role":"Platform Lead"},"title":"Resume","language":"zh-CN"}`,
			want: resume.RegisterInput{UserID: "user-1", IdempotencyKey: "idem-1", SourceType: "guided", GuidedAnswers: map[string]any{"role": "Platform Lead"}, Title: "Resume", Language: "zh-CN"},
		},
	}
	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeRegisterService{out: api.ResumeAssetWithJob{ResumeAssetId: job.ResourceId, Job: job}}
			h := newTestHandler(svc)
			rec := httptest.NewRecorder()

			h.RegisterResume(rec, newRegisterRequest(tc.body))

			if rec.Code != http.StatusAccepted {
				t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
			}
			if !equalRegisterInput(svc.in, tc.want) {
				t.Fatalf("service input = %#v, want %#v", svc.in, tc.want)
			}
		})
	}

	invalidBodies := []string{
		`{"sourceType":"upload","title":"Resume","language":"en"}`,
		`{"sourceType":"paste","title":"Resume","language":"en"}`,
		`{"sourceType":"guided","title":"Resume","language":"en"}`,
		`{"sourceType":"unknown","title":"Resume","language":"en"}`,
		`{"sourceType":"upload","fileObjectId":"01918fa0-0000-7000-8000-000000000301","rawText":"must not mix","title":"Resume","language":"en"}`,
	}
	for _, body := range invalidBodies {
		t.Run("rejects "+body, func(t *testing.T) {
			h := newTestHandler(&fakeRegisterService{})
			rec := httptest.NewRecorder()

			h.RegisterResume(rec, newRegisterRequest(body))

			assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
		})
	}
}

func TestRegisterIdempotency(t *testing.T) {
	t.Run("requires key", func(t *testing.T) {
		h := newTestHandler(&fakeRegisterService{})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes", strings.NewReader(`{"sourceType":"paste","rawText":"resume","title":"Resume","language":"en"}`))
		rec := httptest.NewRecorder()

		h.RegisterResume(rec, req)

		assertAPIError(t, rec, http.StatusBadRequest, sharederrors.CodeValidationFailed)
	})

	t.Run("replays cached response with 24h ttl", func(t *testing.T) {
		store := &fakeIdempotencyStore{reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-1",
			ResponseStatus: http.StatusAccepted,
			ResponseBody:   []byte(`{"resumeAssetId":"asset-replay","job":{"id":"job-replay","jobType":"resume_parse","resourceType":"resume_asset","resourceId":"asset-replay","status":"queued","createdAt":"2026-05-13T01:00:00Z","updatedAt":"2026-05-13T01:00:00Z"}}`),
		}}
		h := newTestHandler(&fakeRegisterService{})
		mw := idempotency.New(idempotency.MiddlewareOptions{
			Store: store,
			Now:   func() time.Time { return time.Date(2026, 5, 13, 1, 0, 0, 0, time.UTC) },
		})
		wrapped := mw.Handler("resume", "registerResume", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.RegisterResume))
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, newRegisterRequest(`{"sourceType":"paste","rawText":"resume","title":"Resume","language":"en"}`))

		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get(idempotency.ReplayHeader) != "true" {
			t.Fatalf("replay header = %q", rec.Header().Get(idempotency.ReplayHeader))
		}
		if store.reserveIn.ExpiresAt.Sub(store.reserveIn.Now) != 24*time.Hour {
			t.Fatalf("idempotency ttl = %s", store.reserveIn.ExpiresAt.Sub(store.reserveIn.Now))
		}
		if !strings.Contains(rec.Body.String(), "asset-replay") {
			t.Fatalf("replay body = %s", rec.Body.String())
		}
	})

	t.Run("rejects mismatch", func(t *testing.T) {
		store := &fakeIdempotencyStore{err: idempotency.ErrFingerprintMismatch}
		h := newTestHandler(&fakeRegisterService{})
		mw := idempotency.New(idempotency.MiddlewareOptions{Store: store})
		wrapped := mw.Handler("resume", "registerResume", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.RegisterResume))
		rec := httptest.NewRecorder()

		wrapped.ServeHTTP(rec, newRegisterRequest(`{"sourceType":"paste","rawText":"resume","title":"Resume","language":"en"}`))

		assertAPIError(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
	})
}

func TestRegisterResumeValidationErrorsReturnUnprocessableEntity(t *testing.T) {
	h := newTestHandler(&fakeRegisterService{err: resume.ErrValidationFailed})
	rec := httptest.NewRecorder()

	h.RegisterResume(rec, newRegisterRequest(`{"sourceType":"upload","fileObjectId":"01918fa0-0000-7000-8000-000000000301","title":"Resume","language":"en"}`))

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestRegisterResumeFixtureParity(t *testing.T) {
	fixture := loadRegisterFixture(t)
	for _, scenario := range []string{"default", "paste-text", "guided-answers"} {
		t.Run(scenario, func(t *testing.T) {
			want := fixture.Scenarios[scenario].Response
			h := newTestHandler(&fakeRegisterService{out: want.Body})
			bodyRaw, _ := json.Marshal(fixture.Scenarios[scenario].Request.Body)
			rec := httptest.NewRecorder()

			h.RegisterResume(rec, newRegisterRequest(string(bodyRaw)))

			if rec.Code != want.Status {
				t.Fatalf("status = %d, want %d body=%s", rec.Code, want.Status, rec.Body.String())
			}
			var gotBody map[string]any
			var wantBody map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &gotBody); err != nil {
				t.Fatalf("decode got: %v", err)
			}
			wantRaw, _ := json.Marshal(want.Body)
			if err := json.Unmarshal(wantRaw, &wantBody); err != nil {
				t.Fatalf("decode want: %v", err)
			}
			if !reflect.DeepEqual(gotBody, wantBody) {
				t.Fatalf("response body mismatch\ngot:  %#v\nwant: %#v", gotBody, wantBody)
			}
		})
	}
}

type fakeRegisterService struct {
	in  resume.RegisterInput
	out api.ResumeAssetWithJob
	err error
}

func (s *fakeRegisterService) RegisterResume(_ context.Context, in resume.RegisterInput) (api.ResumeAssetWithJob, error) {
	s.in = in
	return s.out, s.err
}

func newTestHandler(svc resumehandler.RegisterService) *resumehandler.Handler {
	return resumehandler.New(resumehandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) {
			return "user-1", true
		},
	})
}

func newRegisterRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resumes", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-1")
	return req
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Error.Code != code {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, code)
	}
}

func equalRegisterInput(got, want resume.RegisterInput) bool {
	if got.UserID != want.UserID ||
		got.IdempotencyKey != want.IdempotencyKey ||
		got.SourceType != want.SourceType ||
		got.FileObjectID != want.FileObjectID ||
		got.RawText != want.RawText ||
		got.Title != want.Title ||
		got.Language != want.Language {
		return false
	}
	gotRaw, _ := json.Marshal(got.GuidedAnswers)
	wantRaw, _ := json.Marshal(want.GuidedAnswers)
	return string(gotRaw) == string(wantRaw)
}

type registerFixture struct {
	Scenarios map[string]struct {
		Request struct {
			Body map[string]any `json:"body"`
		} `json:"request"`
		Response struct {
			Status int                    `json:"status"`
			Body   api.ResumeAssetWithJob `json:"body"`
		} `json:"response"`
	} `json:"scenarios"`
}

func loadRegisterFixture(t *testing.T) registerFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Resumes", "registerResume.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture registerFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	for _, scenario := range []string{"default", "paste-text", "guided-answers"} {
		if _, ok := fixture.Scenarios[scenario]; !ok {
			t.Fatalf("fixture missing scenario %q; scenarios=%v", scenario, sortedScenarioKeys(fixture.Scenarios))
		}
	}
	return fixture
}

func sortedScenarioKeys(in map[string]struct {
	Request struct {
		Body map[string]any `json:"body"`
	} `json:"request"`
	Response struct {
		Status int                    `json:"status"`
		Body   api.ResumeAssetWithJob `json:"body"`
	} `json:"response"`
}) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type fakeIdempotencyStore struct {
	reserveIn   idempotency.ReservationInput
	completeIn  idempotency.CompletionInput
	reservation idempotency.Reservation
	err         error
}

func (s *fakeIdempotencyStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	s.reserveIn = in
	if s.err != nil {
		return idempotency.Reservation{}, s.err
	}
	if s.reservation.State == "" {
		return idempotency.Reservation{State: idempotency.StateExecute, RecordID: "idem-rec-1"}, nil
	}
	return s.reservation, nil
}

func (s *fakeIdempotencyStore) MarkSucceeded(_ context.Context, in idempotency.CompletionInput) error {
	s.completeIn = in
	return nil
}

func (s *fakeIdempotencyStore) MarkFailed(_ context.Context, in idempotency.CompletionInput) error {
	s.completeIn = in
	return nil
}
