// Package doubao_speech implements the aiclient.Provider contract for
// Doubao (豆包) speech services. It supports TTS synthesis and STT
// transcription via provider-specific REST/JSON protocols and does not
// assume OpenAI-compatible wire shapes.
//
// Contract source, reviewed 2026-05-08:
//   - Volcengine TTS interface docs: https://www.volcengine.com/docs/6489/81406
//   - Volcengine Doubao ASR overview: https://www.volcengine.com/docs/6561/109880
//   - Volcengine ASR API FAQ: https://www.volcengine.com/docs/6561/111586
//
// The adapter mockserver is the auditable normalized fixture source for
// plan 004 contract tests.
package doubao_speech
