# AI Provider and Model Routing Spec

> **版本**: 2.2
> **状态**: active
> **更新日期**: 2026-05-05

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A3 `ai-provider-and-model-routing` 保留为当前 active Foundation spec。它把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 的 AI provider 抽象落到代码层，决定业务域如何以 provider-neutral 方式使用 LLM / embedding / speech / rerank / judge 等 AI 能力。

基于当前 [product-scope](../product-scope/spec.md) 与 `docs/ui-design/` / `ui-design/` 交互，easyinterview 的 AI 使用面已经超过“单一文本 LLM endpoint”：

- JD 导入解析、Job Picks 匹配解释、公司轻情报摘要；
- 模拟面试首题、追问、轻量观察 / hint、文本输入中的语音转写、voice interview；
- 报告生成、逐题评估、复练当前轮 / 下一轮上下文；
- 简历解析、岗位定制、bullet 改写、用户画像信号更新；
- 真实面试 debrief 文本引导、语音复盘抽取、debrief 分析与复盘练习；
- embedding upsert、retrieval rerank、离线 LLM Judge / eval。

因此本 spec 的目标从“一个全局 AI provider base URL + profile”升级为：

1. **Provider-neutral 抽象**：业务代码 0 厂商 SDK 入侵，只依赖 `AIClient` 接口与 `model_profile_name`；切换供应商、模型、fallback、成本等级不改业务代码。
2. **Provider Registry + Capability Profile**：应用维护 provider connection registry；Model Profile 按 `capability`（如 `chat` / `embed` / `stt` / `realtime` / `rerank` / `judge`）引用 provider ref、model、参数与 fallback。单一 provider 可作为启动配置，但不是最终架构约束。
3. **可观测可计费**：每一次 `AIClient.*` 调用必须产出 A3-owned `AICallMeta`（provider / model_family / model_id / capability / prompt_version / rubric_version / model_profile_version / language / tokens / cost / latency / fallback_chain / route / validation_status / error_code），并由 [F1 `observability-stack`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 统一接入 metric / log / DB（`ai_task_runs`）。
4. **可测试可灰度**：`stub` provider 提供 hash-based 确定性输出，仅用于单元测试、离线契约测试或显式 mock 场景；local deploy / Kind / staging / prod 必须通过 provider registry 解析真实 provider endpoint 与 secret，缺失即 fail-fast。
5. **隐私守约**：AI 调用 payload 在 `audit_events` 写 hash + 长度 + profile，不写明文 prompt / response（与 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) 对齐）。

本 spec 不定义具体 prompt（归 [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md)）、不定义业务调用现场（归各 C / D 域）、不部署或拥有独立 AI 代理服务。若未来使用外部模型代理 / router，它只是 provider registry 中的一个 provider ref，不成为业务语义。

## 2 范围

### 2.1 In Scope

- **AIClient 接口**：Go 包 `backend/internal/ai/aiclient/`。当前同步调用面为 `Complete(ctx, profile, payload) → (response, meta)` / `Embed(ctx, profile, input) → (vector, meta)`；`Stream(ctx, profile, payload) → (<-chan AIStreamEvent, error)` 的事件合同已冻结，完整 provider streaming 消费由 002+ 承接。`Transcribe` / realtime speech / rerank / judge 的可执行 adapter 由后续 plan 打开，但其 capability profile 命名空间由本 spec 锁定。
- **Provider Registry schema**：配置文件 + 启动加载；字段：`name` / `protocol`（`stub` | `openai_compatible` | `realtime_audio` | `rerank_compatible` | `judge_compatible`）/ `base_url_env` / `api_key_env` / `capabilities[]` / `version`。Registry 文件默认落点 `config/ai-providers.yaml`，A4 通过 `AI_PROVIDER_REGISTRY_PATH` 注入；tracked 文件不得包含 secret 明文，只保存 env secret ref。`base_url_env` / `api_key_env` 对 `stub` 可为空；对需要网络出站的 provider protocol 必须声明，且仅在该 provider 被 profile 选中或进入 fallback chain 时解析实际 secret。
- **Model Profile schema**：YAML 文件 + 热加载；字段：`name` / `capability`（`chat` | `embed` | `stt` | `realtime` | `rerank` | `judge`）/ `status`（`active` | `disabled` | `unsupported`）/ `unsupported_reason`（`disabled` / `unsupported` 时必填）/ `default.{provider_ref, model, params}` / `fallback[]`（每项包含 `provider_ref` / `model` / `when[]`）/ `timeout_ms` / `max_tokens` / `rate_limit.{rps, tpm}` / `route` / `version` / 可选 `privacy_policy`。Profile 文件默认落点 `config/ai-profiles/*.yaml`，A4 通过 `AI_MODEL_PROFILE_PATH` 注入。
- **Provider 实现集**：
  - `stub`：hash-based 确定性输出，从 OpenAPI fixtures 反向喂养（与 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 同源）；仅允许在单元测试、离线契约测试或显式 mock 场景启用。
  - `openai_compatible`：P0 已实现 Chat Completions / Embeddings 协议子集；后续 registry 化后通过 provider ref 读取 base URL / API key，不再依赖全局唯一 endpoint 语义；不直接 import 任何厂商 SDK。
  - `realtime_audio` / `rerank_compatible` / `judge_compatible`：本 spec 锁命名空间与 fail-closed 语义；可执行协议 adapter 必须由对应后续 plan 递增 spec 后实现。
- **路由策略**：业务 `feature_key` 由 F3 Resolve 为 `model_profile_name`；A3 由 profile 解析 provider ref、capability、model、参数与 fallback。Fallback 可由 AIClient 在 profile fallback chain 内集中执行，业务代码不得自行 retry-with-different-model；每次 fallback 都必须写入 meta / metric / log。
- **观测埋点契约**：A3 必须注册并暴露 `ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total` 共 7 个 metric family；每次调用递增 run / latency / token / cost，validation failure 与 fallback counter 仅在对应事件发生时递增。同时落 DB 表 `ai_task_runs`，schema 由 [B4](../db-migrations-baseline/spec.md) 落地。
- **Audit hook**：每次调用产出 `audit_events` 行（`action=ai.call`），`metadata` 字段含 `prompt_hash` / `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`；不含明文。

### 2.2 Out of Scope

- 具体 prompt 内容、rubric schema、版本表：归 [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md)。
- 业务调用现场（哪个 C / D 域调用哪种 profile）：归各自 spec / plan。
- 外部 AI provider 服务部署、K8s Secret / Vault / cost cap 策略：归 A4 / E4 / 运维；本 spec 只锁应用侧 registry / profile / provider ref 契约。
- STT / realtime voice 的完整协议 adapter、音频 payload 形态与 HTTP wire：归 002+ 与 C14 / practice voice owner；本 spec 只锁 profile capability 与 fail-closed 规则。
- LLM Judge / 离线评估集实现：归 F3 后续评估 plan。
- DB 表本身：归 B4；本 spec 只引用字段名。
- 错误码与跨语言 AI vocabulary：依赖 B1 已落地的 `AI_*` 前缀错误码、AI capability、provider registry 字段名、model profile 字段名与 AI meta 字段名；新增跨边界字面量必须先改 B1。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | AIClient 接口形态 | 业务调用面继续只接受 `profile name`；`Complete` / `Embed` 已可执行，`Stream` 事件合同锁定，speech / rerank / judge adapter 由后续 plan 打开 | 业务代码绝对零厂商 SDK 入侵 |
| D-2 | Provider Registry | A3 owns provider registry schema；tracked registry 只保存 provider ref、protocol、capabilities 与 secret env ref，不保存 secret 明文 | 单一 provider 可作为启动配置，但多 provider / 多能力不需要改业务代码 |
| D-3 | Model Profile 字段集 | profile 使用 `capability` + `provider_ref`；不再把 profile 绑定为全局 provider endpoint 的 route | provider profile 配置漂移可控 |
| D-4 | 业务引用形态 | 业务只引用 `model_profile_name`（如 `practice.followup.default` / `report.generate.default`），不引用 provider / model 字符串 | 切换 provider / model = 改 registry/profile，不改业务代码 |
| D-5 | Stub 触发条件 | 仅 `APP_ENV=test`、离线契约测试或显式 mock 场景允许走 stub；local deploy / Kind / staging / prod 必须能通过 registry 解析真实 provider secret，缺失即 fail-fast | 单测稳定、可重放，同时防止本地部署静默假数据 |
| D-6 | Fallback 边界 | Fallback 由 AIClient 在 profile fallback chain 内集中执行，最多 2 跳；业务代码不得自行 retry-with-different-model；provider 自身返回的 fallback meta 也必须纳入同一 chain | 防止业务绕开 cost / rate limit / observability |
| D-7 | 观测埋点强制 | A3 注册 7 个 metric family；每次调用必须产出 run / latency / token / cost 指标 + DB 行 + log；fallback / validation failure 指标只在对应事件发生时递增 | F1 dashboard 可信且 counter 语义正确 |
| D-8 | 隐私字段红线 | log / metric / DB metadata 字段中绝不出现明文 prompt / response；只允许 hash / 长度 / profile | 与 ADR-Q5 / logging 标准对齐 |
| D-9 | OpenAI-compatible API 协议子集 | 当前可执行协议仍是 Chat Completions + Embeddings；Audio Transcription / realtime / rerank / judge 进入后续 plan 前必须 fail-closed | 主流 provider 可即插即用，同时避免假承诺 voice 能力 |
| D-10 | F3 profile 覆盖 | F3 12 个 baseline feature_key 必须全部能解析到 A3 profile catalog；P1/P2 capability 可先以 `status=disabled` / `status=unsupported` profile 占位，并写明 `unsupported_reason`，但不得缺命名空间 | 业务域开工前具备完整 AI 调用坐标 |
| D-11 | Product/UI capability inventory | A3 spec 必须维护产品 / UI AI 场景到 capability family 的映射；新增 AI 场景必须先修订本表与 F3 feature_key / profile 字典 | 防止新业务回到单模型假设 |
| D-12 | B1 AI vocabulary 边界 | `chat/embed/stt/realtime/rerank/judge` capability、provider registry/profile 字段名、AI meta 字段名与 provider/profile routing `AI_*` 错误码由 B1 生成；A3 只 alias / consume，不私造跨边界常量 | 防止 Go/TS/OpenAPI 与 runtime 常量漂移 |

### 3.2 待确认事项

- `model_profile_version` 是否独立 SemVer vs 与 prompt_version 联动：默认独立 SemVer（profile 升级不必随 prompt），由 F3 在自己的 plan 里决定如何引用。
- Stream 暴露到 HTTP 时采用 SSE 还是 chunked：内部 `AIStreamEvent` 合同先固定；具体 HTTP wire 由 002+ consumer plan 决定。
- Voice Interview 是使用 `stt + chat + tts` 组合还是 realtime multimodal provider：由 C14 / practice voice 进入实现前与本 spec 联合修订；未决前，UI voice 能力必须 feature-gated 或 fail-closed。
- Rerank / judge 是否使用专用 provider protocol 还是 OpenAI-compatible JSON schema 调用：由 C11 / F3 eval plan 决定；A3 只要求 capability profile 能表达并观测。

## 4 设计约束

### 4.1 接口约束

- `AIClient.Complete` 的入参 `payload` 必须包含 `messages[]` + `metadata`（业务侧的 `feature_key` / `prompt_version` / `rubric_version` / `language`，可选 `output_schema`）；client 不直接接受裸 prompt 字符串。
- `AICallMeta` 字段顺序固定：`provider` / `model_family` / `model_id` / `capability` / `prompt_version` / `rubric_version` / `model_profile_name` / `model_profile_version` / `language` / `input_tokens` / `output_tokens` / `cost_usd_micros` / `latency_ms` / `fallback_chain[]` / `route` / `validation_status` / `error_code`。其中跨 Go/TS/OpenAPI 边界消费的 capability、profile/provider 字段名、fallback label 字段与错误码由 B1 生成；A3 owns runtime 填充与校验。
- `Stream` 返回 `AIStreamEvent` channel，event type 固定为 `delta` / `error` / `done`；`delta` 只携带结构化增量，`error` 携带 B1 错误码，`done` 携带最终 `AICallMeta`。`Stream` 必须可中断（context cancellation）。
- 不支持的 capability 必须 fail-closed：profile 能加载为 `disabled` / `unsupported` 状态，且必须携带 `unsupported_reason`；运行时调用不得静默降级到 chat 模型或 stub。

### 4.2 路由与 fallback 约束

- Provider ref 是 registry 内唯一名；Model Profile 只能引用 registry 中已存在且声明了对应 `capabilities[]` 的 provider ref。
- Registry 中的 `base_url_env` / `api_key_env` 只是 secret 名称；loader 从 A4 SecretSource 解析实际值。tracked YAML 不得出现真实 API key。
- Profile fallback chain 由 A3 client 执行或合并 provider 返回的 fallback meta；每次 fallback 记录 `from_provider/from_model` 与 `to_provider/to_model`，最多 2 跳。
- A/B / 用户分桶仍由 PostHog feature flag 在业务 / F3 Resolve 层决定，不塞入 AIClient；AIClient 只接收最终 profile name。
- `timeout_ms` 是 client 总超时（含网络 + provider 排队 + provider 推理），到期后客户端必须 return `AI_PROVIDER_TIMEOUT`，不能让 ctx 永久挂起。

### 4.3 观测与隐私约束

- 每次调用产生的 log（事件名 `ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`）必须遵守 AI Log 字段集；明文红线见 §1 目标 5。
- DB `ai_task_runs.metadata` 仅允许写入摘要字段（hash / 长度 / profile / provider ref / capability）；原始响应如需保存必须落对象存储并由业务隐私策略覆盖，不能写 PG metadata。
- `audit_events.action='ai.call'` 必须由 client 内部写入，业务代码不得跳过。

### 4.4 测试约束

- Stub provider 的输入 → 输出映射必须 deterministic（相同 input + profile 永远产出相同 output）；不依赖时间 / 随机数。
- 任何单元测试默认走 stub；不允许某测试在本地测试或未来远端 CI 中悄悄打到真 provider。
- 任何 local deploy / Kind / staging / prod 部署都不得在被选中的真实 provider secret 缺失时静默回退到 stub；启动期 config validation 必须直接失败。
- Registry / profile loader 必须有负向 fixture：未知 provider ref、capability 不匹配、secret env 缺失、unsupported capability 被调用、profile fallback 超 2 跳。

### 4.5 Product/UI AI Capability Catalog

| 产品 / UI 场景 | 主要输入 | Capability family | 默认 profile 命名 |
|----------------|----------|-------------------|-------------------|
| JD 导入解析 | JD 文本 / URL 提取文本 | `chat` 结构化抽取 | `target.import.default` |
| Job Picks 匹配解释 | JD + 简历 / 用户画像 | `embed` + `rerank` + `chat` | `embedding.default` / `retrieval.rerank.default` / `target.import.default` |
| 公司轻情报摘要 | source-grounded public info | `chat` source-grounded summarization | `target.intel.default`（P1/P2 占位） |
| 简历解析 | 简历文本 / 上传解析结果 | `chat` 结构化抽取 | `resume.parse.default` |
| 简历定制 / bullet 改写 | JD + 简历证据 | `chat` 写作 / 改写 | `resume.tailor.default` |
| 用户画像信号更新 | 简历 / JD / session / debrief | `chat` 分类摘要 + `embed` | `profile.update.default`（后续占位） |
| 模拟面试首题 | JD / round / resume / role | `chat` 对话生成 | `practice.first_question.default` |
| 模拟面试追问 | transcript / current answer | `chat` 低延迟生成 | `practice.followup.default` |
| assisted hint / turn observe | 当前回答 + rubric | `chat` 低延迟观察 | `practice.turn_observe.default` |
| 文本输入语音转写 | audio chunk | `stt` | `practice.dictation.stt.default`（002+） |
| Voice Interview | audio stream + session state | `realtime` 或 `stt + chat + tts` | `practice.voice.realtime.default`（002+） |
| 报告生成 | full session + JD + resume | `chat` 长上下文结构化推理 | `report.generate.default` |
| 单题评估 | 单题回答 + rubric | `chat` rubric assessment | `report.assessment.default` |
| 复练当前轮 / 下一轮 | report gaps + replay items | `chat` 生成 | `report.generate.default` / `report.assessment.default` / `practice.first_question.default` / `practice.followup.default` |
| Debrief 文本引导 | JD / mock report / resume | `chat` source-grounded generation | `debrief.generate.default` |
| Debrief 语音抽取 | audio + running transcript | `stt` + `chat` 抽取 | `debrief.voice.extract.default`（002+） |
| Debrief 分析 | real questions + mock/JD/resume | `chat` 长上下文分析 | `debrief.generate.default` |
| Embedding upsert | JD / resume / report / debrief text | `embed` | `embedding.default` |
| Retrieval rerank | candidate evidence list | `rerank` | `retrieval.rerank.default` |
| 离线 LLM Judge / eval | prompt output + rubric | `judge` | `judge.default`（F3 eval） |

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `backend/internal/ai/aiclient/` 接口与默认实现 | A3 | `AIClient` / `AICallMeta` / registry loader / profile loader / stub / openai_compatible adapter |
| Provider Registry schema | A3 + A4 | A3 owns schema / validation；A4 owns path/env/SecretSource 注入 |
| Model Profile schema | A3 | `config/ai-profiles/*.yaml` schema 与热加载 |
| Profile 文件内容 | F3 + 各 AI feature owner | F3 owns feature_key -> model_profile_name；A3 owns profile schema；业务 owner 负责新增场景时补 profile |
| Profile 文件路径 / secret 注入 | A4 | `AI_PROVIDER_REGISTRY_PATH` / `AI_MODEL_PROFILE_PATH` 与 provider-specific env secret ref |
| 真实 provider endpoint | E4 + 运维 | 本地部署可直连真实 AI provider；staging / prod 可接运维提供的 provider endpoint；本 spec 不部署独立代理 |
| 业务调用现场 | C4-C7 / C9 / C11 / C14 / D3 | 各业务 spec / plan 引用 profile name，不引用 provider/model |
| 共享约定 | B1 | `AI_*` 错误码、AI capability、provider registry/profile 字段名、AI meta 字段名共享常量、`ApiError` / `ApiErrorResponse` 消费约定 |
| DB 表 | B4 | `ai_task_runs` schema |
| Metric / Dashboard | F1 | 7 个 ai_* metric + AI Cost & Quality Dashboard |
| 测试 stub provider | A3 | 应用内 deterministic stub，仅供单元测试 / 离线契约测试 / 显式 mock 场景 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Stub 单测 | 单测环境（`APP_ENV=test`，无真实 provider secret） | 业务代码调用 `aiclient.Complete(ctx, "practice.followup.default", payload)` | client 路由到 stub provider；返回结构化 response + meta；`meta.provider == "stub"`；同 input 多次调用结果一致 | 001 |
| C-2 | Registry provider route | registry 中 `default-openai-compatible` 声明 `chat` / `embed` capability，并引用 env secret | 调用 `Complete` / `Embed` | 出站 HTTP 请求命中该 provider ref 的 OpenAI-compatible endpoint；header 含 `Authorization`；`meta.provider` / `meta.capability` / `meta.model_profile_name` 正确 | 003 |
| C-3 | Central fallback | profile 声明 primary + fallback provider ref，primary 超时且 fallback 成功 | 调用 `Complete` | AIClient 执行受限 fallback，`fallback_chain` 记录 provider/model hop；`ai_fallback_total` +1；业务代码无 retry-with-different-model 循环 | 003 |
| C-4 | Registry + profile 热加载 | A3 loader 已启动 | `config/ai-providers.yaml` 或 `config/ai-profiles/*.yaml` 修改后保存 | client 在 ≤ 30s 内热加载；正在进行的调用使用旧快照完成；新调用使用新快照 | 003 |
| C-5 | 观测埋点齐全 | 任一无 fallback、无 validation failure 的调用完成 | F1 metric / log / DB 三方查询 | 7 个 metric family 均已注册；run / latency / token / cost 指标增长；fallback / validation failure counter 不增长；`ai_task_runs` + `audit_events` 各写一行，无明文 | 001 + 003 |
| C-6 | 隐私红线 | grep 全部生产代码与 log | 任意调用 | 不出现 `payload.messages[*].content` / `response.content` 明文落 log 或 DB metadata；hash / 长度 / profile 三类摘要必须出现 | 001 + 003 |
| C-7 | 错误码合规 | provider 返回结构化输出非法 | client `validate_output` 失败 | 返回错误码 `AI_OUTPUT_INVALID`；`ai_output_validation_failures_total` +1 | 001 |
| C-8 | active spec relation gate | 本 spec 通过 `/plan-review` | 与当前 active spec 和 future workstream 关系审查 | A3 与 F3 / B1 / A4 / F1 / release gate 引用关系自洽；A3 不重新引入已废弃的 provider-proxy 业务语义 | plan-review |
| C-9 | Registry secret fail-fast | local deploy / Kind / staging / prod 缺失 registry 选中 provider 的 base URL 或 API key | 启动 API / worker | 进程启动失败并报配置错误；不得自动回退到 stub provider | 003 + A4 |
| C-10 | F3 baseline profile coverage | F3 12 个 baseline feature_key 已定义默认 profile name | 运行 profile coverage lint | 每个默认 profile 在 `config/ai-profiles/` 中存在，且 capability / provider_ref / status 合法；允许 P1/P2 profile `disabled` / `unsupported`，但必须携带 `unsupported_reason` 且不得缺文件 | 003 + F3 |
| C-11 | Product/UI capability inventory drift | 新增 AI 场景或 UI 交互依赖 AI | `/plan-review` 或 lint 检查 | 本 spec §4.5、F3 feature_key 字典与 A3 profile catalog 同步更新；不得只在业务代码 hardcode 新 profile | 003 + F3 |
| C-12 | Unsupported capability fail-closed | profile 使用 `stt` / `realtime` / `rerank` / `judge`，但对应 adapter 未激活 | 运行时调用该 profile | 返回明确 unsupported capability 错误并记录 meta/log；不得降级到 chat 或 stub；对应 UI 能力必须 feature-gated | 003 + 002 |

## 7 关联计划

A3 当前计划拆分为两份 completed foundation plan 与一份 draft capability adapter extension plan：

- [001-aiclient-and-profile-bootstrap](./plans/001-aiclient-and-profile-bootstrap/plan.md)（completed）：已落地 P0 `Complete` / `Embed`、`Stream` 事件合同类型、unit-test stub provider、`openai_compatible` Chat / Embeddings provider、基础 Model Profile loader 与 observability / audit decorator。
- [002-tools-streaming-and-stt](./plans/002-tools-streaming-and-stt/plan.md)（draft/blocked）：Tools / full streaming / STT / realtime speech 等协议能力延期占位；必须先触发 ADR / spec 修订，才能切 active。
- [003-provider-registry-and-capability-profiles](./plans/003-provider-registry-and-capability-profiles/plan.md)（completed）：已落地本 spec v2.2 的 provider registry、capability-scoped profile、central fallback、A4 env dictionary、B1 AI vocabulary、F3 12 profile coverage、active anti-stub gate 与 drift gate，为后续业务域实施提供完整 AI provider 配置面。

后续如需扩展，递增本 spec 版本并原地修订对应 plan；不创建 sibling spec。
