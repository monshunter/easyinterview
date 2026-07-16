package observability

import (
	"log/slog"
	"sync"
)

// Log event names. Spec §4.3 + F1 observability-stack logging §4.4 lock the
// four task events; raw-capture adds one content-free local diagnostic event.
const (
	EventTaskCompleted          = "ai.task.completed"
	EventTaskFailed             = "ai.task.failed"
	EventTaskFallback           = "ai.task.fallback"
	EventOutputValidationFailed = "ai.output.validation_failed"
	EventRawCaptureWriteFailed  = "ai.raw.capture.write_failed"
)

// LogFields is the closed allowlist of structured log fields the
// decorator emits. Adding a key requires a logging-spec change. Tests
// inspect a captured copy to assert the privacy red line.
type LogFields struct {
	Provider            string
	ModelID             string
	ModelProfileName    string
	ModelProfileVersion string
	PromptVersion       string
	RubricVersion       string
	Capability          string
	Language            string
	InputTokens         int
	OutputTokens        int
	CostUSDMicros       int64
	LatencyMs           int64
	FallbackChain       []string
	Route               string
	ValidationStatus    string
	ErrorCode           string
}

// Logger is the surface the decorator uses. F1's structured logger will
// satisfy it; tests use NewMemoryLogger.
type Logger interface {
	Log(event string, fields LogFields)
}

// SlogLogger emits the closed observability event/field contract through the
// process structured logger. Raw capture failures deliberately carry no
// underlying error text or path; the stable warning event is the diagnostic.
type SlogLogger struct {
	logger *slog.Logger
}

// NewSlogLogger adapts the process logger to the AI observability contract.
func NewSlogLogger(logger *slog.Logger) *SlogLogger {
	if logger == nil {
		logger = slog.Default()
	}
	return &SlogLogger{logger: logger}
}

// Log implements Logger.
func (l *SlogLogger) Log(event string, fields LogFields) {
	if l == nil || l.logger == nil {
		return
	}
	args := slogArgs(fields)
	switch event {
	case EventTaskFailed, EventOutputValidationFailed, EventRawCaptureWriteFailed:
		l.logger.Warn(event, args...)
	default:
		l.logger.Info(event, args...)
	}
}

func slogArgs(fields LogFields) []any {
	args := make([]any, 0, 36)
	appendString := func(key, value string) {
		if value != "" {
			args = append(args, key, value)
		}
	}
	appendInt := func(key string, value int) {
		if value != 0 {
			args = append(args, key, value)
		}
	}
	appendInt64 := func(key string, value int64) {
		if value != 0 {
			args = append(args, key, value)
		}
	}

	appendString("provider", fields.Provider)
	appendString("modelId", fields.ModelID)
	appendString("modelProfileName", fields.ModelProfileName)
	appendString("modelProfileVersion", fields.ModelProfileVersion)
	appendString("promptVersion", fields.PromptVersion)
	appendString("rubricVersion", fields.RubricVersion)
	appendString("capability", fields.Capability)
	appendString("language", fields.Language)
	appendInt("inputTokens", fields.InputTokens)
	appendInt("outputTokens", fields.OutputTokens)
	appendInt64("costUsdMicros", fields.CostUSDMicros)
	appendInt64("latencyMs", fields.LatencyMs)
	if len(fields.FallbackChain) > 0 {
		args = append(args, "fallbackChain", fields.FallbackChain)
	}
	appendString("route", fields.Route)
	appendString("validationStatus", fields.ValidationStatus)
	appendString("errorCode", fields.ErrorCode)
	return args
}

// LogEntry is one captured log entry.
type LogEntry struct {
	Event  string
	Fields LogFields
}

// MemoryLogger captures log events in order. It is concurrent-safe.
type MemoryLogger struct {
	mu      sync.Mutex
	entries []LogEntry
}

// NewMemoryLogger constructs an empty in-memory logger.
func NewMemoryLogger() *MemoryLogger { return &MemoryLogger{} }

// Log implements Logger.
func (m *MemoryLogger) Log(event string, fields LogFields) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, LogEntry{Event: event, Fields: fields})
}

// Entries returns a copy of recorded entries.
func (m *MemoryLogger) Entries() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]LogEntry, len(m.entries))
	copy(out, m.entries)
	return out
}

// Reset drops captured entries; useful between tests sharing a logger.
func (m *MemoryLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = nil
}
