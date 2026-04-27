# AI Gateway and Model Routing Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 把 A3 `ai-gateway-and-model-routing` 列为 Layer A · Foundation 的第三份 child（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)；间接依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 提供的共享类型）。它把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) 的 9 项硬约束落到代码层，决定了：

- 业务代码（C4–C7 / C9 / C11 / C14）调用 LLM / embedding / STT 的唯一接口形态；
- prompt / rubric / model profile 在调用现场如何串起来；
- 本地 dev / 单元测试 / staging / prod 4 类环境如何切换 provider。

目标是：

1. **Provider-neutral 抽象**：业务代码 0 厂商 SDK 入侵，只依赖 `AIClient` 接口与 `Model Profile` name；切换厂商或加 fallback 不改业务代码。
2. **可观测可计费**：每一次 `AIClient.*` 调用必须产出 `meta`（provider / model_family / model / prompt_version / rubric_version / model_profile_version / tokens / cost / latency / fallback_chain / route），并由 [F1 `observability-stack`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份) 统一接入 metric / log / DB（`ai_task_runs`）。
3. **可测试可灰度**：`stub` provider 提供 hash-based 确定性输出，单元测试默认走 stub；CI / staging 可指 mock gateway；生产指 OpenAI-compatible AI Gateway（Higress / LiteLLM / Kong AI 任一）。
4. **隐私守约**：AI 调用 payload 在 `audit_events` 写 hash + 长度 + profile，不写明文 prompt / response（与 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) 对齐）。

本 spec 不定义具体 prompt（归 [F3 `prompt-rubric-registry`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份)）、不定义业务调用现场（归各 C 域）、不部署 gateway（运维 / E4 承接）。

## 2 范围

### 2.1 In Scope

- **AIClient 接口**：Go 包 `backend/internal/ai/aiclient/`，唯一对外能力 `Complete(ctx, profile, payload) → (response, meta)` / `Embed(ctx, profile, input) → (vector, meta)` / `Stream(ctx, profile, payload) → channel`；`meta` 字段固化为 `AICallMeta` 结构体，由 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 共享类型支撑。
- **Model Profile schema**：YAML 文件 + 热加载；schema 在本 spec 冻结。字段：`name` / `task_type`（`chat` | `embed` | `stt`）/ `default.{provider, model, params}` / `fallback[]`（按序触发条件）/ `timeout_ms` / `max_tokens` / `rate_limit.{rps, tpm}` / `gateway_route`。Profile 文件落点 `config/ai-profiles/*.yaml`（A4 控制 `AI_MODEL_PROFILE_PATH` 指向）。
- **Provider 实现集**：
  - `stub`：hash-based 确定性输出，从 OpenAPI fixtures 反向喂养（与 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#55-layer-e--integration4-份) 同源）。
  - `openai_compatible`：通过 `AI_GATEWAY_BASE_URL` 出站，仅依赖 OpenAI Chat Completions / Embeddings / Audio Transcription 协议子集；不直接 import 任何厂商 SDK。
- **路由策略**：profile name → provider 选择 → model 选择 → fallback 链；fallback 只在 gateway 层触发，业务看到「成功 + fallback meta」或「最终失败」，不允许业务自行重试切换模型。
- **观测埋点契约**：每次调用上报 `ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total` 共 7 个 metric；同时落 DB 表 `ai_task_runs`（schema 由 [B4](../engineering-roadmap/spec.md#52-layer-b--contract4-份全部-p0) 落地，与 [03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md) 一致）。
- **Audit hook**：每次调用产出 `audit_events` 行（`action=ai.call`），`metadata` 字段含 `prompt_hash` / `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`；不含明文。

### 2.2 Out of Scope

- 具体 prompt 内容、rubric schema、版本表：归 [F3 `prompt-rubric-registry`](../engineering-roadmap/spec.md#56-layer-f--quality-横切4-份)。
- 业务调用现场（哪个 C 域调用 `Complete` 还是 `Embed`）：归各自 C 域 spec / plan。
- 真实 gateway 部署（Higress / LiteLLM Helm chart）、路由配置、cost cap、rate limit 规则：归运维 + E4；本 spec 仅锁 OpenAI-compatible API 契约。
- Token 计费成本表：本 spec 把 `cost_usd_micros` 字段定义清楚，具体 provider × model × pricing 由 F3 / F1 维护。
- LLM Judge / 离线评估集：归 F3。
- DB 表本身：归 B4；本 spec 只引用字段名。
- 错误码命名：依赖 B1 已落地的 `AI_*` 前缀错误码（`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` 等），新增错误码必须先改 B1。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | AIClient 接口形态 | `Complete` / `Embed` / `Stream` 三方法；`payload` 与 `response` 为结构化对象（不直接传 string）；`meta` 由 client 返回，业务不能伪造 | 业务代码绝对零厂商 SDK 入侵 |
| D-2 | Model Profile 字段集 | 见 §2.1；新增字段必须递增 spec 版本 | gateway 配置漂移可控 |
| D-3 | 业务引用形态 | 业务只引用 `profile name`（如 `practice.followup.default` / `review.report.default`），不引用 provider / model 字符串 | 切换 provider / model = 改 profile YAML，不改代码 |
| D-4 | Stub 触发条件 | 单元测试默认走 stub；切真模型需在测试 config 显式设置 `AI_GATEWAY_BASE_URL` 与 `AI_PROVIDER_OVERRIDE=openai_compatible` | 单测稳定、可重放 |
| D-5 | Fallback 边界 | fallback 在 gateway 层触发；业务看到「成功 + fallback meta 标记」或「最终失败」；业务代码绝不写 retry-with-different-model 循环 | 防止业务代码绕开 cost cap / rate limit |
| D-6 | 观测埋点强制 | 每次调用必须产出 7 个 metric + DB 行 + log；客户端封装为 middleware-style decorator，不允许业务调用绕过埋点 | F1 dashboard 可信 |
| D-7 | 隐私字段红线 | log / metric / DB metadata 字段中绝不出现明文 prompt / response；只允许 hash / 长度 / profile | 与 ADR-Q5 / [05-logging-standard.md §5](../../../easyinterview-tech-docs/05-logging-standard.md) 对齐 |
| D-8 | OpenAI-compatible API 协议子集 | Chat Completions（`/v1/chat/completions`）+ Embeddings（`/v1/embeddings`）+ Audio Transcription（`/v1/audio/transcriptions`）；不锁 model_id 命名（由 gateway 路由） | 主流 OpenAI-compatible gateway 即插即用 |

### 3.2 待确认事项

- 是否在 `AIClient` 上扩展 `Tools(...)` 接口（function calling / tool use）：默认 P0 不上；如 W3 业务域出现 tool-use 需求，可在本 spec 修订递增版本后加（仍不打破 provider-neutral；ADR-Q6 §5 已记录此触发条件）。
- `model_profile_version` 是否独立 SemVer vs 与 prompt_version 联动：默认独立 SemVer（profile 升级不必随 prompt），由 F3 在自己的 plan 里决定如何引用。
- Stream 模式的 SSE / chunked 协议：默认 SSE（OpenAI-compatible 主流），实际由 001-bootstrap plan 落地时回填。

## 4 设计约束

### 4.1 接口约束

- `AIClient.Complete` 的入参 `payload` 必须包含 `messages[]` + `metadata`（业务侧的 `feature_key` / `prompt_version` / `rubric_version`）；client 不直接接受裸 prompt 字符串。
- `AICallMeta` 字段顺序固定（与 ADR-Q6 §3.1 一致），TS / Go 双端共用 B1 共享类型；任何字段新增由 B1 修订并由本 spec 引用。
- `Stream` 必须可中断（context cancellation），中断后客户端必须仍然产出 partial meta（`tokens_in/out` 截至中断时点）。

### 4.2 路由与 fallback 约束

- Profile fallback 只支持 ordered list（不支持权重路由 / A-B 桶）；A-B / 用户分桶由 PostHog feature flag 在业务层决定（与 ADR-Q3 一致），不入侵 AIClient。
- 单次调用 fallback 最多 2 跳（`primary → fallback[0] → fallback[1]`），超出标记 `AI_FALLBACK_EXHAUSTED`。
- `timeout_ms` 是 client 总超时（含网络 + gateway 排队 + provider 推理），到期后客户端必须 return `AI_PROVIDER_TIMEOUT`，不能让 ctx 永久挂起。

### 4.3 观测与隐私约束

- 每次调用产生的 log（事件名 `ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`）必须遵守 [05-logging-standard.md §4.4](../../../easyinterview-tech-docs/05-logging-standard.md#44-ai-log-额外字段) AI Log 字段集；明文红线见 §5.1。
- DB `ai_task_runs.metadata` 仅允许写入摘要字段（hash / 长度 / profile）；`raw_response_object_key` 字段在 [03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md) 中已预留为可选，如需保留原始响应必须落到对象存储（非 PG），由 F3 / F1 在自己的 spec 中决定是否启用。
- `audit_events.action='ai.call'` 必须由 client 内部写入，业务代码不得跳过。

### 4.4 测试约束

- Stub provider 的输入 → 输出映射必须 deterministic（相同 input + profile 永远产出相同 output）；不依赖时间 / 随机数。
- 任何单元测试默认走 stub；不允许某测试在 CI 中悄悄打到真 gateway（lint 规则在 A5 接入时强制）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `backend/internal/ai/aiclient/` 接口与默认实现 | A3 | `AIClient` / `AICallMeta` / stub / openai_compatible adapter |
| Model Profile 文件 schema | A3 | `config/ai-profiles/*.yaml` schema 与热加载 |
| Profile 文件内容（prompt / rubric / model 三元组） | F3 | A3 只锁 schema 字段，具体值由 F3 + 运维维护 |
| Profile 文件路径 / secret 注入 | A4 | `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` / `AI_MODEL_PROFILE_PATH` |
| 真实 gateway 部署 | E4 + 运维 | Higress / LiteLLM / Kong AI Helm chart 与配置 |
| 业务调用现场 | C4-C7 / C9 / C11 / C14 | 各 C 域 spec / plan 引用 profile name |
| 共享类型 | B1 | `AICallMeta` Go / TS 字段、`AI_*` 错误码 |
| DB 表 | B4 | `ai_task_runs` schema |
| Metric / Dashboard | F1 | 7 个 ai_* metric + AI Cost & Quality Dashboard |
| Local dev gateway 占位 | A2 | compose 中预留 `ai-gateway-mock` 服务名 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Stub 单测 | 单测环境（无 `AI_GATEWAY_BASE_URL`） | 业务代码调用 `aiclient.Complete(ctx, "practice.followup.default", payload)` | client 路由到 stub provider；返回结构化 response + meta；`meta.provider == "stub"`；同 input 多次调用结果一致 | A3 后续 001 |
| C-2 | OpenAI-compatible 路由 | dev/staging 设置 `AI_GATEWAY_BASE_URL=http://gateway:8080/v1` | 调用 `Complete` | 出站 HTTP 请求命中 `/v1/chat/completions`；header 含 `Authorization`；响应被解析为 `response + meta`；不直接 import 任何厂商 SDK（grep `go.mod` 无 `openai-go` / `anthropic-sdk-go` 等） | A3 后续 001 |
| C-3 | Fallback 触发 | profile `default.provider` 超时；`fallback[0]` 可用 | 调用 `Complete` | 客户端透传 fallback 触发；`meta.fallback_chain == [primary, fallback0]`；`ai_fallback_total{from_model=…,to_model=…}` +1；业务代码无任何额外重试 | A3 后续 001 |
| C-4 | Profile 热加载 | A3 后续 001 完成 | `config/ai-profiles/*.yaml` 修改后保存 | client 在 ≤ 30s 内热加载新 profile；正在进行的调用使用旧 profile 完成；新调用使用新 profile | A3 后续 001 |
| C-5 | 观测埋点齐全 | 任一调用完成 | F1 metric / log / DB 三方查询 | 7 个 metric 各 +1；log 含 §4.3 字段；`ai_task_runs` 写一行；`audit_events` 写一行（`action=ai.call`，无明文） | A3 后续 001 + F1 接入 |
| C-6 | 隐私红线 | grep 全部生产代码与 log | 任意调用 | 不出现 `payload.messages[*].content` / `response.content` 明文落 log 或 DB metadata；hash / 长度 / profile 三类摘要必须出现 | A3 后续 001 |
| C-7 | 错误码合规 | provider 返回结构化输出非法 | client `validate_output` 失败 | 返回错误码 `AI_OUTPUT_INVALID`（B1 锁定常量）；`ai_output_validation_failures_total` +1 | A3 后续 001 |
| C-8 | W1 review gate | 本 spec 通过 `/plan-review` | 9 份 W1 spec 集中审查 | A3 与 F3 / B1 / A4 / F1 / E4 引用关系自洽，无 `AIClient` 字段冲突 | engineering-roadmap/001 Phase 3.2 |

## 7 关联计划

A3 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 A3 自身的 `001-bootstrap`（W2 起）承接：

- 落地 `backend/internal/ai/aiclient/` 接口、stub provider、`openai_compatible` provider。
- 落地 Model Profile YAML schema + loader + 热加载。
- 落地 client 内部 metric / log / DB / audit decorator。
- 提供单测（stub 路径）与一份最小 OpenAPI-compatible adapter 集成测试（mock server，由 E1 复用）。

A3 第二份 plan（如 `002-tools-and-streaming`）按需在 W3 / W4 阶段决策（function calling、stream 完整化），届时递增 spec 版本并原地修订；不创建 sibling spec。
