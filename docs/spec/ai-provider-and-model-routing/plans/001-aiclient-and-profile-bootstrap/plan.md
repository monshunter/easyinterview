# AIClient and Profile Bootstrap

> **版本**: 2.9
> **状态**: completed
> **更新日期**: 2026-07-16

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

本 plan 承接 A3 的 AIClient foundation：

- `backend/internal/ai/aiclient/` 暴露 provider-neutral `AIClient` interface、`AICallMeta`、payload types、stream event contract and fail-closed capability behavior。
- Provider Registry 与 Model Profile loaders 读取 `config/ai-providers.yaml` / `config/ai-profiles.yaml`，支持 secret env ref、capability validation、hot reload and active profile resolution。
- deterministic `stub` provider 只在测试或显式 mock 场景启用；非测试 runtime 必须解析真实 provider registry / profile / secret，否则 fail-fast。
- `openai_compatible` / `judge_compatible` adapter 使用官方 `openai-go/v3 v3.43.0` 实现 OpenAI-compatible chat、streaming、STT 协议子集；SDK 只存在于 provider adapter/internal helper。
- observability decorator 注册 7 个 AI metric family，写 `ai_task_runs` 与 `audit_events(action='ai.call')`，并保证 prompt / response / audio / synthesized text 不以明文进入 logs、metrics、DB metadata 或 audit metadata。
- runtime config contract 使用 provider registry/profile paths 与 provider-specific env refs；`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 只作为当前 `deepseek` provider ref 的 env source，不是全局业务契约。

## 2 当前合同

2026-07-16 原地重开时，`AIClient` / registry / profile / fallback / meta / observability / privacy 合同已存在，`openai_compatible` 与 `judge_compatible` 仍使用标准库手写 wire。下表中的 SDK transport 是 Phase 15 目标状态；只有对应 checklist item 与验证 gate 完成后才能视为当前实现事实。

### 2.1 Surface Matrix

| surface | current behavior | runtime/config truth | coverage |
|---------|------------------|----------------------|----------|
| `AIClient` interface | `Complete`, `Stream`, `Transcribe`, `Synthesize` use model profile names and return `AICallMeta`; unsupported capability fails closed | `backend/internal/ai/aiclient/aiclient.go`, `meta.go`, `payload.go` | `go test ./internal/ai/aiclient -count=1` |
| Provider Registry | provider refs declare protocol, capabilities, base URL env and API key env; tracked YAML contains no secret values | `config/ai-providers.yaml`, `providerregistry/` | providerregistry loader tests, `make lint-config` |
| Model Profile catalog | profiles declare capability, status, provider ref, model, timeout, rate limits, route and version | `config/ai-profiles.yaml`, `profile/` | profile loader tests, `make lint-ai-profile-coverage` |
| deterministic stub | hash-based output, unit-test only by default, explicit mock opt-in required elsewhere | `providers/stub/`, `WithStubAllowed`, `stub.WithAppEnv` | stub tests, config fail-fast tests |
| OpenAI-compatible adapter | Chat Completions, streaming SSE, Audio Transcriptions and tool-call wire subset through adapter-private `openai-go/v3`; DeepSeek thinking uses controlled extra JSON | `providers/openai_compatible/`, `providers/judge_compatible/` | contract tests with `mockserver/` and SDK import boundary lint |
| Observability decorator | 7 metric families, structured log events, `ai_task_runs`, `audit_events`, hash/length-only metadata | `observability/`, `writers.go` | observability and privacy tests |
| config fail-fast | selected active provider secrets must resolve outside test/mock paths | `config.go`, `bootstrap/`, A4 SecretSource | config tests, `make lint-config` |
| terminology boundary | active code/config/docs use provider-neutral naming and no provider-proxy vocabulary | `scripts/lint/ai_provider_terminology.py` | `make lint-ai-provider-terminology` |

### 2.2 Ownership Boundary

- A3 owns AIClient, provider registry schema, model profile schema, provider adapters, provider-neutral metadata, observability wrapper and local fail-fast validation.
- F3 owns feature_key -> model_profile_name resolution and prompt/rubric content.
- A4 owns secret source injection and runtime config binding.
- B4 owns `ai_task_runs` and `audit_events` schema.
- Business backend owners call A3 through profile names only and do not import vendor SDKs or provider/model constants；`openai-go/v3` is allowed only in A3 provider adapter/internal helper implementation.

## 3 质量门禁

- **Plan 类型**: `code-internal + contract + platform-foundation`。
- **TDD 策略**: 适用。Focused tests cover AIClient routing, provider registry/profile loaders, stub, OpenAI-compatible adapter, observability/audit, privacy, config fail-fast and terminology lint.
- **BDD 策略**: `BDD-N/A`。本 plan 是内部 AI provider client / profile / observability foundation，不产生浏览器 UI、public HTTP API 或用户业务流程；用户可见 AI flows 由各业务 owner 的 BDD gate 覆盖。本 plan 不创建 `bdd-plan.md` / `bdd-checklist.md`，也不新增或复用 E2E 场景。
- **Operation Matrix**: `N/A`。本 plan 不新增或修改 OpenAPI operation、fixture、frontend consumer、backend handler、persistence 或用户流程；AI provider transport 只在 A3 adapter 内替换，替代证据由下列 contract/lint/root gates 承接。
- **替代验证 gate**:
  - `cd backend && go test ./internal/ai/aiclient/... -count=1`
  - `cd backend && go test ./internal/platform/config ./cmd/api -count=1`
  - `make lint-ai-provider-terminology`
  - `make lint-ai-profile-coverage`
  - `make lint-config`
  - `python3 scripts/lint/ai_provider_terminology_test.py`：SDK import boundary 只允许 `providers/openai_compatible`、`providers/judge_compatible` 与精确的 `providers/internal/openaisdk` helper
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

- Use adapter-private `openai-go/v3` for OpenAI-compatible chat, streaming, STT and tool-call wire subset.
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

### Phase 6: OpenAI-compatible base URL normalization simplification

- Preserve root and `/v1` provider base URL behavior through existing adapter contract tests.
- Remove the redundant suffix guard from `normalizeBaseURL` and require scoped `staticcheck` to stay clean.
- Run the OpenAI-compatible adapter package tests and owner documentation gates before restoring completed state.

### Phase 7: AIClient duplicate writer state removal

- Use whole-program reachability analysis with test executables enabled to identify the unused core-client writer options.
- Delete `Client` / `clientOptions` task-run and audit writer fields, their core `With*Writer` options and zero-consumer getters; the `observability` decorator remains the only persistence injection owner.
- Run the full AIClient package tests, `staticcheck`, reachability rescan and owner documentation gates before restoring completed state.

### Phase 8: Stub provider name wrapper removal

- Delete the zero-consumer `stub.ProviderName` wrapper and keep `stub.Name` as the single typed provider identifier.
- Run the stub and full AIClient package tests, `staticcheck`, reachability rescan and owner documentation gates before restoring completed state.

### Phase 9: Completion dispatch duplication removal

- Keep `Complete` and `CompleteJudge` as the public chat/judge capability boundaries.
- Move their identical validation, dispatch, fallback execution and metadata merge path into one private capability-parameterized helper.
- Use scoped `dupl` RED/GREEN plus existing chat/judge/fallback/validation tests, `staticcheck`, owner contexts and docs/diff/pruning gates; do not add a generic public API.

### Phase 10: Observability latency fallback consolidation

- Keep `Complete`, `Transcribe` and `Synthesize` capability-specific call and record paths unchanged.
- Move their repeated "preserve provider latency, otherwise use measured duration" metadata fallback into one private helper.
- Use scoped `dupl` RED/GREEN plus existing Complete/STT/TTS observability and privacy tests, full AIClient/backend tests, `staticcheck`, owner contexts and docs/diff/pruning gates.

### Phase 11: Observability invalid-schema test harness consolidation

- Preserve the exact top-level test names used by BUG-0095 and prompt-rubric focused gates.
- Move repeated decorator construction and `AI_OUTPUT_INVALID` / validation-status / metric assertions into one test-only helper; keep required-field, trailing-token and enum-mismatch inputs explicit at their named tests.
- Use scoped `dupl` reduction, exact-name focused tests, full observability/AIClient/backend tests, `staticcheck`, owner contexts and docs/diff/pruning gates.

### Phase 12: Observability fallback-label test table consolidation

- Replace the two unreferenced top-level fallback label tests with one table-driven test and two named cases.
- Keep each complete `AICallMeta` input and exact 11-label metric tuple; do not weaken provider/model-family/date-suffix coverage.
- Use scoped `dupl` RED/GREEN, focused fallback tests, full observability/AIClient/backend tests, `staticcheck`, owner contexts and docs/diff/pruning gates.

### Phase 13: AIClient invalid-input assertion consolidation

- Preserve Complete, Transcribe and Synthesize top-level test names and their capability-specific provider-not-called assertions.
- Move the repeated `AI_OUTPUT_INVALID`, `ErrorCode` and invalid validation-status assertions into one test-only helper.
- Use scoped `dupl` RED/GREEN, exact focused tests, full AIClient/backend tests, `staticcheck`, owner contexts and docs/diff/pruning gates.

### Phase 14: Observability privacy leak assertion consolidation

- Preserve the Complete and TTS privacy test names plus their capability/metric/audit sanity checks.
- Move the repeated six-counter, log, task-run and audit metadata plaintext scans into one test-only helper parameterized by planted tokens.
- Use scoped `dupl` RED/GREEN, exact privacy tests, full observability/AIClient/backend tests, `staticcheck`, owner contexts and docs/diff/pruning gates.

### Phase 15: Official openai-go transport migration

#### 15.1 SDK dependency and import boundary

先在 `scripts/lint/ai_provider_terminology_test.py` 为 SDK import boundary 增加 RED：`github.com/openai/openai-go/v3` 只能由 `providers/openai_compatible`、`providers/judge_compatible` 或精确的 `providers/internal/openaisdk` helper import，业务 package、public A3 type、profile/config、observability 和 speech adapters 必须拒绝。随后固定 `openai-go/v3 v3.43.0`，更新 A3 README/package docs，并保持 `AIClient` public surface 与 config schema 不变。

#### 15.2 Complete and judge Chat Completions parity

先扩充现有 mockserver/adapter contract tests，覆盖 custom base URL、注入 HTTP client、authorization、messages、tools/tool choice、output schema、DeepSeek `thinking` extra JSON、usage/finish/tool-call mapping、request/fallback headers、缺 choices、empty/reasoning-only content、4xx/5xx/timeout、同 provider 最多 2 次 retry 与 response body cap。再把 `openai_compatible.Complete` 和 `judge_compatible.Complete` 切到 SDK Chat Completions；通过包裹注入 HTTP transport 的 bounded response reader 保留非 streaming body cap，并保留本地 B1 error mapping、`AICallMeta`、schema validator、privacy 与 AIClient 跨 provider fallback。Focused gate 为 `cd backend && go test ./internal/ai/aiclient/providers/openai_compatible ./internal/ai/aiclient/providers/judge_compatible ./internal/ai/aiclient -count=1`。

#### 15.3 Streaming and Audio Transcriptions parity

先用现有 provider contract tests 锁定 SSE delta/error/done 顺序、context cancel/partial meta、malformed/oversized event、multipart filename/content-type/language/prompt 与 STT response cap。再用 SDK streaming 与 Audio Transcriptions API 替换手写 SSE/multipart transport，并映射回冻结的 `AIStreamEvent` / `TranscriptionResponse`；streaming 单 event 上限与非 streaming 总 body cap 分开处理，不打开 realtime、TTS 或用户可见 HTTP streaming。Focused gate 为 `cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1`。

#### 15.4 Manual wire cleanup and verification

删除被 SDK 覆盖的 request/response/SSE/multipart struct 与手写 post helper，只保留 provider-neutral mapping、DeepSeek controlled extras、project error/meta/privacy/response-cap glue。负向搜索确认业务代码无 SDK import、旧通用 wire/helper 无运行时消费者；运行 focused adapter/AIClient/config/lint、`go vet`、`staticcheck`、根 `make test`、`make build`、context/docs/index/diff gates 后恢复 completed。

### Phase 16: Secure local raw Complete I/O capture

#### 16.1 Recorder RED/GREEN

Add tests for versioned request/response two-phase NDJSON with UUIDv7 call ID, `callId == ai_task_runs.id`, schema-invalid raw response, provider-error request preservation, append reopen/cross-restart ID non-collision, 100 concurrent calls without line interleaving, stable ConfigDir-parent path resolution, per-component symlink/non-regular rejection, existing 0644→0600 tightening, dedicated parent/file `0700/0600`, and explicit absence of header/credential/audio/reasoning/internal-identity fields. Privacy assertions parse the NDJSON schema and plant tokens only in non-allowlisted inputs；they must not reject arbitrary words inside the intentionally preserved message/content/tool-argument values. Implement one mutex-protected process-shared recorder that marshals a complete line before one append write.

#### 16.2 Decorator/runtime wiring

Replace `WithRawOutputDebugWriter` and stderr markers with `WithRawIOCapture`. Capture only `Complete` at the provider-neutral boundary: allocate/fill the TaskRun UUIDv7 before the provider call, write `ai.complete.raw.v1` request, then write response after schema validation/meta enrichment. API startup opens one recorder and passes it to report/practice/resume/targetjob/resume-tailor decorators；resume-tailor must stop hand-writing a duplicate task run. Open/close occurs once. Startup open failure is fatal；runtime write failure emits a process-visible structured WARN using only the stable event and closed safe fields, excludes underlying path/error/raw content, and never changes the AI business result.

#### 16.3 Privacy/config handoff

Consume A4 `ai.debugCaptureRawIO` / `ai.debugRawIOPath`. Dev/test default enabled；staging/prod force disabled and reject an env override. Delete current-scope `AI_DEBUG_PRINT_RAW_OUTPUT` without alias. Raw NDJSON remains under ignored `.test-output`, never enters logs/DB/audit/runtime-config/E2E evidence, and excludes Transcribe/Synthesize/Stream content. Full-container binds the dedicated host raw directory read-write into backend only, so container recreation does not erase evidence.

#### 16.4 Substitute gates

BDD is not applicable because this is an internal developer diagnostic. Run focused observability/config/runtime tests, privacy negative searches, lint-config, P0.099 evidence-isolation static tests, root `make test` and build/context/docs gates.

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-1 | Business code can call AI through profile names without vendor SDK imports | AIClient tests, vendor SDK grep in provider terminology lint |
| A-2 | Provider registry/profile loaders validate current catalogs and hot reload snapshots | providerregistry/profile tests, `make lint-ai-profile-coverage` |
| A-3 | Stub is deterministic and test/mock-gated | stub and config tests |
| A-4 | OpenAI-compatible adapters map chat, judge, streaming and STT through the adapter-local official SDK without leaking SDK types/imports into business code | adapter contract tests, SDK import-boundary lint |
| A-5 | Observability writes metrics/log/task-run/audit with hash/length-only sensitive metadata | observability/privacy tests |
| A-6 | Non-test runtime fails closed when selected provider secrets are unavailable | config/bootstrap tests, `make lint-config` |
| A-7 | Active terminology remains provider-neutral and current | `make lint-ai-provider-terminology` |
| A-8 | Base URL normalization has no redundant conditional path and preserves root plus `/v1` inputs | adapter contract tests, scoped `staticcheck` |
| A-9 | Task-run and audit writers have one injection path through the observability decorator | `deadcode -test`, symbol inventory, AIClient tests and `staticcheck` |
| A-10 | Stub provider identity has one exported source through `stub.Name` | `deadcode -test`, symbol inventory, stub/AIClient tests and `staticcheck` |
| A-11 | Chat and judge completion share one internal execution path while preserving distinct capability dispatch | scoped `dupl`, AIClient and judge dispatch tests, `staticcheck` |
| A-12 | Complete, STT and TTS use one latency fallback rule while retaining capability-specific recording | scoped `dupl`, observability and privacy tests, `staticcheck` |
| A-13 | Invalid output-schema tests share one harness while retaining exact focused gate names and distinct inputs | scoped `dupl`, exact-name observability tests |
| A-14 | Fallback label derivation cases share one table harness while retaining exact metric tuples | scoped `dupl`, focused fallback observability tests |
| A-15 | Complete, STT and TTS invalid-input tests share one error/meta assertion without weakening provider call guards | scoped `dupl`, exact focused AIClient tests |
| A-16 | Complete and TTS privacy tests share one plaintext scan across metrics/log/task-run/audit surfaces | scoped `dupl`, exact privacy tests |
| A-17 | `openai-go/v3 v3.43.0` is pinned and imported only by A3 OpenAI-compatible adapter/internal helper | SDK import boundary test, `go list -m`, repo-wide import search |
| A-18 | Chat and judge Complete preserve tools/JSON/thinking/usage/header/error/meta/privacy contracts while SDK owns same-provider retry and AIClient owns cross-provider fallback | openai/judge adapter contract tests, AIClient fallback tests, privacy tests |
| A-19 | SDK streaming and Audio Transcriptions preserve `AIStreamEvent`, cancellation/partial meta, multipart and response-cap contracts | openai-compatible stream/STT contract tests |
| A-20 | Hand-written generic OpenAI wire/transport is removed without changing public A3/config/business contracts | negative symbol/import search, full AIClient/backend/root gates |
| A-21 | Local Complete request/response can be diagnosed from a secure independent NDJSON file without leaking secrets/content into ordinary observability or production | recorder concurrency/permissions/privacy tests + A4/A2/P0.099 handoff gates |

## 6 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-16 | 2.9 | Make raw-capture runtime write failures process-visible through a content-free structured warning while preserving the AI result. |
| 2026-07-16 | 2.8 | Reopen Phase 16 to replace stderr response-only debug with secure process-shared Complete request/response NDJSON capture. |
| 2026-07-16 | 2.6 | Replace self-maintained OpenAI-compatible transport with adapter-private `openai-go/v3 v3.43.0` while preserving AIClient, profile, fallback, metadata, privacy and response-cap contracts. |
| 2026-07-10 | 2.5 | Consolidate repeated observability privacy leak assertions. |
| 2026-07-10 | 2.4 | Consolidate repeated AIClient invalid-input error and metadata assertions. |
| 2026-07-10 | 2.3 | Consolidate fallback-label observability tests into one table. |
| 2026-07-10 | 2.2 | Consolidate repeated invalid-schema observability test setup and assertions. |
| 2026-07-10 | 2.1 | Consolidate observability latency fallback across Complete, STT and TTS. |
| 2026-07-10 | 2.0 | Consolidate duplicate chat and judge completion execution into one private helper. |
| 2026-07-10 | 1.9 | Remove the zero-consumer stub provider name wrapper. |
| 2026-07-10 | 1.8 | Remove the unreachable duplicate task-run and audit writer state from the core AIClient. |
| 2026-07-10 | 1.7 | Simplify OpenAI-compatible base URL normalization under existing root and `/v1` contract coverage. |
| 2026-07-07 | 1.6 | Compress owner docs to current AIClient, provider registry, model profile, adapter, observability and fail-fast contract. |
| 2026-05-05 | 1.5 | Complete provider terminology remediation and provider-neutral config/profile naming. |
