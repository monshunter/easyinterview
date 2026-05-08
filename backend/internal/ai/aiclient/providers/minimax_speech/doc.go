// Package minimax_speech implements the aiclient.Provider contract for
// MiniMax speech services. It supports TTS synthesis via provider-specific
// REST/JSON protocol and does not assume OpenAI-compatible wire shapes.
// MiniMax STT is not confirmed per plan 004 and must not be declared.
//
// Contract source, reviewed 2026-05-08:
//   - MiniMax T2A HTTP docs: https://platform.minimax.io/docs/api-reference/speech-t2a-http
//   - MiniMax API overview: https://platform.minimax.io/docs/api-reference/api-overview
//
// The adapter mockserver is the auditable normalized fixture source for
// plan 004 contract tests.
package minimax_speech
