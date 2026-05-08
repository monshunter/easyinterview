package mockserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// Behavior controls the mock server response.
type Behavior struct {
	StatusCode  int
	Body        string
	ErrorBody   string
	SleepMs     int
	ContentType string
}

// Server wraps an httptest.Server with configurable endpoint behavior.
type Server struct {
	HTTPServer  *httptest.Server
	ttsBehavior Behavior
}

// New starts a new mock server with default 200 OK behavior.
func New() *Server {
	s := &Server{
		ttsBehavior: Behavior{StatusCode: 200},
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

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/v1/tts/synthesize") || strings.HasPrefix(path, "/v1/tts/synthesize") {
		s.serveBehavior(w, s.ttsBehavior)
		return
	}
	w.WriteHeader(404)
}

func (s *Server) serveBehavior(w http.ResponseWriter, b Behavior) {
	if b.SleepMs > 0 {
		time.Sleep(time.Duration(b.SleepMs) * time.Millisecond)
	}
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
		"audio":        "bWluaW1heCBhdWRpbw==", // "minimax audio" in base64
		"content_type": "audio/mpeg",
		"duration_ms":  1200,
		"char_count":   8,
	}
	b, _ := json.Marshal(resp)
	return string(b)
}
