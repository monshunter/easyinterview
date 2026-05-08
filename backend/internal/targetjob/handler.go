package targetjob

import (
	"net/http"
)

// Handler binds the four B2-defined TargetJob OpenAPI operations into the
// targetjob domain. Phase 1 of plan 001 only locks the contract surface so
// `cmd/api/main.go` can register handlers via the generated
// `api.HandlerFromMux` without us drifting from B2 wire shape; Phase 2 fills
// in the real service / store wiring.
//
// Compile-time guarantee: *Handler satisfies targetJobServerSurface, and
// targetJobServerSurface mirrors the B2 generated ServerInterface for the
// four TargetJob operations (verified by handler_test.go via reflection).
type Handler struct {
	// dependencies wired in Phase 2 (store, service, dispatcher, clock, ...).
}

// NewHandler returns a stub Handler that responds 501 to every operation.
// Phase 2 of the plan will replace this with the real constructor.
func NewHandler() *Handler { return &Handler{} }

// ImportTargetJob is the POST /targets/import binding.
func (h *Handler) ImportTargetJob(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w, "importTargetJob")
}

// ListTargetJobs is the GET /targets binding.
func (h *Handler) ListTargetJobs(w http.ResponseWriter, _ *http.Request) {
	notImplemented(w, "listTargetJobs")
}

// GetTargetJob is the GET /targets/{targetJobId} binding.
func (h *Handler) GetTargetJob(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "getTargetJob")
}

// UpdateTargetJob is the PATCH /targets/{targetJobId} binding.
func (h *Handler) UpdateTargetJob(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "updateTargetJob")
}

// targetJobServerSurface mirrors the B2 generated ServerInterface for the
// four TargetJob operations exactly. handler_test.go pins this against the
// real ServerInterface via reflection so any B2 wire-shape change surfaces
// here as a compile error or test failure.
type targetJobServerSurface interface {
	ImportTargetJob(w http.ResponseWriter, r *http.Request)
	ListTargetJobs(w http.ResponseWriter, r *http.Request)
	GetTargetJob(w http.ResponseWriter, r *http.Request, targetJobId string)
	UpdateTargetJob(w http.ResponseWriter, r *http.Request, targetJobId string)
}

var _ targetJobServerSurface = (*Handler)(nil)

func notImplemented(w http.ResponseWriter, op string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"errors":[{"code":"NOT_IMPLEMENTED","message":"` + op + ` is not yet implemented"}]}`))
}
