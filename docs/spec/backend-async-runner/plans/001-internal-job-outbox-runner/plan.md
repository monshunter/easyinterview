# Internal Job and Outbox Runner

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-22

**关联 Checklist**: [checklist](./checklist.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把当前散落的 `targetjob.Drainer` / `review.Runner+Reaper` / `auth.BackgroundMailDispatcher` / `jdmatchRuntime.Drainer` 与缺失的 outbox dispatcher 收敛为单一 `backend/internal/runner/` kernel；接管 9 个当前可执行 canonical job_type（含 `email_dispatch` 与 `jd_match_agent_scan`，不含当前 501-only 的 `privacy_export` 与 future-async reserved 的 `jd_match_search`）的 lease / retry / dead-letter / reaper / shutdown；落地 B3 D-7 / D-8 / D-9 outbox dispatcher 协议；保持「不建独立 worker 进程」语义。

## 2 背景

[backend-async-runner spec](../../spec.md) §1 / §3 已锁定决策与现状证据。当前实现层证据摘要：

- `cmd/api/main.go` 持有 5 个独立 lifecycle（auth dispatcher / targetJobRuntime / resumeRuntime / reportRuntime / jdmatchRuntime）；
- `review.Runner+Reaper` 是 backend-review spec D-13 / D-16 标注的「等 backend-async-runner 接管」临时形态；
- `targetjob.Drainer` 已抽象但无 reaper，多个域各自实例化；
- `backend/cmd/api/jdmatch_runtime.go` 新增 `jdmatchRuntime.Drainer`，当前真实注册 `jd_match_agent_scan`；B3 `jd_match_search` 只作 future-async reservation，P0 search 仍为 sync HTTP handler；
- `outbox_events` 表只写不读；
- retry backoff 在 review / resume / targetjob 三处不一致；
- `email_dispatch` 还在 `auth.BackgroundMailDispatcher` 进程内 channel，未走 `async_jobs`。

本计划不修改任何业务 handler 的对外行为，只迁移生命周期边界与运行形态。

## 3 质量门禁分类

- **Plan 类型**: `code-internal` + `contract` + `tooling` + `docs`。涉及 backend Go 代码 / SQL / 配置 / lint 脚本 / 多个 owner spec 边界条款同步。
- **TDD 策略**: Code plan requires TDD。通过 `/implement backend-async-runner/001-internal-job-outbox-runner backend` → `/tdd` 执行；每个代码 / 契约 checklist item 先写或调整 focused test（kernel unit / domain handler integration / outbox integration / cmd/api lifecycle test），再实现最小变更；详见 [test-plan](./test-plan.md) 跨 phase 测试映射。
- **BDD 策略**: 不创建本 plan 专属 BDD 文件。本计划是行为保持的 backend-internal 迁移，不新增用户可见 UI / API envelope / 业务入口；但它会触达既有用户可见链路，必须在 Phase 4 以 `BDD-Gate:` 重跑 owner spec 既有场景（auth email / privacy_delete / target_import / report_generate / debrief_generate / resume_parse / resume_tailor / jd_match_agent_scan），并把证据归档到 [test-checklist](./test-checklist.md)。
- **替代验证 gate**:
  - **Contract test**：kernel `Runtime{Register,RunOnce,ReapOnce,Start,Shutdown}` 接口语义 unit test；`BackoffPolicy.Next` table 测试。
  - **Integration test**：outbox dispatcher 5s scan + FOR UPDATE SKIP LOCKED + retry / dead-letter / metrics 在真 PG 上覆盖；`async.queueWeights` typed config 注入测试。
  - **Regression rerun**：`backend/internal/targetjob/pipeline_test.go` / `targetjob/e2e_scenario_test.go` / `privacy/runner/delete_handler_test.go` / `debrief/generate_handler_test.go` / `debrief/service_test.go` / `resume/jobs/*_test.go` / `backend/internal/jdmatch/jobs/agent_scan_test.go` / `cmd/api/resume_parse_drainer_scenario_test.go` / `cmd/api/resume_tailor_drainer_scenario_test.go` / `cmd/api/reports_http_scenario_test.go` / `cmd/api/jdmatch_live_scenario_test.go::TestJDMatchAgentScanDrainerScenario` 在 kernel 接管后必须 PASS；各 owner spec BDD 场景 P4 rerun。
  - **Legacy negative search**：`make lint-runner-legacy`（新增 lint script）扫描 spec D-12 列出的旧 entry point；本 plan 自身 zero-reference gate 在 P4 收口。
  - **Doc reconcile**：spec D-* 决策落实后必须同步修订 `backend-runtime-topology` § 模块边界、`backend-review` D-13 / D-16、`backend-debrief` D-5、`backend-targetjob` D-5、`backend-resume` § 模块边界、`backend-auth` D-* / `email_dispatch` 章节、`event-and-outbox-contract` § 模块边界、`secrets-and-config` 新增 `async.*` typed config 节点。doc reconcile gate 以 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + 针对各 owner spec D-* 边界条款的精确 grep 为准（本仓库未维护独立 `scripts/check_docs/` 工具集）。
  - **Drift gate**：`cd backend && go build ./...`、`cd backend && go vet ./...`、Go race test、`make codegen-check`（如新增 generated 资源）、`validate_context.py`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`git diff --check`。

### 3.1 Operation Matrix

本计划不新增 user-facing operation；以下 9 个当前可执行 job_type 的对外 API envelope、fixture、frontend consumer 与 scenario 入口均**不变**，仅迁移 handler 注册边界与生命周期，因此 matrix 重点标注 backend handler 路径与 persistence/scenario gate 的当前状态。B3 当前 11 个 canonical job_type 中，`privacy_export` 当前仅保留为 DB / OpenAPI 501 contract（`requestPrivacyExport` fixture），`jd_match_search` 为 future-async reserved（P0 `searchJobs` sync HTTP handler），二者无 runner producer / handler，不在本计划注册。

| canonical job_type | 触发 operationId（OpenAPI） | fixture | frontend consumer | 当前 backend handler | 本 plan 迁移目标 handler | persistence | AI dependency | scenario coverage |
|--------------------|---------------------------|---------|-------------------|----------------------|--------------------------|-------------|---------------|-------------------|
| `email_dispatch` | `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json` | passwordless login UI（auth flow） | `backend/internal/auth/mail.go::BackgroundMailDispatcher`（进程内 channel） | `backend/internal/auth/email_dispatch_handler.go` 注册到 kernel；producer 改为 `INSERT INTO async_jobs(job_type='email_dispatch')` | `async_jobs(email_dispatch)`（新接入）+ `DeliveryWriter` sink | stub/fixture（dev `DevMailSink`） | BDD `E2E.P0.003`；smoke `TestAuthEmailEndToEnd`（1 个 scan 周期内 magic link 可见） |
| `privacy_delete` | `deleteMe` (`DELETE /v1/me`) | `openapi/fixtures/Auth/deleteMe.json` | account / privacy 设置面 | `backend/internal/privacy/runner/delete_handler.go` 注册到 `targetJobRuntime.Drainer` | 同 handler 注册到 kernel `runner.Runtime` | `privacy_requests` + `async_jobs(privacy_delete)` + 用户级资源级联删除 | none | BDD `E2E.P0.003` + `E2E.P0.033`；smoke `DELETE /api/v1/me` → `privacy_requests.status='completed'` |
| `report_generate` | source event：`completePracticeSession`（同事务写 `feedback_reports` placeholder + `async_jobs(report_generate)`）；查询：`getFeedbackReport` / `listTargetJobReports` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` / `openapi/fixtures/Reports/getFeedbackReport.json` / `openapi/fixtures/Reports/listTargetJobReports.json` | 报告面 | `backend/internal/review/runner.go` + `reaper.go` + `lease.go`（独立 polling worker） | 新建 `backend/internal/review/generate_handler.go` 实现 `runner.Handler`；删除 `review.Runner` / `Reaper` / `ComputeReportFailureBackoff` | `feedback_reports` + `async_jobs(report_generate)` + `outbox_events('report.generated')` | A3/F3 评审 AI profile | BDD `E2E.P0.041` + `E2E.P0.052` / `054` / `055`；regression `review/runner_test.go`（重写到 kernel）+ `cmd/api/reports_http_scenario_test.go` |
| `target_import` | `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` | 目标导入面 | `backend/internal/targetjob/drainer.go` 实例化在 `cmd/api/main.go::buildTargetJobRuntime` | 同 handler 通过 `runner.FromTargetjobHandler` 注册到 kernel；删除独立 drainer 实例 | `target_jobs` + `async_jobs(target_import)` + `outbox_events('target.import.requested')` | A3 解析 profile | BDD `E2E.P0.010` / `012` / `013`；regression `targetjob/pipeline_test.go` + `e2e_scenario_test.go` |
| `source_refresh` | N/A（internal-only，由调度或目标更新触发，无独立 API surface） | N/A | N/A（无 frontend） | `backend/internal/targetjob/drainer.go` 注册的 `SourceRefreshHandler` | 同 handler 通过 adapter 注册到 kernel | `source_records` + `async_jobs(source_refresh)` | A3 解析 profile（同 target_import） | regression `targetjob/pipeline_test.go::TestDrainer_RunOnceProcessesQueuedJob` |
| `debrief_generate` | `createDebrief` | `openapi/fixtures/Debriefs/createDebrief.json` | 复盘面 | `domaindebrief.NewGenerateHandler(...)` 注册到 `targetJobRuntime.Drainer` | 同 handler 注册到 kernel | `debriefs` + `async_jobs(debrief_generate)` | A3 debrief profile | BDD `E2E.P0.060` / `062`；regression `debrief/generate_handler_test.go` + `service_test.go` |
| `resume_parse` | `registerResume` / `confirmResumeStructuredMaster` | `openapi/fixtures/Resumes/registerResume.json` / `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` | 简历上传面 | `backend/internal/resume/jobs/parse.go` + `resume/store/async.go`（固定 15s retry） | `resumejobs.NewParseHandler(...)` 注册到 kernel；删除 `resumeRuntime.Drainer` 与 store-side 15s retry 副本 | `resume_assets` + `async_jobs(resume_parse)` | A3 简历解析 profile | BDD `E2E.P0.034` / `035`；regression `resume/jobs/parse_test.go` + `cmd/api/resume_parse_drainer_scenario_test.go` |
| `resume_tailor` | `requestResumeTailor`（查询 `getResumeTailorRun` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion`） | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` / `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` / `openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json` / `openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json` | 简历适配面 | `backend/internal/resume/jobs/tailor.go` | `resumejobs.NewTailorHandler(...)` 注册到 kernel；退避走 kernel `BackoffPolicy` | `resume_tailor_runs` + `async_jobs(resume_tailor)` | A3 简历适配 profile | BDD `E2E.P0.077` / `078` / `080`；regression `resume/jobs/tailor_test.go` + `cmd/api/resume_tailor_drainer_scenario_test.go` |
| `jd_match_agent_scan` | source trigger：internal schedule / lazy trigger / source data change；状态查询：`getAgentScanStatus` | `openapi/fixtures/JobMatch/getAgentScanStatus.json` | JD-Match tab polling agent status | `backend/cmd/api/jdmatch_runtime.go` 用 `targetjob.NewDrainer` 注册 `JobTypeJdMatchAgentScan`，调用 `backend/internal/jdmatch/jobs.Run` | `backend/internal/jdmatch/jobs.Run` 通过 `runner.Handler` 注册到 kernel；删除 `jdmatchRuntime.Drainer` 独立实例；确认 `jd_match_search` 不注册 | `agent_scans` + `jd_match_recommendations` + `async_jobs(jd_match_agent_scan)` + `outbox_events('jd_match.recommendation.completed')` | A3/F3 `jd_match.recommendation` profile | BDD `E2E.P0.094` / `097`；regression `backend/internal/jdmatch/jobs/agent_scan_test.go` + `backend/cmd/api/jdmatch_live_scenario_test.go::TestJDMatchAgentScanDrainerScenario` |

矩阵字段 `frontend consumer` 仅列出受影响入口的语义所属，无前端实现修改；本计划属于 backend-internal 重构，所有 user-facing operation 的 OpenAPI envelope / fixture / scenario 行为不变，由 P4 `BDD-Gate:` owner scenario rerun + P2 handler test rerun 提供 evidence handoff。`jd_match_search` 虽已在 `shared/jobs.yaml` 占位并声明 `priority: medium`，但 A4 当前只锁定 `critical/default/low` 三档权重且 P0 search 为 sync HTTP handler；实施期不得注册该 job_type，除非先由 B3/A4 原地修订 medium priority 与 async search owner gate。

## 4 实施步骤

### Phase 1: Kernel 基础设施

#### 1.1 新建 `backend/internal/runner/` package 骨架

落地 `runtime.go`（Runtime 与 Registry）、`lease.go`（lease SQL contract）、`reaper.go`（lease timeout 回收）、`backoff.go`（统一退避表）、`handler.go`（`Handler` / `JobHandlerFunc` / `ClaimedJob` / `JobOutcome` 类型）、`adapter_targetjob.go`（`FromTargetjobHandler` shim）、`doc.go`。所有公共结构必须配套 godoc。

#### 1.2 实现 Lease / Finalize SQL contract

按 [spec §4.2](../../spec.md#42-sql-约束) 实现 `LeaseAsyncJob` / `FinalizeAsyncJob` / `ReclaimExpiredLeases`；查询条件、排序、`FOR UPDATE SKIP LOCKED` 必须与 spec 完全一致；列名固定为 `locked_at` / `attempts` / `available_at` / `status`；retryable exhausted 统一 finalize 为 `dead`，non-retryable failure 统一 finalize 为 `failed`。

#### 1.3 实现 BackoffPolicy

`runner.BackoffPolicy.Next(attempts int32) time.Duration` 返回 `[30s, 2m, 10m, 1h, 6h]`；attempts ≥ len → 返回最后值；attempts < 1 → 返回首值。同时暴露 `MaxAttempts = 5` 常量。

#### 1.4 实现 Reaper

`runner.Reaper.RunOnce(ctx)` 调用 `ReclaimExpiredLeases(jobTypes, leaseTimeout, now)`；attempts **不**递增；刷新 F1 字典内的 `async_job_queue_depth{job_type}` / `async_job_lag_seconds{job_type}`，并产出 `async_jobs_processed_total{result="reaped"}`。

#### 1.5 实现 Graceful Shutdown 协调

`Runtime.Shutdown(ctx)` 顺序：(a) cancel internal lease loop，(b) 等待 in-flight handler（subject to `async.shutdownGraceSeconds` 默认 10s），(c) stop reaper，(d) stop outbox dispatcher（Phase 3 上线后）。超时返回 `ctx.Err()` 并保留 in-flight handler 给下次启动 reap。

#### 1.6 typed config 注入

读取 [A4](../../../secrets-and-config/spec.md) `async.queueWeights`；为新增的 `async.shutdownGraceSeconds` / `async.leaseTimeoutSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds`（[spec D-14](../../spec.md#31-已锁定决策)）在 A4 spec 中 additive 修订并落实 config-only typed config 节点（不新增 env key）；缺失或非正数 fail-fast；不得静默回退为代码常量。

#### 1.7 单元测试基础设施

落地 fake `LeaseStore`（in-memory）+ table-driven test 覆盖 register / lease / finalize / retry / reaper / shutdown 五条路径；用 deterministic clock。

#### 1.8 Trace propagation（runtime handler 侧）

按 [spec D-11](../../spec.md#31-已锁定决策) / [§4.4](../../spec.md#44-红线与可观测性约束)，`Runtime` 在调用 handler 前从 `async_jobs.payload` / envelope 中读取 W3C `traceparent`（如存在）重建 span context；handler 内 slog 输出注入 `trace_id` 字段（缺失时跳过）。覆盖范围与 Phase 3 outbox dispatcher trace 路径同源，但落点在 runtime 一侧，避免 Phase 2 业务 handler 迁移后丢失 trace 字段。

### Phase 2: 业务 handler 迁移

#### 2.1 `target_import` + `source_refresh` 迁移

把 `backend/internal/targetjob/drainer.go` 的 Drainer 实例化点从 `cmd/api/main.go.buildTargetJobRuntime` 迁到统一 `runner.Runtime`；保留 `targetjob.SourceRefreshHandler` / `ParseExecutor` 业务实现；通过 `runner.FromTargetjobHandler` 注册。

#### 2.2 `privacy_delete` 迁移

把 `privacyrunner.NewPrivacyDeleteHandler(...)` 注册到 kernel；删除 `cmd/api` 中 privacy 通过 targetJobRuntime.Drainer 注册的旧路径（adapter 期间允许并存）。

#### 2.3 `debrief_generate` 迁移

`domaindebrief.NewGenerateHandler(...)` 注册到 kernel；删除 `cmd/api` 中 debrief 通过 targetJobRuntime.Drainer 的注册路径（adapter 期间允许并存）。

#### 2.4 `resume_parse` + `resume_tailor` 迁移

`resumejobs.NewParseHandler(...)` + `resumejobs.NewTailorHandler(...)` 注册到 kernel；删除 `cmd/api` 中独立 `resumeRuntime.Drainer`；resume_store `ClaimNextAsyncJob` / `FinalizeAsyncJob` SQL 副本由 kernel-side store 统一持有；resume async.go 中固定 15s retry 删除。

#### 2.5 `report_generate` 迁移；删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff`

新建 `review.GenerateHandler` 实现 `runner.Handler`（迁移 `review/runner.go` 中 `LeaseAsyncJob → service.GenerateReport → UpdateAsyncJobSucceeded/Failed` 逻辑）；删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff` / `review.DefaultReportFailureBackoff`；保留 `review.Service` / `review.LeaseAsyncJob`（如有业务态变更需求，迁到 kernel-side store）。

#### 2.6 `jd_match_agent_scan` 迁移

把 `backend/cmd/api/jdmatch_runtime.go` 中 `jdmatchRuntime.Drainer` 注册的 `JobTypeJdMatchAgentScan` 迁到统一 kernel；保留 `backend/internal/jdmatch/jobs.Run` / recommendation generator / `writeJDMatchRecommendationCompletedOutbox` 业务实现；迁移后 `getAgentScanStatus` / `E2E.P0.097` 语义不变；显式断言 `jd_match_search` 未注册 runner。

#### 2.7 退避收口

确认 P2 完成后所有业务 handler 退避走 kernel `BackoffPolicy`；删除各域 hard-coded 退避（`review.ComputeReportFailureBackoff`、resume `async.go` `now.Add(15 * time.Second)` 等）；focused test 覆盖 attempts 1..5 全部走表。

### Phase 3: Outbox dispatcher + email_dispatch

#### 3.1 实现 `runner.OutboxDispatcher`

新建 `backend/internal/runner/outbox/` (或并入 runner package)：`Dispatcher.RunOnce(ctx)` 按 [spec §4.2](../../spec.md#42-sql-约束) 拉取 pending 行；5s scan + `FOR UPDATE SKIP LOCKED` + batch ≤ 100 + sort by `next_attempt_at asc, created_at asc`；已注册 consumer ack 成功后才置 `published`；临时失败或缺少 runtime consumer 时按 `BackoffPolicy` 后移 `next_attempt_at` 并记录 redacted `last_error_message`，attempts≥5 → `failed`。

#### 3.2 Consumer 注册接口

暴露 `Dispatcher.RegisterConsumer(eventName string, consumer OutboxConsumer)`；dry-run consumer 只能由测试显式注入用于 P3 framework 验证，runtime 默认不得用 dry-run ack 替代真实 consumer。缺少 consumer 的 `event_name` 必须保持 `pending` / retry / failed 路径，不得置为 `published`。

#### 3.3 `source_event_only` 跳过

读取 B3 generated `jobs.IsSourceEventOnly(jobType)` 谓词；对应 outbox 事件 publish 后**不**触发 dispatcher 二次创建 `async_jobs`；focused integration test 覆盖 `practice.session.completed`。

#### 3.4 TraceId 透传

dispatcher 在调用 consumer 前从 outbox payload / envelope 中读取 `traceId`（W3C `traceparent`）重建 span；logger 字段 `trace_id`；缺失时写 warn log 后继续 publish（[B3 D-10](../../../event-and-outbox-contract/spec.md#31-已锁定决策含-jobtype-映射表)）。

#### 3.5 指标暴露

`outbox_events_pending` / `outbox_publish_duration_seconds` / `outbox_publish_failures_total` 注册到既有 metric registry；不引入 Prometheus 实例依赖（[backend-runtime-topology D-4](../../../backend-runtime-topology/spec.md#3-用户决策--待确认事项)）。

#### 3.6 `email_dispatch` 收口

修订 `auth.PasswordlessService` enqueue 路径，把 `MailDispatcher.Enqueue(ctx, payload)` 改为 `INSERT INTO async_jobs(job_type='email_dispatch', payload, ...)` 同事务 enqueue；kernel 注册 `EmailDispatchHandler{writer DeliveryWriter}` 通过既有 `DevMailSink` / 生产 sink 发邮件；保留 `EmailDispatchPayload` validator；删除 `auth.BackgroundMailDispatcher` / `NewBackgroundMailDispatcher` 与 `cmd/api` 对应 defer。

### Phase 4: 收口 / 文档同步 / lint negative gate

#### 4.1 `cmd/api/main.go` 单点 shutdown

`main()` 持有单一 `runtime *runner.Runtime`；所有 handler 在 `buildTargetJobRuntime` / `buildResumeRuntime` / `buildReportRuntime` / `buildDebriefRoutes` / `buildJDMatchRuntime` / `buildAuth*` 中 build 完成后注册到 runtime；signal context 触发 → `runtime.Shutdown(ctx)` 单调 drain。

#### 4.2 删除旧形态

删除 `review.Runner` / `review.Reaper` legacy 实现文件；把 `review/runner_test.go` / `review/reaper_test.go` 中仍有价值的租约、退避、回收断言迁移为 kernel / `GenerateHandler` 测试或结构性负向测试，避免删除测试覆盖；删除 `auth.BackgroundMailDispatcher` 与对应 legacy 测试；删除 `cmd/api` 中独立 `targetJobRuntime.Drainer` / `resumeRuntime.Drainer` / `reportRuntime.Runner` / `reportRuntime.Reaper` / `jdmatchRuntime.Drainer` / `mailDispatcher` 字段；保留 history 证据。

#### 4.3 Legacy negative lint

新增 `scripts/lint/runner_legacy.py`（+ `runner_legacy_test.py`，与既有 `backend_review_legacy.py` / `backend_practice_legacy.py` 同风格的 Python lint），扫描 [spec D-12](../../spec.md#31-已锁定决策) 列出的旧 entry point；接入 `make lint-runner-legacy`；history / tests / lint 自身允许保留历史证据。

#### 4.4 同步 owner spec 边界条款

- `backend-runtime-topology` §5 模块边界 / §6 验收：把「backend internal runner 实现」从 future shorthand 改为 active owner `backend-async-runner`。
- `backend-review` D-13 / D-16：标注 `review.Runner` / `review.Reaper` 已由 `backend-async-runner/001` 接管；保留 history 证据。
- `backend-debrief` D-5 / §模块边界：handler 注册改为 kernel；删除「未来 backend-async-runner 接管」未来时态。
- `backend-targetjob` D-5 / §模块边界：Drainer 抽象迁到 kernel；保留 handler 业务实现。
- `backend-resume` §模块边界：handler 注册改为 kernel。
- `backend-jobs-recommendations` D-12 / D-13 / §模块边界：`jd_match_agent_scan` handler 注册改为 kernel；`jd_match_search` 继续标注为 P0 sync handler / future-async reserved，不由本 plan 注册。
- `backend-auth` D-* / `email_dispatch`：`BackgroundMailDispatcher` 已删除；producer 通过 `async_jobs(email_dispatch)` enqueue。
- `event-and-outbox-contract` §5 模块边界：「backend internal runner 实现」改为 `backend-async-runner` owner。
- `secrets-and-config` D-9 / §config dictionary：additive 新增 `async.leaseTimeoutSeconds` / `async.shutdownGraceSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds` typed config 节点。
- `engineering-roadmap` §5.2 / §6.3 S2：确认 `backend-async-runner` 已作为 active subject 引用，负向 gate 拦截 `backend-async-runner.*未创建` / `未创建.*backend-async-runner` 回流。

#### 4.5 Owner BDD 场景回归

依次 rerun 各 owner spec BDD suite（target_import / report_generate / privacy_delete / debrief_generate / resume_parse / resume_tailor / auth email / jd_match_agent_scan），证据归档到 [test-checklist](./test-checklist.md)；JD Match 至少覆盖 `E2E.P0.094`-`E2E.P0.097`；任一场景失败 → 回到对应 phase 修复。

#### 4.6 Spec / plan 状态收尾

完成全部 acceptance criteria 后，把本 plan 状态从 `active` 改为 `completed`；spec.md / history.md 已在创建时即为 `active`，本步骤不再涉及 `draft → active` 过渡；同步 spec INDEX + plans INDEX。

## 5 验收标准

- 本计划列出的实现 / 测试项全部通过（覆盖 [spec C-1~C-20](../../spec.md#6-验收标准)，含 C-13a missing-consumer safety）。
- 替代验证 gate 全部 PASS：contract / integration / regression rerun / legacy negative lint / doc reconcile / drift gate。
- 不存在新增的用户可见行为缺口；既有 owner spec BDD 场景 rerun 通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| review.Runner 删除后报告生成回归 | Phase 2.5 完成时 rerun `review/runner_test.go` + `cmd/api/reports_http_scenario_test.go`；任何失败必须先修复再进入下一 phase |
| email_dispatch 切到 async_jobs 后 magic link 投递延迟 | 把 `email_dispatch` 列入 `low` priority bucket 但 lease loop 仍≤5s scan；P3 收口阶段 smoke 验证 auth email start → DevMailSink delivery 延迟 ≤ 1 个 scan 周期 |
| Outbox dispatcher 上线后 consumer 缺失导致 outbox 行被误确认或长期积压 | runtime 缺少 consumer 时不得置 `published`；dry-run consumer 仅允许测试显式注入；缺少 consumer 的 event 走 retry/dead-letter 并暴露 `outbox_publish_failures_total`，P3 完成前与 F2 / 各 owner 明确启用边界 |
| 合并后的 JD Match runner 被遗漏，导致 `jd_match_agent_scan` 继续走旧 drainer | P2.6 单独迁移 `jd_match_agent_scan`；P4 lint 必须覆盖 `jdmatchRuntime.Drainer` / `JobTypeJdMatchAgentScan`；P4 BDD rerun 覆盖 E2E.P0.094-P0.097 |
| `jd_match_search` future reservation 被误注册，触发 A4 未支持的 `medium` priority | Operation Matrix 与 P2.6 显式禁止注册 `jd_match_search`；若未来启用 async search，先由 B3/A4 修订 medium priority / queue weight gate |
| 多 owner spec D-* 边界条款同步遗漏 | Phase 4.4 用 checklist 逐项打勾；P4 收尾必须运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` |
| 退避收口期间 in-flight job 行为变化导致已部署环境异常 | 本仓库无线上环境，P0 不需要兼容 layer；本 plan 仅需保证 dev / test scenario 通过 |
| typed config 节点新增导致 A4 owner spec 需要 additive 修订 | Phase 1.6 把 A4 修订作为前置 checklist item；若实施期发现 A4 owner spec 已有冲突决策，停止进入 plan-review / design 修订，不以 kernel default 常量绕过 |
