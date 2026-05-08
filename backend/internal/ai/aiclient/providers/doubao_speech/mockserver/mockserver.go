// Package mockserver provides a deterministic mock HTTP server for
// Doubao speech endpoint contract tests.
package mockserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Behavior controls the mock server response for a single endpoint.
type Behavior struct {
	StatusCode   int
	Body         string        // success response body JSON
	ErrorBody    string        // error response body JSON
	SleepMs      int           // simulated delay for timeout tests
	ContentType  string        // override Content-Type header
}

// Server wraps an httptest.Server with configurable endpoint behavior.
type Server struct {
	HTTPServer          *httptest.Server
	ttsBehavior         Behavior
	sttBehavior         Behavior
}

// New starts a new mock server with default 200 OK behavior.
func New() *Server {
	s := &Server{
		ttsBehavior: Behavior{StatusCode: 200},
		sttBehavior: Behavior{StatusCode: 200},
	}
	s.HTTPServer = httptest.NewServer(http.HandlerFunc(s.handle))
	return s
}

// Close shuts down the mock server.
func (s *Server) Close() {
	if s.HTTPServer != nil {
		s.HTTPServer.Close()
	}
}

// URL returns the base URL of the mock server.
func (s *Server) URL() string {
	return s.HTTPServer.URL
}

// SetTTSBehavior configures the /v1/tts/synthesize endpoint behavior.
func (s *Server) SetTTSBehavior(b Behavior) {
	s.ttsBehavior = b
}

// SetSTTBehavior configures the /v1/audio/recognize endpoint behavior.
func (s *Server) SetSTTBehavior(b Behavior) {
	s.sttBehavior = b
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case strings.HasSuffix(path, "/v1/tts/synthesize") || strings.HasPrefix(path, "/v1/tts/synthesize"):
		s.serveBehavior(w, r, s.ttsBehavior)
	case strings.HasSuffix(path, "/v1/audio/recognize") || strings.HasPrefix(path, "/v1/audio/recognize"):
		s.serveBehavior(w, r, s.sttBehavior)
	default:
		w.WriteHeader(404)
	}
}

func (s *Server) serveBehavior(w http.ResponseWriter, r *http.Request, b Behavior) {
	w.Header().Set("Content-Type", "application/json")
	if b.ContentType != "" {
		w.Header().Set("Content-Type", b.ContentType)
	}
	w.WriteHeader(b.StatusCode)
	if b.StatusCode >= 400 && b.ErrorBody != "" {
		w.Write([]byte(b.ErrorBody))
		return
	}
	if b.Body != "" {
		w.Write([]byte(b.Body))
	}
}

// DefaultTTSSuccessBody returns a standard TTS success response JSON.
func DefaultTTSSuccessBody() string {
	resp := map[string]any{
		"audio":        "dGVzdCBhdWRpbyBieXRlcw==", // "test audio bytes" in base64
		"content_type": "audio/mpeg",
		"duration_ms":  1500,
		"char_count":   10,
	}
	b, _ := json.Marshal(resp)
	return string(b)
}

// DefaultSTTSuccessBody returns a standard STT success response JSON.
func DefaultSTTSuccessBody() string {
	resp := map[string]any{
		"text": "这是转写结果",
	}
	b, _ := json.Marshal(resp)
	return string(b)
}
