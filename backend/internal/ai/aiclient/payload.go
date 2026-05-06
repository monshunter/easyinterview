package aiclient

import "encoding/json"

// Message is one chat turn in a CompletePayload. Role is one of the
// OpenAI-compatible roles ("system" / "user" / "assistant" / "tool"); the
// AIClient does not prescribe which roles a profile may use.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// CallMetadata is the per-call business context required by AIClient. It is
// supplied by the caller and copied into AICallMeta (PromptVersion,
// RubricVersion, Language) plus AI logs / audit metadata.
type CallMetadata struct {
	FeatureKey    string           `json:"featureKey"`
	PromptVersion string           `json:"promptVersion"`
	RubricVersion string           `json:"rubricVersion"`
	Language      string           `json:"language"`
	OutputSchema  json.RawMessage  `json:"outputSchema,omitempty"`
	TaskRun       AITaskRunContext `json:"taskRun,omitempty"`
}

// Tool is the provider-neutral function/tool schema subset accepted by
// AIClient.Complete. Parameters must be a JSON schema object; callers never
// pass provider/model strings through this contract.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// ToolChoiceMode describes how a Complete call selects a tool.
type ToolChoiceMode string

const (
	ToolChoiceModeAuto ToolChoiceMode = "auto"
	ToolChoiceModeNone ToolChoiceMode = "none"
	ToolChoiceModeTool ToolChoiceMode = "tool"
)

// ToolChoice optionally pins tool selection for a Complete call. Name is only
// meaningful when Mode == ToolChoiceModeTool.
type ToolChoice struct {
	Mode ToolChoiceMode `json:"mode"`
	Name string         `json:"name,omitempty"`
}

// CompletePayload is the input to AIClient.Complete and AIClient.Stream.
// Callers cannot pass a bare prompt string; Messages must be non-empty or the
// client returns AI_OUTPUT_INVALID.
type CompletePayload struct {
	Messages   []Message    `json:"messages"`
	Metadata   CallMetadata `json:"metadata"`
	Tools      []Tool       `json:"tools,omitempty"`
	ToolChoice *ToolChoice  `json:"toolChoice,omitempty"`
}

// CompleteResponse is the structured response returned by Complete. Content
// is the assistant message body; FinishReason mirrors the upstream finish
// reason (e.g. "stop" / "length" / "tool_calls") when available.
type CompleteResponse struct {
	Content      string     `json:"content"`
	FinishReason string     `json:"finishReason,omitempty"`
	ToolCalls    []ToolCall `json:"toolCalls,omitempty"`
}

// ToolCall is the provider-neutral tool call result surfaced by Complete.
// Arguments remains JSON so business owners can unmarshal into their own
// request-specific schema after checking the tool name.
type ToolCall struct {
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// EmbedInput is the input to AIClient.Embed.
type EmbedInput struct {
	Texts    []string     `json:"texts"`
	Metadata CallMetadata `json:"metadata"`
}

// EmbedResponse holds the embedding vectors in the same order as
// EmbedInput.Texts.
type EmbedResponse struct {
	Vectors [][]float64 `json:"vectors"`
}
