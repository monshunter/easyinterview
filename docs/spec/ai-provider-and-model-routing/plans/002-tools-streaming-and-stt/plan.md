# AI Tools, Streaming, and STT Extension

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ai-provider-and-model-routing spec](../../spec.md) §3.2 待确认事项与 §7 关联计划末段所列「按需决策」的三类扩展能力提前升级为当前 AI 底座。2026-05-06 用户明确选择“方案 B：AI 提前落地，避免后续业务开发返工”，因此本 plan 从 draft 激活为 active，并按 [ADR-Q6 v2.0](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 进入实施；当前合同以 A3 spec v2.22 为准。覆盖范围：

- **Tools / function calling**：在 `Complete` payload 上追加 provider-neutral `tools[]` / `tool_choice` / `output_schema`，并在 response/meta 中返回 tool-call 摘要，保持 provider-neutral 红线不破。
- **Stream consumer 完整化**：基于 plan 001 已冻结的 `AIStreamEvent` 事件合同，落地 provider 侧 SSE / chunked streaming 消费 + `<-chan AIStreamEvent` 完整生命周期，并把 partial-token meta（context cancellation 场景）补齐；业务 HTTP wire 由调用方 owner 决定。
- **STT (`Transcribe(ctx, profile, audio) → (transcript, meta)`)**：新增 OpenAI-compatible `/v1/audio/transcriptions` provider adapter，使 `capability=stt` 具备可执行 adapter；tracked `practice.voice.stt.default` 继续由产品 owner 显式决定 disabled/active。provider-specific speech / TTS 底座归 004，用户可见电话模式与媒体生命周期归 `practice-voice-mvp`。

本 plan 不修改 [001-aiclient-and-profile-bootstrap](../001-aiclient-and-profile-bootstrap/plan.md) 的 phase 范围；plan 001 锁定的 `AIStreamEvent` 类型保持稳定并被本 plan 复用。

本 plan 的 provider-neutral tools、streaming 与 STT adapter contract 已完成。配置/协议完成条件由代码层 contract、privacy、lint 与根 `make test` 承接，不要求真实 provider smoke，也不得包装成 E2E。

## 2 背景

[spec §3.2 待确认事项](../../spec.md#32-待确认事项) 曾列出三项尚未锁定的边界：`Tools(...)` 是否扩展、`model_profile_version` 是否独立 SemVer、`Stream` HTTP wire 选型。其中 Tools 与 provider-side streaming 已在 [ADR-Q6 v2.0](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 中提前打开为底座能力；HTTP 业务 wire 仍由后续 backend/frontend owner 决定。STT `Transcribe(...)` 当前由 A3 先提供 OpenAI-compatible Audio Transcriptions adapter，production voice / practice voice owner 后续只需消费 `capability=stt` profile，不必再次重做底层 provider adapter。

为避免后续业务域临时需求驱动「就地塞入新接口」破坏 ADR-Q6 9 项硬约束，本 plan 统一落地底座能力：Complete tools payload、provider-side streaming consumer、STT transcription adapter、B1/F1/F3/A4 边界同步与 drift gates。Realtime multimodal 仍显式 fail-closed，不在本 plan 顺手打开。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + provider-adapter + observability/privacy`。本 plan 修改 AIClient contract、provider adapter、metadata、profile/registry catalog、B1/F1/F3/A4 接口边界与验证资产。
- **TDD 策略**: 通过 `/implement` -> `/tdd` 顺序执行。每个 checklist item 必须以接口契约 tests、adapter mockserver tests、stream cancellation tests、STT contract tests、metric/log/privacy assertions、profile/registry lint 和 negative search 作为 Red-Green-Refactor 断言来源。
- **BDD 策略**: BDD 不适用。本 plan 打开的能力仍是内部 AI provider / provider protocol / observability 契约；任何把 streaming、STT 或 tool 调用暴露到用户可见 UI / API workflow 的后续 workstream 必须在自身 plan 维护 BDD gate。
- **替代验证 gate**: focused Go tests、OpenAI-compatible protocol contract tests、F1/F3/B1 compatibility checks、profile coverage lint、privacy grep、startup/config contract、active-scope negative search、根 `make test` 与 `sync-doc-index --check`。

### 3.1 声明式边界锁定

- **配置真理源**: provider registry 以 `config/ai-providers.yaml` 为唯一 registry truth source，Model Profile 以单一 `config/ai-profiles.yaml` catalog 为唯一 profile truth source；runtime 只接受 current `capability`、provider key 与 provider-ref endpoint 口径。
- **运行时派生字段**: `AICallMeta`、fallback chain、stream partial meta、tool invocation summary、STT usage 统计和 observability labels 均由运行时调用派生，不得写回 registry / profile YAML。
- **跨边界字面量**: `capability`、provider/profile 字段名、AI meta 字段名和 `AI_*` 错误码由 B1 生成或登记；本 plan 不私造 Go/TS/OpenAPI 共享常量。
- **兼容策略**: runtime 只接受 current schema；不新增 schema fallback 或 vendor SDK，不允许业务包直接引用 provider/model 字符串。
- **提前激活边界**: Tools / provider-side streaming / OpenAI-compatible STT transcription 已由本 plan 完成；provider-specific speech / TTS 归 004，媒体留存与 UI release gate 归 `practice-voice-mvp`，realtime multimodal / judge 继续 fail-closed。

## 4 实施步骤

### Phase 1: 触发条件复核与 ADR / spec 修订

> **提前激活依据**: 2026-05-06 用户明确选择提前落地完整 AI 底座，避免后续业务开发返工；本依据已写入 ADR-Q6 v2.0、A3 spec v2.5 与工作日志。

#### 1.1 触发证据归档

把触发来源（用户确认、业务 spec id、plan id、事故记录或工作日志条目）汇总到本 plan 的工作日志条目中，明确说明提前激活原因、参考的上游文档版本号和仍不打开的能力。

#### 1.2 ADR-Q6 修订

按 [ADR-Q6 §5](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md#5-失效与修订条件) 提示的修订流程原地修订 ADR-Q6 到 v2.0，显式记录被打开的能力（Tools / Stream / STT）以及与原 9 项硬约束的兼容性论证；零厂商 SDK 入侵、`AIClient` 唯一对外能力、隐私红线必须保留。

#### 1.3 spec 版本递增

把 [ai-provider-and-model-routing spec](../../spec.md) 版本从 `2.4` 递增到 `2.5`，在 §2.1 / §3.1 / §4 / §6 中追加被打开能力对应的接口契约、决策项、设计约束与 AC。同步更新 [history.md](../../history.md)。

#### 1.4 plan 状态切换

完成 1.2 与 1.3 后，把本 plan Header 从 `状态: draft` 调整为 `状态: active` / `版本: 1.0`，并把 [plans/INDEX.md](../INDEX.md) 同步。未完成 1.1–1.3 的情况下严禁切换。

### Phase 2: Tools / function calling 实现

#### 2.1 接口扩展

在 spec §4.1 锁定 `Complete` payload 扩展形态：新增 `tools[]` / `tool_choice`，复用既有 `output_schema` 字段。业务代码继续只引用 profile name，调用现场不感知 provider 工具协议差异。

#### 2.2 OpenAI-compatible 实现

在 `backend/internal/ai/aiclient/` 的 openai_compatible adapter 中追加 `tool_calls` / `tool_choice` 字段映射；stub provider 增加 deterministic tool-call 回放路径。零 vendor SDK 红线保持。

#### 2.3 观测与隐私

`AICallMeta` 新增 `tool_invocations[]`（或等价聚合字段，最终字段名以 spec 修订版为准）；log / DB metadata 仍只写 hash + 长度 + profile，不写 tool args 明文。

#### 2.4 L2 remediation: stub tool replay

修复 L2 发现的 stub tool replay 证明缺口：stub provider 在 payload 携带 provider-neutral `tools[]` / `tool_choice` 时必须返回 deterministic `ToolCalls`，并填充 `AICallMeta.ToolInvocations` 的 name / arguments hash / length 摘要；测试必须证明 stub 不只是返回普通文本响应。

### Phase 3: Stream consumer 完整化

#### 3.1 Provider-side streaming 消费

实现 openai_compatible adapter 的 SSE / chunked 解析，把 provider 增量映射为 plan 001 已冻结的 `delta` / `error` / `done` 事件，channel 生命周期与 spec §4.1 一致。

#### 3.2 取消语义与 partial meta

补齐 context cancellation 路径：截至中断时点的 `input_tokens` / `output_tokens` 必须尽力填充，provider 不支持时填 0；`error_code` 写入取消原因（B1 已锁定的 `AI_*` 错误码集合，不引入新字符串常量除非先改 B1）。

#### 3.3 HTTP wire 选型

落地 provider-side SSE consumer，并把决策写回 spec §3.1 / §4.1。业务 HTTP wire（SSE 或 chunked）由 [frontend-workspace-and-practice](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) / [backend-practice](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 自行 plan，不由本 plan 承担。

#### 3.4 L2 remediation: stream canonical meta

修复 L2 发现的 `Stream` terminal meta 漂移：`AIClient.Stream` 必须像 `Complete` / `Transcribe` 一样对 terminal `done` 的 partial provider meta 执行 canonical merge，确保 profile / capability / prompt / rubric / language / validation status 字段完整；normal done 与 cancellation partial done 都必须有 client-level test 证明。

### Phase 4: STT provider adapter

#### 4.1 `Transcribe` 接口

新增 `Transcribe(ctx, profile, audio) → (transcript, meta)`；`audio` 入参形态在 spec §4.1 锁定为内存字节 + filename + content type + 可选 language / prompt，production voice / practice voice owner 后续如需 object key / streaming audio，必须另行修订。

#### 4.2 openai_compatible `/v1/audio/transcriptions`

落地 audio transcription provider 适配，复用 plan 003 的 provider registry / capability profile / fail-fast / 隐私红线机制；`capability=stt` adapter 可执行，tracked voice profile 是否启用由 voice product owner 显式决定。

#### 4.3 metric / DB 字段

确认 7 个 ai_* metric family 的 label 集对 `capability=stt` 仍成立；如需新增 STT 专属 label，先在 spec §2.1 / [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 同步增量，再实现。

#### 4.4 realtime fail-closed 复核

`practice.voice.realtime.default` 继续保持 `status=unsupported`，除非 production voice / practice voice owner 已明确选择 realtime multimodal provider 并在 A3 spec、Product/UI release gate 与 profile catalog 中同步打开。只实现 STT 时不得顺手打开 realtime adapter。

### Phase 5: 接入 F1 / F3 / B1

#### 5.1 F1 metric / event 字段扩展

把本 plan 引入的新 `AICallMeta` 字段（tool_invocations / partial_meta_reason 等）同步到 [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) dashboard 与日志规范，保持「per-call 指标」与「event-only counter」语义不变。F1 AI metric label 使用 `capability`；任何 label 变更必须先由 F1 spec / plan 承接。

#### 5.2 F3 prompt schema 升级

[F3 prompt-rubric-registry](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) profile schema 新增 `tools[]` / `output_schema` / `stream_wire` 字段时，必须先增量 spec，再让本 plan 消费。

#### 5.3 B1 共享常量扩展

任何新增 `AI_*` 错误码、共享字段名、`AICallMeta` 字段必须先改 [B1 shared-conventions-codified](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec)；本 plan 严禁直接定义跨语言常量。

#### 5.4 L2 remediation: observability enrichment

修复 L2 发现的 F1 label gate 缺口：observability wrapper 在 stream done、stream error 与 pre-dispatch failure 记录前必须使用已注入的 `ProfileResolver` 对 meta 做 profile/capability/route enrichment，避免可解析 profile 的失败路径长期落成 `unknown` label；focused tests 必须覆盖成功 stream、stream error 与 invalid Complete failure labels。

### Phase 6: Verification

#### 6.1 spec §6 AC 增量

每个实现 phase 必须在 spec §6 验收标准表追加 ≥ 1 条 AC（覆盖正常路径 + 错误路径 + 隐私红线 + 观测埋点四类），并在本 plan 工作日志中给出 spec 版本号引用。

#### 6.2 单测 / 离线契约测试

stub provider 增加新能力的 deterministic 路径；离线 contract 测试覆盖 OpenAI-compatible 的 tool / streaming / audio transcription 协议子集。

#### 6.3 非测试本地 app 启动契约

阶段收口运行 provider/AIClient contract、observability/privacy、profile/config lint 与根 `make test`。真实 provider 调用可作为人工诊断，但不属于本 plan 完成条件，也不进入 `test/scenarios/e2e/`。

#### 6.4 Out-of-scope 输入负向搜索与 drift gate

激活后的收口必须执行 active-scope negative search，确认 A3-owned 代码、配置、deploy、generated artifacts、active docs 与本 plan 修订过的 owner docs 只使用 current capability keys、provider keys、单一 profile catalog、provider-ref routing 与当前模块命名。精确 out-of-scope literal 只允许出现在 denylist、rejection validator 或 negative fixture；历史 work journal / reports / bugs 作为只读证据不参与 runtime / contract surface 判定。

### Phase 7: Stream error contract test table consolidation

Replace the duplicate malformed-chunk and provider-error stream tests with one table-driven contract test. Keep named cases, exact SSE chunks, one terminal error event, `AI_OUTPUT_INVALID` for malformed JSON and `AI_PROVIDER_TIMEOUT` for provider error events.

## 5 验收标准

- Provider-neutral tools、streaming 与 STT adapter contracts 已完成；完成条件不包含真实 provider smoke。
- spec §6 AC 表已按 §6.1 同步追加；[history.md](../../history.md) 已记录版本递增。
- ADR-Q6 修订或新 ADR 已 `accepted`；plan 001 phase 范围未被改动。
- 零厂商 SDK 红线、隐私红线、fail-fast on missing `AI_PROVIDER_*` 三条全量保持。
- `context.yaml` 引用的 ADR / Product / F1 / F3 / B1 / A4 文档已随 `/implement` 加载；每个 checklist item 均有实际运行的 focused test、lint、startup contract 或 negative search 证据，阶段完成由根 `make test` 承接。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 后续 owner 把 002 误当成用户可见电话模式 owner | §3.1 锁定 002 只承接 Tools / provider streaming / OpenAI-compatible STT；provider-specific speech/TTS 归 004，电话模式 API/UI/媒体生命周期归 `practice-voice-mvp`，realtime profile 继续 fail-closed |
| Tool / function calling 在不同 provider 间语义差异破坏 provider-neutral | 接口形态在 spec §4.1 锁定为 OpenAI-compatible 的 tool 协议子集；不向上暴露 provider 特定字段 |
| Streaming partial meta 在 provider 不支持 token 增量时被错误填充 | Phase 3.2 明确：填 0 + `error_code` 标注；spec §4.1 已锁定该语义，本 plan 不放宽 |
| STT 启动时 audio payload 形态与 production voice / practice voice owner spec 不一致 | Phase 4.1 要求 spec §4.1 与对应 owner spec 联合锁定后再实现 |
| 新字段绕过 B1 / F1 / F3 引入跨域漂移 | Phase 5 显式编排接入顺序；本 plan 严禁直接落跨语言常量 |
| 002 文案与 A3 003 的 registry/profile catalog 当前合同漂移 | §3.1 锁定当前 declarative truth source；Phase 6.4 做 out-of-scope 输入负向搜索 |
| AI 底座提前落地后未被后续业务域正确消费 | Phase 5 / 6 要求 B1/F1/F3/A4 边界、profile coverage、negative search、startup/config contract 与根 `make test` 同步完成；用户流程由业务 owner 自身决定是否需要真实 API/UI E2E |

## 7 Activation governance

本 plan 已于 2026-05-06 根据用户明确确认提前激活并已完成。当前 owner 范围只包含 Complete tools payload、provider-side streaming SSE consumer、OpenAI-compatible STT Audio Transcriptions、metadata / observability / profile / registry / shared vocabulary gate。provider-specific speech/TTS 由 004 承接，媒体留存与电话模式 UI/API 由 `practice-voice-mvp` 承接；realtime multimodal 与 judge adapter 必须由各自 owner spec / plan 打开。

## 8 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-10 | 1.4 | Consolidate duplicate stream error contract tests into one table. | tech-debt pruning |
| 2026-07-10 | 1.3 | 对齐 spec 2.22 与 003/004 当前 owner 边界，删除 cost-cap 与 real-provider completion gate，并统一 out-of-scope 术语。 | tech-debt pruning |
| 2026-05-22 | 1.2 | Historical deployment-smoke exploration；不再作为当前完成条件。 | local-dev-stack contract sync |
| 2026-05-06 | 1.1 | L2 plan-code-review remediation：追加 stub tool replay、stream canonical meta 与 observability enrichment 三项修复 gate。 | plan-code-review --fix |
| 2026-05-06 | 1.0 | 用户确认方案 B 后提前激活：ADR-Q6 v2.0 与 A3 spec v2.5 打开 Tools、provider-side streaming 与 STT transcription 底座；realtime multimodal 继续 fail-closed。 | user activation |
| 2026-05-06 | 0.5 | L1 plan-review remediation：对齐当前 roadmap subject 命名、F1 out-of-scope label owner handoff、draft `/implement` 候选阻断与 active-scope negative search 范围。 | plan-review --fix |
| 2026-05-05 | 0.4 | L1 plan-review remediation：对齐 A3 spec v2.3 与 provider registry/profile catalog truth source，补 context references、声明式边界、realtime fail-closed gate、逐项验证断言与 out-of-scope 输入负向搜索 gate。 | plan-review --fix |
| 2026-05-05 | 0.3 | 对齐 A3 spec v1.9：STT 从范围外任务分类预留口径改为 `capability=stt` fail-closed profile，真正 adapter 仍需本 plan 激活后实现。 | provider-registry design |
| 2026-05-04 | 0.2 | L1 plan-review remediation：补齐 draft plan 的质量门禁分类，并重申 draft gate 不允许直接 implementation。 | docs-only L1 remediation |
