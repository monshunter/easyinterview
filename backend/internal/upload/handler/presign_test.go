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
	sharederrors "github.com/monshunter/easyinterview/backend/internal/shared/errors"
	uploadhandler "github.com/monshunter/easyinterview/backend/internal/upload/handler"
	uploadservice "github.com/monshunter/easyinterview/backend/internal/upload/service"
)

func TestHandlerImplementsCreateUploadPresignSurface(t *testing.T) {
	var _ interface {
		CreateUploadPresign(http.ResponseWriter, *http.Request)
	} = (*uploadhandler.Handler)(nil)
}

func TestCreateUploadPresignRequiresIdempotencyKey(t *testing.T) {
	h := newTestHandler(&fakePresignService{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", strings.NewReader(`{"purpose":"resume","fileName":"resume.pdf","contentType":"application/pdf","byteSize":128}`))
	rec := httptest.NewRecorder()

	h.CreateUploadPresign(rec, req)

	assertAPIError(t, rec, http.StatusBadRequest, sharederrors.CodeValidationFailed)
}

func TestCreateUploadPresignIdempotencyReplayAndTTL(t *testing.T) {
	store := &fakeIdempotencyStore{
		reservation: idempotency.Reservation{
			State:          idempotency.StateReplay,
			RecordID:       "idem-rec-1",
			ResponseStatus: http.StatusCreated,
			ResponseBody:   []byte(`{"fileObjectId":"file-replay","uploadUrl":"http://upload","method":"PUT","headers":{},"expiresAt":"2026-05-12T00:10:00Z"}`),
		},
	}
	h := newTestHandler(&fakePresignService{})
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store:     store,
		Now:       func() time.Time { return time.Date(2026, 5, 12, 0, 0, 0, 0, time.UTC) },
		KeyPepper: "pepper",
	})
	wrapped := mw.Handler("upload", "createUploadPresign", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.CreateUploadPresign))
	req := newPresignRequest(`{"purpose":"resume","fileName":"resume.pdf","contentType":"application/pdf","byteSize":128}`)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get(idempotency.ReplayHeader) != "true" {
		t.Fatalf("expected replay header, got %q", rec.Header().Get(idempotency.ReplayHeader))
	}
	if store.reserveIn.ExpiresAt.Sub(store.reserveIn.Now) != 24*time.Hour {
		t.Fatalf("idempotency ttl = %s", store.reserveIn.ExpiresAt.Sub(store.reserveIn.Now))
	}
	if !strings.Contains(rec.Body.String(), "file-replay") {
		t.Fatalf("replay body = %s", rec.Body.String())
	}
}

func TestCreateUploadPresignRejectsExpiredOrMismatchedIdempotencyKey(t *testing.T) {
	store := &fakeIdempotencyStore{err: idempotency.ErrFingerprintMismatch}
	h := newTestHandler(&fakePresignService{})
	mw := idempotency.New(idempotency.MiddlewareOptions{
		Store: store,
		Now:   func() time.Time { return time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC) },
	})
	wrapped := mw.Handler("upload", "createUploadPresign", func(*http.Request) (string, bool) { return "user-1", true }, http.HandlerFunc(h.CreateUploadPresign))
	req := newPresignRequest(`{"purpose":"resume","fileName":"resume.pdf","contentType":"application/pdf","byteSize":128}`)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assertAPIError(t, rec, http.StatusConflict, sharederrors.CodeIdempotencyKeyMismatch)
}

func TestCreateUploadPresignPurposeValidation(t *testing.T) {
	h := newTestHandler(&fakePresignService{})
	req := newPresignRequest(`{"purpose":"unknown","fileName":"resume.pdf","contentType":"application/pdf","byteSize":128}`)
	rec := httptest.NewRecorder()

	h.CreateUploadPresign(rec, req)

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestCreateUploadPresignByteSizeLimit(t *testing.T) {
	h := newTestHandler(&fakePresignService{})
	req := newPresignRequest(`{"purpose":"privacy_export","fileName":"privacy.zip","contentType":"application/zip","byteSize":5242881}`)
	rec := httptest.NewRecorder()

	h.CreateUploadPresign(rec, req)

	assertAPIError(t, rec, http.StatusUnprocessableEntity, sharederrors.CodeValidationFailed)
}

func TestCreateUploadPresignReturnsCreatedResponse(t *testing.T) {
	svc := &fakePresignService{out: api.UploadPresign{
		FileObjectId: "01918fa0-0000-7000-8000-000000001100",
		UploadUrl:    "https://uploads.acme.example/presigned/upload?token=abc",
		Method:       "PUT",
		Headers: map[string]any{
			"Content-Type":                 "application/pdf",
			"x-amz-server-side-encryption": "AES256",
		},
		ExpiresAt: "2026-04-28T14:00:00Z",
	}}
	h := newTestHandler(svc)
	req := newPresignRequest(`{"purpose":"resume","fileName":"alice-resume-2026.pdf","contentType":"application/pdf","byteSize":248192}`)
	rec := httptest.NewRecorder()

	h.CreateUploadPresign(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var got api.UploadPresign
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.FileObjectId != svc.out.FileObjectId || got.Method != "PUT" || got.UploadUrl == "" || got.ExpiresAt == "" {
		t.Fatalf("response = %+v", got)
	}
	if svc.in.UserID != "user-1" || svc.in.IdempotencyKey != "idem-1" || svc.in.Purpose != "resume" {
		t.Fatalf("service input = %+v", svc.in)
	}
	if svc.in.PresignTTL != 10*time.Minute || svc.in.MaxBytes != 10485760 {
		t.Fatalf("config input ttl=%s max=%d", svc.in.PresignTTL, svc.in.MaxBytes)
	}
}

func TestCreateUploadPresignFixtureParity(t *testing.T) {
	var fixture struct {
		Scenarios map[string]struct {
			Response struct {
				Status int               `json:"status"`
				Body   api.UploadPresign `json:"body"`
				Header map[string]string `json:"headers"`
				Any    map[string]any    `json:"-"`
			} `json:"response"`
		} `json:"scenarios"`
	}
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "openapi", "fixtures", "Uploads", "createUploadPresign.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	defaultScenario := fixture.Scenarios["default"]
	h := newTestHandler(&fakePresignService{out: defaultScenario.Response.Body})
	req := newPresignRequest(`{"purpose":"resume","fileName":"alice-resume-2026.pdf","contentType":"application/pdf","byteSize":248192}`)
	rec := httptest.NewRecorder()

	h.CreateUploadPresign(rec, req)

	if rec.Code != defaultScenario.Response.Status {
		t.Fatalf("status = %d, fixture = %d", rec.Code, defaultScenario.Response.Status)
	}
	var got map[string]any
	var want map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	wantRaw, _ := json.Marshal(defaultScenario.Response.Body)
	if err := json.Unmarshal(wantRaw, &want); err != nil {
		t.Fatalf("decode fixture body: %v", err)
	}
	if !reflect.DeepEqual(sortedKeys(got), sortedKeys(want)) {
		t.Fatalf("response keys = %v, fixture keys = %v", sortedKeys(got), sortedKeys(want))
	}
	if !reflect.DeepEqual(sortedKeys(got["headers"].(map[string]any)), sortedKeys(want["headers"].(map[string]any))) {
		t.Fatalf("header keys = %v, fixture header keys = %v", sortedKeys(got["headers"].(map[string]any)), sortedKeys(want["headers"].(map[string]any)))
	}
}

func sortedKeys(in map[string]any) []string {
	keys := make([]string, 0, len(in))
	for key := range in {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func newTestHandler(svc uploadhandler.PresignService) *uploadhandler.Handler {
	return uploadhandler.New(uploadhandler.Options{
		Service: svc,
		Session: func(context.Context) (string, bool) {
			return "user-1", true
		},
		PresignTTL: 10 * time.Minute,
		MaxBytesByPurpose: map[string]int64{
			"resume":                10485760,
			"target_job_attachment": 10485760,
			"privacy_export":        5242880,
		},
	})
}

func newPresignRequest(body string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/uploads/presign", strings.NewReader(body))
	req.Header.Set(idempotency.HeaderName, "idem-1")
	return req
}

func assertAPIError(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var out api.ApiErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if out.Error.Code != code {
		t.Fatalf("error code = %q body=%s", out.Error.Code, rec.Body.String())
	}
}

type fakePresignService struct {
	in  uploadservice.CreatePresignInput
	out api.UploadPresign
	err error
}

func (s *fakePresignService) CreateUploadPresign(_ context.Context, in uploadservice.CreatePresignInput) (api.UploadPresign, error) {
	s.in = in
	if s.err != nil {
		return api.UploadPresign{}, s.err
	}
	if s.out.FileObjectId == "" {
		return api.UploadPresign{FileObjectId: "file-1", UploadUrl: "http://upload", Method: "PUT", Headers: map[string]any{}, ExpiresAt: "2026-05-12T00:10:00Z"}, nil
	}
	return s.out, nil
}

type fakeIdempotencyStore struct {
	reserveIn   idempotency.ReservationInput
	reservation idempotency.Reservation
	err         error
}

func (s *fakeIdempotencyStore) Reserve(_ context.Context, in idempotency.ReservationInput) (idempotency.Reservation, error) {
	s.reserveIn = in
	if s.err != nil {
		return idempotency.Reservation{}, s.err
	}
	return s.reservation, nil
}

func (s *fakeIdempotencyStore) MarkSucceeded(context.Context, idempotency.CompletionInput) error {
	return nil
}

func (s *fakeIdempotencyStore) MarkFailed(context.Context, idempotency.CompletionInput) error {
	return nil
}
