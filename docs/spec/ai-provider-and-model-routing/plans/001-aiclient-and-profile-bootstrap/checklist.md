# AIClient and Profile Bootstrap Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 0: foundation preflight

- [x] 0.1 B1 AI errors/capability vocabulary and B4 task-run schema are available（验证：AIClient tests and B1/B4 generated contracts）
- [x] 0.2 A4 config binding exposes provider registry and model profile paths（验证：`make lint-config`）
- [x] 0.3 current profile coverage resolves required feature/profile inventory（验证：`make lint-ai-profile-coverage`）

## Phase 1: AIClient interface and stub

- [x] 1.1 `AIClient` exposes provider-neutral `Complete`, `Stream`, `Transcribe` and `Synthesize` calls by profile name（验证：`cd backend && go test ./internal/ai/aiclient -count=1`）
- [x] 1.2 `AICallMeta` is filled by the client and carries provider/model/capability/profile/version/language/tokens/cost/latency/fallback/route/validation/error metadata（验证：meta builder tests）
- [x] 1.3 unsupported profile/capability calls fail closed without fallback to chat or stub（验证：AIClient unsupported capability tests）
- [x] 1.4 deterministic stub output is allowed only for test/mock paths（验证：stub tests and config fail-fast matrix）

## Phase 2: registry and profile loaders

- [x] 2.1 Provider Registry loader validates provider refs, protocols, capabilities and secret env refs without checked-in secret values（验证：providerregistry loader tests）
- [x] 2.2 Model Profile loader validates capability/status/provider/model/timeout/rate-limit/route/version fields with file/line errors（验证：profile loader tests）
- [x] 2.3 loader hot reload swaps snapshots without affecting in-flight calls（验证：loader concurrency tests）
- [x] 2.4 current catalogs stay aligned with F3 feature profile coverage（验证：`make lint-ai-profile-coverage`）

## Phase 3: provider adapters

- [x] 3.1 `openai_compatible` adapter uses standard library HTTP/JSON and does not import vendor SDKs（验证：provider terminology lint and adapter tests）
- [x] 3.2 adapter supports chat completions, streaming SSE, Audio Transcriptions and tool-call wire subset（验证：`go test ./internal/ai/aiclient/providers/openai_compatible -count=1`）
- [x] 3.3 provider base URLs normalize root and `/v1` inputs without duplicate path prefixes（验证：adapter contract tests）
- [x] 3.4 provider errors map to B1 `AI_*` errors; unknown upstream codes are sanitized（验证：adapter error envelope tests）

## Phase 4: observability, privacy and fail-fast

- [x] 4.1 observability decorator registers 7 AI metric families and emits structured completion/failure/fallback/validation log events（验证：observability tests）
- [x] 4.2 `ai_task_runs` and `audit_events` are written through DI writers with B4-compatible rows（验证：decorator writer tests）
- [x] 4.3 prompt/response/audio/synthesis text is hashed/counted and never stored as raw metadata（验证：observability privacy tests）
- [x] 4.4 selected non-stub provider secrets fail fast outside test/mock paths（验证：config/bootstrap tests and `make lint-config`）

## Phase 5: closeout and handoff

- [x] 5.1 focused AIClient and config tests pass（验证：`cd backend && go test ./internal/ai/aiclient/... ./internal/platform/config ./cmd/api -count=1`）
- [x] 5.2 terminology/profile/config/codegen gates pass（验证：`make lint-ai-provider-terminology`、`make lint-ai-profile-coverage`、`make lint-config`、`make codegen-check`）
- [x] 5.3 owner context and docs indexes are current（验证：`validate_context.py ai-provider-and-model-routing/001 backend`、`sync-doc-index --check`、`make docs-check`）
- [x] 5.4 DI handoff surfaces remain available for A4/B4/F1/F3/business owners（验证：AIClient README and package tests）

## Phase 6: OpenAI-compatible base URL normalization simplification

- [x] 6.1 `normalizeBaseURL` 删除冗余 suffix guard，同时保持 root 与 `/v1` 输入合同（验证：OpenAI-compatible adapter package tests、scoped `staticcheck`、owner context/docs gates）
  <!-- verified: 2026-07-10 method=openai-base-url-normalization-simplification evidence="S1017 red identified the guarded TrimSuffix path. Focused root and /v1 contract tests, full openai_compatible package tests, full AIClient package tests and scoped staticcheck PASS; owner/product contexts, sync-doc-index, docs-check, diff-check and pruning surface PASS real_residuals=0." -->
