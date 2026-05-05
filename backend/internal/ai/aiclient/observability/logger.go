package observability

import "sync"

// Log event names. Spec §4.3 + F1 observability-stack logging §4.4 lock these
// four; new event names require a logging spec revision.
const (
	EventTaskCompleted          = "ai.task.completed"
	EventTaskFailed             = "ai.task.failed"
	EventTaskFallback           = "ai.task.fallback"
	EventOutputValidationFailed = "ai.output.validation_failed"
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
	TaskType            string
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
