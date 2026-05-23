# ADR-Q6 · AI Provider 与模型路由

> **版本**: 2.2
> **状态**: accepted
> **更新日期**: 2026-05-22

## 1 背景

`easyinterview` 当前开发期依赖 LLM / STT 占位 / judge 占位三类外部 AI 能力，覆盖：

- 同步：JD 解析提示词、模拟面试首题与追问（`practice` 域）
- 异步：报告生成（`review` 域）、简历定制（`resume` 域）、报告题目回顾 / 本轮复练上下文物化、debrief 生成
- P2：voice STT、source intel

`engineering-roadmap decisions` §2 把「AI Adapter Layer」标记为「模型供应商抽象、重试、fallback、成本记录」，§5 把 `ai` 模块拆为 `prompt / rubric registry + provider adapters + 调用记录`，§7 已规划 `ai_fallback_model_enabled` 等 feature flag；`F1 observability-stack` §「ai_*」指标与 §「fallback rate」dashboard 早已锁定。当前决策收敛为：**应用内 `AIClient` + Provider Registry + Capability-scoped Model Profile**；项目代码和配置只关心 AI provider 能力与连接，不把独立 provider-proxy 作为业务语义。

仓库现状：

- 没有任何业务代码 import 厂商 SDK
- `engineering-roadmap/spec.md` §5.1 A3 已重命名为 `ai-provider-and-model-routing`，明确 P0 必须交付 `provider-neutral AIClient` + `Provider Registry` + `Capability Model Profile` + OpenAI-compatible / stub provider
- 部署与测试形态由 ADR-Q4 锁定为 Docker Compose 外部依赖 + 宿主机 app runtime + repo-tracked 本地 scenario runner；K8s / Kind / Helm 不再是当前 P0 默认前提

业务约束：

- P0 当前开发期以 DeepSeek V4 Flash/Pro 作为主要 chat provider；STT、realtime、judge 保持 fail-closed 占位。单一全局 base URL / API key 不能作为长期契约
- 隐私（Q-5）要求所有 AI 调用记录可观测、可审计、可关闭
- 成本控制：每次调用必须记录 token / 美元 / 模型 / fallback 次数

## 2 选项与取舍

### 选项 A · 应用内 `AIClient` + Provider Registry + Capability Model Profile

**Pros**：

- 业务代码只依赖 1 个抽象 + profile name，**零厂商 SDK 入侵**
- 多模型 / 多 provider / fallback / token rate limit / cost cap 由 A3 provider registry、profile 与运维 secret 注入共同表达；代码不变
- 单元测试用 `stub` provider（hash-based 确定性输出）；非测试本地 app run、未来 staging / prod 部署必须配置真实 OpenAI-compatible provider 或 provider endpoint
- `AI_PROVIDER_REGISTRY_PATH` + provider-specific secret env ref 是 provider 配置入口；`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 只可作为默认 provider ref 引用的 env 名，不是全局唯一 contract
- F1 metric label（`provider / capability / model_family / model_profile_version / route`）由 AIClient 根据 registry/profile/provider response 统一补全
- 与 F3 `prompt-rubric-registry` 解耦：F3 只负责 prompt + rubric + profile name 的版本表，不涉及 provider

**Cons**：

- provider registry、profile 与 A4 secret env 字典必须保持同步
- central fallback / rate limit / cost cap 需要明确可观测和告警
- profile / registry 配置漂移需要严格 lint 与 ops 流程

### 选项 B · 业务代码直接 import 厂商 SDK（OpenAI Go / Anthropic Go）

**Pros**：

- 起步快

**Cons**：

- 锁死单家供应商；切厂商 = 改代码 + 重新发布
- fallback / rate limit / cost 全部业务实现，违反「provider-neutral」红线
- 无法在不上线情况下调整 model；与 P0 灰度策略冲突

### 选项 C · 业务代码 import provider proxy / router 专用 SDK

**Pros**：

- 借用代理层内部能力（如本地 plugin）

**Cons**：

- 把业务与某个 provider proxy 强耦合；切换 endpoint = 改代码
- 与 §3.2 Q-6 已确认方向「Higress 作为生产部署候选而非业务 SDK」直接冲突

### 选项 D · 自建 AI provider proxy（gRPC / HTTP）从零写

**Pros**：

- 完全自主

**Cons**：

- 把 provider proxy 自身变成业务，开发 + 运维成本巨大
- 与 ROI 不符；社区已有成熟方案（Higress / LiteLLM / Kong AI）

## 3 决策

**P0 锁定选项 A**。2026-05-06 经执行者确认，为避免后续 `backend-practice`、`backend-debrief`、production voice 与 F3 schema 接入时重复改造 AI 底座，原先在 §5 中列为“触发后评估”的 Tools / provider streaming / STT 能力提前纳入 A3 当前底座实施。该激活不推翻 provider-neutral 抽象，仍保持业务只依赖 `AIClient` + profile name，不引入厂商 SDK。

本 ADR 把 §3.2 Q-6 已确认方向固化为以下 9 项硬约束：

1. **AIClient 接口**（A3 owner）
   - 唯一对外能力：`Complete(ctx, profile, payload) → (response, meta)` / `Stream(ctx, profile, payload) → channel` / `Transcribe(ctx, profile, audio) → (transcript, meta)`
   - Tool / function calling 作为 `CompletePayload.tools[]`、`tool_choice`、`output_schema` 与 `CompleteResponse.tool_calls[]` 的 provider-neutral 扩展，不新增业务侧可绕过 profile 的 provider-specific 接口
   - `meta` 携带：`provider`、`model_family`、`model`、`prompt_version`、`rubric_version`、`model_profile_version`、`tokens_in/out`、`cost_usd`、`latency_ms`、`fallback_chain[]`、`route`
2. **Provider Registry + Model Profile**（A3 owner）
   - Provider Registry：`config/ai-providers.yaml` + 热加载；字段：`name` / `protocol` / `base_url_env` / `api_key_env` / `capabilities[]` / `version`；`stub` 可不声明 secret env ref，网络出站 provider 必须声明
   - Model Profile：YAML 文件 + 热加载；schema 在 A3 spec 中冻结
   - 字段：`name`（业务引用）/ `capability`（chat | stt | realtime | judge）/ `status`（active | disabled | unsupported）/ `unsupported_reason`（disabled / unsupported 时必填）/ `default.provider_ref+model+params` / `fallback[]`（provider-aware chain）/ `timeout_ms` / `max_tokens` / `rate_limit`（rps + tpm）/ `route`
   - 业务代码引用 `profile name`，不引用 provider / model 字符串
3. **运行时注入**：非单元测试运行环境通过 `AI_PROVIDER_REGISTRY_PATH` + `AI_MODEL_PROFILE_PATH` + registry 内 provider-specific secret env ref 注入；`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 只可作为默认 OpenAI-compatible provider ref 的 env 名。仓库不保留旧 provider-proxy 连接参数兼容层。
4. **Stub provider**（A3 owner）
   - 仅用于单元测试、离线 contract 测试或显式 mock 场景
   - 输入 → 输出 hash-based 确定性映射；可被 OpenAPI fixtures 反向喂养（与 E1 `mock-contract-suite` 同源）
   - 单元测试默认走 `stub`；非测试本地 app run、未来 staging / prod 不允许默认降级到 stub，缺少 provider registry、model profile path 或选中 provider 的 secret env ref 时必须 fail-fast
5. **Provider endpoint 边界**：本 ADR 不锁死供应商、托管形态或代理实现，只锁 provider registry / profile / OpenAI-compatible API 子集和应用侧 secret ref 连接参数。当前可执行 OpenAI-compatible 子集包含 Chat Completions、chat streaming SSE 与 Audio Transcriptions；realtime multimodal、judge 仍需各自 owner 后续递增 spec 后打开
6. **F3 解耦**：`prompt-rubric-registry` 只持有 `(feature_key, prompt_version, rubric_version, model_profile_name)` 四元组；不持有 provider / model 字符串
7. **可观测性**（F1 owner）
   - 每次 `AIClient.*` 调用必须落 `ai_task_runs_total` + `ai_task_latency_seconds` + `ai_task_input/output_tokens_total` + `ai_task_cost_usd_total` + `ai_output_validation_failures_total` + `ai_fallback_total`
   - dashboard：provider / capability / model 使用量 + fallback rate + cost / day + p95 latency
8. **隐私**（Q-5 关联）：AI 调用 payload 在 `audit_events` 写入 hash + 长度 + profile，**不写明文 prompt / response**；明文只允许保留在 `practice_session_events` 等业务表，受删除链路覆盖
9. **fallback 边界**：fallback 由 AIClient 在 profile fallback chain 内集中执行或合并 provider 返回的 fallback meta，最多 2 跳。业务代码看到的是「成功 response + fallback meta 标记」或「最终失败」；不允许业务自行重试切换模型

## 4 影响范围

- **A3 `ai-provider-and-model-routing`** —— 落地 `AIClient` + Provider Registry + Capability Model Profile schema + stub provider + OpenAI-compatible adapter
- **A4 `secrets-and-config`** —— `AI_PROVIDER_REGISTRY_PATH` / `AI_MODEL_PROFILE_PATH` / provider-specific secret env ref 配置项；非测试本地 app run 与未来部署必须能注入真实 provider 凭证
- **F1 `observability-stack`** —— `ai_*` 指标与 dashboard
- **F3 `prompt-rubric-registry`** —— 引用 `model_profile_name`；baseline prompt/rubric 与后续真实 model profile 切换
- **C4 `backend-targetjob`** / **C5 `backend-practice`** / **C6 `backend-review`** / **C7 `backend-resume`** / **C9 `backend-debrief`** —— 全部仅依赖 `AIClient` + profile name；禁止 import 厂商 SDK
- **C14 `backend-voice-stt`**（P2） —— STT / realtime 走同一 `AIClient` capability profile，profile 路由到 speech provider ref
- **`release-gate-and-rollout`** —— 校验 AI provider 路由可观测性 + fallback alert + cost cap 配置
- **B1 `shared-conventions-codified`** —— AI capability、Provider Registry / Model Profile / AI meta 字段名与 AI 错误码的共享常量或生成类型；A3 仍 owns Model Profile schema、`AIClient` runtime 与 `AICallMeta` 填充语义
- **CLAUDE.md / `test/scenarios/`** —— 当前场景默认使用 repo-tracked 本地 runner；只有离线 contract 测试可显式切 stub / mock provider ref，需要真实 AI provider 的非测试 app run / smoke 必须显式注入 provider registry/profile/secret 组合

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- provider ref 连接故障导致 ≥ 2 次 P0 事故 / 季 → 评估业务侧降级到显式只读/不可用状态或运维 provider ref 切换机制（仍走 `AIClient`，不打破抽象）
- 出现超出当前 OpenAI-compatible tool subset 的 provider-specific 高级特性 → 先递增 A3 / F3 / B1 owner spec，再评估是否扩展 provider-neutral payload；不得把厂商私有字段直接暴露给业务
- 多模型并行评估 / A/B（Q-3 PostHog feature flag 联动）需要按用户分桶 → 由 F3 Resolve + feature flag 选择最终 profile；不入侵业务调用现场
- provider ref 性能成为瓶颈（p95 > business SLA × 1.5）→ 评估更换 provider ref 或自建轻量 router（仍保持 `AIClient`）
- OpenAI-compatible API 不再是行业 lingua franca → 评估 provider adapter 升级；业务无感知
- realtime multimodal voice 进入 P0/P1 发布 → 新增或修订 ADR 明确音频留存、双向流、TTS、成本、隐私删除链路与 UI release gate；不得由当前 STT adapter 顺手打开

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q6-ai-provider-and-model-routing.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-6、§5.1 A3、§4.3 mock-first
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` checklist 1.1（保留 ADR-Q1..Q6 约束）与 checklist 3.3（production voice / personal knowledge search 等 future candidates 延后）
- 参考背景：`engineering-roadmap decisions` §2 §5 §「AI Adapter Layer」、`F1 observability-stack` §「ai_*」§「fallback rate」、`F1 observability-stack logging`
- 下游 child：A3 / A4 / F1 / F3 / C4-C7 / C9 / C11 / C14 / E4 / B1
- 关联 ADR：ADR-Q4-cloud-deploy-target（AI provider registry/profile/secret 注入）、ADR-Q5-privacy-cadence（AI 调用 payload 仅写 hash）

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-22 | 2.2 | 对齐 ADR-Q4 v1.7：当前 P0 测试/部署不再默认 Kind / K8s / Helm；AI fail-fast 边界改写为非测试本地 app run 与未来部署必须注入真实 provider registry/profile/secret，离线 contract 测试仍可显式 stub。 |
| 2026-05-08 | 2.1 | 对齐 A3 003 Phase 6：当前开发期删除向量化 / 重排能力与基础设施，DeepSeek V4 Flash/Pro 成为 repo-tracked chat provider 主力；未来资料检索需求需重新设计。 |
| 2026-05-06 | 2.0 | 提前激活 A3 002：在保持 provider-neutral / 零 SDK / 隐私红线的前提下，将 Tools payload 扩展、provider-side streaming consumer 与 STT Audio Transcriptions 纳入当前 AI 底座实施范围；realtime multimodal 仍 fail-closed。 |
| 2026-05-05 | 1.8 | 收口 active 部署与失效条件措辞：Kind / staging / prod 均以 provider registry/profile/secret 组合和 provider ref 为当前契约，不再用单一 endpoint 作为当前目标架构描述。 |
| 2026-05-05 | 1.7 | 同步 A3 003 Phase 4：运行时注入 fail-fast 口径改为 registry/profile path + 被选中 provider secret；B1 边界扩展为 AI capability、provider/profile 字段名与 provider/profile routing 错误码。 |
| 2026-05-05 | 1.6 | 基于 product-scope 与 UI AI 场景重估，将目标架构从单一 provider endpoint 升级为 Provider Registry + Capability Model Profile；fallback 改由 AIClient 在 profile chain 内集中执行，A4 入口新增 `AI_PROVIDER_REGISTRY_PATH`。 |
| 2026-05-05 | 1.5 | 全面更名并收口 AI provider 口径：A3 目录与 ADR 文件改为 `ai-provider-and-model-routing`，运行时连接参数改为 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY`，并确认不保留旧连接参数或 route schema 兼容层。 |
| 2026-05-04 | 1.4 | 对齐 engineering-roadmap v3.0：关联章节不再指向旧 Phase 5.2，改为引用当前 ADR 保留与 future candidate 延后规则。 |
| 2026-05-03 | 1.3 | 对齐 product-scope v1.1：review 域 AI 物化只服务报告题目回顾和本轮复练上下文，不恢复独立错题本 / Drill。 |
| 2026-04-29 | 1.2 | 收口 A/B spec 全面审查 remediation：明确 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 是 AIClient 的 OpenAI-compatible 连接参数，可指真实 LLM provider 或生产 provider endpoint；fallback 由连接 provider endpoint route 承担，A3 client 不自行切换模型；B1 只提供共享字段/常量，A3 owns runtime。 |
| 2026-04-27 | 1.1 | 明确 stub 只用于单元测试 / 离线 contract 测试；docker compose 与 Kind 本地部署必须使用真实 AI provider 提供的 OpenAI-compatible LLM 服务，不默认降级到 stub，也不要求本地部署 AI provider 组件。 |
