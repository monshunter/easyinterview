// Package doubao_speech implements the aiclient.Provider contract for
// Doubao (豆包) speech services. It supports TTS synthesis and STT
// transcription via provider-specific REST/JSON protocols and does not
// assume OpenAI-compatible wire shapes.
//
// API contract fixture source: adapter mockserver captures the
// documented request/response shapes; official doc version must be
// recorded in the plan's operation matrix before production activation.
package doubao_speech
