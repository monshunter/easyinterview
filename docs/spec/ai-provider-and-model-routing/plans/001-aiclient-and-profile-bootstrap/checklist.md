# AI Provider and Model Routing Bootstrap Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

## Phase 0: 前置契约复核

- [x] 0.1 确认 A1 `repo-scaffold/001-bootstrap` v1.1 已提供 `config/` 根容器与根 README 索引；缺失时先走 A1 remediation，不在 A3 下另起平行配置目录
- [x] 0.2 确认 B1 至少提供 `AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_FALLBACK_EXHAUSTED` baseline 错误码；AI vocabulary 字段名若尚未由 B1 002 生成，A3 001 仅维护包内私有字段表并在 handoff 标注切换点

## Phase 1: AIClient interface + stub provider 骨架

- [x] 1.1 落地 `backend/internal/ai/aiclient/` 包骨架（`aiclient.go` / `meta.go` / `payload.go` / `profile.go` / `doc.go`）；声明 `AIClient` interface 三方法签名（`Complete` / `Embed` 同步面 + `Stream(ctx, profile, payload) → (<-chan AIStreamEvent, error)` 事件合同）；`AIStreamEvent` 仅冻结 `delta` / `error` / `done` 三种 type 与 channel close 语义，不实现 provider 流式消费循环
- [x] 1.2 落地 `AICallMeta` 运行时结构体，字段顺序固定（`Provider` / `ModelFamily` / `ModelID` / `TaskType` / `PromptVersion` / `RubricVersion` / `ModelProfileName` / `ModelProfileVersion` / `Language` / `InputTokens` / `OutputTokens` / `CostUSDMicros` / `LatencyMs` / `FallbackChain[]` / `Route` / `ValidationStatus` / `ErrorCode`）；`metaBuilder` helper 由 client 填充并校验；引用 B1 错误码 / `task_type` enum 共享常量，不重复定义
- [x] 1.3 落地 `providers/stub/` deterministic stub provider（`sha256(profile + canonical(payload))` 截字节作 RNG seed）；factory 在初始化即检查 `APP_ENV=test` 或显式 `WithStubAllowed(true)`，否则拒绝实例化；`Stream` 仅返回一次 `done` 事件并关闭 channel
- [x] 1.4 落地包级单测（`aiclient_test.go` / `providers/stub/stub_test.go`）覆盖 stub 路径：`Complete` 返回 `meta.Provider == "stub"`、同 input 多次调用一致、空 messages 返回 `AI_OUTPUT_INVALID`、`Stream` 收到 `done` 事件且 channel 关闭

## Phase 2: openai_compatible provider + Model Profile loader

- [x] 2.1 落地 `providers/openai_compatible/` adapter：`Complete` 走 `/v1/chat/completions`、`Embed` 走 `/v1/embeddings`；header 含 `Authorization: Bearer ${AI_PROVIDER_API_KEY}` / `Content-Type` / `X-Request-ID` 透传；解析 `usage.*tokens` / `model` / `x-fallback-*` / `x-route` 入 `AICallMeta`；只用标准库 `net/http` + `encoding/json`，零厂商 SDK；timeout 走 `context.WithTimeout` 返回 `AI_PROVIDER_TIMEOUT`；非 2xx 按 B1 错误码映射；A3 client 绝不自行 retry-with-different-model
- [x] 2.2 落地 `profile/` Model Profile schema 类型与 YAML loader：字段集严格对齐 spec §2.1（`name` / `task_type` 含 `stt` 预留值 / `default.*` / `fallback[]` / `timeout_ms` / `max_tokens` / `rate_limit.{rps,tpm}` / `route` / `version`）；落地 `config/ai-profiles/practice.followup.default.yaml` 与 `review.report.default.yaml` 两个最小 fixture profile 仅供本 plan 测试 / 本地验证；解析失败附 file path + line number
- [x] 2.3 落地 ≤30 秒热加载：polling reloader（5s 默认 cadence，plan §2.3 允许的 fsnotify fallback 路径）+ atomic store + RW mutex 保证正在进行的调用使用旧 profile、新调用使用新 profile；暴露 `Reload(ctx) error` 测试入口；落 `loader_concurrency_test.go` `go test -race` 至少 100 轮并发读 + reload 无 race
- [x] 2.4 落地 `providers/openai_compatible/contract_test.go` 离线契约测试 + 可被 E1 复用的 `mockserver/` helper：覆盖正常 chat / embeddings、超时、5xx、fallback meta 注入；断言 token 解析、fallback chain 透传、超时 → `AI_PROVIDER_TIMEOUT`、5xx → B1 错误码语义；mock server interface 稳定供 E1 复用
- [x] 2.5 L2 remediation F1：`openai_compatible` adapter 必须同时支持 root endpoint 与已含 `/v1` 的 `AI_PROVIDER_BASE_URL`，避免拼出 `/v1/v1/chat/completions`；补离线契约测试覆盖 `BaseURL = mockServer.URL() + "/v1"`
- [x] 2.6 L2 remediation F3：`profile.Loader` 的手写 schema 校验错误必须包含 file path + line number；补单测覆盖缺必填字段 / 非法 task_type 等非 YAML 语法错误路径
- [x] 2.7 L2 remediation F4：`openai_compatible` 4xx error envelope 只透传 B1 `CodeRegistry` 登记错误码；未知上游 code fallback 到 `AI_OUTPUT_INVALID`；补契约测试覆盖未知 code fallback 与已登记 code passthrough；验证: 2026-04-30 `go test ./internal/ai/aiclient/providers/openai_compatible -run 'TestComplete_4xx(ParsesErrorEnvelope|UnknownErrorCodeFallsBackToAIOutputInvalid)' -count=1`

## Phase 3: Observability / audit decorator + DB / log / metric 接入

- [x] 3.1 落地 `observability/metrics.go` 注册 7 个 metric family（`ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total`）；label 集对齐 F1（`provider` / `model_family` / `model_profile_name` / `route` / `task_type` / `language` / `result` / `from_model_family` / `to_model_family`）；通过 `prometheus.Registerer` 抽象，业务调用不能绕过 decorator
- [x] 3.2 落地 `observability/decorator.go` middleware 包裹 `Complete` / `Embed`：写入 4 类结构化日志事件名（`ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`）字段集对齐 F1 observability-stack logging §4.4；通过 DI 写入 `ai_task_runs`（A3 只填 typed columns）；写入 `audit_events.action='ai.call'` 行，metadata 仅含 `prompt_hash` / `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`
- [x] 3.3 fallback / validation 计数器语义：`meta.FallbackChain[]` 长度 > 1 时 `ai_fallback_total{from_model_family,to_model_family,result="fallback"}` +1；`validateOutput` 失败时 `ai_output_validation_failures_total` +1 + 发出 `ai.output.validation_failed` 日志，错误码统一 `AI_OUTPUT_INVALID`；`AI_FALLBACK_EXHAUSTED` 仅透传 endpoint / provider 返回值
- [x] 3.4 落地 `observability/privacy_test.go` 白盒测试：构造带敏感内容的 messages / response，使用 in-memory writer + log capture 断言 metric label / log fields / DB row metadata / audit_events metadata 不含明文，仅含 hash 前缀 / 长度数字 / profile 名
- [x] 3.5 L2 remediation F2：`observability` decorator 在 `OutputSchema` 存在时必须校验 schema 的基本 `type` / `required` / `properties` 约束，不能只检查 response content 是合法 JSON；schema 不匹配时返回 `AI_OUTPUT_INVALID` 并递增 validation failure counter
- [x] 3.6 L2 remediation F1：`AITaskRunRow` 覆盖 B4 `ai_task_runs` 必填列与 A3 typed columns，业务 `task_type` 不再写入 `chat` / `embed` / `stt`；缺 task/resource context 或 writer failure 不得静默吞掉；补 focused tests 覆盖 B4-compatible row 与 writer failure propagation；验证: 2026-04-30 `go test ./internal/ai/aiclient/observability -run 'TestDecorator_(SuccessIncrementsRunsAndLogsCompleted|AITaskRunWriterFailureReturned)' -count=1` 与 `go test ./internal/ai/aiclient/... -count=1`

## Phase 4: 配置校验与本地部署 fail-fast

- [x] 4.1 落地 `config.go` 配置 struct（`AppEnv` / `ProviderBaseURL` / `ProviderAPIKey` / `ModelProfilePath`）与 `New(cfg)` 启动期校验：`AppEnv != "test"` 且 provider config 任一字段空时返回 `ErrMissingProviderConfig`；`AppEnv == "test"` 路径允许 `WithStubAllowed(true)` 启用 stub
- [x] 4.2 落地 `config_test.go`：`AppEnv=test` 缺 provider config 但启用 stub → 成功；`AppEnv=production` 缺 provider config → 错误；`AppEnv=test` 但 stub 选项未启用且无 provider config → 错误；提供 `New(cfg)` / DI 构造契约供 A4 / C 域在 `cmd/api` / `cmd/worker` 接入时把 cfg 错误转换为 non-zero exit，本 plan 不创建或重写 entrypoint
- [x] 4.3 落地 `backend/internal/ai/aiclient/README.md`：写明 stub 仅 `APP_ENV=test` 启用、docker compose / Kind / staging / prod 必须真实 OpenAI-compatible endpoint、smoke 验证步骤示意（导出真实 endpoint env 后跑 `go test -tags smoke`，绝不在测试代码 / fixture 中嵌入真实 API key），fsnotify ↔ polling 兜底机制说明

## Phase 5: Verification + handoff

- [x] 5.1 自检 spec C-1 / C-2 / C-3 / C-4：`cd backend && go test ./internal/ai/aiclient/...` 全绿；用 `httptest` 验证 `Authorization: Bearer ...` header 出站；`Reload(ctx)` 触发后新调用使用新 profile、≤30 秒收敛
- [x] 5.2 自检 spec C-5 / C-6 / C-7 / C-9：单次成功调用确认 7 个 metric family 已注册且 run / latency / token / cost 按本次调用增长（fallback / validation failure 不增长）；fake `ai_task_runs` writer 写入一行；fake `audit_events` writer 写入一行（`action='ai.call'`、metadata 仅 hash + 长度 + profile）；`ai.task.completed` 日志事件字段齐全；mock server 触发 `validateOutput` 失败返回 `AI_OUTPUT_INVALID` + `ai_output_validation_failures_total` +1；`AppEnv=production` 缺 provider config 启动失败
- [x] 5.3 grep 红线：`grep -RIn -E '"github.com/(sashabaranov/go-openai|openai/openai-go|anthropic[a-z-]*|cohere[a-z-]*|google/generative-ai-go)"' backend/` 无匹配；`grep -E 'openai-go|anthropic-sdk-go|cohere-go|generative-ai-go' backend/go.mod` 无匹配；隐私 grep（`payload.Messages[*].Content` / `response.Content`）仅出现在 hash / 长度计算路径，禁止出现在 log / metric / DB write 上下文
- [x] 5.4 文档与 INDEX 同步 + handoff：把 plan / checklist Header 切到 completed；运行 `/sync-doc-index --check` 与 `/sync-doc-index --fix-index` 同步 plans/INDEX.md 与根 docs/spec/INDEX.md；不修改 engineering-roadmap/001 已完成的 roadmap checklist；命令日志贴入工作日志；给 F1 / F3 / B4 / 各 C 域 owner 留 handoff 备注（`AITaskRunWriter` / `AuditEventWriter` / `prometheus.Registerer` 三个 DI 入口已暴露）

## Phase 6: AI provider terminology remediation

- [x] 6.1 重命名 A3 subject 目录与 ADR-Q6 文件为 `ai-provider-and-model-routing` / `ADR-Q6-ai-provider-and-model-routing.md`，同步 active spec/plan/context/INDEX/ADR/roadmap/A4/A2/B1/F1/F3 引用；验证 `validate_context.py --context docs/spec/ai-provider-and-model-routing/plans/001-aiclient-and-profile-bootstrap/context.yaml --target backend` 通过，retired subject 路径不存在且 active docs 不引用retired identifier
  <!-- verified: 2026-05-05 method=validate_context+rg evidence=validate_context OK; old subject path absent; no active references to retired subject/ADR path -->
- [x] 6.2 重命名 runtime config contract：`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY`、`ProviderBaseURL` / `ProviderAPIKey` / `ErrMissingProviderConfig`、`ai.providerBaseURL` / `ai.providerApiKey` 全量落地；同步 `.env.example`、`config/config.yaml`、A4 bindings/validator、cmd worker tests、dev-stack env/doctor/docs、shared generated comments；验证 `python3 scripts/lint/env_dict.py --repo-root .` 与 focused config tests 通过，且不保留旧 env key fallback
  <!-- verified: 2026-05-05 method=lint-config+go-test+rg evidence=env_dict OK; go test ./internal/platform/config ./cmd/worker ./internal/ai/aiclient -count=1 OK; no retired env/API/config identifiers in active scope -->
- [x] 6.3 重命名 Model Profile route schema：YAML `route` 与 Go `Route` 使用 provider-neutral 命名；同步 fixtures、loader tests、openai_compatible contract tests、meta builder 与 README；验证 `go test ./internal/ai/aiclient/profile ./internal/ai/aiclient/providers/openai_compatible ./internal/ai/aiclient -count=1` 通过，active fixtures 不含 retired schema key
  <!-- verified: 2026-05-05 method=go-test+rg evidence=go test ./internal/ai/aiclient/profile ./internal/ai/aiclient/providers/openai_compatible ./internal/ai/aiclient -count=1 OK; no retired route schema/API terms in active profile scope -->
- [x] 6.4 增加 provider 旧口径负向 gate：当前代码、配置、active docs、deploy 资产、generated artifacts 不得出现 retired env/schema/API identifiers 或把 AI provider 连接描述为独立转发层；历史 `docs/work-journal/`、`docs/reports/`、`docs/bugs/` 与 history 修订记录仅作为只读例外
  <!-- verified: 2026-05-05 method=lint+unit-test+rg evidence=make lint-ai-provider-terminology OK; python3 scripts/lint/ai_provider_terminology_test.py OK; active-scope retired terminology rg returned no matches -->
- [x] 6.5 Verification + lifecycle sync：运行 focused tests、`make lint-config`、`make codegen-check`、`make docs-check`、`make lint`、`make test`、`make build` 与 retired provider terminology 负向搜索；全部通过后将 plan/checklist 恢复 `completed` 并同步 `docs/spec/INDEX.md` / `plans/INDEX.md`
  <!-- verified: 2026-05-05 method=focused-tests+make-gates+sync-doc-index evidence=go test focused AI/config/worker OK; make lint-config OK; make lint OK; make test OK; make build OK; make codegen-check OK with temporary index seeded for intended conventions/generated AI provider diffs; docs-check run after index sync; active-scope retired terminology rg returned no matches -->
