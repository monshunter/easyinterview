// Package openaicompatible implements aiclient.Provider against any
// OpenAI-compatible HTTP provider endpoint. The P0 protocol surface is
// /v1/chat/completions and
// /v1/audio/transcriptions.
//
// The adapter MUST stay free of vendor SDKs (openai-go, anthropic-sdk-go,
// cohere-go, generative-ai-go, ...) — it composes the wire format from
// net/http and encoding/json only (spec §6 C-2).
package openaicompatible
