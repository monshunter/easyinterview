# Observability Stack Spec

> **版本**: 1.10
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 F1 `observability-stack` 定义为当前 active Quality spec（依赖 [A2 `local-dev-stack`](./../local-dev-stack/spec.md) 与 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md)）。它直接承接当前代码与运维的可观测层。A2 只提供默认本地应用运行时、应用 `/metrics` 与容器日志出口，不默认提供 OTel Collector / Grafana / Loki / Prometheus。

当前 metric、label、log、trace、dashboard 和 alerting 可执行契约由本 spec、F1 后续编码 truth source 与 product-scope 当前范围决定。F1 独立承接 metric 命名、allowed/forbidden labels、log 字段集、PII redaction、trace attributes、dashboard baseline、alert rules 与上线观测 gate。

本 spec 由 `engineering-roadmap/001-decompose-subspecs` 的 contract lock 创建；当前执行口径是固定 baseline 指标命名约定（Prometheus / OTel label / log 字段 / span attributes），防止后续 P0 backend / frontend / analytics / mock workstream 各自取名。真实 helper、lint、dashboard 与 alerting rules 由 F1 `001` plan 验证。

目标是：

1. **指标命名约定锁定**：Counter `*_total` / Histogram `*_duration_seconds` / Gauge `*_in_flight|*_queue_depth` 命名规则，allowed labels 与 forbidden labels 清单（见 §3.1.1）冻结。
2. **OTel / metrics 接入框架**：Backend / backend internal runner / Frontend 暴露或生产 metrics/logs 与 OTel SDK 初始化点；生产或可选观测环境再接 OTel Collector / Prometheus / Loki / Sentry（trace backend P0 不锁，留接口）。
3. **日志字段约束**：本 spec 锁定字段集并落到 Go logger middleware；明文红线以 D-6 和 F1 tooling 为准。
4. **5 个 dashboard baseline**：本 spec 锁定 5 个 dashboard（业务漏斗 / API & Session Health / Report Pipeline / AI Cost & Quality / Privacy & Compliance）并在上线前完整接齐；baseline plan 先交付命名约定 + 接入框架。

本 spec 不实现具体业务指标埋点（归各 C / D / F2 域）、不部署生产观测后端（归 [E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) + 运维）、不写产品分析事件（归 [F2 `analytics-funnel`](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)）。

## 2 范围

### 2.1 In Scope

- **指标命名约定**：见 §3.1.1；锁定后 lint 强制（A5 接入）。
- **OTel 接入框架**：
  - Backend：`backend/internal/platform/otel/`（OTel SDK 初始化 + tracer / meter provider + propagator）。
  - Frontend：`frontend/src/lib/otel/`（轻量 client，Trace 透传 `traceparent`）。
  - 运行时配置：可选 `OTEL_EXPORTER_OTLP_ENDPOINT`（来自 [A4 字典](../secrets-and-config/spec.md#311-p0-必备-env-key-字典)）；普通本地 dev 为空时只暴露 `/metrics` 与日志，不尝试上报。
- **Logger middleware**：`backend/internal/platform/logx/`（基于 `zerolog`，输出 JSON）；自动注入 F1 通用字段；明文红线类型 `RedactedString`（来自 A4）+ `Hashed`（基于 sha256+salt）helper。
- **Sentry SDK 接线**：Backend / Frontend；DSN 由 A4 env 注入；`SENTRY_DSN` 字段在 §3.1.1 字典中追加（A4 待加入）。
- **Trace 规范**：F1 span name / attribute 集合落到 backend 中间件 + [B3 dispatcher](../event-and-outbox-contract/spec.md) 中的 `traceId` 透传协议。
- **告警规则集 baseline**：F1 P1/P2/P3 告警列表落到 Prometheus alerting rules YAML（grafana / alertmanager 部署归运维）。
- **Dashboard JSON 模板**：5 个 dashboard 的 baseline JSON 落 `deploy/observability/dashboards/`；具体 panel 内容由各 C / D / F2 后续增量贡献，provisioning 由 F1/E4 或可选观测 profile 承接，不进入 A2 默认 `make dev-up`。
- **健康检查端点**：所有进程暴露 `GET /healthz`（liveness）+ `GET /readyz`（readiness）；schema 锁定。

### 2.2 Out of Scope

- 产品分析事件 / 漏斗：归 [F2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)。
- 业务域具体指标埋点（如 `target_import_requests_total`）：归各 C 域。
- AI 调用观测埋点：归 [A3 AIClient](./../ai-provider-and-model-routing/spec.md) 内置，本 spec 仅锁 metric 命名。
- 生产 OTel Collector / Prometheus / Loki / Grafana 部署 chart：归 E4 + 运维。
- LLM Judge / 离线评估：归 [F3](./../prompt-rubric-registry/spec.md)。
- DB 表本身（`audit_events` / `ai_task_runs`）：归 B4。
- 审计动作记录：归各 C 域 + [F4 `privacy-and-audit-runtime`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) P1。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策（含命名约定字典）

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | metric 类型与后缀 | Counter `*_total`；Histogram `*_duration_seconds`；Gauge `*_in_flight` / `*_queue_depth` / `*_pending`；Summary 不使用 | F1 命名 baseline |
| D-2 | 单位 | 时间 seconds（不用 ms 作为指标后缀，日志可用 `latencyMs`）/ 大小 bytes / 金额 usd（如必须） | – |
| D-3 | allowed labels | `service` / `route` / `method` / `status_code` / `operation` / `job_type` / `provider` / `model_family` / `model_profile_name` / `capability` / `language` / `feature` / `env` / `result` / `error_code` / `source_type` / `from_provider` / `from_model_family` / `to_provider` / `to_model_family` | 新增 label 必须是有界枚举；`error_code` 必须来自 B1 `ApiErrorCode`，`source_type` 必须来自 B2/B3 有界 source enum，禁止原始 URL 或自由文本 |
| D-4 | forbidden labels | `user_id` / `target_job_id` / `session_id` / `prompt_version` / 原始 URL 全 path / 原始 provider model id / `from_model` / `to_model` / 任意自由文本 | 高基数禁入 metric；可入 log 或 event |
| D-5 | log 字段集 | 通用 12 字段 + access / job / AI 三种额外字段集 | F1 logger 自动注入 |
| D-6 | log 明文红线 | 绝不进 log：`rawJdText` / `answerText` / `resumeRawText` / `thankYouDraft` / `parsedSummary` 全量 / `promptTemplateBody` / `modelRawResponse` / 文件上传 / 下载 URL / token | `Hashed` helper 提供 sha256+salt |
| D-7 | trace propagation | W3C `traceparent` + `tracestate`；浏览器请求带上即透传；OTel SDK 默认 | – |
| D-8 | 健康检查 | `GET /healthz` 仅检自身存活；`GET /readyz` 检 DB / Redis 等必需依赖；OTel endpoint 仅在显式配置时检查 | A2 `make dev-doctor` 也可消费 healthz / readyz / metrics |
| D-9 | dashboard 名称固定 | `easyinterview-business-funnel` / `easyinterview-api-session-health` / `easyinterview-report-pipeline` / `easyinterview-ai-cost-quality` / `easyinterview-privacy-compliance` 共 5 个 | 后续 child 在自己 plan 里贡献 panel |
| D-10 | 告警优先级与阈值 | P1 5 条全部默认开启；P2 / P3 按需 | – |

#### 3.1.1 Backend / Background runner baseline metrics 字典（baseline freeze）

| 模块 | 指标名 | 类型 | Labels |
|------|--------|------|--------|
| HTTP | `http_server_requests_total` | Counter | service,route,method,status_code |
| HTTP | `http_server_request_duration_seconds` | Histogram | service,route,method |
| HTTP | `http_server_in_flight_requests` | Gauge | service,route |
| HTTP | `http_server_response_size_bytes` | Histogram | service,route |
| DB | `db_query_duration_seconds` | Histogram | service,operation |
| DB | `db_queries_total` | Counter | service,operation,result |
| DB | `db_pool_in_use_connections` | Gauge | service |
| DB | `db_pool_idle_connections` | Gauge | service |
| DB | `db_pool_wait_count_total` | Counter | service |
| Async | `async_jobs_enqueued_total` | Counter | job_type |
| Async | `async_jobs_processed_total` | Counter | job_type,result |
| Async | `async_job_duration_seconds` | Histogram | job_type |
| Async | `async_job_queue_depth` | Gauge | job_type |
| Async | `async_job_lag_seconds` | Gauge | job_type |
| Outbox | `outbox_events_pending` | Gauge | – |
| Outbox | `outbox_publish_duration_seconds` | Histogram | – |
| Outbox | `outbox_publish_failures_total` | Counter | – |
| AI | `ai_task_runs_total` | Counter | provider,model_family,model_profile_name,route,capability,language,result |
| AI | `ai_task_latency_seconds` | Histogram | provider,model_family,model_profile_name,route,capability,language,result |
| AI | `ai_task_input_tokens_total` | Counter | provider,model_family,model_profile_name,route,capability,language,result |
| AI | `ai_task_output_tokens_total` | Counter | provider,model_family,model_profile_name,route,capability,language,result |
| AI | `ai_task_cost_usd_total` | Counter | provider,model_family,model_profile_name,route,capability,language,result |
| AI | `ai_output_validation_failures_total` | Counter | provider,model_family,model_profile_name,route,capability,language,result |
| AI | `ai_fallback_total` | Counter | provider,model_family,model_profile_name,route,capability,language,result,from_provider,from_model_family,to_provider,to_model_family |
| Auth | `auth_challenge_started_total` | Counter | service,result |
| Auth | `auth_session_minted_total` | Counter | service,result |
| Auth | `auth_logout_total` | Counter | service,result |
| Auth | `auth_delete_handoff_total` | Counter | service,result |
| Auth | `auth_failure_total` | Counter | service,operation,result |
| TargetJob | `target_job_imports_total` | Counter | service,operation,source_type,result,error_code |
| TargetJob | `target_job_parse_duration_seconds` | Histogram | service,job_type,source_type,language,result |
| TargetJob | `target_job_parse_failures_total` | Counter | service,job_type,source_type,language,error_code,result |

Auth 指标由 C1 `backend-auth/001-email-code-session-bootstrap` 在自身 plan 中接入；F1 仅登记 metric 名和 label contract。Auth metric labels 只能使用 `service` / `operation` / `result`，不得包含 `user_id`、`session_id`、邮箱、token、完整 URL 或任意自由文本。

业务域（target / practice / report / resume / privacy）指标由各 C 域在自己的 plan 中接入。F1 仅锁 label 集合与命名前缀（domain prefix `target_` / `practice_` / `report_` / `resume_` / `privacy_`）；非当前独立域前缀不得作为新增指标前缀。

TargetJob 指标由 C4 `backend-targetjob/001` 接入：`source_type` 只能是 `url` / `text` / `file` / `manual_form` 等有界导入来源，`error_code` 只能是 B1 错误码常量，`language` 只能是 BCP-47 归一化值；不得把 target id、user id、source URL、prompt version 或任意自由文本作为 label。

### 3.2 待确认事项

- Trace backend：默认 OTel Collector → 直送 Tempo / Jaeger 之一；P0 不锁，由后续根据数据量与运维能力决策。
- Sentry self-host vs SaaS：默认 P0 用 SaaS（Sentry.io），如有合规需求由 E4 切 self-host。
- Loki 索引粒度（labels）：默认 `service / env / level`；如查询缓慢由 F1 在自身 plan 调整。
- Prometheus retention：默认 14 天；如需长期由 Mimir / VictoriaMetrics 替换，由 E4 决策。

## 4 设计约束

### 4.1 命名约束

- 任何新增 metric 必须先在本 spec §3.1.1 中登记（spec 修订递增版本）；A5 lint 拦截未登记的 metric 名。
- 任何 metric label 必须落在 §3.1.1 / D-3 allowed labels 中；新增 label 必须 spec 修订（理由：高基数防御）。
- 业务域 metric 命名必须 `<domain>_<noun>_<unit_or_total>` 形式，例：`practice_session_duration_seconds`、`feedback_reports_generated_total`。

### 4.2 logger 约束

- 所有 backend 包必须使用 `logx.Logger`（F1 提供）；禁止裸 `fmt.Println` / `log.Print`（A5 lint 拦截）。
- 日志中包含敏感字段时必须使用 `Hashed(...)` 或 `RedactedString`；F1 提供 `lint-logs` 工具扫描已知敏感字段名。
- Trace context 必须在 logger 中自动注入（middleware 强制），不允许业务代码手动传 `requestId`。

### 4.3 trace 约束

- 每个 HTTP route 自动产生一个根 span，命名由 F1 trace convention 决定。
- 每个异步 job 进入 backend internal runner 时从 `traceId` 字段重建 span context（B3 透传）。
- 每次 AI 调用作为子 span（A3 内部产生），span attributes 由 F1 trace convention 与 A3 AI metadata 决定；不允许写明文 prompt / answer。

### 4.4 性能约束

- logger middleware 性能开销 ≤ 5% 单请求耗时。
- metric 上报 batch 间隔 ≤ 10s（OTel SDK 默认）；不允许同步 push（性能风险）。
- 健康检查端点 P95 ≤ 100ms。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `internal/platform/{otel,logx}/` Go 包 | F1 | OTel + zerolog middleware |
| `frontend/src/lib/otel/` | F1 | trace propagator |
| `deploy/observability/{dashboards,alerts}/` | F1 | 5 dashboard JSON + alerting rules YAML |
| 可选本地观测运行后端 | F1 + E4 | 如需 OTel Collector / Prometheus / Loki / Grafana，本 spec 或 E4 提供可选 profile / chart；A2 默认 `make dev-up` 不启动 |
| 业务域 metric 埋点 | 各 C 域 | 通过 F1 提供的 helper |
| 业务域 log 调用 | 各 C 域 | 必须使用 logx |
| 产品分析事件 | F2 | 与本 spec 命名空间分离（snake_case underscore_event） |
| AI 观测埋点 | A3 内部 | F1 仅锁 metric 名 |
| Sentry DSN 注入 | A4 | F1 提供 SDK 接线 |
| 部署后端（Grafana / Prometheus / Loki / Tempo） | E4 + 运维 | F1 仅交付配置 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | metric 命名 lint | 故意提交一个 `practice_sessionsCompletedCount`（驼峰、缺单位） | CI | `lint-metrics` 失败；Job Summary 提示规范 | F1 后续 001 + A5 |
| C-2 | label 高基数防御 | 提交一个 metric 含 `user_id` label | CI | `lint-metrics` 失败 | F1 后续 001 + A5 |
| C-3 | log 明文红线 | 在 `internal/practice/` 中调用 `logx.Info("answer", "answer", answerText)` | CI | `lint-logs` 失败 | F1 后续 001 + A5 |
| C-4 | trace propagation | 前端 fetch 带 `traceparent` | API → backend internal runner → AI | 同一 traceId 贯穿 4 层 span；配置了 trace backend 时可查询，未配置时不阻塞 A2 本地开发栈 | F1 后续 001 |
| C-5 | dashboard provision | F1 可选观测 profile / E4 部署路径启动 | Grafana | 5 个 dashboard 名称已存在；空 panel 提示「待后续接入」 | F1 后续 001 + E4 |
| C-6 | 健康检查 | 服务运行 | `GET /healthz` 与 `GET /readyz` | 200 + JSON `{status:"ok",components:[...]}` | F1 后续 001 |
| C-7 | 告警 baseline | 制造模拟事件触发 P1 告警 | Prometheus alerting rules | 5 条 P1 告警 fire；可路由到 Slack/Email（运维端） | F1 后续 001 + 运维 |
| C-8 | 业务域 helper | C5 调用 `metrics.PracticeSessionStarted(goal, mode, language)` | 单测 | 内部上报 `practice_sessions_started_total{goal,mode,language}` +1 | F1 后续 001 + C5 |
| C-9 | baseline 验证 | 本 spec 通过 `/plan-review`，F1 后续 001 完成 baseline | active spec 关系已保留 | F1 baseline 指标命名约定、lint / helper / dashboard 框架自洽，后续 workstream 可引用 | F1 后续 001 |
| C-10 | 上线门槛 | C-1..C-9 + 5 个 dashboard 接齐 | F1 release-observability checklist | 全部勾选 | F1 后续 002（dashboard 完善）+ E4 |

## 7 关联计划

F1 当前暂无 active impl plan；后续由 F1 自身的 plans 承接（[engineering-roadmap §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 保留该 active spec）：

- `001-baseline`：`internal/platform/{otel,logx}/` + `frontend/src/lib/otel/` + 5 dashboard 框架 + alerting rules + lint 工具 + 健康检查。
- `002-dashboards-and-alerts`：完成 5 个 dashboard 的 panel 接齐 + 告警阈值校准；上线前闭合 F1 最低上线门槛。

后续如需扩展（profiling / continuous profiling）：递增 spec 版本，原地修订；不创建 sibling spec。
