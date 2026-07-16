# AI Provider and Model Routing Spec

> **版本**: 2.30
> **状态**: active
> **更新日期**: 2026-07-16

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A3 `ai-provider-and-model-routing` 保留为当前 active Foundation spec。它把 [ADR-Q6](../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md) 的 AI provider 抽象落到代码层，决定业务域如何以 provider-neutral 方式使用 LLM / speech / judge 等 AI 能力。

基于当前 [product-scope](../product-scope/spec.md) 与 `docs/ui-design/` / `frontend/` 交互，easyinterview 的 AI 使用面已经超过“单一文本 LLM endpoint”：

- JD 导入解析、workspace 内嵌公司轻情报摘要；
- 模拟面试首题、追问、轻量观察 / hint、电话模式底层 STT / chat / TTS 编排；
- 报告生成、逐题评估、复练当前轮 / 下一轮上下文；
- 简历解析、岗位定制、bullet 改写；
- 离线 LLM Judge / eval。

因此本 spec 的目标从“一个全局 AI provider base URL + profile”升级为：

1. **Provider-neutral 抽象**：业务代码 0 厂商 SDK 入侵，只依赖 `AIClient` 接口与 `model_profile_name`；切换供应商、模型、fallback、成本等级不改业务代码。
2. **Provider Registry + Capability Profile**：应用维护 provider connection registry；Model Profile 按 `capability`（当前为 `chat` / `stt` / `tts` / `realtime` / `judge`）引用 provider ref、model、参数与 fallback。单一 provider 可作为启动配置，但不是最终架构约束。
3. **可观测可计费**：每一次 `AIClient.*` 调用必须产出 A3-owned `AICallMeta`（provider / model_family / model_id / capability / prompt_version / rubric_version / model_profile_version / language / tokens / cost / latency / fallback_chain / route / validation_status / error_code），并由 [F1 `observability-stack`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 统一接入 metric / log / DB（`ai_task_runs`）。
4. **可测试可灰度**：`stub` provider 提供 hash-based 确定性输出，仅用于单元测试、离线契约测试或显式 mock 场景；非测试本地 app run、未来 staging / prod 必须通过 provider registry 解析真实 provider endpoint 与 secret，缺失即 fail-fast。
5. **隐私守约**：AI 调用 payload 在 `audit_events` 写 hash + 长度 + profile，不写明文 prompt / response（与 [ADR-Q5](../engineering-roadmap/decisions/ADR-Q5-privacy-cadence.md) 对齐）。

本 spec 不定义具体 prompt（归 [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md)）、不定义业务调用现场（归各 backend / frontend owner）、不部署或拥有独立 AI 代理服务。若未来使用外部模型代理 / router，它只是 provider registry 中的一个 provider ref，不成为业务语义。

## 2 范围

### 2.1 In Scope

- **AIClient 接口**：Go 包 `backend/internal/ai/aiclient/`。当前调用面为 `Complete(ctx, profile, payload) → (response, meta)` / `Stream(ctx, profile, payload) → (<-chan AIStreamEvent, error)` / `Transcribe(ctx, profile, audio) → (transcript, meta)`；`Synthesize(ctx, profile, text) → (audio, meta)` 由 plan 004 打开。`CompletePayload` 可携带 provider-neutral `tools[]` / `tool_choice` / `output_schema`；业务仍只传 `model_profile_name`，不传 provider/model 字符串。向量化 / 重排能力已从当前开发阶段删除，未来如重新需要必须先递增本 spec、B1/B3/B4/F3 契约与实现计划。
- **Provider Registry schema**：配置文件 + 启动加载；字段：`name` / `protocol`（`stub` | `openai_compatible` | `doubao_speech` | `minimax_speech` | `realtime_audio` | `judge_compatible`）/ `base_url_env` / `api_key_env` / `capabilities[]` / `version`。Registry 文件默认落点 `config/ai-providers.yaml`，A4 通过 `AI_PROVIDER_REGISTRY_PATH` 注入；tracked 文件不得包含 secret 明文，只保存 env secret ref。`base_url_env` / `api_key_env` 对 `stub` 可为空；对需要网络出站的 provider protocol 必须声明，且仅在该 provider 被 profile 选中或进入 fallback chain 时解析实际 secret。
- **Model Profile schema**：单一 YAML catalog 文件 + 热加载；catalog 顶层字段为 `profiles[]`，每项字段为 `name` / `capability`（`chat` | `stt` | `tts` | `realtime` | `judge`）/ `status`（`active` | `disabled` | `unsupported`）/ `unsupported_reason`（`disabled` / `unsupported` 时必填）/ `default.{provider_ref, model, params}` / `fallback[]`（每项包含 `provider_ref` / `model` / `when[]`）/ `timeout_ms` / `max_tokens` / 可选 `context_window_tokens` / `rate_limit.{rps, tpm}` / `route` / `version` / 可选 `privacy_policy`。`context_window_tokens` 是单请求输入+输出容量；`rate_limit.tpm` 只表示吞吐，不得互相替代。Profile catalog 默认落点 `config/ai-profiles.yaml`，A4 通过 `AI_MODEL_PROFILE_PATH` 注入文件路径。
- **Provider 实现集**：
  - `stub`：hash-based 确定性输出，从 OpenAPI fixtures 反向喂养（与 [E1 `mock-contract-suite`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 同源）；仅允许在单元测试、离线契约测试或显式 mock 场景启用。
  - `openai_compatible`：P0 已实现 Chat Completions / chat streaming SSE / Audio Transcriptions 协议子集，并支持 Chat Completions tool-call wire 子集；通过 provider ref 读取 base URL / API key，不依赖全局唯一 endpoint 语义。adapter 内部固定使用官方 `github.com/openai/openai-go/v3` 维护 OpenAI-compatible wire，SDK 不得越过 A3 provider adapter 边界。当前开发主力 provider ref 为 `deepseek`，只使用 `deepseek-v4-flash` / `deepseek-v4-pro` 两个模型 ID。
  - `doubao_speech`：plan 004 打开的 provider-specific speech protocol，优先承载豆包 STT 与豆包 TTS。不得假设其 wire shape 与 OpenAI Audio Transcriptions 或其他 provider 一致。
  - `minimax_speech`：plan 004 打开的 provider-specific speech protocol，优先承载 MiniMax TTS（例如 `speech-02-turbo`）。MiniMax STT 只有在公开接口、权限和测试契约确认后才能加入 `stt` profile。
  - `judge_compatible`：只实现 `Complete` 的 judge 专用 OpenAI Chat Completions 子集，当前 `judge-deepseek` / `deepseek-v4-pro` 由 `judge.default` 使用；请求通过 profile 显式关闭 thinking、要求 JSON object，adapter 对空 final content、缺 choices 和非法响应 fail-closed，且不记录 `reasoning_content` / prompt / output 明文。`Transcribe` / `Stream` / `Synthesize` 始终返回 unsupported。
  - `realtime_audio`：本 spec 只锁命名空间与 fail-closed 语义；可执行协议 adapter 必须由对应后续 plan 递增 spec 后实现。
- **路由策略**：业务 `feature_key` 由 F3 Resolve 为 `model_profile_name`；A3 由 profile 解析 provider ref、capability、model、参数与 fallback。Fallback 可由 AIClient 在 profile fallback chain 内集中执行，业务代码不得自行 retry-with-different-model；每次 fallback 都必须写入 meta / metric / log。
- **观测埋点契约**：A3 必须注册并暴露 `ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total` 共 7 个 metric family；每次调用递增 run / latency / token / cost，validation failure 与 fallback counter 仅在对应事件发生时递增。同时落 DB 表 `ai_task_runs`，schema 由 [B4](../db-migrations-baseline/spec.md) 落地。
- **Audit hook**：每次调用产出 `audit_events` 行（`action=ai.call`），`metadata` 字段含 `prompt_hash` / `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`；不含明文。

### 2.2 Out of Scope

- 具体 prompt 内容、rubric schema、版本表：归 [F3 `prompt-rubric-registry`](../prompt-rubric-registry/spec.md)。
- 业务调用现场（哪个 backend / frontend owner 调用哪种 profile）：归各自 spec / plan。
- 外部 AI provider 服务部署、platform secret / Vault / cost cap 策略：归 A4 / E4 / 运维；本 spec 只锁应用侧 registry / profile / provider ref 契约，K8s Secret 是否使用由后续 release ADR 决定。
- Realtime voice 的完整双向协议 adapter、媒体留存与 HTTP wire：归 production voice / practice voice owner；本 spec 当前只打开 STT / TTS 级联底座，realtime profile 继续 fail-closed。
- LLM Judge 的 prompt、rubric、阈值与离线/真实评估集：归 F3；A3 只拥有 judge capability/profile/provider adapter 与调用 meta/fail-closed。
- DB 表本身：归 B4；本 spec 只引用字段名。
- 错误码与跨语言 AI vocabulary：依赖 B1 已落地的 `AI_*` 前缀错误码、AI capability、provider registry 字段名、model profile 字段名与 AI meta 字段名；新增跨边界字面量必须先改 B1。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | AIClient 接口形态 | 业务调用面继续只接受 `profile name`；`Complete` / `Stream` / `Transcribe` / `Synthesize` 按 capability/profile 执行，Tools 通过 `CompletePayload` provider-neutral 字段表达；judge 使用独立 `CompleteJudge` dispatch 到 `judge_compatible`，realtime 继续 fail-closed | 业务代码绝对零厂商 SDK 入侵，judge 流量不伪装成业务 chat |
| D-2 | Provider Registry | A3 owns provider registry schema；tracked registry 只保存 provider ref、protocol、capabilities 与 secret env ref，不保存 secret 明文 | 单一 provider 可作为启动配置，但多 provider / 多能力不需要改业务代码 |
| D-3 | Model Profile 字段集 | profile 使用 `capability` + `provider_ref`；不再把 profile 绑定为全局 provider endpoint 的 route | provider profile 配置漂移可控 |
| D-3a | Model Profile 物理落点 | repo-tracked profile catalog 使用单一 `config/ai-profiles.yaml`，`AI_MODEL_PROFILE_PATH` 表示 catalog 文件路径；不再使用一 profile 一文件目录作为 active truth source | 降低小规模 profile catalog 的文件碎片和审查成本 |
| D-4 | 业务引用形态 | 业务只引用 `model_profile_name`（如 `practice.chat.default` / `report.generate.default`），不引用 provider / model 字符串 | 切换 provider / model = 改 registry/profile，不改业务代码 |
| D-5 | Stub 触发条件 | 仅 `APP_ENV=test`、离线契约测试或显式 mock 场景允许走 stub；非测试本地 app run、未来 staging / prod 必须能通过 registry 解析真实 provider secret，缺失即 fail-fast | 单测稳定、可重放，同时防止本地运行或部署静默假数据 |
| D-6 | Fallback 边界 | Fallback 由 AIClient 在 profile fallback chain 内集中执行，最多 2 跳；业务代码不得自行 retry-with-different-model；provider 自身返回的 fallback meta 也必须纳入同一 chain | 防止业务绕开 cost / rate limit / observability |
| D-7 | 观测埋点强制 | A3 注册 7 个 metric family；每次调用必须产出 run / latency / token / cost 指标 + DB 行 + log；fallback / validation failure 指标只在对应事件发生时递增 | F1 dashboard 可信且 counter 语义正确 |
| D-8 | 隐私字段红线 | log / metric / DB metadata 字段中绝不出现明文 prompt / response；只允许 hash / 长度 / profile | 与 ADR-Q5 / logging 标准对齐 |
| D-9 | OpenAI-compatible API 协议子集 | 当前业务 chat 可执行 Chat Completions + chat streaming SSE + Audio Transcriptions + Chat tool-call wire 子集；provider-specific speech 由 `doubao_speech` / `minimax_speech` 独立实现；judge 通过 capability 隔离的 `judge_compatible` Complete-only adapter 执行，realtime 继续 fail-closed | 主流 chat provider 可即插即用，同时不把 speech / realtime / judge 伪装成同一业务 wire |
| D-9a | 当前开发 provider / model | 当前开发主力 provider ref 为 `deepseek`；chat profile 的 model ID 集合固定为 `deepseek-v4-flash` / `deepseek-v4-pro`，其他 alias 由配置 lint 拒绝 | 本地开发与未来部署前的 AI 调用口径稳定且可审计 |
| D-9b | OpenAI-compatible SDK 边界 | `openai_compatible` / `judge_compatible` adapter 内部固定使用 `openai-go/v3 v3.43.0` 的 Chat Completions、streaming 与 Audio Transcriptions client；自定义 base URL、注入 HTTP client、同 provider retry 和兼容扩展字段均封装在 adapter 内。业务包、AIClient 接口、profile、meta、schema validator 与跨 provider fallback 不 import 或暴露 SDK 类型 | 删除自维护通用 OpenAI wire，同时保留 provider-neutral 业务边界、DeepSeek `thinking`、响应上限、错误映射、隐私和可测试性 |
| D-10 | F3 profile 覆盖 | F3 baseline feature_key 必须全部能解析到 A3 profile catalog；P1/P2 capability 可先以 `status=disabled` / `status=unsupported` fail-closed profile 登记，并写明 `unsupported_reason`，但不得缺命名空间 | 业务域开工前具备完整 AI 调用坐标 |
| D-11 | Product/UI capability inventory | A3 spec 必须维护产品 / UI AI 场景到 capability family 的映射；新增 AI 场景必须先修订本表与 F3 feature_key / profile 字典 | 防止新业务回到单模型假设 |
| D-12 | B1 AI vocabulary 边界 | `chat/stt/tts/realtime/judge` capability、provider registry/profile 字段名、AI meta 字段名与 provider/profile routing `AI_*` 错误码由 B1 生成；A3 只 alias / consume，不私造跨边界常量 | 防止 Go/TS/OpenAPI 与 runtime 常量漂移 |
| D-13 | Provider-side streaming consumer | A3 固定消费 OpenAI-compatible SSE `data:` frames 并映射为 `AIStreamEvent` 的 `delta` / `error` / `done`；context cancel 以 B1 `AI_*` 错误码和 `partial_meta_reason` 形成 terminal event；业务 HTTP SSE / chunked wire 由 backend / frontend owner 在自身 API plan 决定 | 后续业务域可复用 provider streaming 底座，同时不提前承诺用户可见 HTTP wire |
| D-14 | 电话模式级联语音底座 | P0 电话模式优先采用 `stt -> chat -> tts` 级联方案替代 S2S / realtime voice；`stt`、`chat`、`tts` 必须是独立 profile，可选择不同 provider，不绑定同一家供应商 | 降低成本并保留 provider 切换能力 |
| D-15 | TTS capability | `tts` 是独立 capability，不混入 `realtime`；TTS 失败只能影响语音播放，不得丢失已生成文本回复或用户 transcript | 防止把低成本级联语音误标为 realtime S2S |
| D-15a | Active profile minimum output budget | 六个 active profile `judge.default`、`practice.chat.default`、`report.generate.default`、`resume.parse.default`、`resume.tailor.default`、`target.import.default` 的 `max_tokens` 均不得低于 16,384；typed code defaults 当前使用 16,384。 | canonical coverage lint 只锁定 active 集合与 16K floor；profile loader owner package 保留一处 default / override / invalid 契约，不在 composition、domain 或 scenario 层复制配置传播测试。 |
| D-16 | Report generation profile contract | `report.generate.default` 保持 `capability=chat`、1M context default、至少 16K output budget 与 `thinking=disabled`；`response_format` 仍只由调用 payload 的 object `output_schema` 驱动。profile loader 对缺失字段使用 typed code default，显式非法值 fail-closed。 | 配置合法性由 loader owner 契约与 active-budget floor lint 承接；bytes 与 tokens 不直接相加，不设离线容量公式或真实 provider 配置 gate。四个 provider adapter 的 response body 统一由 A4 `ai.maxResponseBodyBytes` 注入。 |
| D-17 | Context-aware judge final-content contract | `judge.default` 使用 judge capability、至少 16K output budget、non-thinking JSON wire 与 fail-closed 空 final-content 处理。 | adapter contract test 验证 wire 和 reasoning-only/empty final failure，只保留脱敏 finish/token/presence 元数据，不以 exact profile coordinate 或真实 provider smoke 作为完成 gate。 |

### 3.2 待确认事项

- `model_profile_version` 是否独立 SemVer vs 与 prompt_version 联动：默认独立 SemVer（profile 升级不必随 prompt），由 F3 在自己的 plan 里决定如何引用。
- Stream 暴露到 HTTP 时采用 SSE 还是 chunked：A3 provider 侧固定消费 OpenAI-compatible SSE；业务 HTTP wire 仍由后续 backend / frontend owner 在自身 API plan 决定。
- MiniMax 是否承担 STT：当前只把 MiniMax 作为优先 TTS provider 候选；若后续 MiniMax STT 接口、权限、价格和契约测试确认，再通过 plan 004 或后续增量修订打开。
- Judge 已锁定为 capability 隔离的 `judge_compatible` Complete-only protocol；F3 owns judge prompt/rubric/阈值，A3 owns profile、wire、meta 与 fail-closed。后续若切换到其他 judge provider protocol，必须递增 A3/F3 合同。

## 4 设计约束

### 4.1 接口约束

- `AIClient.Complete` 的入参 `payload` 必须包含 `messages[]` + `metadata`（业务侧的 `feature_key` / `prompt_version` / `rubric_version` / `language`，可选 `output_schema`）；client 不直接接受裸 prompt 字符串。OpenAI-compatible provider 在存在 `output_schema` 时必须请求 JSON object response mode，并继续用本地 schema validator 做二次校验。
- `CompletePayload.tools[]` 只表达 OpenAI-compatible tool schema 的 provider-neutral 子集：`name` / `description` / JSON schema `parameters`；`tool_choice` 只允许 `auto` / `none` / 指定 tool name。`CompleteResponse.tool_calls[]` 只返回 tool name 与 arguments JSON，业务不得读取 provider 私有字段。
- `openai-go/v3` 只允许由 `providers/openai_compatible`、`providers/judge_compatible` 及两者共享的 provider-internal helper import；任何 public A3 type、业务 package、profile/config schema、log/metric/DB/audit contract 都不得暴露 SDK type。当前继续使用 Chat Completions，不迁移到 Responses API。
- `Transcribe` 的入参 `audio` 固定为内存字节 + filename + content type + 可选 language / prompt，provider adapter 以 multipart/form-data 调 `/v1/audio/transcriptions`；原始音频、转写全文和 tool args 明文不得写入 log / DB metadata / metric label。
- `Synthesize` 的入参固定为文本 + voice / format / speaking rate 等 provider-neutral 参数 + metadata；provider adapter 返回音频 bytes 或 chunk metadata。TTS 输入文本和输出音频不得以明文写入 log / DB metadata / metric label。
- `AICallMeta` 字段顺序固定：`provider` / `model_family` / `model_id` / `capability` / `prompt_version` / `rubric_version` / `model_profile_name` / `model_profile_version` / `language` / `input_tokens` / `output_tokens` / `cost_usd_micros` / `latency_ms` / `fallback_chain[]` / `route` / `validation_status` / `error_code` / `tool_invocations[]` / `partial_meta_reason`。其中跨 Go/TS/OpenAPI 边界消费的 capability、profile/provider 字段名、fallback label 字段、tool/partial meta 字段与错误码由 B1 生成；A3 owns runtime 填充与校验。
- `Stream` 返回 `AIStreamEvent` channel，event type 固定为 `delta` / `error` / `done`；`delta` 只携带结构化增量，`error` 携带 B1 错误码，`done` 携带最终 `AICallMeta`。`Stream` 必须消费 provider-side SSE，支持 context cancellation，并在取消时尽力填充 partial meta；业务 HTTP wire 由后续 backend/frontend owner 自行决定。
- 不支持的 capability 必须 fail-closed：profile 能加载为 `disabled` / `unsupported` 状态，且必须携带 `unsupported_reason`；运行时调用不得静默降级到 chat 模型或 stub。
- 电话模式级联语音上下文提交必须由业务 owner 记录已播放边界：AI 文本回复与 TTS audio chunk 先保持 draft，只有前端确认完整播放的 chunk 才能进入下一轮 prompt；被打断后未播放内容必须丢弃，不得写入 committed context。

### 4.2 路由与 fallback 约束

- Provider ref 是 registry 内唯一名；Model Profile 只能引用 registry 中已存在且声明了对应 `capabilities[]` 的 provider ref。
- Registry 中的 `base_url_env` / `api_key_env` 只是 secret 名称；loader 从 A4 SecretSource 解析实际值。tracked YAML 不得出现真实 API key。
- Profile fallback chain 由 A3 client 执行或合并 provider 返回的 fallback meta；每次 fallback 记录 `from_provider/from_model` 与 `to_provider/to_model`，最多 2 跳。
- SDK 只负责同一 provider endpoint 内的瞬时连接错误、408、409、429 与 5xx 重试，固定最多 2 次且受 profile `timeout_ms` 的单一 context 总预算约束；跨 provider/model fallback 仍只由 AIClient profile chain 执行，业务代码不得感知或叠加重试。
- A/B / 用户分桶仍由 PostHog feature flag 在业务 / F3 Resolve 层决定，不塞入 AIClient；AIClient 只接收最终 profile name。
- `timeout_ms` 是 client 总超时（含网络 + provider 排队 + provider 推理），到期后客户端必须 return `AI_PROVIDER_TIMEOUT`，不能让 ctx 永久挂起。

### 4.3 观测与隐私约束

- 每次调用产生的 log（事件名 `ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`）必须遵守 AI Log 字段集；明文红线见 §1 目标 5。
- DB `ai_task_runs.metadata` 仅允许写入摘要字段（hash / 长度 / profile / provider ref / capability）；原始响应如需保存必须落对象存储并由业务隐私策略覆盖，不能写 PG metadata。
- `audit_events.action='ai.call'` 必须由 client 内部写入，业务代码不得跳过。

### 4.4 测试约束

- Stub provider 的输入 → 输出映射必须 deterministic（相同 input + profile 永远产出相同 output）；不依赖时间 / 随机数。
- 任何单元测试默认走 stub；不允许某测试在本地测试或未来远端 CI 中悄悄打到真 provider。
- 任何非测试本地 app run、未来 staging / prod 部署都不得在被选中的真实 provider secret 缺失时静默回退到 stub；启动期 config validation 必须直接失败。
- Registry / profile loader 必须有负向 fixture：未知 provider ref、capability 不匹配、secret env 缺失、unsupported capability 被调用、profile fallback 超 2 跳。
- 六个 active profile 的 `max_tokens` 不得低于 16,384；canonical coverage lint 只锁定 active 集合与 floor。纯配置测试只在 profile loader owner package 保留一组 default / override / invalid 表驱动契约，不在 bootstrap、domain、frontend 或 scenario 层复制同值传播断言。
- `report.generate.default` 必须显式使用 `thinking=disabled`，1M context 与 16K output 使用 typed default；不通过 bytes+tokens 算术、默认尺寸材料、exact-profile lint 或真实 provider smoke 证明配置容量。四个 provider adapter 的 response body cap 必须消费 A4 注入，禁止 adapter-local hardcode。
- `judge.default` 的 adapter request contract test 必须证明 non-thinking + JSON object wire；reasoning-only / empty-content 响应只返回脱敏 finish/token/presence 元数据并以 `AI_OUTPUT_INVALID` fail-closed。不要求真实 provider smoke 作为完成条件。
- SDK import boundary test 必须证明 `openai-go/v3` 只存在于 A3 provider adapter/internal helper，且请求/响应日志、错误包装与 debug hook 不泄漏 API key、prompt、response、audio、tool arguments 或 reasoning content。

### 4.5 Product/UI AI Capability Catalog

| 产品 / UI 场景 | 主要输入 | Capability family | 默认 profile 命名 |
|----------------|----------|-------------------|-------------------|
| JD 导入解析 | JD 文本 / URL 提取文本 | `chat` 结构化抽取 | `target.import.default` |
| 公司轻情报摘要 | source-grounded public info | `chat` source-grounded summarization | `target.intel.default`（P1/P2 fail-closed） |
| 简历解析 | 简历文本 / 上传解析结果 | `chat` 结构化抽取 | `resume.parse.default` |
| 简历定制 / bullet 改写 | JD + 简历证据 | `chat` 写作 / 改写 | `resume.tailor.default` |
| 模拟面试连续聊天 | JD / round / resume / ordered messages | `chat` 对话生成 | `practice.chat.default` |
| 电话模式 | 当前暂时禁用 | `stt` / `tts` / `realtime` 均 fail-closed | `practice.voice.stt.default` / `practice.voice.tts.default` disabled；`practice.voice.realtime.default` unsupported |
| 报告生成 | full session + JD + resume | `chat` 长上下文结构化推理 | `report.generate.default` |
| 会话级报告 | frozen JD/resume/round context + ordered transcript | `chat` report generation | `report.generate.default` |
| 复练当前轮 / 下一轮 | report competency gaps + round context | `chat` 生成 | `report.generate.default` / `practice.chat.default` |
| 离线 LLM Judge / eval | prompt output + rubric | `judge` | `judge.default`（F3 eval） |

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `backend/internal/ai/aiclient/` 接口与默认实现 | A3 | `AIClient` / `AICallMeta` / registry loader / profile loader / stub / openai_compatible adapter；`openai-go/v3` 仅为 adapter 私有实现依赖 |
| Provider Registry schema | A3 + A4 | A3 owns schema / validation；A4 owns path/env/SecretSource 注入 |
| Model Profile schema | A3 | `config/ai-profiles.yaml` catalog schema 与热加载 |
| Profile 文件内容 | F3 + 各 AI feature owner | F3 owns feature_key -> model_profile_name；A3 owns profile schema；业务 owner 负责新增场景时补 profile |
| Profile 文件路径 / secret 注入 | A4 | `AI_PROVIDER_REGISTRY_PATH` / `AI_MODEL_PROFILE_PATH` 与 provider-specific env secret ref |
| 真实 provider endpoint | A3 + A4 + E4/运维 | 非测试本地 app run 可直连真实 AI provider；未来 staging / prod 可接运维提供的 provider endpoint；本 spec 不部署独立代理 |
| 业务调用现场 | `backend-targetjob` / `backend-practice` / `backend-review` / `backend-resume` / future retrieval / production voice owners | 各业务 spec / plan 引用 profile name，不引用 provider/model；`backend-debrief` 按 product-scope D-22 不在当前范围 |
| 共享约定 | B1 | `AI_*` 错误码、AI capability、provider registry/profile 字段名、AI meta 字段名共享常量、`ApiError` / `ApiErrorResponse` 消费约定 |
| DB 表 | B4 | `ai_task_runs` schema |
| Metric / Dashboard | F1 | 7 个 ai_* metric + AI Cost & Quality Dashboard；AI metric label 使用 `capability`，任何 label 变更必须先由 F1 spec / plan 承接 |
| 测试 stub provider | A3 | 应用内 deterministic stub，仅供单元测试 / 离线契约测试 / 显式 mock 场景 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Stub 单测 | 单测环境（`APP_ENV=test`，无真实 provider secret） | 业务代码调用 `aiclient.Complete(ctx, "practice.chat.default", payload)` | client 路由到 stub provider；返回结构化 response + meta；`meta.provider == "stub"`；同 input 多次调用结果一致 | 001 |
| C-2 | Registry provider route | registry 中 `deepseek` 声明 `chat` capability，并引用 env secret | 调用 `Complete` / `Stream` | 出站 HTTP 请求命中该 provider ref 的 OpenAI-compatible endpoint；header 含 `Authorization`；`meta.provider` / `meta.capability` / `meta.model_profile_name` 正确；模型 ID 只使用 `deepseek-v4-flash` / `deepseek-v4-pro` | 003 + 002 |
| C-3 | Central fallback | profile 声明 primary + fallback provider ref，primary 超时且 fallback 成功 | 调用 `Complete` | AIClient 执行受限 fallback，`fallback_chain` 记录 provider/model hop；`ai_fallback_total` +1；业务代码无 retry-with-different-model 循环 | 003 |
| C-4 | Registry + profile 热加载 | A3 loader 已启动 | `config/ai-providers.yaml` 或 `config/ai-profiles.yaml` 修改后保存 | client 在 ≤ 30s 内热加载；正在进行的调用使用前一快照完成；新调用使用新快照 | 003 |
| C-5 | 观测埋点齐全 | 任一无 fallback、无 validation failure 的调用完成 | F1 metric / log / DB 三方查询 | 7 个 metric family 均已注册；run / latency / token / cost 指标增长；fallback / validation failure counter 不增长；`ai_task_runs` + `audit_events` 各写一行，无明文 | 001 + 003 |
| C-6 | 隐私红线 | grep 全部生产代码与 log | 任意调用 | 不出现 `payload.messages[*].content` / `response.content` 明文落 log 或 DB metadata；hash / 长度 / profile 三类摘要必须出现 | 001 + 003 |
| C-7 | 错误码合规 | provider 返回结构化输出非法 | client `validate_output` 失败 | 返回错误码 `AI_OUTPUT_INVALID`；`ai_output_validation_failures_total` +1 | 001 |
| C-8 | active spec relation gate | 本 spec 通过 `/plan-review` | 与当前 active spec 和 future workstream 关系审查 | A3 与 F3 / B1 / A4 / F1 / release gate 引用关系自洽；provider 只通过 registry ref 表达，A3 不拥有独立 provider-proxy 业务语义 | plan-review |
| C-9 | Registry secret fail-fast | 非测试本地 app run、未来 staging / prod 缺失 registry 选中 provider 的 base URL 或 API key | 启动 backend runtime | 进程启动失败并报配置错误；不得自动回退到 stub provider | 003 + A4 |
| C-10 | F3 baseline profile coverage | F3 6 个 baseline feature_key 已定义默认 profile name | 运行 profile coverage lint | 每个默认 profile 在 `config/ai-profiles.yaml` catalog 中存在，且 capability / provider_ref / status 合法；允许 P1/P2 profile `disabled` / `unsupported`，但必须携带 `unsupported_reason` 且不得缺 catalog entry | 003 + F3 |
| C-11 | Product/UI capability inventory drift | 新增 AI 场景或 UI 交互依赖 AI | `/plan-review` 或 lint 检查 | 本 spec §4.5、F3 feature_key 字典与 A3 profile catalog 同步更新；不得只在业务代码 hardcode 新 profile | 003 + F3 |
| C-12 | Unsupported capability fail-closed | profile 使用未激活的 `realtime` 或 disabled/unsupported speech capability | 运行时调用该 profile | 返回明确 unsupported capability 错误并记录 meta/log；不得降级到 chat 或 stub；对应 UI 能力必须 feature-gated | 003 + 002 |
| C-13 | Tool call provider-neutral | profile 使用 `chat` capability 且 payload 携带 `tools[]` / `tool_choice` | 调用 `Complete` | openai_compatible adapter 映射 tool wire；响应返回 `tool_calls[]` 与 `finish_reason=tool_calls`；`AICallMeta.tool_invocations[]` 只含 tool name / argument hash / argument length，不含 args 明文 | 002 |
| C-14 | Active profile configuration | `config/ai-profiles.yaml` 声明当前 active profiles | 运行 loader owner contract 与 coverage lint | active 集合完整、`max_tokens >= 16384`，显式非法配置 fail closed；不锁 exact profile 坐标或真实 provider smoke | 003 |
| C-14 | Provider-side streaming | profile 使用 `chat` capability | 调用 `Stream` 且 provider 返回 SSE delta / done | channel 按顺序发 `delta`，最终发 `done` 并关闭；malformed chunk / provider error / context cancel 发 `error` 或带 partial meta 的 terminal event，错误码来自 B1 `AI_*` | 002 |
| C-15 | STT transcription | profile 使用 `stt` capability 且 provider ref 支持 OpenAI-compatible Audio Transcriptions | 调用 `Transcribe` | adapter 调 `/v1/audio/transcriptions`；返回 transcript + meta；缺 secret / provider error / unsupported profile fail-fast；log / DB / audit / metric label 不含音频或转写全文明文 | 002 |
| C-16 | TTS synthesis | profile 使用 `tts` capability 且 provider ref 支持 provider-specific synthesis wire | 调用 `Synthesize` | adapter 返回音频 bytes 或 chunk metadata + meta；缺 secret / provider error / unsupported profile fail-fast；log / DB / audit / metric label 不含待合成文本或音频明文 | 004 |
| C-17 | Independent cascaded voice profiles | 电话模式配置分别选择 `stt`、`chat`、`tts` profile | 业务编排一轮 `stt -> chat -> tts` | STT/TTS 可指向不同 provider；TTS 失败不丢失 transcript / chat text；STT 失败不调用 chat/TTS；任何一步不得静默回退到 `realtime` 或 stub | 004 + practice-voice-mvp |
| C-19 | Active profile budgets and response cap | canonical catalog 包含六个 active profile | 运行 coverage floor lint、一处 loader default/override/invalid contract与 shared bounded-reader tests | 六个 active profile `max_tokens >= 16,384`；report 1M context 与 16K output 使用 typed default；不做 bytes+tokens 公式、exact-profile lint 或真实 provider配置 gate；四 adapter 共享注入 response cap | 003 Phase 11 + A4 Phase 13 |
| C-20 | Judge final-content reliability | context-aware judge 需要 strict JSON | 运行 adapter request/response contract tests | 请求关闭 thinking 并要求 JSON object；reasoning-only/length/empty final fail-closed 且不泄漏 CoT | 003 + F3/004 |
| C-21 | Official SDK transport boundary | OpenAI-compatible chat/judge/STT provider 使用自定义 base URL 与注入 HTTP client | 运行 adapter contract、SDK import boundary 与隐私测试 | Complete/Stream/Transcribe/tools/JSON/DeepSeek thinking/header/meta/error/timeout/response-cap 合同保持；同 provider retry 最多 2 次，跨 provider fallback 仍归 AIClient；`openai-go/v3` 不越过 provider adapter/internal helper | 001 Phase 15 |
| C-18 | Barge-in committed context | AI TTS 正在播放且用户插话 | 前端发出 barge-in / played chunk 事件 | 后端只把已完整播放 chunk 的 assistant 文本写入 committed context；未播放 draft 不进入下一轮 prompt；event log 可追溯 interrupted 状态 | practice-voice-mvp |

## 7 关联计划

A3 当前计划已完成 foundation transport migration 与 capability/speech foundation：

- [001-aiclient-and-profile-bootstrap](./plans/001-aiclient-and-profile-bootstrap/plan.md)（completed）：在既有 `AIClient`、profile、fallback、meta、observability 与 privacy 合同不变的前提下，以官方 `openai-go/v3` 替换 `openai_compatible` / `judge_compatible` 自维护通用 wire。
- [002-tools-streaming-and-stt](./plans/002-tools-streaming-and-stt/plan.md)（completed）：已落地 Tools payload 扩展、provider-side streaming consumer 与 STT Audio Transcriptions 底座；realtime multimodal 仍保持 fail-closed。
- [003-provider-registry-and-capability-profiles](./plans/003-provider-registry-and-capability-profiles/plan.md)（completed）：已完成 provider registry、capability profile、DeepSeek V4 Flash/Pro、typed defaults、thinking 与 shared response-cap 合同。
- [004-cascaded-speech-provider-foundation](./plans/004-cascaded-speech-provider-foundation/plan.md)（completed）：已完成 `stt -> chat -> tts` provider-specific speech foundation；realtime S2S 继续 fail-closed。

后续如需扩展，递增本 spec 版本并原地修订对应 plan；不创建 sibling spec。
