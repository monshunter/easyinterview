# AI Gateway and Model Routing Bootstrap Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-29

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

- [ ] 2.1 落地 `providers/openai_compatible/` adapter：`Complete` 走 `/v1/chat/completions`、`Embed` 走 `/v1/embeddings`；header 含 `Authorization: Bearer ${AI_GATEWAY_API_KEY}` / `Content-Type` / `X-Request-ID` 透传；解析 `usage.*tokens` / `model` / `x-fallback-*` / `x-route` 入 `AICallMeta`；只用标准库 `net/http` + `encoding/json`，零厂商 SDK；timeout 走 `context.WithTimeout` 返回 `AI_PROVIDER_TIMEOUT`；非 2xx 按 B1 错误码映射；A3 client 绝不自行 retry-with-different-model
- [ ] 2.2 落地 `profile/` Model Profile schema 类型与 YAML loader：字段集严格对齐 spec §2.1（`name` / `task_type` 含 `stt` 预留值 / `default.*` / `fallback[]` / `timeout_ms` / `max_tokens` / `rate_limit.{rps,tpm}` / `gateway_route` / `version`）；落地 `config/ai-profiles/practice.followup.default.yaml` 与 `review.report.default.yaml` 两个最小 fixture profile 仅供本 plan 测试 / 本地验证；解析失败附 file path + line number
- [ ] 2.3 落地 ≤30 秒热加载：fsnotify 监听 + 30s polling 兜底；atomic store + RW mutex 保证正在进行的调用使用旧 profile、新调用使用新 profile；暴露 `Reload(ctx) error` 测试入口；落 `loader_concurrency_test.go` `go test -race` 至少 100 轮并发读 + reload 无 race
- [ ] 2.4 落地 `providers/openai_compatible/contract_test.go` 离线契约测试 + 可被 E1 复用的 `mockserver/` helper：覆盖正常 chat / embeddings、超时、5xx、fallback meta 注入；断言 token 解析、fallback chain 透传、超时 → `AI_PROVIDER_TIMEOUT`、5xx → B1 错误码语义；mock server interface 稳定供 E1 复用

## Phase 3: Observability / audit decorator + DB / log / metric 接入

- [ ] 3.1 落地 `observability/metrics.go` 注册 7 个 metric family（`ai_task_runs_total` / `ai_task_latency_seconds` / `ai_task_input_tokens_total` / `ai_task_output_tokens_total` / `ai_task_cost_usd_total` / `ai_output_validation_failures_total` / `ai_fallback_total`）；label 集对齐 F1（`provider` / `model_family` / `model_profile_name` / `route` / `task_type` / `language` / `result` / `from_model_family` / `to_model_family`）；通过 `prometheus.Registerer` 抽象，业务调用不能绕过 decorator
- [ ] 3.2 落地 `observability/decorator.go` middleware 包裹 `Complete` / `Embed`：写入 4 类结构化日志事件名（`ai.task.completed` / `ai.task.failed` / `ai.task.fallback` / `ai.output.validation_failed`）字段集对齐 05-logging-standard.md §4.4；通过 DI 写入 `ai_task_runs`（A3 只填 typed columns）；写入 `audit_events.action='ai.call'` 行，metadata 仅含 `prompt_hash` / `response_hash` / `prompt_char_length` / `response_char_length` / `profile_name`
- [ ] 3.3 fallback / validation 计数器语义：`meta.FallbackChain[]` 长度 > 1 时 `ai_fallback_total{from_model_family,to_model_family,result="fallback"}` +1；`validateOutput` 失败时 `ai_output_validation_failures_total` +1 + 发出 `ai.output.validation_failed` 日志，错误码统一 `AI_OUTPUT_INVALID`；`AI_FALLBACK_EXHAUSTED` 仅透传 endpoint / gateway 返回值
- [ ] 3.4 落地 `observability/privacy_test.go` 白盒测试：构造带敏感内容的 messages / response，使用 in-memory writer + log capture 断言 metric label / log fields / DB row metadata / audit_events metadata 不含明文，仅含 hash 前缀 / 长度数字 / profile 名

## Phase 4: 配置校验与本地部署 fail-fast

- [ ] 4.1 落地 `config.go` 配置 struct（`AppEnv` / `GatewayBaseURL` / `GatewayAPIKey` / `ModelProfilePath`）与 `New(cfg)` 启动期校验：`AppEnv != "test"` 且 gateway 任一字段空时返回 `ErrMissingGatewayConfig`；`AppEnv == "test"` 路径允许 `WithStubAllowed(true)` 启用 stub
- [ ] 4.2 落地 `config_test.go`：`AppEnv=test` 缺 gateway 但启用 stub → 成功；`AppEnv=production` 缺 gateway → 错误；`AppEnv=test` 但 stub 选项未启用且无 gateway → 错误；提供 `New(cfg)` / DI 构造契约供 A4 / C 域在 `cmd/api` / `cmd/worker` 接入时把 cfg 错误转换为 non-zero exit，本 plan 不创建或重写 entrypoint
- [ ] 4.3 落地 `backend/internal/ai/aiclient/README.md`：写明 stub 仅 `APP_ENV=test` 启用、docker compose / Kind / staging / prod 必须真实 OpenAI-compatible endpoint、smoke 验证步骤示意（导出真实 endpoint env 后跑 `go test -tags smoke`，绝不在测试代码 / fixture 中嵌入真实 API key），fsnotify ↔ polling 兜底机制说明

## Phase 5: Verification + handoff

- [ ] 5.1 自检 spec C-1 / C-2 / C-3 / C-4：`cd backend && go test ./internal/ai/aiclient/...` 全绿；用 `httptest` 验证 `Authorization: Bearer ...` header 出站；`Reload(ctx)` 触发后新调用使用新 profile、≤30 秒收敛
- [ ] 5.2 自检 spec C-5 / C-6 / C-7 / C-9：单次成功调用确认 7 个 metric family 已注册且 run / latency / token / cost 按本次调用增长（fallback / validation failure 不增长）；fake `ai_task_runs` writer 写入一行；fake `audit_events` writer 写入一行（`action='ai.call'`、metadata 仅 hash + 长度 + profile）；`ai.task.completed` 日志事件字段齐全；mock server 触发 `validateOutput` 失败返回 `AI_OUTPUT_INVALID` + `ai_output_validation_failures_total` +1；`AppEnv=production` 缺 gateway 启动失败
- [ ] 5.3 grep 红线：`grep -RIn -E '"github.com/(sashabaranov/go-openai|openai/openai-go|anthropic[a-z-]*|cohere[a-z-]*|google/generative-ai-go)"' backend/` 无匹配；`grep -E 'openai-go|anthropic-sdk-go|cohere-go|generative-ai-go' backend/go.mod` 无匹配；隐私 grep（`payload.Messages[*].Content` / `response.Content`）仅出现在 hash / 长度计算路径，禁止出现在 log / metric / DB write 上下文
- [ ] 5.4 文档与 INDEX 同步 + handoff：把 plan / checklist Header 切到 completed；运行 `/sync-doc-index --check` 与 `/sync-doc-index --fix-index` 同步 plans/INDEX.md 与根 docs/spec/INDEX.md；不修改 engineering-roadmap/001 Phase 3.2；命令日志贴入工作日志；给 F1 / F3 / B4 / 各 C 域 owner 留 handoff 备注（`AITaskRunWriter` / `AuditEventWriter` / `prometheus.Registerer` 三个 DI 入口已暴露）
