package targetjob_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	api "github.com/monshunter/easyinterview/backend/internal/api/generated"
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
