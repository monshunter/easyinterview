# Backend Async Runner Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-21

## 1 背景与目标

[engineering-roadmap §6.3 S2](../engineering-roadmap/spec.md#63-s2--backend-domain-implementation) 把 `backend-async-runner` 列为「backend 内部 job、outbox、retry 和删除链路执行；P0 不拆独立 worker 进程」的 owner subject；[backend-runtime-topology](../backend-runtime-topology/spec.md) D-1~D-4 已锁定 P0 拓扑为 `frontend` + `backend` 两个应用进程，后台任务由 backend internal runner 承接。当前仓库已经在不同业务域生长出多套局部 runner / drainer：

- `backend/internal/targetjob/drainer.go`：通用 `Drainer` 抽象 + `JobHandler` 接口，注册 `target_import` / `source_refresh` / `privacy_delete` / `debrief_generate` 等 job_type；无 reaper。
- `backend/internal/review/{runner,reaper,lease}.go`：[backend-review D-13 / D-16](../backend-review/spec.md) 显式承认为 P0 临时形态的独立 polling worker，仅服务 `report_generate`；spec 文本明确写明「未来 `backend-async-runner` plan 接管时统一抽象 review.Runner + drainer」。
- `backend/internal/auth/mail.go` `BackgroundMailDispatcher`：进程内 channel 直推 `DeliveryWriter` sink，绕过 `async_jobs`，与 [ADR-Q2](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 锁定的 `email_dispatch` canonical jobType 不一致。
- `backend/cmd/api/main.go` 维护 4 个独立 `defer ... Shutdown(ctx)` lifecycle（auth dispatcher / targetJobRuntime / resumeRuntime / reportRuntime），无统一 graceful drain 顺序。
- `outbox_events` 表由各域业务事务写入（resume / targetjob / review / practice store），但仓库**没有任何 publisher / dispatcher 代码**消费这张表；[B3 D-7 / D-8 / D-9](../event-and-outbox-contract/spec.md) 锁定的 dispatcher 协议未实现。
- retry/backoff 不一致：`review.ComputeReportFailureBackoff` 用 `2^attempts * 30s`，resume `async.go` 用固定 15s，与 [ADR-Q2 §3.4](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 锁定的 `30s/2m/10m/1h/6h, max 5` 不符。
- lease 回收 reaper 仅覆盖 `report_generate`；其它 6 个 job_type 若 backend 中途崩溃将 stuck 在 `running`/`locked_at` 状态。

本 subject 的目标是把上述局部形态收敛为单一 backend in-process runtime kernel，统一 lease / retry / dead-letter / reaper / shutdown 协议，落地 B3 dispatcher 缺口，并保留「不建独立 worker 进程」语义；未来如需拆运行单元、切 Asynq server 或引入新调度器，新 ADR 即可在不动业务 producer / handler 的前提下替换底层实现。

## 2 范围

### 2.1 In Scope

- 新增 `backend/internal/runner/` package：单一 `runner.Runtime`，负责 handler registry、lease（`SELECT ... FOR UPDATE SKIP LOCKED`）、finalize、retry（ADR-Q2 退避）、reaper（lease timeout 回收）、graceful shutdown drain。
- 统一 `runner.Handler` 接口；提供 `targetjob.JobHandler` adapter shim 以支持 P2 渐进迁移。
- 接管 8 个当前可执行 canonical job_type 的运行：`target_import` / `source_refresh` / `privacy_delete` / `debrief_generate` / `resume_parse` / `resume_tailor` / `report_generate` / `email_dispatch`；handler 业务实现仍由各 owner 域持有，本 subject 只承接注册与生命周期。B3 第 9 项 `privacy_export` 当前只保留为 DB / OpenAPI 501 contract，未有 producer / handler，本计划不注册 handler。
- 落地 `runner.OutboxDispatcher`：按 [B3 D-7](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 5s scan + `FOR UPDATE SKIP LOCKED` + batch≤100 + sort by `next_attempt_at asc, created_at asc`；按 D-7 退避 30s/2m/10m/1h/6h；按 D-8 dead-letter at attempt 5；按 D-9 暴露 `outbox_events_pending` / `outbox_publish_duration_seconds` / `outbox_publish_failures_total` 指标；按 D-10 透传 `traceId`；按 D-15 通过 `IsSourceEventOnly()` 谓词跳过 `source_event_only` job 创建；未注册 consumer 的 runtime 事件不得被 dry-run 误标为 `published`。
- 删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff` / `auth.BackgroundMailDispatcher` 等局部形态；删除多 `targetjob.NewDrainer` 实例化点；`cmd/api` 单点持有 `runner.Runtime` 并编排 shutdown。
- 接入 [A4 D-9](../secrets-and-config/spec.md) `async.queueWeights` typed config 为 kernel 优先级权重源；同步按 A4 additive 流程新增 `async.shutdownGraceSeconds` / `async.leaseTimeoutSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds` config-only typed 节点（不新增 env key）。
- C1 `email_dispatch` producer 由进程内 channel 切换到 `INSERT INTO async_jobs(job_type='email_dispatch')`；kernel 注册 `EmailDispatchHandler` 通过 [C1 `DeliveryWriter`](../backend-auth/spec.md) sink 发邮件。
- 负向 lint gate `make lint-runner-legacy`（或等价 scripts）：旧 `review.NewRunner` / `review.NewReaper` / `ComputeReportFailureBackoff` / `auth.BackgroundMailDispatcher` / 散落 `targetjob.NewDrainer` 实例化（除 kernel adapter 外）必须 0 命中。

### 2.2 Out of Scope

- 不替换 PG-backed `async_jobs` / `outbox_events` 存储为 Asynq Redis 队列；本 subject 把 kernel 设计为「task contract 不变前提下可被 Asynq 实现替换」，但实际 Redis/Asynq 引入由 ADR-Q2 §5 触发条件成立后另起 plan。
- 不引入独立 worker 进程；P0 拓扑沿用 [backend-runtime-topology D-1](../backend-runtime-topology/spec.md)。
- 不重写各域 handler 业务行为：业务签名（resolveActive、AI complete、handler 状态机、outbox 红线）保留原样；本 subject 仅迁移生命周期边界，不动业务断言。
- 不新增 producer / B3 event；本 subject 不修改 `shared/events.yaml` / `shared/jobs.yaml` / B4 baseline。
- 不实现 `privacy_export` handler 或 producer；P0 privacy export API 仍按 B2 fixture 返回 501，直到 privacy export owner plan 启用真实导出链路。
- 不实现独立 outbox 重放工具 / Schema Registry（沿用 B3 §3.2 默认决策）。
- 不接入 Prometheus / Grafana / OTel Collector 默认运行；指标只暴露在应用 `/metrics` 端点，由 F1 / 生产部署可选 profile 消费（[backend-runtime-topology D-4](../backend-runtime-topology/spec.md)）。
- 不修改 frontend；本 subject 无 UI 真理源。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 单一 kernel | 新建 `backend/internal/runner/`，所有 in-process async job 执行收敛到一个 `runner.Runtime` | 删除 `review.Runner` / `review.Reaper` / 多 `targetjob.NewDrainer` 实例化；`cmd/api` 持有 1 个 runtime |
| D-2 | 统一 Handler 接口 | 引入 `runner.Handler{Handle(ctx, runner.ClaimedJob) runner.JobOutcome}`；同包提供 `runner.JobHandlerFunc` adapter 与 `runner.FromTargetjobHandler(targetjob.JobHandler) runner.Handler` shim | P2 迁移期允许同名调用；P4 完成后 `targetjob.JobHandler` 可降级为 alias 或被删除 |
| D-3 | Lease SQL contract | kernel `LeaseAsyncJob`/`FinalizeAsyncJob` 直接拥有 SQL；列名 `locked_at` / `attempts` / `available_at` 与 [B4 baseline](../db-migrations-baseline/spec.md) 一致；P2 各域 store 不再独立持有 ClaimNext / Finalize SQL 副本，统一调用 kernel-side store | 收口 review / resume / targetjob 三处重复 SQL |
| D-4 | Retry backoff | 唯一退避表 `[30s, 2m, 10m, 1h, 6h]`，max 5 attempts；retryable 失败在 attempts < max 时回到 `queued`，attempts >= max 时进入 `dead`；non-retryable 失败进入 `failed`；与 [ADR-Q2 §3.4](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) / B3 D-7 一致 | 删除 `review.ComputeReportFailureBackoff` (2^n) 与 resume `async.go` 固定 15s；统一 targetjob/review/resume 的终态语义 |
| D-5 | Reaper 覆盖 | 单一 reaper，覆盖全部已注册 job_type；lease timeout 默认 `async.leaseTimeoutSeconds`（typed config 提供，默认 300s）；reaper interval 默认 `async.reaperIntervalSeconds`（默认 60s） | review.Reaper 删除；target_import / privacy_delete / debrief / resume parse/tailor / email_dispatch 死锁 row 可恢复 |
| D-6 | Outbox dispatcher | 新增 `runner.OutboxDispatcher` 模块；5s scan + `FOR UPDATE SKIP LOCKED` + batch≤100 + sort by `next_attempt_at asc, created_at asc`；按 B3 D-7 退避；attempt≥5 → `publish_status='failed'` + redacted `last_error_message`；未注册 consumer 的 runtime 事件保持 `pending` 并按 retry/dead-letter 路径记录错误，不允许默认 dry-run ack；产出 B3 D-9 指标 | 落实 B3 dispatcher 协议；为 F2 analytics / 各域 audit consumer 提供入口，同时防止 outbox 行被无消费者确认后丢失 |
| D-7 | `source_event_only` 跳过 | dispatcher 必须读取 B3 generated `IsSourceEventOnly(jobType)` 谓词，对应 outbox 事件**不**触发 dispatcher 二次创建 `async_jobs`；具体 binding 由业务事务自身在产生 source event 时同事务写 job row（[B3 D-15](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表)） | 对齐 backend-practice/002 D-32 `practice.session.completed` ↔ `report_generate` forward-binding；防止 outbox 重放造成第二个 report job |
| D-8 | Shutdown 协调 | `cmd/api` 监听 SIGINT/SIGTERM；signal context 触发后 kernel 按以下顺序 drain：(a) 停止接受新的 lease，(b) 等待 in-flight handler 完成（subject to `async.shutdownGraceSeconds`，默认 10s），(c) 停止 reaper，(d) 停止 outbox dispatcher；超时 SLA 内未完成的 in-flight handler 由下一次启动 reaper 回收 | 替代当前 4 个独立 `defer Shutdown`；mailDispatcher channel 与 reportRuntime defer 一并删除 |
| D-9 | Queue weights | 注入 [A4 D-9](../secrets-and-config/spec.md) `async.queueWeights{critical:6, default:3, low:1}` typed config；priority 划分按 ADR-Q2 §3.5：`critical=report_generate, privacy_delete`；`default=target_import, resume_tailor, resume_parse, debrief_generate, source_refresh`；`low=email_dispatch` | P0 仅作为 lease 选择的次级排序键；DB 列默认按 `available_at asc, created_at asc`，priority 用于 fair scheduling 实现 |
| D-10 | Email dispatch 收口 | P3 把 `email_dispatch` 从 C1 进程内 channel 迁移到 `async_jobs` 行：C1 `MailDispatcher.Enqueue` 改为 `INSERT INTO async_jobs(job_type='email_dispatch', payload, ...)`；kernel 注册 `EmailDispatchHandler` 通过既有 `DeliveryWriter` sink 发邮件；payload 仍受 [B3 D-12 redaction redline](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表) 与 `shared/jobs/jobs.go` `EmailDispatchPayload` validator 约束 | 删除 `auth.BackgroundMailDispatcher`；C1 `DevMailSink` 写入路径不变 |
| D-11 | TraceId 透传 | kernel 在 handler 调用前重建 W3C `traceparent` span（如 outbox payload 中存在 `traceId`）；logger 在 handler / dispatcher 日志中注入 `trace_id` 字段；缺失 `traceId` 时 dispatcher 写 warn log 后继续发布（[B3 D-10](../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表)） | 与 F1 trace propagation 对齐 |
| D-12 | Legacy zero-reference gate | 新增 lint script `scripts/lint/runner_legacy.py`（与既有 `backend_review_legacy.py` / `backend_practice_legacy.py` 同风格的 Python lint，配套 `runner_legacy_test.py`），Makefile target 名 `lint-runner-legacy`；扫描旧 `review.NewRunner` / `review.NewReaper` / `review.ComputeReportFailureBackoff` / `review.DefaultReportFailureBackoff` / `auth.BackgroundMailDispatcher` / `auth.NewBackgroundMailDispatcher` / 多处 `targetjob.NewDrainer(targetjob.DrainerOptions{...})` 实例化（除 kernel adapter shim 与本文件历史断言外）必须 0 命中 | 防止旧 runner 形态回流；history/tests/lint 自身允许保留历史证据；脚本扩展名与 `scripts/lint/` 既有 Python 工具集对齐，便于复用 grep 排除规则 |
| D-13 | `Runtime.RunOnce` 必须暴露 | kernel `Runtime` 对外暴露 `RunOnce(ctx) (processed bool, err error)`（与 §4.1 接口约束一致），供 unit test / integration test 同步驱动 lease + finalize 单条流程；与既有 `targetjob.Drainer.RunOnce` / `review.Runner.RunOnce` 习惯保持一致 | 锁定测试驱动 API，避免实施期重新讨论；与 §4.1 「`Runtime` 必须暴露」清单冗余但显式 |
| D-14 | Runtime lease loop scan 间隔 | kernel `Runtime` 默认 lease loop scan 间隔由新增 typed config `async.scanIntervalSeconds` 提供（默认 5s，与 outbox dispatcher 5s scan 对齐）；缺失或非正数 fail-fast；与 D-9 priority bucket 共同决定 lease 选择 | 锁定 `email_dispatch` 等 low-priority job 的可见延迟上限「≤1 个 scan 周期」；配合 plan §6 风险表对 magic link 投递延迟的承诺 |

### 3.2 待确认事项

当前无待用户确认项。`async.leaseTimeoutSeconds` / `async.reaperIntervalSeconds` / `async.shutdownGraceSeconds` / `async.scanIntervalSeconds` 已按 D-5 / D-8 / D-14 锁定为 A4 additive config-only typed 节点，不新增 env key；实施期若发现 A4 owner spec 已有冲突决策，必须先回到 plan-review / design 修订，不得静默回退为代码常量。

## 4 设计约束

### 4.1 Kernel 接口约束

- `runner.Runtime` 必须暴露：`Register(jobType string, handler runner.Handler)`、`Start(ctx context.Context)`、`Shutdown(ctx context.Context) error`、`RunOnce(ctx context.Context) (processed bool, err error)`、`ReapOnce(ctx context.Context) (reclaimed int64, err error)`。
- `runner.Handler.Handle` 返回 `runner.JobOutcome{Succeeded, Retryable, ErrorCode, ErrorMessage, AsyncJobFinalized}`；`AsyncJobFinalized` 用于既有 review handler 内部自行 finalize 的兼容路径（P2 迁移完成后可移除）。
- kernel 不持有任何业务状态机；handler 业务态由各域 handler / store 独立维护。

### 4.2 SQL 约束

- 列名仅允许 `async_jobs.locked_at` / `attempts` / `max_attempts` / `available_at` / `status` ∈ {`queued`,`running`,`succeeded`,`failed`,`cancelled`,`dead`}；与 [B4 baseline](../db-migrations-baseline/spec.md) 完全一致；禁止引入 `worker_id` / `leased_at` / `attempt_count` 等同义列。
- `outbox_events` retry 列名沿用 [B3 §2.1](../event-and-outbox-contract/spec.md#21-in-scope)：`publish_attempts`、`next_attempt_at`、`locked_at`、`last_error_code`、`last_error_message`；不引入新 retry 列。
- Lease 查询固定为 `SELECT ... FROM async_jobs WHERE status='queued' AND available_at <= now() AND job_type IN (...) ORDER BY available_at ASC, created_at ASC FOR UPDATE SKIP LOCKED LIMIT 1`；outbox dispatcher 同样使用 `FOR UPDATE SKIP LOCKED`。
- Reaper 查询固定为 `UPDATE async_jobs SET status='queued', locked_at=NULL, available_at=now()+backoff WHERE status='running' AND locked_at <= now() - leaseTimeout`；attempts 不递增（视为 lease 超时回收，非业务失败）。

### 4.3 退避与 dead-letter 约束

- 业务失败 retry 退避表：`[30s, 2m, 10m, 1h, 6h]`，由 `runner.BackoffPolicy` 单点暴露；retryable outcome 在 `attempts < max_attempts` 时 finalize 为 `queued` 并推进 `available_at`，在 `attempts >= max_attempts` 时 finalize 为 `dead`；non-retryable outcome finalize 为 `failed`；不得让各域 handler 私自选择不同终态。
- Lease 超时回收**不**算 attempt 次数；回收后 `available_at` 仍按业务退避表前移避免 thrash。
- Outbox dispatcher 退避表与 job retry 一致（均沿用 [ADR-Q2 §3.4](../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) / B3 D-7）。
- Dead-letter 不允许 silent drop：必须 emit `async_jobs_processed_total{result="dead"}` Counter 与 audit_event；outbox dead-letter 必须 emit `outbox_publish_failures_total` 并写 redacted `last_error_message`。

### 4.4 红线与可观测性约束

- handler payload 红线沿用各域 owner spec（B3 §4.2 + `shared/jobs/jobs.go` 现有 validator）；kernel 不在 finalize 时反序列化业务 payload，避免成为新红线泄漏点。
- `outbox_events.last_error_message` 与 `async_jobs.error_message` 必须 redacted；不允许写入 raw provider response / prompt body / answer text / file URL / email plaintext。
- 必产指标（与 ADR-Q2 §3.6 + B3 D-9 对齐）：
  - `async_job_duration_seconds{job_type}` Histogram
  - `async_jobs_processed_total{job_type, result}` Counter
  - `async_job_queue_depth{job_type}` Gauge
  - `async_job_lag_seconds{job_type}` Gauge
  - `outbox_events_pending` Gauge
  - `outbox_publish_duration_seconds` Histogram
  - `outbox_publish_failures_total` Counter
- 必产结构化日志字段：`job_type` / `job_id` / `attempts` / `outcome` / `trace_id`（如存在）。

### 4.5 兼容与迁移约束

- P2 迁移期允许 `runner.FromTargetjobHandler(targetjob.JobHandler)` 适配 shim 继续将既有 handler 注册到 kernel；P4 完成后可移除该 shim。
- P3 之前各域产生 outbox 行的写法不变（业务事务直插 `outbox_events`）；P3 dispatcher 上线后自动消费。
- P4 完成后必须保证：`backend/cmd/api/main.go` 中 `mailDispatcher` 变量、`reportRuntime.Runner` / `reportRuntime.Reaper` 字段、`resumeRuntime.Drainer` 字段、`targetJobRuntime.Drainer` 字段全部由 `runner.Runtime` 单实例替换；4 个独立 `defer Shutdown(ctx)` 收敛为 1 个。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| Kernel 实现 | `backend-async-runner`（本 spec） | `backend/internal/runner/`：Runtime / Registry / Lease / Reaper / OutboxDispatcher / BackoffPolicy / Adapter shim |
| Handler 业务实现 | 各 backend-* domain spec | `target_import`/`source_refresh` 归 [backend-targetjob](../backend-targetjob/spec.md)；`privacy_delete` 归 [backend-targetjob](../backend-targetjob/spec.md) §「删除链路」+ `backend/internal/privacy/runner/`；`debrief_generate` 归 [backend-debrief](../backend-debrief/spec.md)；`resume_parse`/`resume_tailor` 归 [backend-resume](../backend-resume/spec.md)；`report_generate` 归 [backend-review](../backend-review/spec.md)；`email_dispatch` handler 归 [backend-auth](../backend-auth/spec.md) + 本 spec D-10 |
| Async jobs DDL | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) | `async_jobs` / `outbox_events` schema + retry columns |
| Event / job 契约 | [B3 event-and-outbox-contract](../event-and-outbox-contract/spec.md) | canonical `job_type` ↔ dotted name 映射、payload schema、`IsSourceEventOnly` 谓词 |
| 异步配置 | [A4 secrets-and-config](../secrets-and-config/spec.md) | `async.queueWeights` + 本 spec 待确认的 `async.leaseTimeoutSeconds` / `async.shutdownGraceSeconds` / `async.reaperIntervalSeconds` typed config |
| Trace / 指标 / 日志 | [F1 observability-stack](../observability-stack/spec.md) | metric 命名字典 + trace propagation；kernel 仅生产指标，不持有 Prometheus 实例 |
| Outbox consumer 注册 | 各 domain spec + 本 spec | 本 spec 暴露 consumer registration API；具体 consumer 实现（如 analytics 双发）归对应 owner；本 spec 负责保证未注册 consumer 不被默认 ack |
| Cmd/api bootstrap | 本 spec D-1 / D-8 | 单点持有 `runner.Runtime`；handler 注册保留在各域 build 函数中 |
| Runtime topology | [backend-runtime-topology](../backend-runtime-topology/spec.md) | 「不建独立 worker 进程」前置；本 spec 沿用 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | Kernel 接口最小可用 | 新建 `backend/internal/runner/` package | unit test 调用 `Runtime.Register/Start/Shutdown/RunOnce/ReapOnce` 通过 fake store | 注册 / lease / finalize / retry / reaper / shutdown 五条路径均产出预期断言 | 001 Phase 1 |
| C-2 | 统一退避表 | kernel `BackoffPolicy` 配置默认值 | `BackoffPolicy.Next(attempts)` for attempts 1..5 | 返回 `[30s,2m,10m,1h,6h]`；retryable attempts≥5 finalize 为 `dead`，non-retryable finalize 为 `failed` | 001 Phase 1 |
| C-3 | Reaper 覆盖全部 job_type | `async_jobs` 中存在 `target_import` / `privacy_delete` / `debrief_generate` 等 stuck `running` 行 | `Runtime.ReapOnce` | 全部超过 lease timeout 的行被回收为 `queued` + `locked_at=NULL`；attempts 不变 | 001 Phase 1 |
| C-4 | Graceful shutdown | handler 在 in-flight 中收到 ctx cancel | `Runtime.Shutdown(ctx)` with grace timeout | in-flight handler 在 timeout 内完成；新 lease 在 Shutdown 后被拒；超时后 ctx.Err 返回上层 | 001 Phase 1 |
| C-5 | target_import / source_refresh 迁移 | kernel 已上线 | 移除 `targetJobRuntime.Drainer` 独立实例，注册到 kernel | `targetjob/pipeline_test.go` / `targetjob/e2e_scenario_test.go` rerun PASS | 001 Phase 2 |
| C-6 | privacy_delete 迁移 | privacy handler 已注册到 kernel | `DELETE /api/v1/me` 触发 privacy 链路 | `privacy/runner/delete_handler_test.go` PASS；用户端可观察 privacy.request → completed | 001 Phase 2 |
| C-7 | debrief_generate 迁移 | debrief handler 注册到 kernel | `debrief.created` outbox + debrief job row | `debrief/generate_handler_test.go` / `service_test.go` PASS；`debriefs.status` 进入 `completed` | 001 Phase 2 |
| C-8 | resume_parse / resume_tailor 迁移 | resume handlers 注册到 kernel | 触发 resume 注册 + tailor run | `resume/jobs/*_test.go` + `cmd/api/resume_parse_drainer_scenario_test.go` / `resume_tailor_drainer_scenario_test.go` PASS | 001 Phase 2 |
| C-9 | report_generate 迁移；review.Runner/Reaper 删除 | report job row queued | kernel lease 接管 | `review/runner_test.go` / `review/reaper_test.go` 重写到 kernel 后 PASS；旧 `review.NewRunner` / `review.NewReaper` 在 lint 中 0 命中 | 001 Phase 2 |
| C-10 | Outbox dispatcher 协议 | `outbox_events` 中存在 pending 行 | `OutboxDispatcher.RunOnce` against 真 PG | 5s scan + SKIP LOCKED + batch≤100 + sort by `next_attempt_at asc, created_at asc` 行为可断言；publish 成功后 `publish_status='published'` | 001 Phase 3 |
| C-11 | Outbox 退避 + dead-letter | dispatcher publish 临时失败 5 次 | 5 次 retry 后 | `publish_attempts=5` → `publish_status='failed'` + redacted `last_error_message`；emit `outbox_publish_failures_total` | 001 Phase 3 |
| C-12 | Outbox at-least-once + idempotency | 同一 `eventId` 被 dispatcher 投递两次 | consumer | 业务表只更新一次；duplicate eventId 不破坏业务状态 | 001 Phase 3 |
| C-13 | `source_event_only` 跳过 | `practice.session.completed` outbox 行写入 | OutboxDispatcher | 不创建新 `report_generate` job row；`async_jobs(report_generate)` 仍由 `completePracticeSession` 同事务创建 | 001 Phase 3 |
| C-13a | 未注册 consumer 不被确认 | runtime dispatcher 没有注册某 `event_name` 的 consumer | `OutboxDispatcher.RunOnce` 处理该 outbox 行 | 行不得置为 `published`；按 retry / failed dead-letter 语义记录 redacted missing-consumer 错误；test-only dry-run consumer 必须显式注入 | 001 Phase 3 |
| C-14 | Email dispatch 收口 | C1 `MailDispatcher.Enqueue` 调用 | kernel lease `email_dispatch` | `async_jobs(email_dispatch)` 行被消费；`DeliveryWriter.Write` 收到 payload；旧 `auth.BackgroundMailDispatcher` 在 lint 中 0 命中 | 001 Phase 3 |
| C-15 | TraceId 透传 | outbox payload 含 `traceId` | dispatcher → handler | handler / dispatcher log 中可见 `trace_id` 字段；F1 trace 串联 | 001 Phase 3 |
| C-16 | Cmd/api 单点 shutdown | SIGTERM 到达 cmd/api | signal context | 单一 `runner.Runtime.Shutdown(ctx)` 调用替代 4 个独立 defer；`cmd/api/main_test.go` lifecycle 断言通过 | 001 Phase 4 |
| C-17 | Queue weights 注入 | `async.queueWeights{critical:6,default:3,low:1}` 已 typed config 化 | kernel boot | `Runtime` 初始化读取 weights；priority bucket 按 D-9 划分；缺失或非正数 fail-fast（沿用 A4 C-12） | 001 Phase 1 |
| C-18 | 既有 BDD 场景回归 | 各 owner spec BDD 场景已存在 | P2 / P3 / P4 完成后 rerun owner BDD suite | auth email / privacy_delete / target_import / report_generate / debrief_generate / resume_parse / resume_tailor 用户路径全部 PASS | 001 Phase 4 |
| C-19 | Legacy zero-reference gate | spec D-12 列出的旧形态 | `make lint-runner-legacy` | 旧 `review.NewRunner` / `review.NewReaper` / `ComputeReportFailureBackoff` / `auth.BackgroundMailDispatcher` / 多 `targetjob.NewDrainer` 实例化（除 kernel adapter 与 history 证据外）0 命中；lint 失败时输出文件 / 行号 | 001 Phase 4 |
| C-20 | Runtime lease scan + grace + lease timeout + reaper interval typed config | 新增 `async.scanIntervalSeconds` / `leaseTimeoutSeconds` / `shutdownGraceSeconds` / `reaperIntervalSeconds` typed config 节点（A4 additive config-only） | kernel boot + reaper loop | scanIntervalSeconds 默认 5s 与 outbox dispatcher 一致；缺失或非正数 fail-fast（沿用 A4 C-12）；`email_dispatch` 等 low-priority job 可见延迟 ≤1 个 scan 周期 | 001 Phase 1 |

## 7 关联计划

- [001-internal-job-outbox-runner](./plans/001-internal-job-outbox-runner/plan.md)：本 spec 的唯一 active 实施计划，覆盖 D-1~D-14 全部决策。Phase 1 落地 kernel；Phase 2 迁移 7 个业务 handler；Phase 3 落地 outbox dispatcher + email_dispatch；Phase 4 收口 cmd/api shutdown + lint negative gate + 旧 owner spec D-* 边界条款同步。

未来如需新增独立 Asynq server / 拆运行单元 / 引入新调度器，必须先 supersede ADR-Q2 + 新 plan 显式承接本 spec D-* 决策。
