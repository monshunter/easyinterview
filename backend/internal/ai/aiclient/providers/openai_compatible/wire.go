package openaicompatible

import "encoding/json"

type wireMessage struct {
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	ToolCalls []wireToolCall `json:"tool_calls,omitempty"`
}

type chatCompletionsRequest struct {
	Model       string        `json:"model"`
	Messages    []wireMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
	Tools       []wireTool    `json:"tools,omitempty"`
	ToolChoice  any           `json:"tool_choice,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
	TopP        *float64      `json:"top_p,omitempty"`
}

type chatCompletionsResponse struct {
	ID      string                  `json:"id"`
	Model   string                  `json:"model"`
	Choices []chatCompletionsChoice `json:"choices"`
	Usage   wireUsage               `json:"usage"`
}

type chatCompletionsChoice struct {
	Index        int         `json:"index"`
	Message      wireMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

type wireTool struct {
	Type     string       `json:"type"`
	Function wireFunction `json:"function"`
}

type wireFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
	Arguments   string          `json:"arguments,omitempty"`
}

type wireToolCall struct {
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type"`
	Function wireFunction `json:"function"`
}

type transcriptionResponse struct {
	Text string `json:"text"`
}

type wireUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

type streamChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage wireUsage `json:"usage"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
