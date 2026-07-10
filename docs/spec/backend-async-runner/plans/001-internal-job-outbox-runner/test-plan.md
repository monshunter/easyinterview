# Internal Job and Outbox Runner Test Plan

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)
**关联 Checklist**: [checklist](./checklist.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 测试目标

为 backend in-process runtime kernel + outbox dispatcher 提供跨 phase 的覆盖矩阵；保证 D-1~D-14 决策的实现路径都有可执行测试断言，并把既有 owner 域 handler 的回归基线纳入本计划，避免迁移过程中行为漂移。

## 2 覆盖矩阵

### 2.1 Primary path

| 行 | 决策 / spec 引用 | Phase | 测试入口 | 类型 |
|----|------------------|-------|----------|------|
| P1-1 | spec D-1 / D-2 kernel 接口 | 1 | `backend/internal/runner/runtime_test.go::TestRuntime_RegisterAndRunOnce` | unit |
| P1-2 | spec D-3 lease SQL contract | 1 | `backend/internal/runner/lease_integration_test.go::TestLeaseAsyncJob_ClaimsQueuedRow` | integration (PG) |
| P1-3 | spec D-3 finalize succeed/retry/fail | 1 | `backend/internal/runner/lease_integration_test.go::TestFinalizeAsyncJob_*` | integration (PG) |
| P1-4 | spec D-4 退避表 | 1 | `backend/internal/runner/backoff_test.go::TestBackoffPolicy_Next_Table` | unit |
| P1-5 | spec D-5 reaper | 1 | `backend/internal/runner/reaper_test.go::TestReaper_ReclaimsExpiredLeases` | unit + integration |
| P1-6 | spec D-8 shutdown 顺序 | 1 | `backend/internal/runner/runtime_test.go::TestRuntime_GracefulShutdown` | unit |
| P2-1 | spec C-5 target_import / source_refresh | 2 | `backend/internal/targetjob/pipeline_test.go` + `e2e_scenario_test.go` rerun | regression |
| P2-2 | spec C-6 privacy_delete | 2 | `backend/internal/privacy/runner/delete_handler_test.go` rerun + cmd/api smoke | regression + smoke |
| P2-3 | D-22 out-of-scope module guard | 2 | Out-of-scope module negative search / current package absence check | regression-negative |
| P2-4 | spec C-8 resume_parse / resume_tailor | 2 | `backend/internal/resume/jobs/*_test.go` + `backend/cmd/api/resume_parse_runner_scenario_test.go` + `resume_tailor_runner_scenario_test.go` rerun | regression + scenario |
| P2-5 | spec C-9 report_generate；review.Runner/Reaper 删除 | 2 | `backend/internal/review/runner_test.go` / `reaper_test.go` 有价值断言迁移到 kernel / `GenerateHandler` tests + `backend/cmd/api/reports_http_scenario_test.go` rerun | unit + regression |
| P2-6 | D-22 out-of-scope jobs-recommendations guard | 2 | `make lint-runner-out-of-scope` + out-of-scope module negative search | regression-negative |
| P3-1 | spec C-10 outbox 5s/skip-locked/batch/sort | 3 | `backend/internal/runner/outbox_integration_test.go::TestOutboxDispatcher_ClaimsPendingBatch` | integration (PG) |
| P3-2 | spec C-15 trace_id 透传 | 3 | `backend/internal/runner/outbox_trace_test.go::TestOutboxDispatcher_PropagatesTraceParent` | unit |
| P3-3 | spec C-14 email_dispatch 收口 | 3 | `backend/internal/auth/email_dispatch_handler_test.go::TestStartAuthEmailChallenge_EnqueuesEmailDispatchJob` + `email_dispatch_handler_test.go` | unit + integration |
| P4-1 | spec C-16 cmd/api 单点 shutdown | 4 | `backend/cmd/api/main_test.go::TestMain_SingleRuntimeShutdown` | unit + integration |
| P4-2 | spec C-22 / BUG-0106 privacy identity cleanup | 4 | `backend/internal/auth` DeleteMe store/service tests + `backend/internal/privacy/runner/delete_handler_test.go` + `backend/cmd/api` privacy integration regression | unit + integration |
| P4-3 | canonical runner contract cleanup | 4 | `scripts/lint/runner_out_of_scope_test.py` + native handler compile assertions + targetjob/resume cmd/api runner scenarios | contract + regression + scenario |

### 2.2 Alternate path

| 行 | 决策 | Phase | 入口 | 类型 |
|----|------|-------|------|------|
| A-2 | spec D-9 queue priority bucket（critical / default / low）与 production per-job lease loop 防饥饿 | 1 / 4 | `backend/internal/runner/lease_test.go::TestLeaseAsyncJob_RespectsPriorityBuckets` + `backend/internal/runner/runtime_test.go::TestRuntime_StartDoesNotLetCriticalJobStarveEmailDispatch` | unit |
| A-3 | spec D-11 outbox payload 缺失 traceId 时降级 warn log | 3 | `backend/internal/runner/outbox_trace_test.go::TestOutboxDispatcher_WarnsOnMissingTrace` | unit |

### 2.3 Failure / recovery path

| 行 | 决策 | Phase | 入口 | 类型 |
|----|------|-------|------|------|
| F-1 | spec D-4 attempts 达 max 后 retryable finalize 为 `dead`，non-retryable finalize 为 `failed` | 1 | `backend/internal/runner/lease_test.go::TestFinalizeAsyncJob_PermanentFailureAtMax` + `TestFinalizeAsyncJob_NonRetryableFailure` | unit |
| F-2 | spec D-5 reaper 超时回收不递增 attempts | 1 | `backend/internal/runner/reaper_test.go::TestReaper_DoesNotIncrementAttempts` | unit |
| F-3 | spec C-4 handler 卡住时 Shutdown 在 timeout 内返回 ctx.Err | 1 | `backend/internal/runner/runtime_test.go::TestRuntime_ShutdownTimeoutPropagates` | unit |
| F-4 | spec C-11 outbox dead-letter | 3 | `backend/internal/runner/outbox_dead_letter_test.go::TestOutboxDispatcher_DeadLetterAtAttemptFive` | integration (PG) |
| F-5 | spec C-11 outbox dead-letter 写 redacted `last_error_message` | 3 | `backend/internal/runner/outbox_dead_letter_test.go::TestOutboxDispatcher_RedactsLastError` | integration (PG) |
| F-6 | spec C-21 handler 执行时间超过 backoff 时 finalize 使用 handler 返回后的 fresh timestamp | 4 | `backend/internal/runner/runtime_test.go::TestRuntime_FinalizeUsesTimestampAfterHandlerReturns` | unit |

### 2.4 Boundary / idempotency

| 行 | 决策 | Phase | 入口 | 类型 |
|----|------|-------|------|------|
| B-1 | spec C-2 退避表 attempts<1 / attempts≥len 边界 | 1 | `backend/internal/runner/backoff_test.go::TestBackoffPolicy_BoundaryAttempts` | unit |
| B-2 | spec C-10 outbox batch≤100 边界 | 3 | `backend/internal/runner/outbox_integration_test.go::TestOutboxDispatcher_BatchSizeLimit` | integration |
| B-3 | spec C-12 outbox at-least-once + consumer idempotency | 3 | `backend/internal/runner/outbox_integration_test.go::TestOutboxDispatcher_DuplicateEventIdHandledIdempotently` | integration |
| B-4 | spec C-13 `source_event_only` skip | 3 | `backend/internal/runner/outbox_source_event_only_test.go::TestOutboxDispatcher_SkipsSourceEventOnly` | integration |
| B-5 | spec C-13a 缺少 consumer 不得 ack | 3 | `backend/internal/runner/outbox_consumer_test.go::TestDispatcherMissingConsumerDoesNotAck` + `TestDispatcherDryRunConsumerRequiresExplicitRegistration` | unit + integration |
| B-6 | out-of-scope async job names 不回流 | 2 / 4 | `make lint-runner-out-of-scope` 与 active-doc negative grep | lint |

### 2.5 Cross-layer contract

| 行 | 决策 | Phase | 入口 | 类型 |
|----|------|-------|------|------|
| X-1 | spec D-3 列名必须沿用 B4 baseline | 1 | `backend/internal/runner/lease_integration_test.go::TestLeaseAsyncJob_ColumnNames` + `backend/internal/migrations/sql_contract_test.go` rerun | integration |
| X-2 | spec C-17 `async.queueWeights` typed config 注入 | 1 | `backend/internal/platform/config/loader_test.go::TestAsyncSection` + `backend/internal/runner/config_test.go` | unit + integration |
| X-3 | spec D-10 `email_dispatch` payload validator 与 B3 `shared/jobs/jobs.go` 一致 | 3 | `backend/internal/shared/jobs/jobs_test.go::TestEmailDispatchPayloadValidator` + `backend/internal/runner/email_dispatch_integration_test.go` | unit + integration |
| X-4 | spec D-14 新增 `async.scanIntervalSeconds` / `leaseTimeoutSeconds` / `shutdownGraceSeconds` / `reaperIntervalSeconds` typed config 注入（A4 additive config-only，不新增 env key） | 1 | `backend/internal/platform/config/loader_test.go::TestAsyncSection` + `backend/internal/runner/config_test.go::TestRuntimeConfig_AsyncTimings` | unit + integration |
| X-5 | B3 D-18 / `shared/jobs.yaml` 8 canonical job_type 当前事实 | 2 / 4 | `python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml` + `make lint-runner-out-of-scope` + `backend/internal/shared/jobs/jobs_test.go` | generated contract + lint |
| X-6 | spec C-9 / C-21 `report_generate` retryable failure 通过 kernel shared `BackoffPolicy` finalize | 4 | `backend/internal/review/generate_handler_test.go::TestGenerateHandler_NormalizesFinalizedRetryableFailureThroughKernel` + `backend/cmd/api/reports_http_scenario_test.go::TestE2EP0054ReportAIFailureAndRetry` + `backend/internal/store/review/persist_report_integration_test.go::TestPersistReportFailureRetryAndPermanent` | unit + scenario + integration (PG) |

### 2.6 Privacy / security / observability

| 行 | 决策 | Phase | 入口 | 类型 |
|----|------|-------|------|------|
| O-1 | spec C-11 outbox 失败不写 raw provider response / prompt body / answer text | 3 | `backend/internal/runner/outbox_dead_letter_test.go::TestOutboxDispatcher_RedactsLastError`（同 F-5） | integration |
| O-2 | spec §4.4 指标 family 命名与 label | 3 | `backend/internal/runner/metrics_test.go::TestKernelMetrics_FamilyAndLabels` | unit |
| O-3 | spec D-11 trace_id 字段写入 slog（dispatcher 侧） | 3 | `backend/internal/runner/outbox_trace_test.go::TestOutboxDispatcher_PropagatesTraceParent`（同 P3-2） | unit |
| O-4 | spec D-10 / B3 D-12 email_dispatch 红线（不写 raw token / URL / 邮件正文） | 3 | `backend/internal/runner/email_dispatch_integration_test.go::TestEmailDispatchHandler_PayloadRedaction` | integration |
| O-5 | spec D-11 trace_id 字段写入 slog（runtime handler 侧） | 1 | `backend/internal/runner/runtime_trace_test.go::TestRuntime_HandlerInheritsTraceparent` + `TestRuntime_HandlerLogsTraceIdField` | unit |

### 2.7 Regression / out-of-scope-negative

| 行 | 决策 | Phase | 入口 | 类型 |
|----|------|-------|------|------|
| R-1 | spec D-12 zero-reference grep | 4 | `make lint-runner-out-of-scope`（`scripts/lint/runner_out_of_scope.py`，与既有 `backend_review_out_of_scope.py` 同风格） | lint |
| R-2 | spec C-18 owner BDD 场景 rerun | 4 | Script BDD: `E2E.P0.003` / `010` / `011` / `012` / `013` / `033` / `034` / `035` / `077` / `078` / `080`；Go HTTP BDD: `E2E.P0.041` / `052` / `053` / `054` / `055` | BDD regression |
| R-3 | `git ls-files backend/internal/review` 不含 `runner.go` / `reaper.go` / `lease.go` | 4 | `backend/internal/review/structure_test.go::TestNoOutOfScopeRunnerFiles` 或 lint | structure test |
| R-4 | `git ls-files backend/internal/auth` 不含 `BackgroundMailDispatcher` 引用 | 4 | `backend/internal/auth/mail_test.go::TestNoBackgroundDispatcher` | unit |

### 2.8 UI source parity

不适用。本计划是纯后端基建，无 UI 真理源；`ui-design/` 没有对应 prototype。N/A 理由记录在 spec §2.2 与本节。

### 2.9 高风险 N/A

- **跨日 SLA / saga 编排**：N/A，仍走 ADR-Q2 锁定的「秒级-分钟级 + 幂等」范围。
- **Asynq Redis 队列**：N/A，本 plan 沿用 PG `async_jobs`；future Asynq 替换由 ADR 触发条件成立后另起 plan。
- **`privacy_export` handler**：N/A，B3 / B4 保留 canonical job_type 以支持 future 导出链路，但当前 B2 requestPrivacyExport 为 501 fixture，无 producer / handler；本 plan 不注册 `privacy_export`。

## 3 测试命令清单

- `cd backend && go test ./internal/runner/...`
- `cd backend && go test ./internal/targetjob/... ./internal/privacy/runner/... ./internal/resume/jobs/... ./internal/review/... ./internal/auth/...`
- `cd backend && go test ./cmd/api/...`
- `cd backend && go test ./internal/runner ./internal/review ./cmd/api -run '^(TestRuntime_FinalizeUsesTimestampAfterHandlerReturns|TestRuntime_StartDoesNotLetCriticalJobStarveEmailDispatch|TestGenerateHandler_NormalizesFinalizedRetryableFailureThroughKernel|TestE2EP0052ReportGenerationHappyPath|TestE2EP0054ReportAIFailureAndRetry)$' -count=1`
- `cd backend && go test -tags=integration ./internal/store/review -run '^TestPersistReportFailure' -count=1 -v`
- `make lint-runner-out-of-scope`
- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-async-runner/plans/001-internal-job-outbox-runner/context.yaml --docs-root docs --target backend`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`（A5 标准 doc Header / INDEX 漂移 gate）+ 针对 P4.5–4.13 列表的精确 grep（如 owner spec 用 `grep -n "未来 .backend-async-runner" docs/spec/{backend-runtime-topology,backend-review,backend-targetjob,backend-resume,backend-auth,event-and-outbox-contract}/spec.md` 期望 0 命中；roadmap 用 `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner" docs/spec/engineering-roadmap/spec.md` 期望 0 命中；本仓库未维护独立 `scripts/check_docs/` 工具集）
- 各 owner BDD suite 命令以各 `bdd-checklist.md` 为准

## 4 收口标准

- 所有 §2 行有对应通过证据（unit / integration / scenario / lint / doc reconcile）。
- §2.7 R-1 lint 通过；R-2 BDD 场景全 PASS。
- spec §6 acceptance criteria C-1~C-21（含 C-13a）全部有对应测试入口。
