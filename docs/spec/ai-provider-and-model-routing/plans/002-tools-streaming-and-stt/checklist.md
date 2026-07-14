# AI Tools, Streaming, and STT Extension Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

> 本 plan 已于 2026-05-06 根据用户明确确认提前激活。每个实现 item 必须在 `/tdd` 中先写 Red test 或执行文档声明的替代 gate，再记录实际验证证据。provider-specific speech / TTS 归 004，媒体留存与电话模式 UI/API 归 `practice-voice-mvp`；realtime multimodal 与 judge 不在本 checklist 范围。

## Phase 1: 触发条件复核与 ADR / spec 修订

- [x] 1.1 在工作日志中归档触发证据（用户确认 / 业务 spec id / plan id / 事故记录 / 上游版本号）；验证: 触发来源可追溯到 active spec / plan / bug / work-journal，且 `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/ai-provider-and-model-routing/plans/002-tools-streaming-and-stt/context.yaml --docs-root docs --target backend` 通过
  <!-- verified: 2026-05-06 work-journal=docs/work-journal/2026-05-06.md#1345-工作记录 command="python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/ai-provider-and-model-routing/plans/002-tools-streaming-and-stt/context.yaml --docs-root docs --target backend" -->
- [x] 1.2 完成 ADR-Q6 v2.0 修订（保留零 SDK / 隐私 / 唯一对外能力红线）；验证: ADR Header 合法、状态为 `accepted`，并通过 `make docs-check`
  <!-- verified: 2026-05-06 docs=docs/spec/engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md command="make docs-check" -->
- [x] 1.3 把 spec 版本从 2.4 递增到 2.5 并同步 history.md；验证: `docs/spec/ai-provider-and-model-routing/spec.md`、`history.md` 与 `docs/spec/INDEX.md` 版本一致，`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 通过
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md docs/spec/ai-provider-and-model-routing/history.md docs/spec/INDEX.md" command="python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check" -->
- [x] 1.4 把本 plan Header 切换为 `状态: active` + `版本: 1.0`，并同步 plans/INDEX.md；验证: `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 显示本 plan 位于 Active 分组且无 Header/INDEX drift
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/plans/002-tools-streaming-and-stt/plan.md docs/spec/ai-provider-and-model-routing/plans/INDEX.md" command="python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check" -->

## Phase 2: Tools / function calling 实现

- [x] 2.1 在 spec §4.1 锁定 `Complete` payload tools 扩展形态；验证: 新增/调整的 Go interface contract test 先 Red 后 Green，且业务调用仍只传 `model_profile_name`，不传 provider/model 字符串
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient -run TestComplete_ToolsPayloadRemainsProviderNeutral -count=1 (missing Tools/ToolChoice contract)" green="cd backend && go test ./internal/ai/aiclient -run TestComplete_ToolsPayloadRemainsProviderNeutral -count=1" regression="cd backend && go test ./internal/ai/aiclient -count=1" -->
- [x] 2.2 openai_compatible adapter + stub provider 落地 tool 调用与 deterministic 回放；验证: focused adapter mockserver tests 覆盖 `tool_calls` / `tool_choice` / structured output happy path 与 provider 4xx/5xx error path，stub provider deterministic replay test 通过
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestComplete_MapsToolsAndParsesToolCalls -count=1 (tool_calls empty)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestComplete_MapsToolsAndParsesToolCalls -count=1 && cd backend && go test ./internal/ai/aiclient/providers/stub -run TestStubCompleteWithToolsIsDeterministic -count=1" regression="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1 && cd backend && go test ./internal/ai/aiclient/providers/stub -count=1" -->
- [x] 2.3 `AICallMeta` 扩展 tool 相关字段，log / DB 守住 hash / 长度 / profile 红线；验证: observability/privacy tests 断言 tool args 明文不进入 log / DB / audit / metric label，B1 vocabulary/codegen drift gate 通过
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestComplete_MapsToolsAndParsesToolCalls -count=1 (missing ToolInvocations)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestComplete_MapsToolsAndParsesToolCalls -count=1 && cd backend && go test ./internal/shared/ai -count=1 && cd backend && go test ./internal/ai/aiclient/observability -run 'TestPrivacy_NoPlaintextLeaksAnywhere|TestDecorator_SuccessIncrementsRunsAndLogsCompleted' -count=1" drift="make codegen-conventions && python3 scripts/lint/conventions_drift.py --repo-root ." -->
- [x] 2.4 L2 remediation: stub provider 必须真正 replay provider-neutral tool call，而不是只证明带 tool payload 的文本响应 deterministic；验证: focused Red-Green test 断言 tool choice 返回 deterministic `ToolCalls`，arguments 由 hash/长度摘要进入 `AICallMeta.ToolInvocations` 且不含明文
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/stub -run TestStubCompleteWithToolsIsDeterministic -count=1 (expected deterministic stub tool call replay, got empty ToolCalls)" green="cd backend && go test ./internal/ai/aiclient/providers/stub -run TestStubCompleteWithToolsIsDeterministic -count=1" regression="cd backend && go test ./internal/ai/aiclient/providers/stub -count=1" -->

## Phase 3: Stream consumer 完整化

- [x] 3.1 openai_compatible SSE / chunked 解析映射到 plan 001 锁定的 delta / error / done 事件；验证: provider-side stream parser tests 覆盖多 chunk、malformed chunk、provider error event 与 done event，channel close 语义通过
  <!-- verified: 2026-07-10 method=current-stream-contract-gate evidence="TestStream_ParsesSSEDeltaAndDone plus TestStream_ErrorChunksEmitSharedError named malformed/provider cases cover delta/done, malformed AI_OUTPUT_INVALID, provider AI_PROVIDER_TIMEOUT and channel close; full openai_compatible package passes." -->
- [x] 3.2 context cancellation 路径补齐 partial token meta 与 B1 错误码；验证: focused cancellation test 断言 context cancel 后 channel 收到 error/done 终态、partial token meta 尽力填充且错误码来自 B1 `AI_*`
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestStream_ContextCancelEmitsPartialDoneMeta -count=1 (missing partial done meta)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestStream_ContextCancelEmitsPartialDoneMeta -count=1" regression="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1" -->
- [x] 3.3 provider-side SSE consumer 选型落地，并把业务 HTTP wire handoff 写回 spec §3.1；验证: spec/history 更新通过 `make docs-check`，adapter contract tests 证明 provider SSE 形态一致，且后续 frontend-workspace-and-practice / backend API 用户可见入口仍需自身 BDD gate
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md docs/spec/ai-provider-and-model-routing/history.md docs/spec/INDEX.md" tests="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1" command="make docs-check" -->
- [x] 3.4 L2 remediation: `AIClient.Stream` terminal `done` meta 必须经过 canonical meta merge，补齐 `Capability`、`ModelProfileName`、`ModelProfileVersion`、`PromptVersion`、`RubricVersion`、`Language` 与 `ValidationStatus`；验证: client-level stream Red-Green test 覆盖 normal done 与 cancellation partial done
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient -run 'TestStream_(DoneEventAndChannelClose|PartialDoneMetaIsCanonicalMerged)' -count=1 (done meta missing canonical fields)" green="cd backend && go test ./internal/ai/aiclient -run 'TestStream_(DoneEventAndChannelClose|PartialDoneMetaIsCanonicalMerged)' -count=1" regression="cd backend && go test ./internal/ai/aiclient -count=1" -->

## Phase 4: STT provider adapter

- [x] 4.1 在 A3 spec §4.1 锁定 `Transcribe` 入参形态为 bytes + filename + content type + optional language/prompt；验证: A3 docs 引用同一 audio payload contract，`make docs-check` 与 `sync-doc-index --check` 通过
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md docs/spec/ai-provider-and-model-routing/history.md docs/spec/INDEX.md" command="make docs-check && python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check" -->
- [x] 4.2 落地 openai_compatible `/v1/audio/transcriptions` 适配，使 `capability=stt` adapter 可执行；tracked voice profile 状态由 voice product owner 决定；验证: STT adapter mockserver tests 覆盖 multipart bytes 形态 happy path、provider error path、secret missing fail-fast 与 unsupported profile fail-closed。
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestTranscribe_PostsMultipartAudioAndReturnsTranscript|TestTranscribe_ProviderErrorReturnsSharedCode|TestTranscribe_MissingTextReturnsAIOutputInvalid' -count=1 (missing Transcribe/multipart support)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestTranscribe_PostsMultipartAudioAndReturnsTranscript|TestTranscribe_ProviderErrorReturnsSharedCode|TestTranscribe_MissingTextReturnsAIOutputInvalid|TestNew_RequiresOpenAICompatibleResolvedProviderSecret' -count=1 && cd backend && go test ./internal/ai/aiclient -run 'TestTranscribe_RoutesSTTProfileThroughProvider|TestTranscribe_RequiresAudioBytesFilenameAndContentType|TestTranscribe_RealtimeProfileFailsClosed' -count=1" regression="cd backend && go test ./internal/ai/aiclient/... -count=1" -->
- [x] 4.3 校验或扩展 7 个 ai_* metric family 的 label 集合，确保 STT 可观测；验证: focused metric/log tests 断言 `capability=stt` 有界 label、无 audio/transcript 明文，F1 allowed/forbidden label gate 通过
  <!-- verified: 2026-05-06 docs="docs/spec/observability-stack/spec.md docs/spec/observability-stack/history.md" test="cd backend && go test ./internal/ai/aiclient/observability -run 'TestDecorator_TranscribeRecordsSTTWithoutPlaintext|TestPrivacy_NoPlaintextLeaksAnywhere' -count=1" command="make docs-check" -->
- [x] 4.4 复核 realtime fail-closed：只实现 STT 时不得打开 `practice.voice.realtime.default`；验证: `make lint-ai-profile-coverage` 断言 realtime profile 仍为 `unsupported`，除非 production voice / practice voice owner 已完成联合修订并记录触发证据
  <!-- verified: 2026-05-06 config="config/ai-profiles.yaml" test="cd backend && go test ./internal/ai/aiclient -run TestTranscribe_RealtimeProfileFailsClosed -count=1" command="make lint-ai-profile-coverage" -->

## Phase 5: 接入 F1 / F3 / B1

- [x] 5.1 F1 metric / log / dashboard 字段扩展同步；验证: F1 spec / generated lint gate 与 focused observability tests 对新增字段、allowed labels、forbidden labels 均通过，AI metric label 使用 `capability`
  <!-- verified: 2026-05-06 docs="docs/spec/observability-stack/spec.md docs/spec/observability-stack/history.md docs/spec/INDEX.md" test="cd backend && go test ./internal/ai/aiclient/observability -run 'TestDecorator_TranscribeRecordsSTTWithoutPlaintext|TestPrivacy_NoPlaintextLeaksAnywhere' -count=1" command="make docs-check" -->
- [x] 5.2 F3 profile schema 增量（tools / output_schema / stream_wire）先行落地，再被本 plan 消费；验证: F3 owner spec 或 plan 先行记录字段，`make lint-ai-profile-coverage` 覆盖 `config/ai-profiles.yaml` catalog 中新增 profile 字段和 status 语义
  <!-- verified: 2026-05-06 docs="docs/spec/prompt-rubric-registry/spec.md docs/spec/prompt-rubric-registry/history.md docs/spec/INDEX.md" command="make lint-ai-profile-coverage && make docs-check" -->
- [x] 5.3 B1 共享常量 / 错误码扩展先行合入，再在本 plan 引用；验证: `make codegen-check`、Go/TS AI vocabulary parity tests 与 repo-wide negative search 确认未在 A3 私造跨边界常量
  <!-- verified: 2026-05-06 command="make codegen-check" notes="B1 AI vocabulary/codegen remained in sync; no new A3-owned cross-boundary literal was introduced for STT beyond existing B1 capability/error constants." -->
- [x] 5.4 L2 remediation: observability wrapper 在 stream done / stream error / pre-dispatch failure 记录前必须用 ProfileResolver enrichment 补齐 profile/capability/route label，不得把可解析 profile 的失败路径长期落为 `unknown`；验证: focused observability Red-Green tests 覆盖 stream done、stream error 与 invalid Complete failure labels
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/observability -run 'TestDecorator_(PreDispatchFailureUsesResolvedProfileLabels|StreamDoneUsesResolvedProfileLabels|StreamErrorUsesResolvedProfileLabels)' -count=1 (expected enriched labels, got 0)" green="cd backend && go test ./internal/ai/aiclient/observability -run 'TestDecorator_(PreDispatchFailureUsesResolvedProfileLabels|StreamDoneUsesResolvedProfileLabels|StreamErrorUsesResolvedProfileLabels)' -count=1" regression="cd backend && go test ./internal/ai/aiclient/observability -count=1" -->

## Phase 6: Verification

- [x] 6.1 spec §6 AC 表为每个被激活 phase 追加 ≥ 1 条 AC（含正常 / 错误 / 隐私 / 观测）；验证: AC 行引用本 plan 与被激活 capability，`make docs-check` 通过
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md#6-验收标准 C-13/C-14/C-15" command="make docs-check" -->
- [x] 6.2 单测 + 离线契约测试覆盖被激活的 tool / streaming / STT 协议子集；验证: `cd backend && go test ./internal/ai/aiclient/... -count=1`、新增 focused tests 与 adapter contract tests 均通过
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/ai/aiclient/... -count=1" focused="openai_compatible tool/stream/STT contract tests; aiclient Transcribe interface tests; observability privacy tests" -->
- [x] 6.3 BDD-N/A/REGRESSION: 本 plan 不要求真实 provider smoke；运行 provider/AIClient contract、observability/privacy、profile/config lint 与根 `make test`，并保持 E2E 目录不包装代码测试。
- [x] 6.4 active-scope out-of-scope 输入负向搜索通过；验证: 搜索确认 A3-owned 代码、配置、deploy、generated artifacts、active docs 与本 plan 修订过的 owner docs 只使用 current capability keys、provider keys、单一 profile catalog、provider-ref routing 与当前模块命名；精确 out-of-scope literal 仅在 denylist / rejection validator / negative fixture 中作为防回归证据保留（历史 work journal / reports / bugs 只读例外）
  <!-- verified: 2026-05-06 commands="active-scope out-of-scope-literal rg sweeps over config/deploy/A3/F1/F3 current docs" result="no matches" allowed-exception="loader rejection tests and lint fixtures retain exact out-of-scope keys." -->

## Phase 7: Stream error contract test table consolidation

- [x] 7.1 Record scoped OpenAI-compatible `dupl` RED and confirm the two exact old test names have no consumers outside this owner checklist.
  <!-- verified: 2026-07-10 method=openai-stream-error-test-dupl evidence="Scoped dupl -t 100 reports malformed-chunk and provider-error tests as the provider package's only clone group; repo search finds only their declarations and this owner gate." -->
- [x] 7.2 Replace both old tests and owner references with one table-driven test retaining both named chunks and exact shared error codes.
  <!-- verified: 2026-07-10 method=openai-stream-error-test-table evidence="TestStream_ErrorChunksEmitSharedError runs named malformed_chunk and provider_error_event cases with exact chunks and expected AI_OUTPUT_INVALID/AI_PROVIDER_TIMEOUT codes. Focused/full provider tests pass, old names are absent, scoped dupl is zero and staticcheck passes." -->
- [x] 7.3 Run focused/full provider and AIClient tests, root `make test`, vet/staticcheck and 002/product/docs/pruning gates；plan 收口为 completed。
