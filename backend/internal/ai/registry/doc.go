// Package registry is the F3 prompt-rubric-registry runtime owner.
//
// It loads the on-disk truth source under config/prompts/ and config/rubrics/
// (see config/prompts/README.md for the canonical hash algorithm shared with
// scripts/lint/prompt_lint.py), resolves prompt and rubric coordinates for
// business callers, and exposes the LLM Judge interface for future eval
// implementations.
//
// Boundary red lines (spec §4.2 + plan §2.1 + plan §2.9):
//
//   - This package MUST NOT import backend/internal/targetjob or any C-domain
//     business package. The targetjob domain receives the registry through
//     RegistryAdapter, which translates registry.PromptResolution into the
//     local targetjob.PromptResolution shape.
//   - This package MUST NOT call into the AI client package (avoiding a
//     circular dependency). Business callers resolve registry coordinates
//     first, then issue the AI completion call themselves.
//   - This package MUST NOT emit business metric counters or business
//     audit logs. The loader may log its own startup/load state at warn or
//     info level, and the resolver may bump warn-level fallback counters,
//     but it does not write to the audit_events pipeline owned by the AI
//     client observability decorator.
//   - This package MUST NOT read the process environment, hold secrets, or
//     speak to any network. The loader is filesystem-only; secrets and DB
//     connections are A4-owned and injected by cmd/api wiring.
//
// Package files:
//
//   - types.go    — PromptResolution, PromptMeta, RubricSchema, Judge interface
//   - loader.go   — filesystem walk, hash verification, snapshot building
//   - resolver.go — ResolveActive / GetPrompt / GetRubric with language fallback
//   - cache.go    — atomic.Value snapshot + 30s TTL + Reload hook
//   - judge.go    — NotImplementedJudge default implementation
//   - registry.go — NewRegistryClient constructor and Client struct
package registry
