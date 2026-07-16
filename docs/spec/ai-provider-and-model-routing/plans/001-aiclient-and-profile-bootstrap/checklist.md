# AIClient and Profile Bootstrap Checklist

> **版本**: 2.6
> **状态**: completed
> **更新日期**: 2026-07-16

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

- [x] 3.1 `openai_compatible` adapter keeps `openai-go/v3` imports private to the Phase 15 adapter/internal-helper allowlist and does not expose vendor SDK types or imports to business packages（验证：provider terminology lint、SDK import-boundary tests and adapter tests）
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

## Phase 7: AIClient duplicate writer state removal

- [x] 7.1 删除 core `aiclient.Client` 中零消费者的 task-run/audit writer fields、options 与 getters，使 `observability` decorator 成为唯一 writer 注入路径；验证：`deadcode -test` RED/GREEN、symbol inventory、`go test ./internal/ai/aiclient/... -count=1`、scoped `staticcheck` 与 owner docs gates
  <!-- verified: 2026-07-10 method=aiclient-duplicate-writer-state-removal evidence="deadcode -test RED reported both core writer options. Removed the duplicate fields, options and zero-consumer getters while retaining the observability decorator writer interfaces and options. AIClient/full backend tests, staticcheck, go vet, reachability rescan and symbol inventory PASS." -->

## Phase 8: Stub provider name wrapper removal

- [x] 8.1 删除零消费者 `stub.ProviderName` wrapper，以 `stub.Name` 作为唯一 exported provider identity；验证：`deadcode -test` RED/GREEN、symbol inventory、stub/AIClient tests、scoped `staticcheck` 与 owner docs gates
  <!-- verified: 2026-07-10 method=stub-provider-name-wrapper-removal evidence="deadcode -test RED reported ProviderName as unreachable. Deleted the wrapper without changing stub.Name or any consumer. Stub/full AIClient tests, staticcheck, reachability rescan and exact declaration inventory PASS." -->

## Phase 9: completion dispatch duplication removal

- [x] 9.1 Record scoped `dupl` RED evidence for the identical `Complete` and `CompleteJudge` execution bodies.
  <!-- verified: 2026-07-10 method=aiclient-completion-dupl-red evidence="Scoped golangci-lint dupl reported only client.go lines 68-90 and 99-121 as reciprocal duplicates for Complete and CompleteJudge." -->
- [x] 9.2 Extract one private capability-parameterized helper while preserving both public methods and their chat/judge fail-close behavior.
  <!-- verified: 2026-07-10 method=aiclient-completion-dispatch-green evidence="Complete and CompleteJudge now pass CapabilityChat/CapabilityJudge to one private complete helper. Focused AIClient tests and subtree staticcheck pass, while the scoped client.go dupl gate reports zero findings." -->
- [x] 9.3 Run focused/full AIClient, judge/bootstrap, `staticcheck`, scoped `dupl`, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=aiclient-completion-dispatch-consolidation evidence="Scoped client.go dupl is clean; focused AIClient and complete aiclient subtree plus registry judge/bootstrap tests pass. Full backend go test ./... -count=1, go vet and staticcheck pass. Both contexts, links and final docs/index/diff/pruning gates pass with real_residuals=0. Public methods, chat/judge capability fail-close, fallback and metadata contracts are unchanged. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 10: observability latency fallback consolidation

- [x] 10.1 Record scoped `dupl` RED evidence for the identical Transcribe and Synthesize timing/metadata wrappers.
  <!-- red: 2026-07-10 method=observability-speech-latency-dupl evidence="Scoped dupl -t 100 reported decorator.go lines 155-166 and 195-206 as the file's only clone group." -->
- [x] 10.2 Extract one private latency fallback helper and reuse it from Complete, Transcribe and Synthesize without changing capability-specific record paths.
  <!-- verified: 2026-07-10 method=observability-latency-fallback-helper evidence="Complete, Transcribe and Synthesize call one withLatencyFallback helper after the same completion timestamp sample. The helper contains the only LatencyMs==0/measured-duration fallback; scoped dupl reports zero groups and observability tests plus staticcheck pass." -->
- [x] 10.3 Run scoped `dupl`, observability/privacy, full AIClient/backend, vet/staticcheck, owner/product contexts and docs/index/diff/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=observability-latency-fallback-consolidation evidence="Scoped decorator.go dupl is clean; observability/privacy, full AIClient/registry and full backend tests pass, as do go vet/staticcheck. A3/product contexts and final docs/index/link/diff/pruning gates pass with real_residuals=0. Provider latency precedence, measured fallback, capability-specific record paths and privacy behavior are unchanged. No Bug/retrospective report, environment restart or data cleanup was needed." -->

## Phase 11: observability invalid-schema test harness consolidation

- [x] 11.1 Record scoped `dupl` RED and exact-name consumer inventory for the invalid-schema tests.
  <!-- red: 2026-07-10 method=observability-invalid-schema-test-dupl evidence="Scoped dupl -t 100 reported trailing-token and enum-mismatch tests as one clone group; BUG-0095 and prompt-rubric owner docs reference the exact top-level test names." -->
- [x] 11.2 Extract one test-only invalid-schema harness, preserve all three top-level test names and keep each content/schema case explicit.
  <!-- verified: 2026-07-10 method=observability-invalid-schema-test-harness evidence="The required-field, trailing-token and enum-mismatch top-level tests retain their exact names and explicit content/schema inputs while sharing one helper for decorator setup and invalid/error/metric assertions. Exact-name -run/-list and staticcheck pass; decorator_test.go clone groups drop from two to one unrelated group." -->
- [x] 11.3 Run exact-name focused tests, scoped dupl reduction, observability/AIClient/full backend, vet/staticcheck and owner/product/docs/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=observability-invalid-schema-test-harness-consolidation evidence="All three exact top-level tests remain listed and pass. decorator_test.go clone groups drop from two to one unrelated fallback group; observability, full AIClient/registry and full backend tests plus go vet/staticcheck pass. A3/product contexts and final docs/index/link/diff/pruning gates pass with real_residuals=0. No test input or assertion was removed, and no Bug/retrospective report or environment operation was needed." -->

## Phase 12: observability fallback-label test table consolidation

- [x] 12.1 Record scoped `dupl` RED and verify the two top-level test names have no external consumers.
  <!-- red: 2026-07-10 method=observability-fallback-label-test-dupl evidence="After Phase 11, scoped dupl -t 100 reports only the date-suffix and central-chain fallback counter tests as one clone group; repo search finds no external exact-name references." -->
- [x] 12.2 Replace both tests with one table-driven test while retaining both meta inputs and exact label tuples.
  <!-- verified: 2026-07-10 method=observability-fallback-label-table evidence="One TestDecorator_FallbackCounterLabelDerivation now runs named date-suffix and central-chain cases with complete meta and exact 11-label tuples. Both subtests pass, old top-level names are absent, scoped dupl reports zero clone groups and staticcheck passes." -->
- [x] 12.3 Run focused fallback tests, scoped dupl, observability/AIClient/full backend, vet/staticcheck and owner/product/docs/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=observability-fallback-label-table-consolidation evidence="Both named fallback label cases pass with their exact 11-label tuples; decorator_test.go scoped dupl reports zero groups. Full observability/AIClient/registry/backend tests and go vet/staticcheck pass, as do A3/product contexts and final docs/index/link/diff/pruning gates with real_residuals=0. No coverage was removed and no Bug/retrospective report or environment operation was needed." -->

## Phase 13: AIClient invalid-input assertion consolidation

- [x] 13.1 Record scoped `dupl` RED for the Transcribe and Synthesize invalid-input tests and inventory the matching Complete assertion.
  <!-- red: 2026-07-10 method=aiclient-invalid-input-test-dupl evidence="Scoped dupl -t 100 reports the Transcribe and Synthesize invalid-input tests as a clone; Complete empty-messages repeats the same shared error/meta contract." -->
- [x] 13.2 Extract one test-only `AI_OUTPUT_INVALID` assertion helper while preserving all three top-level tests and provider call-count guards.
  <!-- verified: 2026-07-10 method=aiclient-invalid-input-assertion-helper evidence="Complete, Transcribe and Synthesize invalid-input tests keep their exact names and capability-specific provider guards while sharing one error/meta helper. All focused tests pass, scoped aiclient_test.go dupl is zero, and staticcheck passes." -->
- [x] 13.3 Run exact focused tests, scoped dupl, full AIClient/backend, vet/staticcheck and owner/product/docs/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=aiclient-invalid-input-assertion-consolidation evidence="All three exact Complete/STT/TTS invalid-input tests pass and retain provider call-count guards; aiclient_test.go scoped dupl is zero. Full AIClient/registry/backend tests and go vet/staticcheck pass, as do A3/product contexts and final docs/index/link/diff/pruning gates with real_residuals=0. No Bug/retrospective report or environment operation was needed." -->

## Phase 14: observability privacy leak assertion consolidation

- [x] 14.1 Record scoped `dupl` RED and exact-name consumer inventory for Complete/TTS privacy tests.
  <!-- red: 2026-07-10 method=observability-privacy-scan-dupl evidence="Scoped dupl -t 100 reports the Complete and TTS metric/log/task-run/audit token scans as privacy_test.go's only clone group; external gates reference TestPrivacy_NoPlaintextLeaksAnywhere by exact name." -->
- [x] 14.2 Extract one token-parameterized plaintext scan helper while preserving both top-level tests and their scenario-specific sanity assertions.
  <!-- verified: 2026-07-10 method=observability-privacy-scan-helper evidence="Complete and TTS tests retain their exact names and capability/metric/audit sanity assertions while one helper scans all planted tokens across six counter families, logs, task runs and audit metadata. Focused tests pass, privacy_test.go dupl is zero and staticcheck passes." -->
- [x] 14.3 Run exact privacy tests, scoped dupl, observability/AIClient/full backend, vet/staticcheck and owner/product/docs/pruning gates; then restore the owner to `completed`.
  <!-- verified: 2026-07-10 method=observability-privacy-scan-consolidation evidence="Exact Complete/TTS privacy tests pass with their capability/metric/audit sanity checks intact; privacy_test.go scoped dupl is zero. Full observability/AIClient/registry/backend tests and go vet/staticcheck pass, as do A3/product contexts and final docs/index/link/diff/pruning gates with real_residuals=0. No privacy surface was removed and no Bug/retrospective report or environment operation was needed." -->

## Phase 15: Official openai-go transport migration

- [x] 15.1 RED/GREEN SDK dependency boundary：在 `scripts/lint/ai_provider_terminology_test.py` 新增失败测试证明 `github.com/openai/openai-go/v3` 在 allowlist 外会被拒绝；固定 `v3.43.0`，只允许 `providers/openai_compatible`、`providers/judge_compatible` 与精确的 `providers/internal/openaisdk` helper import，并同步 A3 README/package docs（验证：`python3 scripts/lint/ai_provider_terminology_test.py` RED/GREEN、`cd backend && go list -m github.com/openai/openai-go/v3`、repo-wide import search）
  <!-- verified: 2026-07-16 method=tdd-red-green evidence="RED: the new out-of-boundary openai-go import test returned 0. GREEN: 5 lint tests and the 534-file active-surface scan pass; go list resolves github.com/openai/openai-go/v3 v3.43.0 with Go 1.22, and production Go imports remain empty before adapter migration." -->
- [x] 15.2 RED/GREEN Complete + judge：先补 contract tests 锁定 custom base URL、HTTP client、auth、messages/tools/JSON/DeepSeek thinking、usage/headers、4xx/5xx/timeout/retry、missing/empty/reasoning-only 与 response cap，再切换到 SDK Chat Completions（验证：`cd backend && go test ./internal/ai/aiclient/providers/openai_compatible ./internal/ai/aiclient/providers/judge_compatible ./internal/ai/aiclient -count=1`）
  <!-- verified: 2026-07-16 method=openai-go-chat-completions-green evidence="Both adapters fail RED on transient 503 without retry, then pass through openai-go v3 Chat Completions with custom base URL/client/auth, initial-plus-two same-provider retries inside the profile timeout, tools/JSON/DeepSeek thinking, canonical error privacy, trusted metadata and bounded non-streaming response bodies. Focused OpenAI-compatible, judge, core AIClient tests and active SDK import-boundary lint PASS." -->
- [x] 15.3 RED/GREEN streaming + STT：先补 SSE delta/error/done、cancel/partial meta、malformed/oversized event 与 multipart filename/content-type/language/prompt/response-cap tests，再切换 SDK streaming / Audio Transcriptions 并映射回冻结 public contract；stream event limit 与 non-streaming body cap 分开保持（验证：`cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1`）
  <!-- verified: 2026-07-16 method=openai-go-streaming-stt-green evidence="RED proves manual streaming/STT stop at the first transient 503 and the old stream cap incorrectly applies to the whole response. GREEN uses SDK NewStreaming and Audio Transcriptions with initial-plus-two retries, frozen delta/error/done and cancellation partial-meta mapping, multipart filename/content-type/language/prompt, per-event SSE bounds and separate non-streaming body bounds. Full provider tests and active import-boundary lint PASS." -->
- [x] 15.4 删除被 SDK 覆盖的手写通用 OpenAI wire/post/SSE/multipart transport，保留 project mapping/extras/error/meta/privacy/response-cap glue；运行 SDK import/旧符号负向搜索、AIClient/config/lint、`go vet`、`staticcheck`、根 `make test`、`make build`、owner context/docs/index/diff gates并恢复 completed
  <!-- verified: 2026-07-16 method=openai-go-cleanup-and-baseline-remediation evidence="Manual generic wire/post/SSE/multipart runtime symbols are absent outside contract mockserver fixtures; openai-go v3.43.0 imports remain restricted to the two compatible adapters and their internal helper. Full staticcheck and UI Demo pruning were restored to zero by relocating an integration-only helper, deleting two zero-reference helpers, applying one equivalent score-level conversion and replacing one stale UI truth-source heading. Focused packages, integration-tag compile, go vet, make test (Python 567/4481 subtests, Go all packages, frontend 126 files/1004 tests), make build, AI terminology/profile/config lint, context/docs/index/mod/diff and pruning gates PASS." -->
