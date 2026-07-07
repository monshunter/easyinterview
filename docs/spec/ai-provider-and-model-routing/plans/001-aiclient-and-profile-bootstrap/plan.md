# AIClient and Profile Bootstrap

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 A3 的 AIClient foundation：

- `backend/internal/ai/aiclient/` 暴露 provider-neutral `AIClient` interface、`AICallMeta`、payload types、stream event contract and fail-closed capability behavior。
- Provider Registry 与 Model Profile loaders 读取 `config/ai-providers.yaml` / `config/ai-profiles.yaml`，支持 secret env ref、capability validation、hot reload and active profile resolution。
- deterministic `stub` provider 只在测试或显式 mock 场景启用；非测试 runtime 必须解析真实 provider registry / profile / secret，否则 fail-fast。
- `openai_compatible` adapter 使用标准库 HTTP/JSON 实现 chat、streaming、STT 协议子集，不引入 vendor SDK。
- observability decorator 注册 7 个 AI metric family，写 `ai_task_runs` 与 `audit_events(action='ai.call')`，并保证 prompt / response / audio / synthesized text 不以明文进入 logs、metrics、DB metadata 或 audit metadata。
- runtime config contract 使用 provider registry/profile paths 与 provider-specific env refs；`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 只作为当前 `deepseek` provider ref 的 env source，不是全局业务契约。

## 2 当前合同

### 2.1 Surface Matrix

| surface | current behavior | runtime/config truth | coverage |
|---------|------------------|----------------------|----------|
| `AIClient` interface | `Complete`, `Stream`, `Transcribe`, `Synthesize` use model profile names and return `AICallMeta`; unsupported capability fails closed | `backend/internal/ai/aiclient/aiclient.go`, `meta.go`, `payload.go` | `go test ./internal/ai/aiclient -count=1` |
| Provider Registry | provider refs declare protocol, capabilities, base URL env and API key env; tracked YAML contains no secret values | `config/ai-providers.yaml`, `providerregistry/` | providerregistry loader tests, `make lint-config` |
| Model Profile catalog | profiles declare capability, status, provider ref, model, timeout, rate limits, route and version | `config/ai-profiles.yaml`, `profile/` | profile loader tests, `make lint-ai-profile-coverage` |
| deterministic stub | hash-based output, unit-test only by default, explicit mock opt-in required elsewhere | `providers/stub/`, `WithStubAllowed`, `stub.WithAppEnv` | stub tests, config fail-fast tests |
| OpenAI-compatible adapter | Chat Completions, streaming SSE, Audio Transcriptions and tool-call wire subset through standard library HTTP | `providers/openai_compatible/` | contract tests with `mockserver/` |
| Observability decorator | 7 metric families, structured log events, `ai_task_runs`, `audit_events`, hash/length-only metadata | `observability/`, `writers.go` | observability and privacy tests |
| config fail-fast | selected active provider secrets must resolve outside test/mock paths | `config.go`, `bootstrap/`, A4 SecretSource | config tests, `make lint-config` |
| terminology boundary | active code/config/docs use provider-neutral naming and no provider-proxy vocabulary | `scripts/lint/ai_provider_terminology.py` | `make lint-ai-provider-terminology` |

### 2.2 Ownership Boundary

- A3 owns AIClient, provider registry schema, model profile schema, provider adapters, provider-neutral metadata, observability wrapper and local fail-fast validation.
- F3 owns feature_key -> model_profile_name resolution and prompt/rubric content.
- A4 owns secret source injection and runtime config binding.
- B4 owns `ai_task_runs` and `audit_events` schema.
- Business backend owners call A3 through profile names only and do not import vendor SDKs or provider/model constants.

## 3 质量门禁

- **Plan 类型**: `code-internal + contract + platform-foundation`。
- **TDD 策略**: 适用。Focused tests cover AIClient routing, provider registry/profile loaders, stub, OpenAI-compatible adapter, observability/audit, privacy, config fail-fast and terminology lint.
- **BDD 策略**: 不适用。本 plan 是内部 AI provider client / profile / observability foundation，不产生浏览器 UI、public HTTP API 或用户业务流程；用户可见 AI flows由各业务 owner 的 BDD gate 覆盖。
- **替代验证 gate**:
  - `cd backend && go test ./internal/ai/aiclient/... -count=1`
  - `cd backend && go test ./internal/platform/config ./cmd/api -count=1`
  - `make lint-ai-provider-terminology`
  - `make lint-ai-profile-coverage`
  - `make lint-config`
  - `make codegen-check`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/ai-provider-and-model-routing/plans/001-aiclient-and-profile-bootstrap/context.yaml --target backend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施步骤

### Phase 0: foundation preflight

- Confirm B1 AI errors/capability vocabulary and B4 task-run schema are available.
- Confirm config roots and A4 env binding expose provider registry and model profile paths.
- Confirm active profile coverage resolves the current 9 chat feature keys plus enabled speech/judge profiles.

### Phase 1: AIClient interface and stub

- Define provider-neutral request/response payloads and `AICallMeta`.
- Route by model profile name, not provider/model literals.
- Keep deterministic stub output gated to test/mock paths.
- Keep unsupported profile/capability fail-closed.

### Phase 2: provider registry and model profile loaders

- Load provider refs, capabilities and secret env names from provider registry YAML.
- Load profile catalog, status, capability, provider refs, fallback and timeout/rate-limit values from profile YAML.
- Hot reload snapshots without changing in-flight calls.
- Report file/line validation errors for actionable config repair.

### Phase 3: provider adapters

- Use standard library HTTP/JSON for OpenAI-compatible chat, streaming, STT and tool-call wire subset.
- Normalize provider base URLs without duplicate `/v1`.
- Map provider errors to B1 `AI_*` errors and sanitize unknown upstream codes.
- Keep fallback metadata in `AICallMeta` and bounded metrics.

### Phase 4: observability, privacy and fail-fast

- Register AI metric families and emit structured log events.
- Write `ai_task_runs` and `audit_events` through DI writers.
- Hash and count sensitive payloads; do not persist raw prompt/response/audio/synthesis text.
- Fail fast when a selected non-stub provider cannot resolve required secrets outside test/mock paths.

### Phase 5: closeout and handoff

- Run focused AIClient/config/lint/codegen/docs gates.
- Sync owner plan index.
- Handoff DI surfaces to A4/B4/F1 and profile names to F3/business owners.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Business code can call AI through profile names without vendor SDK imports | AIClient tests, vendor SDK grep in provider terminology lint |
| A-2 | Provider registry/profile loaders validate current catalogs and hot reload snapshots | providerregistry/profile tests, `make lint-ai-profile-coverage` |
| A-3 | Stub is deterministic and test/mock-gated | stub and config tests |
| A-4 | OpenAI-compatible adapter parses tokens, errors, fallback headers, streaming and STT wire without SDKs | adapter contract tests |
| A-5 | Observability writes metrics/log/task-run/audit with hash/length-only sensitive metadata | observability/privacy tests |
| A-6 | Non-test runtime fails closed when selected provider secrets are unavailable | config/bootstrap tests, `make lint-config` |
| A-7 | Active terminology remains provider-neutral and current | `make lint-ai-provider-terminology` |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-07 | 1.6 | Compress owner docs to current AIClient, provider registry, model profile, adapter, observability and fail-fast contract. |
| 2026-05-05 | 1.5 | Complete provider terminology remediation and provider-neutral config/profile naming. |
