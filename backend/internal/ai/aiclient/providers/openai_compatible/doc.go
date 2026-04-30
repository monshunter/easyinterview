// Package openaicompatible implements aiclient.Provider against any
// OpenAI-compatible HTTP gateway (a real provider, Higress, LiteLLM,
// Kong AI, ...). The P0 protocol surface is /v1/chat/completions and
// /v1/embeddings; /v1/audio/transcriptions is reserved for plan 002 / C14.
//
// The adapter MUST stay free of vendor SDKs (openai-go, anthropic-sdk-go,
// cohere-go, generative-ai-go, ...) — it composes the wire format from
// net/http and encoding/json only (spec §6 C-2).
package openaicompatible
