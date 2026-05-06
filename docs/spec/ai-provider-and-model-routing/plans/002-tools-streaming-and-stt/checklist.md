# AI Tools, Streaming, and STT Extension Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-06

**关联计划**: [plan](./plan.md)

> 本 plan 已于 2026-05-06 根据用户明确确认提前激活。每个实现 item 必须在 `/tdd` 中先写 Red test 或执行文档声明的替代 gate，再记录实际验证证据。Realtime multimodal / TTS / 媒体留存 / UI voice release gate 不在本 checklist 范围。

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

## Phase 3: Stream consumer 完整化

- [x] 3.1 openai_compatible SSE / chunked 解析映射到 plan 001 锁定的 delta / error / done 事件；验证: provider-side stream parser tests 覆盖多 chunk、malformed chunk、provider error event 与 done event，channel close 语义通过
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestStream_ParsesSSEDeltaAndDone|TestStream_MalformedChunkEmitsAIOutputInvalid|TestStream_ProviderErrorEventEmitsSharedError' -count=1 (stream returned single done event)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestStream_ParsesSSEDeltaAndDone|TestStream_MalformedChunkEmitsAIOutputInvalid|TestStream_ProviderErrorEventEmitsSharedError' -count=1" regression="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1" -->
- [x] 3.2 context cancellation 路径补齐 partial token meta 与 B1 错误码；验证: focused cancellation test 断言 context cancel 后 channel 收到 error/done 终态、partial token meta 尽力填充且错误码来自 B1 `AI_*`
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestStream_ContextCancelEmitsPartialDoneMeta -count=1 (missing partial done meta)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run TestStream_ContextCancelEmitsPartialDoneMeta -count=1" regression="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1" -->
- [x] 3.3 provider-side SSE consumer 选型落地，并把业务 HTTP wire handoff 写回 spec §3.1；验证: spec/history 更新通过 `make docs-check`，adapter contract tests 证明 provider SSE 形态一致，且后续 frontend-workspace-and-practice / backend API 用户可见入口仍需自身 BDD gate
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md docs/spec/ai-provider-and-model-routing/history.md docs/spec/INDEX.md" tests="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -count=1" command="make docs-check" -->

## Phase 4: STT provider adapter

- [x] 4.1 在 A3 spec §4.1 锁定 `Transcribe` 入参形态为 bytes + filename + content type + optional language/prompt；验证: A3 docs 引用同一 audio payload contract，`make docs-check` 与 `sync-doc-index --check` 通过
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md docs/spec/ai-provider-and-model-routing/history.md docs/spec/INDEX.md" command="make docs-check && python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check" -->
- [x] 4.2 落地 openai_compatible `/v1/audio/transcriptions` 适配，`capability=stt` 从 unsupported profile 升级为可执行；验证: STT adapter mockserver tests 覆盖 multipart bytes 形态 happy path、provider error path、secret missing fail-fast 与 unsupported profile fail-closed
  <!-- verified: 2026-05-06 red="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestTranscribe_PostsMultipartAudioAndReturnsTranscript|TestTranscribe_ProviderErrorReturnsSharedCode|TestTranscribe_MissingTextReturnsAIOutputInvalid' -count=1 (missing Transcribe/multipart support)" green="cd backend && go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestTranscribe_PostsMultipartAudioAndReturnsTranscript|TestTranscribe_ProviderErrorReturnsSharedCode|TestTranscribe_MissingTextReturnsAIOutputInvalid|TestNew_RequiresOpenAICompatibleResolvedProviderSecret' -count=1 && cd backend && go test ./internal/ai/aiclient -run 'TestTranscribe_RoutesSTTProfileThroughProvider|TestTranscribe_RequiresAudioBytesFilenameAndContentType|TestTranscribe_RealtimeProfileFailsClosed' -count=1" regression="cd backend && go test ./internal/ai/aiclient/... -count=1" -->
- [x] 4.3 校验或扩展 7 个 ai_* metric family 的 label 集合，确保 STT 可观测；验证: focused metric/log tests 断言 `capability=stt` 有界 label、无 audio/transcript 明文，F1 allowed/forbidden label gate 通过
  <!-- verified: 2026-05-06 docs="docs/spec/observability-stack/spec.md docs/spec/observability-stack/history.md" test="cd backend && go test ./internal/ai/aiclient/observability -run 'TestDecorator_TranscribeRecordsSTTWithoutPlaintext|TestPrivacy_NoPlaintextLeaksAnywhere' -count=1" command="make docs-check" -->
- [x] 4.4 复核 realtime fail-closed：只实现 STT 时不得打开 `practice.voice.realtime.default`；验证: `make lint-ai-profile-coverage` 断言 realtime profile 仍为 `unsupported`，除非 production voice / practice voice owner 已完成联合修订并记录触发证据
  <!-- verified: 2026-05-06 config="config/ai-profiles.yaml" test="cd backend && go test ./internal/ai/aiclient -run TestTranscribe_RealtimeProfileFailsClosed -count=1" command="make lint-ai-profile-coverage" -->

## Phase 5: 接入 F1 / F3 / B1

- [x] 5.1 F1 metric / log / dashboard 字段扩展同步；验证: F1 spec / generated lint gate 与 focused observability tests 对新增字段、allowed labels、forbidden labels 均通过；若 F1 仍使用旧任务分类 label，本 phase 原地修订为 AI metric `capability` label 口径
  <!-- verified: 2026-05-06 docs="docs/spec/observability-stack/spec.md docs/spec/observability-stack/history.md docs/spec/INDEX.md" test="cd backend && go test ./internal/ai/aiclient/observability -run 'TestDecorator_TranscribeRecordsSTTWithoutPlaintext|TestPrivacy_NoPlaintextLeaksAnywhere' -count=1" command="make docs-check" -->
- [x] 5.2 F3 profile schema 增量（tools / output_schema / stream_wire）先行落地，再被本 plan 消费；验证: F3 owner spec 或 plan 先行记录字段，`make lint-ai-profile-coverage` 覆盖 `config/ai-profiles.yaml` catalog 中新增 profile 字段和 status 语义
  <!-- verified: 2026-05-06 docs="docs/spec/prompt-rubric-registry/spec.md docs/spec/prompt-rubric-registry/history.md docs/spec/INDEX.md" command="make lint-ai-profile-coverage && make docs-check" -->
- [x] 5.3 B1 共享常量 / 错误码扩展先行合入，再在本 plan 引用；验证: `make codegen-check`、Go/TS AI vocabulary parity tests 与 repo-wide negative search 确认未在 A3 私造跨边界常量
  <!-- verified: 2026-05-06 command="make codegen-check" notes="B1 AI vocabulary/codegen remained in sync; no new A3-owned cross-boundary literal was introduced for STT beyond existing B1 capability/error constants." -->

## Phase 6: Verification

- [x] 6.1 spec §6 AC 表为每个被激活 phase 追加 ≥ 1 条 AC（含正常 / 错误 / 隐私 / 观测）；验证: AC 行引用本 plan 与被激活 capability，`make docs-check` 通过
  <!-- verified: 2026-05-06 docs="docs/spec/ai-provider-and-model-routing/spec.md#6-验收标准 C-13/C-14/C-15" command="make docs-check" -->
- [x] 6.2 单测 + 离线契约测试覆盖被激活的 tool / streaming / STT 协议子集；验证: `cd backend && go test ./internal/ai/aiclient/... -count=1`、新增 focused tests 与 adapter contract tests 均通过
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/ai/aiclient/... -count=1" focused="openai_compatible tool/stream/STT contract tests; aiclient Transcribe interface tests; observability privacy tests" -->
- [ ] 6.3 本地部署 + Kind 场景端到端 smoke 通过，无明文泄漏，埋点齐全；验证: 按 `test/scenarios/README.md` 与 active suite README 执行 smoke，记录真实 provider registry/profile/secret 组合，privacy grep 无明文
  <!-- blocked: 2026-05-06 scenario-readme="test/scenarios/README.md test/scenarios/e2e/README.md test/scenarios/e2e/INDEX.md" reason="当前 scenario framework 只有 Planned 索引；不存在 env-setup/env-verify 脚本、Kind 部署入口、Ready/Verified AI tool/stream/STT 场景目录或 consuming business flow，无法执行本地部署 + Kind e2e smoke；已用 focused contract/privacy/profile gates 作为当前内部底座替代验证，但本项保持未完成。" -->
- [x] 6.4 active-scope 旧口径负向搜索通过；验证: 搜索确认 A3-owned 代码、配置、deploy、generated artifacts、active docs 与被本 plan 激活并修订过的 owner docs 不再把旧任务分类 key、旧 provider key、一 profile 一目录 truth source、retired AI routing 术语、独立语音路由、独立旧模块口径作为 active runtime truth source；denylist / rejection validator / negative fixture 中的旧 literal 只作为防回归证据保留（历史 work journal / reports / bugs 只读例外；其他 referenced active spec 必须先完成 owner handoff 后再计入 pass）
  <!-- verified: 2026-05-06 commands="active-scope retired-literal rg sweeps over config/deploy/A3/F1/F3 current docs" result="no matches" allowed-exception="backend/internal/ai/aiclient/profile/loader.go keeps retired key literals only to reject them; scripts/tests keep denylist fixtures." -->
