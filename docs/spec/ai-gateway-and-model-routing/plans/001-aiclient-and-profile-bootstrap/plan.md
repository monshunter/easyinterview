# AI Gateway and Model Routing Bootstrap

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-04

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ai-gateway-and-model-routing spec](../../spec.md#21-in-scope) §2.1 In Scope 与 §7 关联计划列出的 P0 范围一次性落地：在 `backend/internal/ai/aiclient/` 写出 provider-neutral 的 `AIClient` 接口（`Complete` / `Embed` 同步面 + `Stream` 事件合同类型）、A3-owned `AICallMeta` 运行时结构体、`stub` 与 `openai_compatible` 两个 provider、Model Profile YAML schema + loader + ≤30 秒热加载、client-internal observability / audit decorator（7 个 `ai_*` metric family + 4 类结构化事件 + `ai_task_runs` 行 + `audit_events` 行）、配置启动校验 fail-fast，以及覆盖 stub 路径的单元测试和可被 [E1 mock-contract-suite](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 复用的离线 adapter 契约测试，最终通过本 plan Phase 5 的本地命令证明 spec [§6 验收标准](../../spec.md#6-验收标准) 中 C-1 / C-2 / C-3 / C-4 / C-5 / C-6 / C-7 / C-9 全部成立。AC C-8 是当前 active spec relation gate，已由 engineering-roadmap 保留 A3 与 F3 / B1 / A4 / F1 / release gate 的边界关系，本 plan 仅引用不替代。

本 plan 的写入边界是 `backend/internal/ai/aiclient/` 与 `config/ai-profiles/` fixture。`backend/cmd/api` / `backend/cmd/worker` 运行时 entrypoint 由 A4 / C 域在自身 plan 中接入，本 plan 只提供 AIClient 构造、配置校验 API 与 DI handoff，不创建或重写 API/worker main。

本 plan 不实现 `Stream` 的完整 provider 消费循环（事件类型在本 plan 冻结，完整流式由 002+ 承接，对应 [ADR-Q6 §3.1](../../../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md#3-决策) D-1）；不实现 `Tools(...)` / function calling（[spec §3.2](../../spec.md#32-待确认事项)）；不实现 Audio Transcription `/v1/audio/transcriptions` 与 `Transcribe(...)`（C14 P2 预留，[spec §2.2 / D-8](../../spec.md#22-out-of-scope)）；不实现真实 gateway 部署 / cost cap / rate limit policy（归 E4 + 运维）；不维护具体 prompt / rubric 内容（归 [F3 prompt-rubric-registry](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec)）；不维护 `ai_task_runs` 表 schema（归 [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md)）；不接入远端 CI required check（A5 deferred-CI 边界，本 plan 全部 gate 在本地 `make` / `go test` 命令运行）。后续如需扩展 Streaming 完整化、function calling、STT，则递增本 spec 与本 plan 版本，原地修订或 spawn `002-tools-streaming-and-stt`。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 把 A3 列为 Layer A · Foundation 第 3 份 child，依赖 [A1 repo-scaffold](../../../repo-scaffold/spec.md) 提供的 `backend/` / `config/` 根容器与 root `Makefile`，间接依赖 [B1 shared-conventions-codified](../../../shared-conventions-codified/spec.md) 已落地或已计划落地的 `AI_*` 错误码、Model Profile / AI meta 字段名共享常量与 `ApiError` / `ApiErrorResponse` 结构。执行本 plan 前必须确认 A1 `001-bootstrap` v1.1 已提供 `config/` 根容器；B1 必须至少提供 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` 三个错误码常量。若 B1 002 的 AI vocabulary 字段名尚未落地，A3 001 可在 A3 包内使用私有字段名校验表，但不得把这些字段导出为跨语言常量；B1 002 完成后再切换 import。

每个 phase 是可独立验证的纵向切片：Phase 1 起来即可 `go test ./backend/internal/ai/aiclient/...` 走 stub 路径；Phase 2 起来即可对 OpenAI-compatible mock server 跑契约测试；Phase 3 起来即可在测试中验证 7 个 metric family 注册、`ai_task_runs` 写入、`audit_events` hash + 长度落盘；Phase 4 起来即可在缺失 `AI_GATEWAY_*` 时 fail-fast；Phase 5 收口 8 项 AC 自检 + grep 红线 + 文档与 INDEX 同步。本 plan 不引入 BDD 资产（场景覆盖由后续 [e2e-scenarios-p0](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) workstream 承接），所有 AC 验证完全由 `go test` / `go build` / 本地 `make` / `grep` / 配置注入 + 启动检查驱动。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + platform-foundation`。本 plan 修改 A3 AIClient runtime package、OpenAI-compatible adapter、profile loader、observability decorator、config fail-fast 与测试资产；不引入用户可感知 UI、HTTP API 行为或端到端业务流程。
- **TDD 策略**: 历史实现通过本 checklist 的每个 phase `自检` / `验证` 子句驱动 Red-Green-Refactor；重进本 plan 时必须通过 `/implement` -> `/tdd` 顺序执行，focused assertions 来源为 Go package tests、OpenAI-compatible mockserver contract tests、profile loader / config / privacy tests、grep 红线与 `make` gate。
- **BDD 策略**: BDD 不适用。本 plan 是内部 AI gateway client / profile / observability 契约交付，不产生浏览器 UI、外部 API 行为或用户业务工作流；后续 P0 用户行为由具体 C/D/E workstream 维护 BDD gate。
- **替代验证 gate**: `go test ./backend/internal/ai/aiclient/...`、OpenAI-compatible mockserver contract tests、profile hot-reload tests、privacy redaction tests、config fail-fast tests、零厂商 SDK grep、明文 prompt/response grep、`go build ./...`、`sync-doc-index --check`。

## 4 实施步骤

### Phase 0: 前置契约复核

#### 0.1 A1 `config/` 根容器复核

确认 [repo-scaffold 001-bootstrap](../../../repo-scaffold/plans/001-bootstrap/plan.md) 已按 v1.1 提供 `config/README.md` 与根 README 索引；若缺失，先走 A1 Phase 4 remediation，禁止 A3 在仓库根另起 `ai-profiles/` 或其它平行配置目录。

#### 0.2 B1 AI vocabulary 复核

确认 B1 已提供 `AI_*` baseline 错误码。Model Profile / AI meta 字段名若尚未由 B1 002 生成，A3 001 仅在 `backend/internal/ai/aiclient/` 内部维护 runtime 字段表，并在 Phase 5 handoff 中列出需切换到 B1 生成常量的字段清单；不得在 A3 下创建跨语言 shared 常量。

### Phase 1: AIClient 接口 + stub provider 骨架

#### 1.1 包骨架与对外类型

在 `backend/internal/ai/aiclient/` 下落地 Go 包骨架：`aiclient.go`（接口定义）、`meta.go`（`AICallMeta` 与 stream event 类型）、`payload.go`（`Complete` / `Embed` 入参与响应结构）、`profile.go`（Model Profile schema 类型）、`doc.go`（包注释，明确 provider-neutral 与零厂商 SDK 红线）。`AIClient` interface 至少声明 `Complete(ctx context.Context, profileName string, payload CompletePayload) (CompleteResponse, AICallMeta, error)` / `Embed(ctx context.Context, profileName string, input EmbedInput) (EmbedResponse, AICallMeta, error)` / `Stream(ctx context.Context, profileName string, payload CompletePayload) (<-chan AIStreamEvent, error)` 三个方法签名，对应 [spec D-1 / §4.1](../../spec.md#41-接口约束)。`Complete` 入参强制要求 `payload.Messages[]` + `payload.Metadata`（`feature_key` / `prompt_version` / `rubric_version` / `language` / 可选 `output_schema`），不接受裸 prompt 字符串；如业务侧传空 `messages`，client 必须返回 `AI_OUTPUT_INVALID`。`AIStreamEvent` 联合体仅声明 `delta` / `error` / `done` 三种 type 与对应 payload 字段，`done` 携带最终 `AICallMeta`，`error` 携带 B1 错误码字符串；本 plan 只冻结类型与 channel close 语义，不实现 provider 流式消费循环。

#### 1.2 `AICallMeta` 字段顺序与填充契约

`AICallMeta` 结构体字段顺序固定为 `Provider` / `ModelFamily` / `ModelID` / `TaskType` / `PromptVersion` / `RubricVersion` / `ModelProfileName` / `ModelProfileVersion` / `Language` / `InputTokens` / `OutputTokens` / `CostUSDMicros` / `LatencyMs` / `FallbackChain[]` / `Route` / `ValidationStatus` / `ErrorCode`，与 [spec §4.1](../../spec.md#41-接口约束) / [ADR-Q6 §3.1](../../../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md#3-决策) 一致并补齐 [B4](../../../db-migrations-baseline/spec.md) / [F1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 消费字段。每个字段引用 [B1 shared-conventions-codified](../../../shared-conventions-codified/spec.md) 的共享常量名（`task_type` 取自 B1 `TaskType` enum，错误码取自 `errors.Code`）。`AICallMeta` 由 client 填充，业务代码不能伪造；client 内部封装一个 `metaBuilder` helper，把 profile 解析结果、provider 返回 meta 与 fallback meta 合流，并校验字段顺序与必填字段。新增字段必须先递增本 spec 版本；如需跨前后端共享则同步追加到 B1。

#### 1.3 stub provider 实现

在 `backend/internal/ai/aiclient/providers/stub/` 落地 deterministic stub：输入 → 输出走 hash-based 映射（`sha256(profile + canonical(payload))` 截前 N 字节作为 RNG seed），结果可被 [E1 mock-contract-suite](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) OpenAPI fixtures 反向喂养（[spec §2.1 stub 段](../../spec.md#21-in-scope)）；不依赖时间 / 随机数，相同 input + profile 永远产出相同 output（[spec §4.4 测试约束](../../spec.md#44-测试约束)）。`stub` provider 的入口 factory 在初始化时检测 `APP_ENV`，仅 `APP_ENV=test` 或显式 `WithStubAllowed(true)` 选项允许实例化；其它环境直接返回 `AI_PROVIDER_TIMEOUT` 之外的明确启动错误（仅 in-memory，不混入 client meta）。`Complete` 与 `Embed` 各自返回固定形状的 stub response（带 deterministic token 数与 latency 占位值），`Stream` 仅返回一次 `done` 事件 + 关闭 channel，作为 002+ 完整化前的兼容 shim。

#### 1.4 包级单元测试（stub 路径）

在 `backend/internal/ai/aiclient/aiclient_test.go` 与 `providers/stub/stub_test.go` 落地最小单测：调用 `Complete(ctx, "practice.followup.default", payload)` 时 client 路由到 stub；返回结构化 response + meta；`meta.Provider == "stub"`；同 input 多次调用产出一致 response 与一致 meta；`payload.Messages` 为空时返回 `AI_OUTPUT_INVALID`（[spec C-1 / C-7](../../spec.md#6-验收标准)）。`Stream` 路径只断言 channel 收到一次 `done` 事件且关闭，不断言 `delta` 顺序（留给 002+）。

### Phase 2: openai_compatible provider + Model Profile loader

#### 2.1 OpenAI-compatible Chat / Embeddings adapter

在 `backend/internal/ai/aiclient/providers/openai_compatible/` 落地 P0 协议子集 adapter：`Complete` 走 `POST {AI_GATEWAY_BASE_URL}/v1/chat/completions`，`Embed` 走 `POST {AI_GATEWAY_BASE_URL}/v1/embeddings`，request header 含 `Authorization: Bearer ${AI_GATEWAY_API_KEY}` / `Content-Type: application/json` / `X-Request-ID` 透传；response 解析 `usage.prompt_tokens` / `usage.completion_tokens` / `model` / 任何 gateway-extension 字段（`x-fallback-from` / `x-fallback-to` / `x-route`），由 client 填入 `AICallMeta`（[spec C-2 / C-3](../../spec.md#6-验收标准)）。Audio Transcription `/v1/audio/transcriptions` 不进入本 plan 验收（[spec D-8](../../spec.md#31-已锁定决策)）。adapter 内部使用标准库 `net/http` + `encoding/json`，不允许 import 任何厂商 SDK（`openai-go` / `anthropic-sdk-go` / `cohere-go` 等），grep gate 在 Phase 5 再验证。`timeout_ms` 用 `context.WithTimeout` 封装，超时返回 B1 `AI_PROVIDER_TIMEOUT`（[spec §4.2](../../spec.md#42-路由与-fallback-约束)）；非 2xx 响应按 B1 错误码映射（5xx → `AI_PROVIDER_TIMEOUT` / 4xx → 解析 body 中 error_code 字段，未识别落到通用 wire error）；fallback meta 仅消费 endpoint / gateway 返回的 `x-fallback-*` header 或 body 字段，A3 client 绝不自行重试切换 model（[spec D-5 / §4.2](../../spec.md#41-接口约束)）。

#### 2.2 Model Profile schema 与 YAML loader

在 `backend/internal/ai/aiclient/profile/` 落地 Model Profile schema 类型与 YAML loader：字段集严格对应 [spec §2.1 Model Profile schema](../../spec.md#21-in-scope) / [ADR-Q6 §3.1 D-2](../../../engineering-roadmap/decisions/ADR-Q6-ai-gateway-and-model-routing.md#3-决策)，包含 `name` / `task_type`（`chat` | `embed` | `stt`，`stt` 仅作为 schema 兼容预留值，loader 解析 OK 但 client 在 task_type=stt 调用时返回明确 not-implemented 错误而非走真实 adapter）/ `default.{provider, model, params}` / `fallback[]`（按序触发条件）/ `timeout_ms` / `max_tokens` / `rate_limit.{rps, tpm}` / `gateway_route` / `version`。loader 读取 `AI_MODEL_PROFILE_PATH` 指向的目录（由 [A4 secrets-and-config](../../../secrets-and-config/spec.md) 注入）下的 `*.yaml` 文件，使用 `gopkg.in/yaml.v3` 反序列化；解析失败立即报错并附 file path + line number。Profile 文件落点 `config/ai-profiles/*.yaml`（[spec §2.1](../../spec.md#21-in-scope)），本 plan 落地 `config/ai-profiles/practice.followup.default.yaml` 与 `config/ai-profiles/review.report.default.yaml` 两个最小 fixture profile，仅供本 plan 测试与本地验证使用；真实 profile 内容由 F3 维护。

#### 2.3 ≤30 秒热加载

loader 启动一个内部 goroutine 监听 `AI_MODEL_PROFILE_PATH` 目录变更（首选 `fsnotify`；如平台不支持则降级到 30 秒 polling timer，但本 plan 默认走 fsnotify 并在 README 写明 fallback），变更后在 ≤30 秒窗口内重新解析整个目录并替换内部 `map[name]*Profile`（[spec C-4](../../spec.md#6-验收标准)）。替换使用 atomic store + RW mutex，确保正在进行的调用使用旧 profile 完成、新调用使用新 profile（不允许同一调用中途切换 profile）。loader 暴露 `Reload(ctx) error` 测试入口，便于单测显式触发 reload；hot-reload race 在 Phase 5 通过并发测试覆盖。

#### 2.4 离线契约测试（mock server）

在 `backend/internal/ai/aiclient/providers/openai_compatible/contract_test.go` 落地离线契约测试：使用 `httptest.NewServer` 启动一个 OpenAI-compatible mock server，覆盖 chat / embeddings 两类正常响应、超时模拟、5xx 错误、fallback meta 注入（response header / body 中带 `x-fallback-from` / `x-fallback-to`）。测试断言 client 能正确解析 `usage.*tokens` 入 `AICallMeta`、能透传 fallback chain、超时返回 `AI_PROVIDER_TIMEOUT`、5xx 走 B1 错误码语义且不自行重试 model。mock server 需以可导出 helper 形式提供给 [E1 mock-contract-suite](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 后续复用（导出包路径 `backend/internal/ai/aiclient/providers/openai_compatible/mockserver/`，供测试用）；本 plan 不实现 E1 自身的 fixture 反向喂养，但保持 mock server interface 稳定。

#### 2.5 L2 remediation: OpenAI-compatible BaseURL 归一化

修复 plan-code-review F1：adapter 必须支持 `AI_GATEWAY_BASE_URL` 既传 endpoint root（如 `https://provider.example`）也传已含 OpenAI-compatible API 前缀的形式（如 `https://provider.example/v1`）。离线契约测试需新增 `BaseURL = mockServer.URL() + "/v1"` 的用例，确保实际请求仍命中 `/v1/chat/completions` / `/v1/embeddings`，不会拼出 `/v1/v1/...`。

#### 2.6 L2 remediation: Profile 校验错误行号

修复 plan-code-review F3：`profile.Loader` 除 YAML 语法错误外，手写 schema 校验错误（缺 `name` / `task_type` / `default.provider` / `default.model` / `timeout_ms` / `version`、非法 `task_type` 等）也必须在错误信息中包含 file path + line number，便于 F3 / 运维快速定位 profile 漂移。

#### 2.7 L2 remediation: B1 error-code registry fallback

修复 plan-code-review F3：`openai_compatible` adapter 解析 4xx error envelope 时，只允许透传 B1 `CodeRegistry` 中登记的错误码；未登记的上游 `error.code` 必须降级为 `AI_OUTPUT_INVALID` 或本 plan 已声明的通用 wire error，避免 provider 私有字符串进入 `AICallMeta.ErrorCode`、日志、指标或 API 错误面。新增离线契约测试覆盖未知 4xx code fallback 与已登记 code passthrough 两条路径。

### Phase 3: Observability / audit decorator + DB / log / metric 接入

#### 3.1 Metric 注册（7 个 family）

在 `backend/internal/ai/aiclient/observability/metrics.go` 注册 7 个 metric family（[spec §2.1 / §4.3 / D-6](../../spec.md#21-in-scope)）：`ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total`。label 集与 [F1 observability-stack spec](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 对齐：`provider` / `model_family` / `model_profile_name` / `route` / `task_type` / `language` / `result`（`success` / `failure` / `fallback`）/ `from_model_family` / `to_model_family`（仅 `ai_fallback_total`）。Counter 语义遵守 [spec D-6](../../spec.md#31-已锁定决策)：每次调用 run / latency / token / cost 指标按本次调用增长；fallback / validation failure counter 仅在对应事件发生时递增。Metric 注册采用 `prometheus.Registerer` 抽象，本 plan 默认使用 `prometheus.DefaultRegisterer` 但允许 F1 在自己 plan 里替换为统一 registry；不允许业务调用绕过 decorator 直接访问 raw provider。

#### 3.2 结构化日志 + DB / audit decorator

在 `backend/internal/ai/aiclient/observability/decorator.go` 落地 middleware-style decorator，包裹 `Complete` / `Embed` 同步调用面（`Stream` 完整化日志由 002+ 承接，本 plan 仅在 `done` 事件时触发一次 decorator 收敛）：调用前后写入结构化 log 事件名 `ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`，字段集严格遵守 [05-logging-standard.md §4.4 AI Log 额外字段](../../../../../easyinterview-tech-docs/05-logging-standard.md#44-ai-log-额外字段)（`provider` / `model_id` / `model_profile_name` / `model_profile_version` / `prompt_version` / `rubric_version` / `task_type` / `language` / `input_tokens` / `output_tokens` / `cost_usd_micros` / `latency_ms` / `fallback_chain` / `route` / `validation_status` / `error_code`）。decorator 同时向 [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md) 提供的 `ai_task_runs` 表写入一行（A3 只填 typed columns，不参与 schema 演进；本 plan 通过 DI 注入 `AITaskRunWriter` interface，由调用方在 wire 时绑定真实 store，本 plan 测试用 fake writer）。同一调用还向 `audit_events` 写一行 `action='ai.call'`，`metadata` 字段仅包含 `prompt_hash`（sha256 hex）/ `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`，不包含明文（[spec D-7 / §4.3](../../spec.md#43-观测与隐私约束)）。

#### 3.3 fallback / validation status 语义

decorator 在 `meta.FallbackChain[]` 长度 > 1 时对 `ai_fallback_total{from_model_family, to_model_family, result="fallback"}` 递增一次（[spec C-3](../../spec.md#6-验收标准)）；在 client 内部 `validateOutput` 返回失败时（如调用方提供 `output_schema` 但 response 不符）对 `ai_output_validation_failures_total` 递增一次并发出 `ai.output.validation_failed` 日志，错误码统一为 `AI_OUTPUT_INVALID`（[spec C-7 / B1 错误码](../../spec.md#22-out-of-scope)），不允许 A3 在此处新增错误码。`AI_FALLBACK_EXHAUSTED` 只透传 endpoint / gateway 返回的语义，A3 client 不主动产出该错误（[spec §4.2](../../spec.md#42-路由与-fallback-约束)）。

#### 3.4 隐私红线测试

在 `backend/internal/ai/aiclient/observability/privacy_test.go` 落地白盒测试：构造一组带敏感内容的 `payload.Messages[*].Content` 与 mock response content，跑完 decorator 后断言 metric label values / log fields / DB row metadata / audit_events metadata 均不包含明文（仅可包含 hash 前缀样本、长度数字与 profile 名）。该测试不依赖真实 PG / Prometheus 后端，使用 in-memory writer + log capture。

#### 3.5 L2 remediation: `output_schema` 基础约束校验

修复 plan-code-review F2：当调用方提供 `OutputSchema` 时，decorator 不能只检查 response content 是合法 JSON；必须至少校验 JSON Schema 的基础 `type`、`required`、`properties` 约束。schema 不匹配时统一返回 `AI_OUTPUT_INVALID`、标记 `ValidationStatusInvalid`、递增 `ai_output_validation_failures_total` 并发出 `ai.output.validation_failed` 日志。错误信息不得包含 response 明文内容。

#### 3.6 L2 remediation: B4-compatible `ai_task_runs` row contract

修复 plan-code-review F1：`AITaskRunRow` 必须覆盖 B4 `ai_task_runs` schema 的必填列与 A3 typed columns，包括 `id`、业务 `task_type`、`resource_type`、`resource_id`、`status`、`started_at`、`completed_at` 与 A3 meta 字段；`chat` / `embed` / `stt` 仅保留为 Model Profile call kind，不得写入 B4 `task_type`。调用方通过 `CallMetadata` 提供业务 task/resource context；decorator 在缺失必填上下文或 writer 返回错误时不能静默吞掉错误。新增 focused tests 断言成功调用写出的 fake row 可满足 B4 required columns、业务 task_type 属于 B4 enum，且 writer failure 会返回给调用方。

### Phase 4: 配置校验与本地部署 fail-fast

#### 4.1 配置 struct 与启动期校验

在 `backend/internal/ai/aiclient/config.go` 定义 client 启动时所需配置 struct：`AppEnv` / `GatewayBaseURL` / `GatewayAPIKey` / `ModelProfilePath`，由 [A4 secrets-and-config](../../../secrets-and-config/spec.md) 在 `cmd/api` / `cmd/worker` wire 阶段注入。client 在 `New(cfg)` 时必须执行启动校验：当 `AppEnv != "test"` 且 `GatewayBaseURL == ""` 或 `GatewayAPIKey == ""` 时立即返回明确错误（建议常量名 `ErrMissingGatewayConfig`），由 `cmd/api` / `cmd/worker` 在 main 中转换为 fail-fast exit（[spec D-4 / C-9](../../spec.md#31-已锁定决策)）。`AppEnv == "test"` 路径不强制 GatewayBaseURL / API key，但允许通过 `WithStubAllowed(true)` 选项显式启用 stub；其它任何路径（dev / staging / prod / docker compose / Kind）缺凭证即失败，绝不静默回退到 stub。

#### 4.2 cmd 入口接入 + 单测

落地 `backend/internal/ai/aiclient/config_test.go` 用例：`AppEnv=test` 且 missing gateway + 显式 stub → 成功；`AppEnv=production` 且 missing → 错误；`AppEnv=test` 但 stub 选项未启用且无 gateway → 错误。本 plan 不在 `cmd/api` 与 `cmd/worker` 实现完整 wire，也不要求创建这些 entrypoint；只提供 `New(cfg)` / DI 构造契约，让 [A4](../../../secrets-and-config/spec.md) / 各 C 域在自身 plan 中把 cfg 错误传播为 non-zero exit。

#### 4.3 docker compose / Kind smoke 校验

在 [A2 dev-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) docker compose 配置与 Kind manifest 中确认未默认设置 stub fallback；当 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 缺失时启动失败。本 plan 不修改 A2 / Kind 资产本身（由各自 owner 维护），但需要在 `backend/internal/ai/aiclient/README.md`（包级 README）中明确记录上述 fail-fast 协议和本地部署要求（接真实 OpenAI-compatible LLM 服务），并写出本 plan smoke 验证步骤示意（导出真实 endpoint env 后跑 `go test -tags smoke ./backend/internal/ai/aiclient/...`，本 plan 不开 smoke build tag 的 CI 集成）。本地 smoke 测试时绝不在测试代码或 fixture 中嵌入真实 API key（[spec §7](../../spec.md#7-关联计划)）。

### Phase 5: Verification + handoff

#### 5.1 Spec C-1 / C-2 / C-3 / C-4 自检

在本地依次运行：`cd backend && go test ./internal/ai/aiclient/...`（覆盖 stub 路径 + openai_compatible mock server 契约测试 + profile loader hot reload 测试 + decorator metric / log / DB / audit 测试 + config fail-fast 测试，[spec C-1 / C-2 / C-3 / C-4](../../spec.md#6-验收标准)）；用 `httptest` 模拟 `https://provider.example/v1/chat/completions` 验证 `Authorization: Bearer ...` header 出站；hot reload 测试对 `config/ai-profiles/*.yaml` 临时改动并触发 `Reload(ctx)`，断言新调用使用新 profile、≤30 秒收敛。把命令日志贴入工作日志。

#### 5.2 Spec C-5 / C-6 / C-7 / C-9 自检

C-5：单次成功调用后查询 in-memory metric registry，确认 7 个 family 已注册且 run / latency / token / cost 按本次调用增长，fallback / validation failure 不增长；查询 fake `ai_task_runs` writer 确认写入一行；查询 fake audit writer 确认 `audit_events` 写入一行（`action='ai.call'`、metadata 仅 hash + 长度 + profile）；查询 in-memory log capture 确认含 `ai.task.completed` 事件且字段集齐全。C-6：`grep -RIn -E '(payload\.messages|response\.content|prompt|response)' backend/internal/ai/aiclient/observability/` 输出仅出现在 hash / 长度计算路径，不出现在 log / metric label / DB metadata 写入路径；同时跑 Phase 3.4 的 `privacy_test.go`。C-7：构造 mock server 返回非法结构 / 触发 client `validateOutput` 失败的用例，断言返回 `AI_OUTPUT_INVALID`、`ai_output_validation_failures_total` +1、`ai.output.validation_failed` 日志事件已发出。C-9：跑 `config_test.go` 中 `AppEnv=production` 缺 gateway 的失败路径用例。

#### 5.3 零厂商 SDK + 隐私 grep 红线

跑 `cd backend && grep -RIn -E '"github.com/(sashabaranov/go-openai|openai/openai-go|anthropic[a-z-]*|cohere[a-z-]*|google/generative-ai-go)"' .` 必须无匹配（[spec C-2 grep 段](../../spec.md#6-验收标准)）；跑 `grep -E 'openai-go|anthropic-sdk-go|cohere-go|generative-ai-go' backend/go.mod` 必须无匹配。跑 `cd backend && grep -RIn -E '(payload\.Messages\[.*\]\.Content|response\.Content|payload\.messages\[\*\]\.content|response\.content)' internal/ai/aiclient/observability/ internal/ai/aiclient/providers/` 仅允许出现在 hash 计算 / 长度计算位置，禁止出现在任何 log / metric / DB write 调用上下文。任何匹配必须在 PR 中显式标注例外原因或修复。

#### 5.4 文档与 INDEX 同步 + handoff

把 plan / checklist Header 在所有验收通过后由 active 切到 completed，运行 `/sync-doc-index --check` 与 `/sync-doc-index --fix-index` 同步 [ai-gateway-and-model-routing/plans/INDEX.md](../INDEX.md) 与根 [docs/spec/INDEX.md](../../../INDEX.md)。不修改 [engineering-roadmap/001-decompose-subspecs](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md) 已完成的 roadmap checklist；C-8 不在本 plan 范畴。把 Phase 5.1 / 5.2 / 5.3 命令输出贴入工作日志。给 [F1 observability-stack](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) / [F3 prompt-rubric-registry](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) / [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md) / 各 C 域 owner 留出 handoff 备注：A3 已暴露 `AITaskRunWriter` / `AuditEventWriter` / `prometheus.Registerer` 三个 DI 入口，依赖注入由各 owner 在自己 plan 中绑定真实实现。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) 中 C-1 / C-2 / C-3 / C-4 / C-5 / C-6 / C-7 / C-9 全部成立；C-8 是 active spec relation gate，本 plan 不重复关闭。
- 本 plan checklist 全部勾选；Phase 5 关键命令日志（`go test` / grep / hot reload / fail-fast）贴入工作日志。
- 业务代码零厂商 SDK 入侵：`backend/go.mod` 与 `backend/internal/ai/aiclient/` 树内不出现 `openai-go` / `anthropic-sdk-go` / `cohere-go` / `generative-ai-go` 等厂商 SDK import；调用方仅依赖 `aiclient` 包 + profile name 字符串。
- 隐私红线：log / metric label / DB metadata / `audit_events.metadata` 中绝不出现明文 prompt / response，仅出现 hash + 长度 + profile 三类摘要；Phase 5.3 grep 与 Phase 3.4 白盒测试同步通过。
- Stub 严格仅在 `APP_ENV=test` 触发；docker compose / Kind / staging / prod 缺 `AI_GATEWAY_BASE_URL` / `AI_GATEWAY_API_KEY` 时进程启动失败，不静默回退到 stub。
- A3 client 不自行执行 retry-with-different-model：fallback chain 仅消费 endpoint / gateway 返回的 meta；`AI_FALLBACK_EXHAUSTED` 仅透传，不主动构造。
- 7 个 metric family 完整注册且 counter 语义正确：每次调用 run / latency / token / cost 增长；fallback / validation failure counter 仅在事件发生时增长。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 业务代码或新 provider adapter 误 import 厂商 SDK（`openai-go` / `anthropic-sdk-go` 等），打破 provider-neutral 红线 | Phase 5.3 强制 `grep go.mod + 源码树` 红线；新增 provider 必须仅依赖 `net/http` + `encoding/json`；CR 时若发现厂商 SDK import 直接拒绝合入；包级 `doc.go` 与 `backend/internal/ai/aiclient/README.md` 显式声明禁令 |
| Stub provider 在非 test 环境被误启用（如错误的 `APP_ENV` / 默认值漂移），导致 prod 静默走假数据 | Phase 1.3 stub factory 在初始化即检查 `APP_ENV`；Phase 4.1 启动期校验在 non-test 缺凭证时 fail-fast；Phase 5.2 C-9 自检覆盖；`backend/internal/ai/aiclient/README.md` 写明白名单条件 |
| Profile 热加载与正在进行的调用形成 race（loader 替换内部 map 时 reader 拿到 partial 状态） | Phase 2.3 使用 atomic store + RW mutex 保证读写一致；正在进行的调用持有当时的 profile pointer，不被替换影响；落 `loader_concurrency_test.go` 跑并发读 + reload 至少 100 轮无 race（`go test -race`） |
| Decorator 漏埋点：某个错误分支或 fallback 路径未触发对应 metric / log / DB / audit 写入，造成 F1 dashboard 失真 | Phase 3.2 decorator 采用 middleware 式包裹同步面，所有出口必须经过 decorator；Phase 3.4 + 5.2 用 fake writer 覆盖 success / failure / fallback / validation 四条路径；新增错误码或调用面必须同步扩 decorator 测试 |
| Audit redaction 失败：日志或 metric label 不慎包含明文 prompt / response（如 debug 日志、错误信息中带 message body） | Phase 3.4 落 white-box `privacy_test.go` 覆盖 metric label / log fields / DB row / audit metadata 四个出口；Phase 5.3 grep 红线兜底；error wrapping helper 强制只接受 `error_code` + `category` 字段，禁止把 raw payload 内容塞进 error string |
| OpenAI-compatible mock server 与真实 provider 行为漂移（mock 总是 200，掩盖了真实 5xx / fallback header / token 字段命名差异） | Phase 2.4 mock server 覆盖 timeout / 5xx / fallback header / 缺失 usage 等异常路径；预留 smoke 验证步骤要求在本地用真实 endpoint 至少跑一次 `Complete` + `Embed` 并核对 `AICallMeta` 字段非空（Phase 4.3 README 写明），但绝不在自动化测试中嵌入真实 API key |
| Profile YAML schema 漂移与 F3 / B1 共享字段名不一致（如 `task_type` 字面量改名）导致 loader 解析失败或 client 误归类 | Phase 1.2 / 2.2 引用 B1 共享常量名而非自定义字符串；Profile schema 字段新增必须先递增 spec 与本 plan 版本；loader 在解析未知字段时 warn 但不丢弃，便于 F3 灰度添加可选字段 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-04 | 1.4 | L1 plan-review remediation：补齐当前强制的质量门禁分类，不改变已完成实现范围。 | historical-spec-implementation-review/001 |
| 2026-04-30 | 1.3 | L2 code-review remediation：补 B4-compatible `ai_task_runs` row contract 与 B1 error-code registry fallback。 | plan-code-review --fix |
| 2026-04-30 | 1.2 | L2 code-review remediation：补 BaseURL `/v1` 归一化、`OutputSchema` 基础校验、Profile 校验错误行号。 | plan-code-review remediation |
| 2026-04-29 | 1.1 | 收口 plan-review：补 Phase 0 前置契约复核，明确 A3 001 不 owns API/worker entrypoint；B1 AI vocabulary 未完全生成时只允许 A3 内部字段表，不得导出跨语言常量。 | plan-review remediation |
