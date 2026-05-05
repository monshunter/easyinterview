# AI Tools, Streaming, and STT Extension

> **版本**: 0.4
> **状态**: draft
> **更新日期**: 2026-05-05

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ai-provider-and-model-routing spec](../../spec.md) §3.2 待确认事项与 §7 关联计划末段所列「按需决策」的三类扩展能力，落到一份**显式延期**的 plan 中作为占位。本 plan 仅在 [ADR-Q6 §5 失效与修订条件](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md#5-失效与修订条件) 任一触发条件成立、且对应 ADR / spec 已完成修订后，才被升级为 active 进入实施。覆盖范围：

- **Tools / function calling**：在 `AIClient` 上扩展 `Tools(ctx, profile, payload) → (response, meta)`（或在 `Complete` payload 上追加 `tools[]` 与结构化 output schema），保持 provider-neutral 红线不破。
- **Stream consumer 完整化**：基于 plan 001 已冻结的 `AIStreamEvent` 事件合同，落地 provider 侧 SSE / chunked streaming 消费 + `<-chan AIStreamEvent` 完整生命周期，并把 partial-token meta（context cancellation 场景）补齐；同时确定 HTTP wire（SSE 或 chunked）。
- **STT (`Transcribe(ctx, profile, audio) → (transcript, meta)`)**：新增 OpenAI-compatible `/v1/audio/transcriptions` provider adapter，把 spec §2.1 中保留的 `capability=stt` 从 fail-closed profile 升级为可执行 adapter，归 [C14 backend-voice-stt](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) P2。
- **成本上限 / 高级 rate-limit 策略**：当 provider endpoint 委派的 cost cap / per-tenant rate-limit 不足以满足业务时，评估在 AIClient 上加可选 middleware（仍不打破 fallback / retry 边界）。

本 plan 不修改 [001-aiclient-and-profile-bootstrap](../001-aiclient-and-profile-bootstrap/plan.md) 的 phase 范围；plan 001 锁定的 `AIStreamEvent` 类型保持稳定并被本 plan 复用。

## 2 背景

[spec §3.2 待确认事项](../../spec.md#32-待确认事项) 列出三项尚未锁定的边界：`Tools(...)` 是否扩展、`model_profile_version` 是否独立 SemVer、`Stream` HTTP wire 选型。其中 `Tools(...)` 与 stream wire 选型已在 [ADR-Q6 §5](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md#5-失效与修订条件) 明确为「触发即修订」。STT `Transcribe(...)` 在 spec §2.2 与 §3.1 D-9 中显式归 [C14 backend-voice-stt](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) P2；当前 A3 spec v2.3 仅允许 `capability=stt` / `capability=realtime` profile 以 `disabled` / `unsupported` 方式存在，运行时必须 fail-closed，直到本 plan 或 C14 / practice voice owner 联合修订后打开可执行 adapter。

为避免在 plan 001 实施边界内被业务侧临时需求驱动「就地塞入新接口」破坏 ADR-Q6 9 项硬约束，本 plan 提前把这些扩展项的入口、激活条件与治理流程固化为一份 `draft` 占位。每个 phase 必须在自己的 `激活条件` 满足后才被解冻，否则 checklist 项保持未勾选并不得开始编码。

## 3 质量门禁分类

- **Plan 类型**: `draft + code-internal + contract + deferred-extension`。本 plan 仅占位 A3 Tools / streaming / STT / cost cap 扩展治理，当前 draft 状态下不得进入实现；激活后会修改 AIClient contract、provider adapter、metadata、F1/F3/B1 接口边界与验证资产。
- **TDD 策略**: 当前不执行实现。只有 Phase 1 的触发证据、ADR-Q6 / spec 修订和 Header active 切换完成后，才能通过 `/implement` -> `/tdd` 顺序执行；届时每个 checklist item 必须以接口契约 tests、adapter mockserver tests、stream cancellation tests、STT contract tests、metric/log/privacy assertions 作为 Red-Green-Refactor 断言来源。
- **BDD 策略**: BDD 不适用。本 plan 打开的能力仍是内部 AI provider / provider protocol / observability 契约；任何把 streaming、STT 或 tool 调用暴露到用户可见 UI / API workflow 的后续 workstream 必须在自身 plan 维护 BDD gate。
- **替代验证 gate**: 激活前 gate 为 ADR/spec trigger evidence、context validation、Header/INDEX sync；激活后 gate 为 focused Go tests、OpenAI-compatible protocol contract tests、F1/F3/B1 compatibility checks、privacy grep、deployment smoke 与 `sync-doc-index --check`。

### 3.1 声明式边界锁定

- **配置真理源**: provider registry 继续以 `config/ai-providers.yaml` 为唯一 registry truth source，Model Profile 继续以单一 `config/ai-profiles.yaml` catalog 为唯一 profile truth source；不得恢复一 profile 一 YAML 目录、`task_type`、`default.provider` 或全局唯一 provider endpoint 口径。
- **运行时派生字段**: `AICallMeta`、fallback chain、stream partial meta、tool invocation summary、STT usage 统计和 observability labels 均由运行时调用派生，不得写回 registry / profile YAML。
- **跨边界字面量**: `capability`、provider/profile 字段名、AI meta 字段名和 `AI_*` 错误码由 B1 生成或登记；本 plan 不私造 Go/TS/OpenAPI 共享常量。
- **兼容策略**: 本 plan 激活后仍执行 clean-break 迁移；不新增旧 schema fallback、不新增 vendor SDK、不允许业务包直接引用 provider/model 字符串。

## 4 实施步骤

### Phase 1: 触发条件复核与 ADR / spec 修订

> **激活条件**: 任一下游触发被记录到工作日志或对应 spec / plan：
> - 后续业务域（典型为 [C5 backend-practice](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 的 followup planner）显式要求 tool-based 结构化抽取；
> - [D3 frontend-workspace-and-practice](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 或 [C5 backend-practice](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 要求实时 token streaming 进入产品 UX；
> - [C14 backend-voice-stt](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 进入 P2 实施；
> - provider endpoint 委派的 cost cap / rate-limit 出现 ≥ 2 次 P0 级别失效证据。

#### 1.1 触发证据归档

把触发来源（业务 spec id、plan id、事故记录或工作日志条目）汇总到本 plan 的工作日志条目中，明确说明哪一项触发条件成立、参考的上游文档版本号。

#### 1.2 ADR-Q6 修订

按 [ADR-Q6 §5](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md#5-失效与修订条件) 提示的修订流程，新增一份 ADR（或将 ADR-Q6 状态切到 `superseded`），显式记录被打开的能力（Tools / Stream / STT / cost cap）以及与原 9 项硬约束的兼容性论证；零厂商 SDK 入侵、`AIClient` 唯一对外能力、隐私红线必须保留。

#### 1.3 spec 版本递增

把 [ai-provider-and-model-routing spec](../../spec.md) 版本从当前基线递增到下一版本（当前基线为 `2.3` 时即 `2.4+`，若 A3 已继续递增则使用当时的下一版本），在 §2.1 / §3.1 / §4 / §6 中追加被打开能力对应的接口契约、决策项、设计约束与 AC。同步更新 [history.md](../../history.md)。

#### 1.4 plan 状态切换

完成 1.2 与 1.3 后，把本 plan Header 从 `状态: draft` 调整为 `状态: active` / `版本: 1.0`，并把 [plans/INDEX.md](../INDEX.md) 同步。未完成 1.1–1.3 的情况下严禁切换。

### Phase 2: Tools / function calling 实现

> **激活条件**: Phase 1 已收口，且触发记录中包含「后续业务域 tool 调用需求」。

#### 2.1 接口扩展

二选一并在 spec §4.1 锁定：(a) 在 `AIClient` 上新增 `Tools(ctx, profile, payload) → (response, meta)`；或 (b) 扩展 `Complete` payload，新增 `tools[]` 与 `output_schema` 字段。无论哪种形态，业务代码继续只引用 profile name，调用现场不感知 provider 工具协议差异。

#### 2.2 OpenAI-compatible 实现

在 `backend/internal/ai/aiclient/` 的 openai_compatible adapter 中追加 `tool_calls` / `tool_choice` 字段映射；stub provider 增加 deterministic tool-call 回放路径。零 vendor SDK 红线保持。

#### 2.3 观测与隐私

`AICallMeta` 新增 `tool_invocations[]`（或等价聚合字段，最终字段名以 spec 修订版为准）；log / DB metadata 仍只写 hash + 长度 + profile，不写 tool args 明文。

### Phase 3: Stream consumer 完整化

> **激活条件**: Phase 1 已收口，且触发记录中包含「D3 / C5 实时 token streaming UX 需求」。

#### 3.1 Provider-side streaming 消费

实现 openai_compatible adapter 的 SSE / chunked 解析，把 provider 增量映射为 plan 001 已冻结的 `delta` / `error` / `done` 事件，channel 生命周期与 spec §4.1 一致。

#### 3.2 取消语义与 partial meta

补齐 context cancellation 路径：截至中断时点的 `input_tokens` / `output_tokens` 必须尽力填充，provider 不支持时填 0；`error_code` 写入取消原因（B1 已锁定的 `AI_*` 错误码集合，不引入新字符串常量除非先改 B1）。

#### 3.3 HTTP wire 选型

按 spec §3.2 待确认事项最终决策结果，落地 SSE 或 chunked 实现，并把决策写回 spec §3.1 决策表。前端消费由 [D3 frontend-workspace-and-practice](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 自行 plan，不由本 plan 承担。

### Phase 4: STT provider adapter

> **激活条件**: Phase 1 已收口，且 [C14 backend-voice-stt](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 进入 P2 实施。

#### 4.1 `Transcribe` 接口

新增 `Transcribe(ctx, profile, audio) → (transcript, meta)`；`audio` 入参形态（multipart / object key / 字节流）需先在 spec §4.1 + C14 spec 联合锁定，本 plan 不预设。

#### 4.2 openai_compatible `/v1/audio/transcriptions`

落地 audio transcription provider 适配，复用 plan 003 的 provider registry / capability profile / fail-fast / 隐私红线机制；`capability=stt` 从 unsupported profile 升级为可执行。

#### 4.3 metric / DB 字段

确认 7 个 ai_* metric family 的 label 集对 `capability=stt` 仍成立；如需新增 STT 专属 label，先在 spec §2.1 / [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 同步增量，再实现。

#### 4.4 realtime fail-closed 复核

`practice.voice.realtime.default` 继续保持 `status=unsupported`，除非 C14 / practice voice owner 已明确选择 realtime multimodal provider 并在 A3 spec、Product/UI release gate 与 profile catalog 中同步打开。只实现 STT 时不得顺手打开 realtime adapter。

### Phase 5: 接入 F1 / F3 / B1

> **激活条件**: Phase 2 / 3 / 4 中至少一项已激活。

#### 5.1 F1 metric / event 字段扩展

把本 plan 引入的新 `AICallMeta` 字段（tool_invocations / partial_meta_reason 等）同步到 [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) dashboard 与日志规范，保持「per-call 指标」与「event-only counter」语义不变。

#### 5.2 F3 prompt schema 升级

[F3 prompt-rubric-registry](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) profile schema 新增 `tools[]` / `output_schema` / `stream_wire` 字段时，必须先增量 spec，再让本 plan 消费。

#### 5.3 B1 共享常量扩展

任何新增 `AI_*` 错误码、共享字段名、`AICallMeta` 字段必须先改 [B1 shared-conventions-codified](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec)；本 plan 严禁直接定义跨语言常量。

### Phase 6: Verification

> **激活条件**: 上述任一 phase 已激活。

#### 6.1 spec §6 AC 增量

被激活的每个 phase 必须在 spec §6 验收标准表追加 ≥ 1 条 AC（覆盖正常路径 + 错误路径 + 隐私红线 + 观测埋点四类），并在本 plan 工作日志中给出 spec 版本号引用。

#### 6.2 单测 / 离线契约测试

stub provider 增加新能力的 deterministic 路径；离线 contract 测试覆盖 OpenAI-compatible 的 tool / streaming / audio transcription 协议子集。

#### 6.3 部署 smoke

本地部署 + Kind 场景至少跑一次端到端 smoke：tool / stream / STT 各自串通业务 → AIClient → endpoint，确认观测埋点齐全且无明文泄漏。

#### 6.4 旧口径负向搜索与 drift gate

激活后的收口必须执行 active-scope negative search，确认代码、配置、deploy、generated artifacts 与 active docs 不含 `task_type`、`default.provider`、`config/ai-profiles/` 作为运行时 truth source、旧 AI gateway 术语、独立 `voice` route、独立 Mistakes / Growth / Drill 口径；历史 work journal / reports / bugs 可作为只读历史例外。

## 5 验收标准

- 本 plan 处于 `active` 时，所有被激活 phase 的 checklist 项全部勾选。
- spec §6 AC 表已按 §6.1 同步追加；[history.md](../../history.md) 已记录版本递增。
- ADR-Q6 修订或新 ADR 已 `accepted`；plan 001 phase 范围未被改动。
- 零厂商 SDK 红线、隐私红线、fail-fast on missing `AI_PROVIDER_*` 三条全量保持。
- `context.yaml` 引用的 ADR / Product / F1 / F3 / B1 / A4 文档已随 `/implement` 加载；每个 checklist item 均有实际运行的 focused test、lint、smoke 或 negative search 证据。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 在 trigger 未成立时被业务侧推动直接进入 /implement | 本 plan Header 锁 `状态: draft`；§6 治理章节明确禁止；进入需先按 Phase 1 完成 ADR / spec 修订 |
| Tool / function calling 在不同 provider 间语义差异破坏 provider-neutral | 接口形态在 spec §4.1 锁定为 OpenAI-compatible 的 tool 协议子集；不向上暴露 provider 特定字段 |
| Streaming partial meta 在 provider 不支持 token 增量时被错误填充 | Phase 3.2 明确：填 0 + `error_code` 标注；spec §4.1 已锁定该语义，本 plan 不放宽 |
| STT 启动时 audio payload 形态与 C14 spec 不一致 | Phase 4.1 要求 spec §4.1 与 C14 spec 联合锁定后再实现 |
| 新字段绕过 B1 / F1 / F3 引入跨域漂移 | Phase 5 显式编排接入顺序；本 plan 严禁直接落跨语言常量 |
| A3 003 后的 registry/profile catalog 口径被旧 002 文案覆盖 | §3.1 锁定当前 declarative truth source；Phase 6.4 做旧口径负向搜索 |
| draft plan 被 `/implement` 候选误当作可编码计划 | §7 继续要求先完成 trigger evidence + ADR/spec 修订；若质量门禁不完整，必须回到 `/plan-review --fix` |

## 7 Activation governance

本 plan 不允许直接进入 `/implement`；必须先按 Phase 1 完成 ADR-Q6 / spec 修订并把 Header 调整为 `状态: active` + `版本: 1.0`，再由 leader 显式触发。任何在 `draft` 状态下对 Phase 2 / 3 / 4 的代码改动均视为越权变更，必须回退并按本节流程重做。

## 8 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-05 | 0.4 | L1 plan-review remediation：对齐 A3 spec v2.3 与 provider registry/profile catalog truth source，补 context references、声明式边界、realtime fail-closed gate、逐项验证断言与旧口径负向搜索 gate。 | plan-review --fix |
| 2026-05-05 | 0.3 | 对齐 A3 spec v1.9：STT 从旧 `task_type` 预留口径改为 `capability=stt` fail-closed profile，真正 adapter 仍需本 plan 激活后实现。 | provider-registry design |
| 2026-05-04 | 0.2 | L1 plan-review remediation：补齐 draft plan 的质量门禁分类，并重申 draft gate 不允许直接 implementation。 | historical-spec-implementation-review/001 |
