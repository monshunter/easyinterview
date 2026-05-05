package aiclient

import sharedai "github.com/monshunter/easyinterview/backend/internal/shared/ai"

// Capability enumerates the AI capability families recognized by Provider
// Registry entries and Model Profiles. B1 owns the literal set; A3 aliases it
// so runtime code cannot drift from the cross-language vocabulary.
type Capability = sharedai.Capability

const (
	CapabilityChat     = sharedai.CapabilityChat
	CapabilityEmbed    = sharedai.CapabilityEmbed
	CapabilitySTT      = sharedai.CapabilityStt
	CapabilityRealtime = sharedai.CapabilityRealtime
	CapabilityRerank   = sharedai.CapabilityRerank
	CapabilityJudge    = sharedai.CapabilityJudge
)

// ProviderProtocol identifies the protocol adapter a Provider Registry entry
// uses.
type ProviderProtocol string

const (
	ProviderProtocolStub             ProviderProtocol = "stub"
	ProviderProtocolOpenAICompatible ProviderProtocol = "openai_compatible"
	ProviderProtocolRealtimeAudio    ProviderProtocol = "realtime_audio"
	ProviderProtocolRerankCompatible ProviderProtocol = "rerank_compatible"
	ProviderProtocolJudgeCompatible  ProviderProtocol = "judge_compatible"
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
	Capability          Capability
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
