package targetjob_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
	sharedtypes "github.com/monshunter/easyinterview/backend/internal/shared/types"
	"github.com/monshunter/easyinterview/backend/internal/targetjob"
)

// targetJobOperationNames are the four B2-defined operations the targetjob
// domain owns. Renaming any of these (e.g., due to B2 churn) must surface
// as a test failure here so the wiring in cmd/api/main.go stays in sync.
var targetJobOperationNames = []string{
	"ImportTargetJob",
	"ListTargetJobs",
	"GetTargetJob",
	"UpdateTargetJob",
}

func TestHandlerSignaturesMatchB2ServerInterface(t *testing.T) {
	apiType := reflect.TypeFor[api.ServerInterface]()
	handlerType := reflect.TypeFor[*targetjob.Handler]()
	for _, name := range targetJobOperationNames {
		apiMethod, ok := apiType.MethodByName(name)
		if !ok {
			t.Fatalf("B2 ServerInterface missing %q — Phase 0 owner contract drifted", name)
		}
		handlerMethod, ok := handlerType.MethodByName(name)
		if !ok {
			t.Fatalf("targetjob.Handler missing %q — Phase 1.3 surface incomplete", name)
		}
		// apiMethod.Type is iface method type; handlerMethod.Type includes the
		// pointer receiver as the first argument. Strip it before comparing.
		want := apiMethod.Type
		gotIn := handlerMethod.Type.NumIn() - 1
		if want.NumIn() != gotIn {
			t.Errorf("%s: in-arg count mismatch: B2=%d, handler=%d", name, want.NumIn(), gotIn)
			continue
		}
		for i := 0; i < want.NumIn(); i++ {
			if want.In(i) != handlerMethod.Type.In(i+1) {
				t.Errorf("%s: arg %d mismatch: B2=%v, handler=%v", name, i, want.In(i), handlerMethod.Type.In(i+1))
			}
		}
		if want.NumOut() != handlerMethod.Type.NumOut() {
			t.Errorf("%s: out-arg count mismatch: B2=%d, handler=%d", name, want.NumOut(), handlerMethod.Type.NumOut())
		}
	}
}

func newWiredHandler(t *testing.T, ids ...string) (*targetjob.Handler, *fakeStore) {
	t.Helper()
	store := &fakeStore{}
	idx := 0
	gen := func() string {
		if idx >= len(ids) {
			t.Fatalf("ran out of injected ids; have %d, requesting more", len(ids))
		}
		v := ids[idx]
		idx++
		return v
	}
	now := time.Date(2026, 5, 9, 20, 0, 0, 0, time.UTC)
	svc := targetjob.NewService(targetjob.ServiceOptions{
		Store:        store,
		NewID:        gen,
		Now:          func() time.Time { return now },
		DedupePepper: "test-pepper",
	})
	h := targetjob.NewHandler(targetjob.HandlerOptions{
		Service: svc,
		Session: func(ctx context.Context) (string, bool) {
			if v, ok := ctx.Value(testUserKey{}).(string); ok && v != "" {
				return v, true
			}
			return "", false
		},
	})
	return h, store
}

type testUserKey struct{}

func TestHandler_ImportTargetJob_Returns202WithTargetJobWithJob(t *testing.T) {
	h, _ := newWiredHandler(t,
		"018f2a40-0000-7000-9000-0000000000a1", // target id
		"018f2a40-0000-7000-9000-0000000000f1", // job id
		"018f2a40-0000-7000-9000-0000000000c1", // source id
		"018f2a40-0000-7000-9000-0000000000e1", // outbox id
	)

	body := api.ImportTargetJobRequest{
		Source: map[string]any{
			"type":    "manual_text",
			"rawText": "Backend Engineer needed.",
		},
		TargetLanguage: "en",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(raw))
	req.Header.Set("Idempotency-Key", "fresh-key")
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "018f2a40-0000-7000-9000-0000000000b1"))

	rec := httptest.NewRecorder()
	h.ImportTargetJob(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out api.TargetJobWithJob
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.TargetJobId != "018f2a40-0000-7000-9000-0000000000a1" {
		t.Fatalf("unexpected targetJobId: %s", out.TargetJobId)
	}
	if out.Job.Status != sharedtypes.JobStatusQueued || out.Job.JobType != api.JobTypeTargetImport {
		t.Fatalf("unexpected job: %+v", out.Job)
	}
}

func TestHandler_ImportTargetJob_RejectsMissingIdempotencyKey(t *testing.T) {
	h, _ := newWiredHandler(t, "a", "b", "c", "d")
	body, _ := json.Marshal(api.ImportTargetJobRequest{
		Source:         map[string]any{"type": "manual_text", "rawText": "x"},
		TargetLanguage: "en",
	})
	req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), testUserKey{}, "user-1"))
	rec := httptest.NewRecorder()
	h.ImportTargetJob(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHandler_ImportTargetJob_RejectsMissingSession(t *testing.T) {
	h, _ := newWiredHandler(t, "a", "b", "c", "d")
	body, _ := json.Marshal(api.ImportTargetJobRequest{
		Source:         map[string]any{"type": "manual_text", "rawText": "x"},
		TargetLanguage: "en",
	})
	req := httptest.NewRequest(http.MethodPost, "/targets/import", bytes.NewReader(body))
	req.Header.Set("Idempotency-Key", "k")
	rec := httptest.NewRecorder()
	h.ImportTargetJob(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestHandlerStubReturns501UntilPhase2Lands(t *testing.T) {
	h := targetjob.NewHandler()
	cases := []struct {
		name string
		exec func(w http.ResponseWriter, r *http.Request)
	}{
		{"ImportTargetJob", h.ImportTargetJob},
		{"ListTargetJobs", h.ListTargetJobs},
		{"GetTargetJob", func(w http.ResponseWriter, r *http.Request) { h.GetTargetJob(w, r, "018f2a40-0000-7000-9000-0000000000a1") }},
		{"UpdateTargetJob", func(w http.ResponseWriter, r *http.Request) { h.UpdateTargetJob(w, r, "018f2a40-0000-7000-9000-0000000000a1") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.exec(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("%s: status = %d, want %d", tc.name, rec.Code, http.StatusNotImplemented)
			}
		})
	}
}
