# ADR-Q6 · AI 网关与模型路由

> **版本**: 1.3
> **状态**: accepted
> **更新日期**: 2026-05-03

## 1 背景

`easyinterview` 全链路依赖 LLM / embedding / STT 三类外部 AI 能力，覆盖：

- 同步：JD 解析提示词、模拟面试首题与追问（`practice` 域）
- 异步：报告生成（`review` 域）、简历定制（`resume` 域）、报告题目回顾 / 本轮复练上下文物化、debrief 生成、retrieval 召回
- P2：voice STT、source intel

`easyinterview-tech-docs/01-technical-architecture.md` §2 把「AI Adapter Layer」标记为「模型供应商抽象、重试、fallback、成本记录」，§5 把 `ai` 模块拆为 `prompt / rubric registry + provider adapters + 调用记录`，§7 已规划 `ai_fallback_model_enabled` 等 feature flag；`04-metrics-observability.md` §「ai_*」指标与 §「fallback rate」dashboard 早已锁定。`engineering-roadmap/spec.md` §3.2 Q-6 已确认总体方向：**应用内 `AIClient` + Model Profile，生产经外部 AI Gateway**；本 ADR 把 W0 hard gate 落到具体边界。

仓库现状：

- 没有任何业务代码 import 厂商 SDK
- `engineering-roadmap/spec.md` §5.1 A3 已重命名为 `ai-gateway-and-model-routing`，明确 P0 必须交付 `provider-neutral AIClient` + `Model Profile` + `OpenAI-compatible provider/gateway route` + 单元测试 `stub`
- 部署形态由 ADR-Q4 锁定为 K8s

业务约束：

- P0 单一供应商 OpenAI-compatible（含 OpenAI / Azure OpenAI / DeepSeek / Together / 国内厂商兼容路由）即可覆盖；多厂商 fallback 留给运维配置，不绑代码
- 隐私（Q-5）要求所有 AI 调用记录可观测、可审计、可关闭
- 成本控制：每次调用必须记录 token / 美元 / 模型 / fallback 次数

## 2 选项与取舍

### 选项 A · 应用内 `AIClient` + Model Profile + 外部 AI Gateway（Higress 等）

**Pros**：

- 业务代码只依赖 1 个抽象 + profile name，**零厂商 SDK 入侵**
- 多模型 / 多 provider / fallback / token rate limit / cost cap 全部归运维通过 gateway 配置；代码不变
- 单元测试用 `stub` provider（hash-based 确定性输出）；本地 docker-compose / Kind / staging / prod 部署必须配置真实 OpenAI-compatible provider 或 gateway endpoint
- `AI_GATEWAY_BASE_URL` 是 OpenAI-compatible endpoint 的唯一切换点；本地可指真实 AI provider，生产可指 ADR-Q4 K8s cluster-internal gateway
- F1 metric label（`provider / model_family / model_profile_version / route`）由 gateway 透传 + 客户端补全
- 与 F3 `prompt-rubric-registry` 解耦：F3 只负责 prompt + rubric + profile name 的版本表，不涉及 provider

**Cons**：

- 多一层 gateway 部署 + 监控
- gateway 故障即全员故障（需 HA + 双副本）
- profile 配置漂移需要严格的 ops 流程

### 选项 B · 业务代码直接 import 厂商 SDK（OpenAI Go / Anthropic Go）

**Pros**：

- 起步快

**Cons**：

- 锁死单家供应商；切厂商 = 改代码 + 重新发布
- fallback / rate limit / cost 全部业务实现，违反「provider-neutral」红线
- 无法在不上线情况下调整 model；与 P0 灰度策略冲突

### 选项 C · 业务代码 import Higress SDK 或其它 gateway-specific SDK

**Pros**：

- 借用 gateway 内部能力（如本地 plugin）

**Cons**：

- 把业务与 gateway 强耦合；切 gateway = 改代码
- 与 §3.2 Q-6 已确认方向「Higress 作为生产部署候选而非业务 SDK」直接冲突

### 选项 D · 自建 gateway（gRPC / HTTP）从零写

**Pros**：

- 完全自主

**Cons**：

- 把 gateway 自身变成业务，开发 + 运维成本巨大
- 与 ROI 不符；社区已有成熟方案（Higress / LiteLLM / Kong AI）

## 3 决策

**P0 锁定选项 A**。本 ADR 把 §3.2 Q-6 已确认方向固化为以下 9 项硬约束：

1. **AIClient 接口**（A3 owner）
   - 唯一对外能力：`Complete(ctx, profile, payload) → (response, meta)` / `Embed(ctx, profile, input) → (vector, meta)` / `Stream(ctx, profile, payload) → channel`
   - `meta` 携带：`provider`、`model_family`、`model`、`prompt_version`、`rubric_version`、`model_profile_version`、`tokens_in/out`、`cost_usd`、`latency_ms`、`fallback_chain[]`、`route`
2. **Model Profile**（A3 owner）
   - YAML 文件 + 热加载；schema 在 A3 spec 中冻结
   - 字段：`name`（业务引用）/ `task_type`（chat | embed | stt）/ `default.provider+model+params` / `fallback[]`（按序触发条件）/ `timeout_ms` / `max_tokens` / `rate_limit`（rps + tpm）/ `gateway_route`
   - 业务代码引用 `profile name`，不引用 provider / model 字符串
3. **运行时注入**：非单元测试运行环境唯一注入点 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY`（保留 `GATEWAY` 命名作为兼容字段名，语义是 AIClient 的 OpenAI-compatible 连接参数；可指真实 LLM provider，也可指生产 gateway）；所有 `AIClient.*` 经此 URL 出站
4. **Stub provider**（A3 owner）
   - 仅用于单元测试、离线 contract 测试或显式 mock 场景
   - 输入 → 输出 hash-based 确定性映射；可被 OpenAPI fixtures 反向喂养（与 E1 `mock-contract-suite` 同源）
   - 单元测试默认走 `stub`；docker compose / Kind / staging / prod 不允许默认降级到 stub，缺少真实 provider endpoint / API key 时必须 fail-fast
5. **gateway 选型**：生产推荐 Higress（与 ADR-Q4 K8s cluster-internal Deployment 一致）；LiteLLM / Kong AI plugin 为可选；本 ADR 不锁死 gateway 实现，只锁 OpenAI-compatible API 契约
6. **F3 解耦**：`prompt-rubric-registry` 只持有 `(feature_key, prompt_version, rubric_version, model_profile_name)` 四元组；不持有 provider / model 字符串
7. **可观测性**（F1 owner）
   - 每次 `AIClient.*` 调用必须落 `ai_task_runs_total` + `ai_task_latency_seconds` + `ai_task_input/output_tokens_total` + `ai_task_cost_usd_total` + `ai_output_validation_failures_total` + `ai_fallback_total`
   - dashboard：provider / model 使用量 + fallback rate + cost / day + p95 latency / task_type
8. **隐私**（Q-5 关联）：AI 调用 payload 在 `audit_events` 写入 hash + 长度 + profile，**不写明文 prompt / response**；明文只允许保留在 `practice_session_events` 等业务表，受删除链路覆盖
9. **fallback 边界**：fallback 只在 AIClient 连接的 OpenAI-compatible endpoint / gateway route 层触发；如果该 endpoint 是真实 LLM provider 且不提供 fallback，A3 client 不自行切换模型。业务代码看到的是「成功 response + fallback meta 标记」或「最终失败」；不允许业务自行重试切换模型

## 4 影响范围

- **A3 `ai-gateway-and-model-routing`** —— 落地 `AIClient` + Model Profile schema + stub provider + OpenAI-compatible adapter
- **A4 `secrets-and-config`** —— `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` / `AI_MODEL_PROFILE_PATH` 配置项；local deploy 与 Kind 必须能注入真实 provider 凭证
- **F1 `observability-stack`** —— `ai_*` 指标与 dashboard
- **F3 `prompt-rubric-registry`** —— 引用 `model_profile_name`；W1 baseline + W3 真实 model profile 切换
- **C4 `backend-targetjob`** / **C5 `backend-practice`** / **C6 `backend-review`** / **C7 `backend-resume`** / **C9 `backend-debrief`** / **C11 `backend-retrieval`** —— 全部仅依赖 `AIClient` + profile name；禁止 import 厂商 SDK
- **C14 `backend-voice-stt`**（P2） —— STT 走同一 `AIClient`（task_type=stt），profile 路由到 STT 专用 gateway route
- **E4 `release-gate-and-rollout`** —— W4 gate 校验 AI Gateway 路由可观测性 + fallback alert + cost cap 配置
- **B1 `shared-conventions-codified`** —— Model Profile / AI meta 字段名与 AI 错误码的共享常量或生成类型；A3 仍 owns Model Profile schema、`AIClient` runtime 与 `AICallMeta` 填充语义
- **CLAUDE.md / `test/scenarios/`** —— Kind 场景默认使用真实 AI provider endpoint；只有离线 contract 测试可显式切 stub / mock gateway

## 5 失效与修订条件

触发推翻或升级本 ADR 的具体阈值：

- gateway 单点故障导致 ≥ 2 次 P0 事故 / 季 → 评估业务侧降级到 stub 的自动 circuit breaker（仍走 `AIClient`，不打破抽象）
- 出现需要业务感知 model 内部状态的高级特性（function calling / tool use 跨厂商差异极大）→ 评估在 `AIClient` 上扩展 `Tools(...)` 接口；不打破 provider-neutral
- 多模型并行评估 / A/B（Q-3 PostHog feature flag 联动）需要按用户分桶 → 在 Profile 上加 `routing_rules` 字段，不入侵业务
- gateway 性能成为瓶颈（p95 > business SLA × 1.5）→ 评估直连 + 自建轻量 router（仍保持 `AIClient`）
- OpenAI-compatible API 不再是行业 lingua franca → 评估 gateway 适配层升级；业务无感知

修订流程：本 ADR 状态由 `accepted` → `superseded`，新 ADR 显式标注 `supersedes: ADR-Q6-ai-gateway-and-model-routing.md`。

## 6 关联

- `engineering-roadmap/spec.md` §3.2 Q-6、§5.1 A3、§4.3 mock-first
- `engineering-roadmap/plans/001-decompose-subspecs/plan.md` Phase 1.1、Phase 5.2（F3 切真 Model Profile）
- 上游：`easyinterview-tech-docs/01-technical-architecture.md` §2 §5 §「AI Adapter Layer」、`04-metrics-observability.md` §「ai_*」§「fallback rate」、`05-logging-standard.md`
- 下游 child：A3 / A4 / F1 / F3 / C4-C7 / C9 / C11 / C14 / E4 / B1
- 关联 ADR：ADR-Q4-cloud-deploy-target（gateway 作为 cluster-internal Deployment）、ADR-Q5-privacy-cadence（AI 调用 payload 仅写 hash）

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-05-03 | 1.3 | 对齐 product-scope v1.1：review 域 AI 物化只服务报告题目回顾和本轮复练上下文，不恢复独立错题本 / Drill。 |
| 2026-04-29 | 1.2 | 收口 A/B spec 全面审查 remediation：明确 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 是 AIClient 的 OpenAI-compatible 连接参数，可指真实 LLM provider 或生产 gateway；fallback 由连接 endpoint / gateway route 承担，A3 client 不自行切换模型；B1 只提供共享字段/常量，A3 owns runtime。 |
| 2026-04-27 | 1.1 | 明确 stub 只用于单元测试 / 离线 contract 测试；docker compose 与 Kind 本地部署必须使用真实 AI provider 提供的 OpenAI-compatible LLM 服务，不默认降级到 stub，也不要求本地部署 AI gateway 组件。 |
