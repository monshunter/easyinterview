// Package mockserver provides a deterministic OpenAI-compatible mock HTTP
// server used by openai_compatible contract tests and reused by E1
// mock-contract-suite. The handler covers chat completions,
// timeout simulation, 5xx errors, and fallback header injection so plan
// 001 can validate every meta-mapping path without standing up a real
// provider endpoint.
//
// Plan 001 freezes the helper interface — E1 may compose new fixtures on
// top, but field-level breaking changes here require a spec or plan revision.
package mockserver
