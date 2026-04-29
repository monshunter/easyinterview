package aiclient

// TaskType enumerates the AI task families recognized by Model Profiles.
//
// B1 shared-conventions-codified has not yet generated a cross-language
// AITaskType enum; until B1 002 lands, A3 owns this private type. The Phase
// 5.4 handoff lists the consumers that must switch to the B1-generated
// constant once it exists.
type TaskType string

const (
	TaskTypeChat  TaskType = "chat"
	TaskTypeEmbed TaskType = "embed"
	// TaskTypeSTT is reserved for C14 P2 backend-voice-stt; loader accepts
	// it but Complete/Embed/Stream return ErrTaskTypeNotImplemented when a
	// profile resolves to stt.
	TaskTypeSTT TaskType = "stt"
)

// ValidationStatus marks whether the response passed client-side
// validateOutput. ValidationStatusOK and ValidationStatusInvalid are the only
// values the decorator currently emits; other layers must not extend this
// alphabet without revising spec §4.1.
type ValidationStatus string

const (
	ValidationStatusOK      ValidationStatus = "ok"
	ValidationStatusInvalid ValidationStatus = "invalid"
)

// AICallMeta is the runtime meta returned by every AIClient call. The field
// order is fixed by spec §4.1 / ADR-Q6 §3.1; new fields require a spec
// version bump and, for cross-language sharing, a B1 update.
//
// Callers cannot construct this struct themselves — the AIClient owns
// metaBuilder which fills, validates, and freezes the value.
type AICallMeta struct {
	Provider            string
	ModelFamily         string
	ModelID             string
	TaskType            TaskType
	PromptVersion       string
	RubricVersion       string
	ModelProfileName    string
	ModelProfileVersion string
	Language            string
	InputTokens         int
	OutputTokens        int
	CostUSDMicros       int64
	LatencyMs           int64
	FallbackChain       []string
	Route               string
	ValidationStatus    ValidationStatus
	ErrorCode           string
}

// StreamEventType identifies the variant of an AIStreamEvent. Plan 001
// freezes exactly three event types; provider-side streaming consumption is
// implemented by plan 002.
type StreamEventType string

const (
	StreamEventDelta StreamEventType = "delta"
	StreamEventError StreamEventType = "error"
	StreamEventDone  StreamEventType = "done"
)

// AIStreamEvent is the union returned over the channel by AIClient.Stream.
// Exactly one of the variant fields is meaningful per Type:
//
//   - Type == "delta": Delta is the incremental content fragment.
//   - Type == "error": ErrorCode is a B1 AI_* code; the channel will close
//     after this event.
//   - Type == "done":  Meta carries the final AICallMeta; the channel closes
//     after this event.
type AIStreamEvent struct {
	Type      StreamEventType
	Delta     string
	ErrorCode string
	Meta      *AICallMeta
}
