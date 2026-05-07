// Package aiclient is the provider-neutral entry point for every LLM,
// STT call inside the easyinterview backend.
//
// Business code MUST depend on this package only — never on a vendor SDK such
// as openai-go, anthropic-sdk-go, cohere-go, or generative-ai-go. ADR-Q6 §3.1
// and the AI provider spec §6 (C-2) treat any vendor SDK import inside
// backend/ as a hard violation.
//
// AICallMeta is owned by this package; callers receive it as the second
// return value alongside the structured response and cannot construct or
// mutate it themselves. The meta carries provider, model, profile, prompt,
// rubric, language, token, cost, latency, fallback, route, validation, and
// error_code fields needed by F1 dashboards, B4 ai_task_runs, and the audit
// pipeline.
package aiclient
