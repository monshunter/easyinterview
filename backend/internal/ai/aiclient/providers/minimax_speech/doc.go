// Package minimax_speech implements the aiclient.Provider contract for
// MiniMax speech services. It supports TTS synthesis via provider-specific
// REST/JSON protocol and does not assume OpenAI-compatible wire shapes.
// MiniMax STT is not confirmed per plan 004 and must not be declared.
//
// API contract fixture source: adapter mockserver captures the
// documented request/response shapes; official doc version must be
// recorded in the plan's operation matrix before production activation.
package minimax_speech
