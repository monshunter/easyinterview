// Package aiclient is the provider-neutral entry point for every LLM,
// STT call inside the easyinterview backend.
//
// Business code MUST depend on this package only and never on a vendor SDK.
// The official openai-go/v3 dependency is private to the OpenAI-compatible
// provider adapters and their exact provider-internal helper; it must not leak
// into this package's public types, business packages, config, or observability.
// Other vendor SDKs remain forbidden by ADR-Q6 and the A3 spec.
//
// AICallMeta is owned by this package; callers receive it as the second
// return value alongside the structured response and cannot construct or
// mutate it themselves. The meta carries provider, model, profile, prompt,
// rubric, language, token, cost, latency, fallback, route, validation, and
// error_code fields needed by F1 dashboards, B4 ai_task_runs, and the audit
// pipeline.
package aiclient
