package mockserver

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

// Behavior controls per-request response shaping.
type Behavior struct {
	// SleepBeforeRespond simulates upstream latency. The adapter's per-call
	// timeout (profile.timeout_ms) tests timeout behavior by setting this
	// higher than the timeout.
	SleepBeforeRespond time.Duration
	// StatusCode overrides the response status. When 0, defaults to 200.
	StatusCode int
	// ErrorBody, when non-empty, is returned verbatim as the response body
	// alongside StatusCode. Used to test 4xx error parsing.
	ErrorBody string
	// FallbackFrom and FallbackTo populate the X-Fallback-{From,To} response
	// headers, exercising the fallback chain meta path.
	FallbackFrom string
	FallbackTo   string
	// Route populates the X-Route response header.
	Route string
	// MissingChoices forces an empty choices[] array to test the
	// AI_OUTPUT_INVALID validation path.
	MissingChoices bool
}

// CapturedRequest summarizes a single request for assertions.
type CapturedRequest struct {
	Path          string
	Method        string
	Authorization string
	ContentType   string
	RequestID     string
	Body          json.RawMessage
}

// Server wraps httptest.Server with helper accessors.
type Server struct {
	HTTPServer *httptest.Server

	mu               sync.Mutex
	captured         []CapturedRequest
	chatBehavior     Behavior
	embedBehavior    Behavior
	chatBodyOverride func() string
	chatStreamChunks []string
	chatStreamDelay  time.Duration
}

// New starts a server with default Behavior (200 OK responses).
func New() *Server {
	s := &Server{}
	s.HTTPServer = httptest.NewServer(http.HandlerFunc(s.handle))
	return s
}

// URL returns the base URL of the test server.
func (s *Server) URL() string { return s.HTTPServer.URL }

// Close shuts the server down.
func (s *Server) Close() { s.HTTPServer.Close() }

// SetChatBehavior atomically replaces the chat-completions behavior.
func (s *Server) SetChatBehavior(b Behavior) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatBehavior = b
}

// SetEmbedBehavior atomically replaces the embeddings behavior.
func (s *Server) SetEmbedBehavior(b Behavior) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.embedBehavior = b
}

// SetChatBodyOverride installs a function that produces the raw response
// body for chat completions. Used to test malformed payloads.
func (s *Server) SetChatBodyOverride(fn func() string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatBodyOverride = fn
}

// SetChatStreamChunks configures a text/event-stream response for chat
// completion requests that set stream=true. Each chunk is written as one
// `data: ...` SSE frame.
func (s *Server) SetChatStreamChunks(chunks []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatStreamChunks = append([]string(nil), chunks...)
}

// SetChatStreamDelay configures the delay before each stream chunk after the
// first one. This keeps the first delta observable while tests cancel the
// request before the terminal frame arrives.
func (s *Server) SetChatStreamDelay(delay time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.chatStreamDelay = delay
}

// Captured returns a copy of all captured requests in arrival order.
func (s *Server) Captured() []CapturedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]CapturedRequest, len(s.captured))
	copy(out, s.captured)
	return out
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	s.mu.Lock()
	s.captured = append(s.captured, CapturedRequest{
		Path:          r.URL.Path,
		Method:        r.Method,
		Authorization: r.Header.Get("Authorization"),
		ContentType:   r.Header.Get("Content-Type"),
		RequestID:     r.Header.Get("X-Request-ID"),
		Body:          json.RawMessage(body),
	})
	chat := s.chatBehavior
	embed := s.embedBehavior
	bodyOverride := s.chatBodyOverride
	streamChunks := append([]string(nil), s.chatStreamChunks...)
	streamDelay := s.chatStreamDelay
	s.mu.Unlock()

	switch r.URL.Path {
	case "/v1/chat/completions":
		applyHeaders(w, chat)
		if chat.SleepBeforeRespond > 0 {
			time.Sleep(chat.SleepBeforeRespond)
		}
		if chat.StatusCode >= 400 {
			w.WriteHeader(chat.StatusCode)
			io.WriteString(w, chat.ErrorBody)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if len(streamChunks) > 0 && chatRequestWantsStream(body) {
			w.Header().Set("Content-Type", "text/event-stream")
			writeChatStream(w, streamChunks, streamDelay)
			return
		}
		if bodyOverride != nil {
			io.WriteString(w, bodyOverride())
			return
		}
		writeChatResponse(w, body, chat)
	case "/v1/embeddings":
		applyHeaders(w, embed)
		if embed.SleepBeforeRespond > 0 {
			time.Sleep(embed.SleepBeforeRespond)
		}
		if embed.StatusCode >= 400 {
			w.WriteHeader(embed.StatusCode)
			io.WriteString(w, embed.ErrorBody)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		writeEmbedResponse(w, body, embed)
	default:
		http.NotFound(w, r)
	}
}

func applyHeaders(w http.ResponseWriter, b Behavior) {
	if b.FallbackFrom != "" {
		w.Header().Set("X-Fallback-From", b.FallbackFrom)
	}
	if b.FallbackTo != "" {
		w.Header().Set("X-Fallback-To", b.FallbackTo)
	}
	if b.Route != "" {
		w.Header().Set("X-Route", b.Route)
	}
}

type chatRequest struct {
	Model    string `json:"model"`
	Stream   bool   `json:"stream"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

func chatRequestWantsStream(body []byte) bool {
	var req chatRequest
	_ = json.Unmarshal(body, &req)
	return req.Stream
}

func writeChatStream(w http.ResponseWriter, chunks []string, delay time.Duration) {
	flusher, _ := w.(http.Flusher)
	for i, chunk := range chunks {
		if i > 0 && delay > 0 {
			time.Sleep(delay)
		}
		io.WriteString(w, "data: ")
		io.WriteString(w, chunk)
		io.WriteString(w, "\n\n")
		if flusher != nil {
			flusher.Flush()
		}
	}
}

type chatResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func writeChatResponse(w http.ResponseWriter, body []byte, b Behavior) {
	var req chatRequest
	_ = json.Unmarshal(body, &req)
	resp := chatResponse{
		ID:    "mock-id-1",
		Model: req.Model,
	}
	resp.Usage.PromptTokens = inputTokens(req)
	if !b.MissingChoices {
		choice := struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{}
		choice.Message.Role = "assistant"
		choice.Message.Content = "mock response for " + req.Model
		choice.FinishReason = "stop"
		resp.Choices = append(resp.Choices, choice)
		resp.Usage.CompletionTokens = len(choice.Message.Content)
		resp.Usage.TotalTokens = resp.Usage.PromptTokens + resp.Usage.CompletionTokens
	}
	_ = json.NewEncoder(w).Encode(resp)
}

func inputTokens(req chatRequest) int {
	n := 0
	for _, m := range req.Messages {
		n += len(strings.Fields(m.Content))
	}
	return n
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Model string `json:"model"`
	Data  []struct {
		Index     int       `json:"index"`
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

func writeEmbedResponse(w http.ResponseWriter, body []byte, _ Behavior) {
	var req embedRequest
	_ = json.Unmarshal(body, &req)
	resp := embedResponse{Model: req.Model}
	for i, text := range req.Input {
		item := struct {
			Index     int       `json:"index"`
			Embedding []float64 `json:"embedding"`
		}{Index: i}
		item.Embedding = []float64{0.1, 0.2, 0.3, float64(len(text))}
		resp.Data = append(resp.Data, item)
		resp.Usage.PromptTokens += len(text)
	}
	resp.Usage.TotalTokens = resp.Usage.PromptTokens
	_ = json.NewEncoder(w).Encode(resp)
}
