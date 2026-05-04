# AI Gateway and Model Routing Spec

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-04-29

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将历史 A3 `ai-gateway-and-model-routing` 保留为当前 active Foundation spec（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md)；间接依赖 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 提供的错误码与通用 API 约定）。它把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md) 的 9 项硬约束落到代码层，决定了：

- 业务代码（C4–C7 / C9 / C11 / C14）调用 LLM / embedding / STT 的唯一接口形态；
- prompt / rubric / model profile 在调用现场如何串起来；
- 单元测试 / 离线契约测试 / 本地部署（docker compose 与 Kind）/ staging / prod 如何切换 provider；默认本地部署不依赖独立 AI gateway 服务，但必须接真实 AI provider 的 OpenAI-compatible LLM 服务。

目标是：

1. **Provider-neutral 抽象**：业务代码 0 厂商 SDK 入侵，只依赖 `AIClient` 接口与 `Model Profile` name；切换厂商或加 fallback 不改业务代码。
2. **可观测可计费**：每一次 `AIClient.*` 调用必须产出 A3-owned `AICallMeta`（provider / model_family / model_id / prompt_version / rubric_version / model_profile_version / task_type / language / tokens / cost / latency / fallback_chain / route / validation_status / error_code），并由 [F1 `observability-stack`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 统一接入 metric / log / DB（`ai_task_runs`）。
3. **可测试可灰度**：`stub` provider 提供 hash-based 确定性输出，仅用于单元测试、离线契约测试或显式 mock 场景；docker compose / Kind / staging / prod 部署必须通过 `AI_GATEWAY_BASE_URL` + `AI_GATEWAY_API_KEY` 连接 OpenAI-compatible endpoint（真实 LLM provider 或生产 AI Gateway 均可），不允许默认降级到 stub。
4. **隐私守约**：AI 调用 payload 在 `audit_events` 写 hash + 长度 + profile，不写明文 prompt / response（与 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) 对齐）。

本 spec 不定义具体 prompt（归 [F3 `prompt-rubric-registry`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)）、不定义业务调用现场（归各 C 域）、不部署 gateway（运维 / E4 承接）。

## 2 范围

### 2.1 In Scope

- **AIClient 接口**：Go 包 `backend/internal/ai/aiclient/`，P0 唯一对外能力为 `Complete(ctx, profile, payload) → (response, meta)` / `Embed(ctx, profile, input) → (vector, meta)`；`Stream(ctx, profile, payload) → (<-chan AIStreamEvent, error)` 的事件合同在本 spec 冻结，但完整 provider streaming 消费由 002+ 承接。`AICallMeta` 是 A3-owned 运行时结构体；B1 提供共享错误码、Model Profile / AI meta 字段名等跨语言常量或生成类型，A3 owns runtime 填充与校验语义。
- **Model Profile schema**：YAML 文件 + 热加载；schema 在本 spec 冻结。字段：`name` / `task_type`（`chat` | `embed` | `stt`，其中 `stt` 为 C14 P2 预留值，A3 001 不实现音频转写调用）/ `default.{provider, model, params}` / `fallback[]`（按序触发条件）/ `timeout_ms` / `max_tokens` / `rate_limit.{rps, tpm}` / `gateway_route` / `version`。Profile 文件落点 `config/ai-profiles/*.yaml`（A4 控制 `AI_MODEL_PROFILE_PATH` 指向）。
- **Provider 实现集**：
  - `stub`：hash-based 确定性输出，从 OpenAPI fixtures 反向喂养（与 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 同源）；仅允许在单元测试、离线契约测试或显式 mock 场景启用。
  - `openai_compatible`：通过 `AI_GATEWAY_BASE_URL` 出站，P0 仅依赖 OpenAI Chat Completions / Embeddings 协议子集；Audio Transcription 协议为 C14 P2 预留，不进入 A3 001 验收。本地部署可直连真实 AI provider，生产可指向 Higress / LiteLLM / Kong AI 等 gateway；不直接 import 任何厂商 SDK。
- **路由策略**：profile name → endpoint / gateway route → provider/model；fallback 只允许在 AIClient 连接的 OpenAI-compatible endpoint / gateway route 层触发。如果该 endpoint 是真实 LLM provider 且不提供 fallback，A3 client 不自行切换模型。业务看到「成功 + fallback meta」或「最终失败」，不允许业务自行重试切换模型。
- **观测埋点契约**：A3 必须注册并暴露 `ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total` 共 7 个 metric family；每次调用递增 run / latency / token / cost，validation failure 与 fallback counter 仅在对应事件发生时递增。同时落 DB 表 `ai_task_runs`，schema 由 [B4](../db-migrations-baseline/spec.md#311-field-level-enum--check-来源矩阵) 落地，并在 [03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md) baseline 外补齐 A3/F1 需要的 typed meta columns。
- **Audit hook**：每次调用产出 `audit_events` 行（`action=ai.call`），`metadata` 字段含 `prompt_hash` / `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`；不含明文。

### 2.2 Out of Scope

- 具体 prompt 内容、rubric schema、版本表：归 [F3 `prompt-rubric-registry`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)。
- 业务调用现场（哪个 C 域调用 `Complete` 还是 `Embed`）：归各自 C 域 spec / plan。
- 真实 gateway 部署（Higress / LiteLLM Helm chart）、路由配置、cost cap、rate limit 规则：归运维 + E4；本 spec 仅锁 OpenAI-compatible API 契约。
- Token 计费成本表：本 spec 把 `cost_usd_micros` 字段定义清楚，具体 provider × model × pricing 由 F3 / F1 维护。
- LLM Judge / 离线评估集：归 F3。
- STT / Audio Transcription provider adapter、音频 payload 与 `Transcribe(...)` 接口：归 C14 P2；本 spec 只保留 `task_type=stt` 作为 profile schema 兼容预留值。
- DB 表本身：归 B4；本 spec 只引用字段名。
- 错误码命名：依赖 B1 已落地的 `AI_*` 前缀错误码（`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED`），新增错误码必须先改 B1。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | AIClient 接口形态 | P0 调用面为 `Complete(ctx, profile, payload) → (response, meta)` / `Embed(ctx, profile, input) → (vector, meta)`；`Stream(ctx, profile, payload) → (<-chan AIStreamEvent, error)` 的事件合同锁定但完整 streaming provider 消费由 002+ 承接；`payload` 与 `response` 为结构化对象（不直接传 string）；`meta` 由 client 返回，业务不能伪造 | 业务代码绝对零厂商 SDK 入侵 |
| D-2 | Model Profile 字段集 | 见 §2.1；新增字段必须递增 spec 版本 | gateway 配置漂移可控 |
| D-3 | 业务引用形态 | 业务只引用 `profile name`（如 `practice.followup.default` / `review.report.default`），不引用 provider / model 字符串 | 切换 provider / model = 改 profile YAML，不改代码 |
| D-4 | Stub 触发条件 | 仅 `APP_ENV=test`、离线契约测试或显式 mock 场景允许走 stub；docker compose / Kind / staging / prod 必须配置 `AI_GATEWAY_BASE_URL` 与 `AI_GATEWAY_API_KEY` 指向 OpenAI-compatible endpoint（真实 AI provider 或生产 gateway），缺失即 fail-fast | 单测稳定、可重放，同时保证本地部署验证真实 LLM 服务 |
| D-5 | Fallback 边界 | fallback 只在 AIClient 连接的 endpoint / gateway route 层触发；A3 client 不自行按 profile 多次请求不同 provider/model；业务看到「成功 + fallback meta 标记」或「最终失败」；业务代码绝不写 retry-with-different-model 循环 | 防止业务代码绕开 cost cap / rate limit |
| D-6 | 观测埋点强制 | A3 注册 7 个 metric family；每次调用必须产出 run / latency / token / cost 指标 + DB 行 + log；fallback / validation failure 指标只在对应事件发生时递增；客户端封装为 middleware-style decorator，不允许业务调用绕过埋点 | F1 dashboard 可信且 counter 语义正确 |
| D-7 | 隐私字段红线 | log / metric / DB metadata 字段中绝不出现明文 prompt / response；只允许 hash / 长度 / profile | 与 ADR-Q5 / [05-logging-standard.md §5](../../../easyinterview-tech-docs/05-logging-standard.md) 对齐 |
| D-8 | OpenAI-compatible API 协议子集 | P0：Chat Completions（`/v1/chat/completions`）+ Embeddings（`/v1/embeddings`）；P2/C14 才能启用 Audio Transcription（`/v1/audio/transcriptions`）并新增 `Transcribe` 合同；不锁 model_id 命名（由 profile / gateway 路由） | 主流 OpenAI-compatible gateway 即插即用，同时避免 P0 假承诺 STT |

### 3.2 待确认事项

- 是否在 `AIClient` 上扩展 `Tools(...)` 接口（function calling / tool use）：默认 P0 不上；如后续业务域出现 tool-use 需求，可在本 spec 修订递增版本后加（仍不打破 provider-neutral；ADR-Q6 §5 已记录此触发条件）。
- `model_profile_version` 是否独立 SemVer vs 与 prompt_version 联动：默认独立 SemVer（profile 升级不必随 prompt），由 F3 在自己的 plan 里决定如何引用。
- Stream 暴露到 HTTP 时采用 SSE 还是 chunked：内部 `AIStreamEvent` 合同先固定；具体 HTTP wire 由后续 consumer plan 决定。

## 4 设计约束

### 4.1 接口约束

- `AIClient.Complete` 的入参 `payload` 必须包含 `messages[]` + `metadata`（业务侧的 `feature_key` / `prompt_version` / `rubric_version` / `language`，可选 `output_schema`）；client 不直接接受裸 prompt 字符串。
- `AICallMeta` 字段顺序固定（与 ADR-Q6 §3.1 一致并补齐 B4/F1 消费字段）：`provider` / `model_family` / `model_id` / `task_type` / `prompt_version` / `rubric_version` / `model_profile_name` / `model_profile_version` / `language` / `input_tokens` / `output_tokens` / `cost_usd_micros` / `latency_ms` / `fallback_chain[]` / `route` / `validation_status` / `error_code`。任何字段新增由本 spec 修订；如需跨前后端共享再追加到 B1。
- `Stream` 返回 `AIStreamEvent` channel，event type 固定为 `delta` / `error` / `done`；`delta` 只携带结构化增量，`error` 携带 B1 错误码，`done` 携带最终 `AICallMeta`。`Stream` 必须可中断（context cancellation），中断后客户端必须尽力产出 partial meta（`input_tokens` / `output_tokens` 截至中断时点），若 provider 不支持 token 增量则填 0，并通过 `error_code` / log `errorCode` 记录取消原因。

### 4.2 路由与 fallback 约束

- Profile fallback 只描述 endpoint / gateway route 可执行的 ordered fallback contract（不支持权重路由 / A-B 桶）；A-B / 用户分桶由 PostHog feature flag 在业务层决定（与 ADR-Q3 一致），不入侵 AIClient。
- A3 client 只消费 endpoint / gateway 返回的 fallback chain / model family meta；当本地直连真实 provider 且 provider 不支持 fallback 时，本次调用不做 fallback，只按 provider 返回成功或失败记录 meta。
- 单次调用 fallback 最多 2 跳，超出由 endpoint / gateway 标记 `AI_FALLBACK_EXHAUSTED`；A3 client 只透传并记录该错误码。
- `timeout_ms` 是 client 总超时（含网络 + gateway 排队 + provider 推理），到期后客户端必须 return `AI_PROVIDER_TIMEOUT`，不能让 ctx 永久挂起。

### 4.3 观测与隐私约束

- 每次调用产生的 log（事件名 `ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`）必须遵守 [05-logging-standard.md §4.4](../../../easyinterview-tech-docs/05-logging-standard.md#44-ai-log-额外字段) AI Log 字段集；明文红线见 §5.1。
- DB `ai_task_runs.metadata` 仅允许写入摘要字段（hash / 长度 / profile）；`raw_response_object_key` 字段在 [03-db-definition.md §5.8](../../../easyinterview-tech-docs/03-db-definition.md) 中已预留为可选，如需保留原始响应必须落到对象存储（非 PG），由 F3 / F1 在自己的 spec 中决定是否启用。
- `audit_events.action='ai.call'` 必须由 client 内部写入，业务代码不得跳过。

### 4.4 测试约束

- Stub provider 的输入 → 输出映射必须 deterministic（相同 input + profile 永远产出相同 output）；不依赖时间 / 随机数。
- 任何单元测试默认走 stub；不允许某测试在本地测试或未来远端 CI 中悄悄打到真 provider / gateway（当前由本地 lint / test gate 强制；远端 CI 仅在 A5 触发条件成立后再接入）。
- 任何 docker compose / Kind / staging / prod 部署都不得在缺少真实 provider endpoint / API key 时静默回退到 stub；启动期 config validation 必须直接失败。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `backend/internal/ai/aiclient/` 接口与默认实现 | A3 | `AIClient` / `AICallMeta` / stub / openai_compatible adapter |
| Model Profile 文件 schema | A3 | `config/ai-profiles/*.yaml` schema 与热加载 |
| Profile 文件内容（prompt / rubric / model 三元组） | F3 | A3 只锁 schema 字段，具体值由 F3 + 运维维护 |
| Profile 文件路径 / secret 注入 | A4 | `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` / `AI_MODEL_PROFILE_PATH`；`AI_GATEWAY_*` 是连接参数名，不表示必须部署 gateway |
| 真实 provider / gateway endpoint | E4 + 运维 | 本地部署可直连真实 AI provider；staging / prod 可经 Higress / LiteLLM / Kong AI 等 gateway；本 spec 只锁 OpenAI-compatible 契约 |
| 业务调用现场 | C4-C7 / C9 / C11 / C14 | 各 C 域 spec / plan 引用 profile name |
| 共享约定 | B1 | `AI_*` 错误码、Model Profile / AI meta 字段名共享常量、`ApiError` / `ApiErrorResponse` 消费约定；`AICallMeta` runtime 由 A3 拥有 |
| DB 表 | B4 | `ai_task_runs` schema |
| Metric / Dashboard | F1 | 7 个 ai_* metric + AI Cost & Quality Dashboard |
| 测试 stub provider | A3 | 应用内 deterministic stub，仅供单元测试 / 离线契约测试 / 显式 mock 场景；A2 不预留 `ai-gateway-mock` compose 服务，本地部署不默认使用 stub |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Stub 单测 | 单测环境（`APP_ENV=test`，无 `AI_GATEWAY_BASE_URL`） | 业务代码调用 `aiclient.Complete(ctx, "practice.followup.default", payload)` | client 路由到 stub provider；返回结构化 response + meta；`meta.provider == "stub"`；同 input 多次调用结果一致 | A3 后续 001 |
| C-2 | OpenAI-compatible 路由 | docker compose / Kind / staging 设置 `AI_GATEWAY_BASE_URL=https://provider.example/v1` 与 `AI_GATEWAY_API_KEY` | 调用 `Complete` | 出站 HTTP 请求命中真实 OpenAI-compatible `/v1/chat/completions`；header 含 `Authorization`；响应被解析为 `response + meta`；`meta.provider != "stub"`；不直接 import 任何厂商 SDK（grep `go.mod` 无 `openai-go` / `anthropic-sdk-go` 等） | A3 后续 001 |
| C-3 | Fallback 触发 | 连接 endpoint / gateway route 对 `default.provider` 超时后成功切到 fallback provider/model | 调用 `Complete` | A3 client 接收并记录 endpoint / gateway 返回的 fallback meta；`meta.fallback_chain == [primary, fallback0]`；`ai_fallback_total{from_model_family=…,to_model_family=…,result="fallback"}` +1；业务代码与 A3 client 均无 retry-with-different-model 循环 | A3 后续 001 |
| C-4 | Profile 热加载 | A3 后续 001 完成 | `config/ai-profiles/*.yaml` 修改后保存 | client 在 ≤ 30s 内热加载新 profile；正在进行的调用使用旧 profile 完成；新调用使用新 profile | A3 后续 001 |
| C-5 | 观测埋点齐全 | 任一无 fallback、无 validation failure 的调用完成 | F1 metric / log / DB 三方查询 | 7 个 metric family 均已注册；run / latency / token / cost 指标按本次调用增长；fallback / validation failure counter 不增长；log 含 §4.3 字段；`ai_task_runs` 写一行；`audit_events` 写一行（`action=ai.call`，无明文） | A3 后续 001 + F1 接入 |
| C-6 | 隐私红线 | grep 全部生产代码与 log | 任意调用 | 不出现 `payload.messages[*].content` / `response.content` 明文落 log 或 DB metadata；hash / 长度 / profile 三类摘要必须出现 | A3 后续 001 |
| C-7 | 错误码合规 | provider 返回结构化输出非法 | client `validate_output` 失败 | 返回错误码 `AI_OUTPUT_INVALID`（B1 锁定常量）；`ai_output_validation_failures_total` +1 | A3 后续 001 |
| C-8 | active spec relation gate | 本 spec 通过 `/plan-review` | 与当前 active spec 和 future workstream 关系审查 | A3 与 F3 / B1 / A4 / F1 / release gate 引用关系自洽；ADR-Q6 为 AI routing 真理源；`AICallMeta` runtime 归 A3，B1 提供共享字段 / 常量，无字段冲突 | plan-review |
| C-9 | 本地部署缺 AI provider fail-fast | docker compose 或 Kind 未设置 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY`，且启用了需要 AIClient 的组件 | 启动 API / worker | 进程启动失败并报配置错误；不得自动回退到 stub provider | A3 后续 001 + A4 + A2 |

## 7 关联计划

A3 当前计划拆分为一份 active P0 bootstrap plan 与一份 draft extension plan：

- [001-aiclient-and-profile-bootstrap](./plans/001-aiclient-and-profile-bootstrap/plan.md)（active）：落地 `backend/internal/ai/aiclient/` 的 P0 `Complete` / `Embed` 接口、`Stream` 事件合同类型、unit-test stub provider、`openai_compatible` Chat / Embeddings provider。
- 落地 Model Profile YAML schema + loader + 热加载；`task_type=stt` 仅作为保留值，不实现音频调用。
- 落地 client 内部 metric / log / DB / audit decorator，并按本 spec 区分 per-call metric 与 event-only counter。
- 提供单测（stub 路径）、离线 adapter 契约测试（mock server，由 E1 复用）与本地部署 smoke 的真实 provider 配置校验（不在测试中泄漏真实 key）。
- 该 plan owns `backend/internal/ai/aiclient/` 与 `config/ai-profiles/` fixture；`backend/cmd/api` / `backend/cmd/worker` 只作为 DI handoff，不要求 A3 001 创建或重写运行时 entrypoint。

- [002-tools-streaming-and-stt](./plans/002-tools-streaming-and-stt/plan.md)（draft/blocked）：仅作为 Tools / Streaming / STT 的延期占位；必须先触发 ADR-Q6 / 本 spec 修订，才能切 active。Function calling、stream 完整化、C14 音频转写不得塞回 001。

后续如需扩展，递增本 spec 版本并原地修订对应 plan；不创建 sibling spec。
