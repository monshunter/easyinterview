// Package openaicompatible implements aiclient.Provider against any
// OpenAI-compatible HTTP provider endpoint. The P0 protocol surface is
// /v1/chat/completions and
// /v1/audio/transcriptions.
//
// The adapter uses the pinned official openai-go/v3 SDK as a private transport
// implementation detail. SDK types must not cross the provider boundary into
// aiclient public contracts, profiles, business packages, or observability.
package openaicompatible
