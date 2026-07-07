# AI Tools, Streaming, and STT Extension

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-22

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ai-provider-and-model-routing spec](../../spec.md) §3.2 待确认事项与 §7 关联计划末段所列「按需决策」的三类扩展能力提前升级为当前 AI 底座。2026-05-06 用户明确选择“方案 B：AI 提前落地，避免后续业务开发返工”，因此本 plan 从 draft 激活为 active，并按 [ADR-Q6 v2.0](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 与 A3 spec v2.5 进入实施。覆盖范围：

- **Tools / function calling**：在 `Complete` payload 上追加 provider-neutral `tools[]` / `tool_choice` / `output_schema`，并在 response/meta 中返回 tool-call 摘要，保持 provider-neutral 红线不破。
- **Stream consumer 完整化**：基于 plan 001 已冻结的 `AIStreamEvent` 事件合同，落地 provider 侧 SSE / chunked streaming 消费 + `<-chan AIStreamEvent` 完整生命周期，并把 partial-token meta（context cancellation 场景）补齐；同时确定 HTTP wire（SSE 或 chunked）。
- **STT (`Transcribe(ctx, profile, audio) → (transcript, meta)`)**：新增 OpenAI-compatible `/v1/audio/transcriptions` provider adapter，把 spec §2.1 中保留的 `capability=stt` 从 fail-closed profile 升级为可执行 adapter；production voice / practice voice owner 仍负责 realtime multimodal、TTS、媒体留存、UI release gate 与隐私删除链路。
- **成本上限 / 高级 rate-limit 策略**：当 provider endpoint 委派的 cost cap / per-tenant rate-limit 不足以满足业务时，评估在 AIClient 上加可选 middleware（仍不打破 fallback / retry 边界）。

本 plan 不修改 [001-aiclient-and-profile-bootstrap](../001-aiclient-and-profile-bootstrap/plan.md) 的 phase 范围；plan 001 锁定的 `AIStreamEvent` 类型保持稳定并被本 plan 复用。

## 2 背景

[spec §3.2 待确认事项](../../spec.md#32-待确认事项) 曾列出三项尚未锁定的边界：`Tools(...)` 是否扩展、`model_profile_version` 是否独立 SemVer、`Stream` HTTP wire 选型。其中 Tools 与 provider-side streaming 已在 [ADR-Q6 v2.0](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 中提前打开为底座能力；HTTP 业务 wire 仍由后续 backend/frontend owner 决定。STT `Transcribe(...)` 当前由 A3 先提供 OpenAI-compatible Audio Transcriptions adapter，production voice / practice voice owner 后续只需消费 `capability=stt` profile，不必再次重做底层 provider adapter。

为避免后续业务域临时需求驱动「就地塞入新接口」破坏 ADR-Q6 9 项硬约束，本 plan 统一落地底座能力：Complete tools payload、provider-side streaming consumer、STT transcription adapter、B1/F1/F3/A4 边界同步与 drift gates。Realtime multimodal 仍显式 fail-closed，不在本 plan 顺手打开。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + provider-adapter + observability/privacy`。本 plan 修改 AIClient contract、provider adapter、metadata、profile/registry catalog、B1/F1/F3/A4 接口边界与验证资产。
- **TDD 策略**: 通过 `/implement` -> `/tdd` 顺序执行。每个 checklist item 必须以接口契约 tests、adapter mockserver tests、stream cancellation tests、STT contract tests、metric/log/privacy assertions、profile/registry lint 和 negative search 作为 Red-Green-Refactor 断言来源。
- **BDD 策略**: BDD 不适用。本 plan 打开的能力仍是内部 AI provider / provider protocol / observability 契约；任何把 streaming、STT 或 tool 调用暴露到用户可见 UI / API workflow 的后续 workstream 必须在自身 plan 维护 BDD gate。
- **替代验证 gate**: focused Go tests、OpenAI-compatible protocol contract tests、F1/F3/B1 compatibility checks、profile coverage lint、privacy grep、deployment smoke、active-scope negative search 与 `sync-doc-index --check`。

### 3.1 声明式边界锁定

- **配置真理源**: provider registry 继续以 `config/ai-providers.yaml` 为唯一 registry truth source，Model Profile 继续以单一 catalog 为唯一 profile truth source；不得恢复一 profile 一 YAML 目录、非当前任务分类 key、非当前 provider key 或全局唯一 provider endpoint 口径。
- **运行时派生字段**: `AICallMeta`、fallback chain、stream partial meta、tool invocation summary、STT usage 统计和 observability labels 均由运行时调用派生，不得写回 registry / profile YAML。
- **跨边界字面量**: `capability`、provider/profile 字段名、AI meta 字段名和 `AI_*` 错误码由 B1 生成或登记；本 plan 不私造 Go/TS/OpenAPI 共享常量。
- **兼容策略**: 本 plan 激活后仍执行 clean-break 迁移；不新增旧 schema fallback、不新增 vendor SDK、不允许业务包直接引用 provider/model 字符串。
- **提前激活边界**: Header `状态: active` 只打开 Tools / provider-side streaming / STT transcription 底座；realtime multimodal、TTS、媒体留存、UI release gate、judge 仍保持 fail-closed 或归各 owner 后续 plan。

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

修复 L2 发现的 `Stream` terminal meta 漂移：`AIClient.Stream` 必须像 `Complete` / `Embed` / `Transcribe` 一样对 terminal `done` 的 partial provider meta 执行 canonical merge，确保 profile / capability / prompt / rubric / language / validation status 字段完整；normal done 与 cancellation partial done 都必须有 client-level test 证明。

### Phase 4: STT provider adapter

#### 4.1 `Transcribe` 接口

新增 `Transcribe(ctx, profile, audio) → (transcript, meta)`；`audio` 入参形态在 spec §4.1 锁定为内存字节 + filename + content type + 可选 language / prompt，production voice / practice voice owner 后续如需 object key / streaming audio，必须另行修订。

#### 4.2 openai_compatible `/v1/audio/transcriptions`

落地 audio transcription provider 适配，复用 plan 003 的 provider registry / capability profile / fail-fast / 隐私红线机制；`capability=stt` 从 unsupported profile 升级为可执行。

#### 4.3 metric / DB 字段

确认 7 个 ai_* metric family 的 label 集对 `capability=stt` 仍成立；如需新增 STT 专属 label，先在 spec §2.1 / [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 同步增量，再实现。

#### 4.4 realtime fail-closed 复核

`practice.voice.realtime.default` 继续保持 `status=unsupported`，除非 production voice / practice voice owner 已明确选择 realtime multimodal provider 并在 A3 spec、Product/UI release gate 与 profile catalog 中同步打开。只实现 STT 时不得顺手打开 realtime adapter。

### Phase 5: 接入 F1 / F3 / B1

#### 5.1 F1 metric / event 字段扩展

把本 plan 引入的新 `AICallMeta` 字段（tool_invocations / partial_meta_reason 等）同步到 [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) dashboard 与日志规范，保持「per-call 指标」与「event-only counter」语义不变。若 F1 active spec 仍使用非当前任务分类 label 口径，本 plan 必须在本 phase 内原地修订 F1 spec，明确 AI metric label 迁移到 `capability`，不得把非当前口径计入 pass。

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

#### 6.3 非测试本地 app smoke

非测试本地 app run 或 repo-tracked scenario runner 至少跑一次端到端 smoke：tool / stream / STT 各自串通业务 → AIClient → endpoint，确认观测埋点齐全且无明文泄漏。默认不要求 Kind / K8s / Helm；若未来 release owner 引入部署级 smoke，必须先修订对应环境契约。

#### 6.4 非当前输入负向搜索与 drift gate

激活后的收口必须执行 active-scope negative search，确认 A3-owned 代码、配置、deploy、generated artifacts、active docs 与被本 plan 激活并修订过的 owner docs 不再把非当前任务分类 key、非当前 provider key、一 profile 一目录 truth source、non-current AI routing 术语、独立语音路由、独立旧命名模块口径作为 active runtime truth source；只有 denylist / rejection validator / negative fixture 可以保留精确非当前 literal。历史 work journal / reports / bugs 可作为只读历史例外。对仍由其他 owner 保持的 referenced active spec（例如 F1 在完成迁移前的非当前任务分类 label），必须回到 Phase 5.1 owner handoff 后再计入 pass。

## 5 验收标准

- 本 plan 处于 `active` 时，所有被激活 phase 的 checklist 项全部勾选。
- spec §6 AC 表已按 §6.1 同步追加；[history.md](../../history.md) 已记录版本递增。
- ADR-Q6 修订或新 ADR 已 `accepted`；plan 001 phase 范围未被改动。
- 零厂商 SDK 红线、隐私红线、fail-fast on missing `AI_PROVIDER_*` 三条全量保持。
- `context.yaml` 引用的 ADR / Product / F1 / F3 / B1 / A4 文档已随 `/implement` 加载；每个 checklist item 均有实际运行的 focused test、lint、smoke 或 negative search 证据。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 提前激活导致后续 owner 误以为 realtime / TTS / UI voice 已可用 | §3.1 与 Phase 4.4 锁定只打开 STT；realtime profile 继续 fail-closed，production voice / practice voice owner 仍需自身 release gate |
| Tool / function calling 在不同 provider 间语义差异破坏 provider-neutral | 接口形态在 spec §4.1 锁定为 OpenAI-compatible 的 tool 协议子集；不向上暴露 provider 特定字段 |
| Streaming partial meta 在 provider 不支持 token 增量时被错误填充 | Phase 3.2 明确：填 0 + `error_code` 标注；spec §4.1 已锁定该语义，本 plan 不放宽 |
| STT 启动时 audio payload 形态与 production voice / practice voice owner spec 不一致 | Phase 4.1 要求 spec §4.1 与对应 owner spec 联合锁定后再实现 |
| 新字段绕过 B1 / F1 / F3 引入跨域漂移 | Phase 5 显式编排接入顺序；本 plan 严禁直接落跨语言常量 |
| A3 003 后的 registry/profile catalog 口径被旧 002 文案覆盖 | §3.1 锁定当前 declarative truth source；Phase 6.4 做非当前输入负向搜索 |
| AI 底座提前落地后未被后续业务域正确消费 | Phase 5 / 6 要求 B1/F1/F3/A4 边界、profile coverage、negative search 和 deployment smoke 同步完成；业务域仍必须在自身 plan 维护 BDD / API gate |

## 7 Activation governance

本 plan 已于 2026-05-06 根据用户明确确认提前激活。激活范围只包含 Complete tools payload、provider-side streaming SSE consumer、STT Audio Transcriptions、metadata / observability / profile / registry / shared vocabulary gate；任何 realtime multimodal、TTS、媒体留存、UI voice release 或 judge adapter 改动仍视为越权变更，必须另行修订对应 owner spec / plan。

## 8 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-06 | 1.1 | L2 plan-code-review remediation：追加 stub tool replay、stream canonical meta 与 observability enrichment 三项修复 gate。 | plan-code-review --fix |
| 2026-05-06 | 1.0 | 用户确认方案 B 后提前激活：ADR-Q6 v2.0 与 A3 spec v2.5 打开 Tools、provider-side streaming 与 STT transcription 底座；realtime multimodal 继续 fail-closed。 | user activation |
| 2026-05-06 | 0.5 | L1 plan-review remediation：对齐当前 roadmap subject 命名、F1 非当前任务分类 label owner handoff、draft `/implement` 候选阻断与 active-scope negative search 范围。 | plan-review --fix |
| 2026-05-05 | 0.4 | L1 plan-review remediation：对齐 A3 spec v2.3 与 provider registry/profile catalog truth source，补 context references、声明式边界、realtime fail-closed gate、逐项验证断言与非当前输入负向搜索 gate。 | plan-review --fix |
| 2026-05-05 | 0.3 | 对齐 A3 spec v1.9：STT 从非当前任务分类预留口径改为 `capability=stt` fail-closed profile，真正 adapter 仍需本 plan 激活后实现。 | provider-registry design |
| 2026-05-04 | 0.2 | L1 plan-review remediation：补齐 draft plan 的质量门禁分类，并重申 draft gate 不允许直接 implementation。 | docs-only L1 remediation |
