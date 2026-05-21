# Internal Job and Outbox Runner Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-21

**关联计划**: [plan](./plan.md)

## Phase 1: Kernel 基础设施

- [ ] 1.1 新建 `backend/internal/runner/` package 骨架（`runtime.go` / `lease.go` / `reaper.go` / `backoff.go` / `handler.go` / `adapter_targetjob.go` / `doc.go`），go build PASS；test 来源 `backend/internal/runner/runtime_test.go::TestRuntime_RegisterAndRunOnce`
- [ ] 1.2 实现 `LeaseAsyncJob` / `FinalizeAsyncJob` / `ReclaimExpiredLeases` SQL contract（列名 `locked_at`/`attempts`/`available_at`/`status`，`FOR UPDATE SKIP LOCKED`，排序 `available_at asc, created_at asc`；retryable exhausted → `dead`，non-retryable → `failed`）；test 来源 `backend/internal/runner/lease_test.go` + `lease_integration_test.go`
- [ ] 1.3 实现 `BackoffPolicy.Next(attempts)` 返回 `[30s,2m,10m,1h,6h]`、`MaxAttempts=5` 常量；test 来源 `backend/internal/runner/backoff_test.go` table-driven
- [ ] 1.4 实现 `Reaper.RunOnce(ctx)` 调用 `ReclaimExpiredLeases`；attempts 不递增；test 来源 `backend/internal/runner/reaper_test.go` + 多 jobType fixture
- [ ] 1.5 实现 `Runtime.Shutdown(ctx)` 顺序协调（stop lease → wait in-flight → stop reaper → stop outbox dispatcher）；test 来源 `backend/internal/runner/runtime_test.go::TestRuntime_GracefulShutdown` + `TestRuntime_ShutdownTimeoutPropagates`
- [ ] 1.6 typed config 注入：A4 spec additive 新增 `async.leaseTimeoutSeconds` / `async.shutdownGraceSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds`（[spec D-14](../../spec.md#31-已锁定决策)，config-only，不新增 env key，不允许静默 fallback 为代码常量）；test 来源 `backend/internal/runner/config_test.go` + `backend/internal/platform/config/loader_test.go::TestAsyncSection`
- [ ] 1.7 单元测试基础设施：fake `LeaseStore` + deterministic clock；test 来源 `backend/internal/runner/fakestore_test.go`
- [ ] 1.8 Runtime handler trace 透传：从 `async_jobs.payload` / envelope 读取 W3C `traceparent` 重建 span context；slog 输出注入 `trace_id` 字段（缺失时跳过）；test 来源 `backend/internal/runner/runtime_trace_test.go::TestRuntime_HandlerInheritsTraceparent` + `TestRuntime_HandlerLogsTraceIdField`

## Phase 2: 业务 handler 迁移

- [ ] 2.1 `target_import` + `source_refresh` 注册到 kernel；rerun `backend/internal/targetjob/pipeline_test.go::TestDrainer_RunOnceProcessesQueuedJobAndFinalizes` + `e2e_scenario_test.go`；删除 `targetJobRuntime.Drainer` 独立实例
- [ ] 2.2 `privacy_delete` 注册到 kernel；rerun `backend/internal/privacy/runner/delete_handler_test.go::TestPrivacyDeleteHandler_DeletesUploadFilesForUser`；smoke `DELETE /api/v1/me` 触发链路 → `privacy_requests.status='completed'`
- [ ] 2.3 `debrief_generate` 注册到 kernel；rerun `backend/internal/debrief/generate_handler_test.go` + `service_test.go::TestService_CreateDebrief_GeneratesAndCompletes`
- [ ] 2.4 `resume_parse` + `resume_tailor` 注册到 kernel；删除 `resumeRuntime.Drainer` 独立实例；删除 `backend/internal/resume/store/async.go` 中固定 15s retry；rerun `backend/internal/resume/jobs/parse_test.go` + `tailor_test.go` + `backend/cmd/api/resume_parse_drainer_scenario_test.go` + `resume_tailor_drainer_scenario_test.go`
- [ ] 2.5 `report_generate` 注册到 kernel；新建 `backend/internal/review/generate_handler.go` 实现 `runner.Handler`；删除 `backend/internal/review/runner.go` / `reaper.go` / `lease.go` 中 `ComputeReportFailureBackoff`+`DefaultReportFailureBackoff`；rerun `backend/internal/review/runner_test.go`（重写到 kernel）+ `reaper_test.go`（重写到 kernel reaper）+ `backend/cmd/api/reports_http_scenario_test.go`
- [ ] 2.6 退避收口：focused test `backend/internal/runner/backoff_integration_test.go::TestAllHandlersUseSharedBackoff` 验证全部 handler 走 `BackoffPolicy.Next`；负向 grep 旧 hard-coded 退避（除 history / tests / lint 自身）0 命中

## Phase 3: Outbox dispatcher + email_dispatch

- [ ] 3.1 实现 `runner.OutboxDispatcher.RunOnce(ctx)`：5s scan + `FOR UPDATE SKIP LOCKED` + batch≤100 + sort by `next_attempt_at asc, created_at asc`；consumer ack 成功才置 `published`，临时失败或缺少 consumer 按 `BackoffPolicy` 后移 `next_attempt_at`；test 来源 `backend/internal/runner/outbox_test.go` + `outbox_integration_test.go`（真 PG）
- [ ] 3.2 `Dispatcher.RegisterConsumer(eventName, consumer)` API；dry-run consumer 只能测试显式注入，runtime 缺少 consumer 时不得置 `published`；test 来源 `backend/internal/runner/outbox_consumer_test.go::TestDispatcherMissingConsumerDoesNotAck` + `TestDispatcherDryRunConsumerRequiresExplicitRegistration`
- [ ] 3.3 `source_event_only` 跳过：dispatcher 读取 `jobs.IsSourceEventOnly(jobType)`；test 来源 `backend/internal/runner/outbox_source_event_only_test.go` 覆盖 `practice.session.completed` 不二次造 job
- [ ] 3.4 TraceId 透传：dispatcher 重建 W3C `traceparent` span，logger 字段 `trace_id`；缺失时 warn log + 继续 publish；test 来源 `backend/internal/runner/outbox_trace_test.go` + slog asserter
- [ ] 3.5 指标暴露：`outbox_events_pending` Gauge / `outbox_publish_duration_seconds` Histogram / `outbox_publish_failures_total` Counter / `async_job_queue_depth{job_type}` / `async_job_lag_seconds{job_type}` / `async_job_duration_seconds{job_type}` / `async_jobs_processed_total{job_type,result}`；test 来源 `backend/internal/runner/metrics_test.go` 断言指标 family 命名与 label
- [ ] 3.6 dead-letter 路径：attempt≥5 → `publish_status='failed'` + redacted `last_error_message` + emit failure counter；test 来源 `backend/internal/runner/outbox_dead_letter_test.go`（真 PG）
- [ ] 3.7 `email_dispatch` producer 切换：修订 `backend/internal/auth/passwordless.go` / `mail.go` 使 `MailDispatcher.Enqueue` 改为 `INSERT INTO async_jobs(job_type='email_dispatch')` 同事务 enqueue；test 来源 `backend/internal/auth/passwordless_test.go::TestStartAuthEmailChallenge_EnqueuesEmailDispatchJob`
- [ ] 3.8 `email_dispatch` handler：新建 `backend/internal/auth/email_dispatch_handler.go` 实现 `runner.Handler`，通过 `DeliveryWriter` 发邮件；保留 `EmailDispatchPayload` validator；删除 `auth.BackgroundMailDispatcher`/`NewBackgroundMailDispatcher`；test 来源 `backend/internal/auth/email_dispatch_handler_test.go` + `backend/cmd/api/main_test.go::TestAuthEmailEndToEnd`

## Phase 4: 收口 / 文档同步 / lint negative gate

- [ ] 4.1 `backend/cmd/api/main.go` 单点 `runtime *runner.Runtime`；删除 `mailDispatcher` / `targetJobRuntime.Drainer` / `resumeRuntime.Drainer` / `reportRuntime.Runner` / `reportRuntime.Reaper` 字段；signal context 触发统一 `runtime.Shutdown(ctx)`；test 来源 `backend/cmd/api/main_test.go::TestMain_SingleRuntimeShutdown`
- [ ] 4.2 删除 `backend/internal/review/runner.go` / `reaper.go` / `lease.go` legacy 实现；把 `runner_test.go` / `reaper_test.go` 中仍有价值的断言迁移到 kernel / `GenerateHandler` / structure negative tests；test 来源 `git ls-files backend/internal/review` 断言 legacy 实现文件不存在
- [ ] 4.3 删除 `backend/internal/auth/mail.go` 中 `BackgroundMailDispatcher` / `NewBackgroundMailDispatcher` 及 `BackgroundMailDispatcherOptions`；保留 `DevMailSink` / `ImmediateMailDispatcher` / `DeliveryWriter` 接口；test 来源 `backend/internal/auth/mail_test.go::TestNoBackgroundDispatcher`
- [ ] 4.4 新增 `scripts/lint/runner_legacy.py`（+ `runner_legacy_test.py`，与 `backend_review_legacy.py` / `backend_practice_legacy.py` 同风格 Python lint）；扫描 [spec D-12](../../spec.md#31-已锁定决策) 列出旧 entry point；接入 `make lint-runner-legacy`；test 来源 `scripts/lint/runner_legacy_test.py`
- [ ] 4.5 同步 `backend-runtime-topology` § 模块边界 / § 验收：把「backend internal runner 实现」owner 改为 `backend-async-runner`；test 来源 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + 精确 grep（如 `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner\|未来 .backend-async-runner" docs/spec/backend-runtime-topology/spec.md` 期望 0 命中）
- [ ] 4.6 同步 `backend-review` D-13 / D-16：标注 review.Runner / Reaper 已由本 plan 接管；保留 history 证据；test 来源 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + `grep -n "未来 .backend-async-runner" docs/spec/backend-review/spec.md` 期望 0 命中（D-13 / D-16 / Q-4 已改为 active 引用）
- [ ] 4.7 同步 `backend-debrief` D-5 / § 模块边界：删除「未来 backend-async-runner 接管」时态；test 来源 `grep -n "未来 .backend-async-runner" docs/spec/backend-debrief/spec.md` 期望 0 命中
- [ ] 4.8 同步 `backend-targetjob` D-5 / § 模块边界：Drainer 抽象迁到 kernel；test 来源 `grep -n "未来 .backend-async-runner" docs/spec/backend-targetjob/spec.md` 期望 0 命中
- [ ] 4.9 同步 `backend-resume` § 模块边界：handler 注册改为 kernel；test 来源 `grep -n "未来 .backend-async-runner" docs/spec/backend-resume/spec.md` 期望 0 命中
- [ ] 4.10 同步 `backend-auth` D-* / `email_dispatch`：`BackgroundMailDispatcher` 已删除；producer 通过 `async_jobs(email_dispatch)` enqueue；test 来源 `grep -n "BackgroundMailDispatcher\|进程内 channel" docs/spec/backend-auth/spec.md` 期望 0 命中
- [ ] 4.11 同步 `event-and-outbox-contract` § 模块边界：「backend internal runner 实现」owner 改为 `backend-async-runner`；test 来源 `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner\|未来 .backend-async-runner" docs/spec/event-and-outbox-contract/spec.md` 期望 0 命中
- [ ] 4.12 同步 `secrets-and-config` D-9 / § config dictionary：additive 新增 `async.leaseTimeoutSeconds` / `async.shutdownGraceSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds`；test 来源 `backend/internal/platform/config/loader_test.go` + `scripts/lint/env_dict.py`
- [ ] 4.13 同步 `engineering-roadmap` §5.2 / §6.3 S2：确认 `backend-async-runner` active subject 引用；test 来源 `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner" docs/spec/engineering-roadmap/spec.md` 期望 0 命中
- [ ] 4.14 BDD-Gate: Owner BDD rerun 全 PASS：auth email + privacy_delete `E2E.P0.003` / `033`；target_import `E2E.P0.010` / `012` / `013`；report_generate `E2E.P0.041` / `052` / `054` / `055`；debrief_generate `E2E.P0.060` / `062`；resume_parse `E2E.P0.034` / `035`；resume_tailor `E2E.P0.077` / `078` / `080`；evidence 归档到 [test-checklist Phase 4](./test-checklist.md)
- [ ] 4.15 全局 drift gate：`cd backend && go build ./...` / `cd backend && go vet ./...` / `cd backend && go test ./...` / `validate_context.py` / `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` / `make lint-runner-legacy` / `git diff --check` 全部 PASS
- [ ] 4.16 状态收尾：plan 状态 `active`→`completed`（spec.md / history.md 已在创建时即 `active`，无 `draft` 中间态）；plans INDEX + spec INDEX 同步；提交工作日志
