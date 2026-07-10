# Internal Job and Outbox Runner

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

保持单一 `backend/internal/runner/` kernel：统一承接 7 个可执行 handler（`email_dispatch` / `privacy_delete` / `report_generate` / `target_import` / `source_refresh` / `resume_parse` / `resume_tailor`）的 lease / retry / dead-letter / reaper / shutdown 与 B3 outbox dispatcher 协议。所有业务 handler 直接实现 `runner.Handler`，业务域不持有重复 async job contract/SQL 或测试专用 runtime；`privacy_export` 仅保留 DB / OpenAPI 501 contract，无 runner handler。

## 2 背景

[backend-async-runner spec](../../spec.md) §1 / §3 是当前 runtime 真理源。`cmd/api` 只持有一个 `runner.Runtime`，7 个可执行 handler 直接注册；kernel 单点拥有 lease/finalize/reaper SQL 与统一 backoff；`runner.OutboxDispatcher` 消费业务事务写入的 outbox；场景通过 `Runtime.RunOnce` 同步驱动真实 kernel。计划维护以 [spec v1.11](../../spec.md)、[checklist](./checklist.md) 和 [test-checklist](./test-checklist.md) 的当前 gate 为准。

## 2.1 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-10 | 1.11 | Remove the test-only targetjob drainer, duplicate async job contracts and SQL, and the targetjob-to-kernel adapter; run all retained handlers and scenarios directly through runner.Runtime. |
| 2026-07-10 | 1.10 | Remove unused logger dependencies from report and resume runtime builders and all composition call sites. |
| 2026-07-10 | 1.9 | 技术债口径清理：把 `report_generate` 的初始 `feedback_reports` 行描述为 pending row，不再使用 placeholder 口径。 |
| 2026-07-10 | 1.8 | 技术债口径清理：把 review runner 描述改为实施前基线与当前 kernel owner 事实，不再使用旧交接口径。 |
| 2026-07-07 | 1.7 | Wording cleanup：收敛 runner out-of-scope gate 与 auth email smoke 为当前 out-of-scope / email-code 口径，不改变 runner 可执行契约。 |
| 2026-07-06 | 1.6 | D-22 后裁剪复查：当前 runner 正向范围收敛为 7 个可执行 handler + `privacy_export` contract-only；`debrief_generate` / `jd_match_agent_scan` / `jd_match_search` 不再作为当前 plan/context/test 的正向 target surface。 |

## 3 质量门禁分类

- **Plan 类型**: `code-internal` + `contract` + `tooling` + `docs`。涉及 backend Go 代码 / SQL / 配置 / lint 脚本 / 多个 owner spec 边界条款同步。
- **TDD 策略**: Code plan requires TDD。通过 `/implement backend-async-runner/001-internal-job-outbox-runner backend` → `/tdd` 执行；每个代码 / 契约 checklist item 先写或调整 focused test（kernel unit / domain handler integration / outbox integration / cmd/api lifecycle test），再实现最小变更；详见 [test-plan](./test-plan.md) 跨 phase 测试映射。
- **BDD 策略**: 不创建本 plan 专属 BDD 文件。本计划是行为保持的 backend-internal 迁移，不新增用户可见 UI / API envelope / 业务入口；但它会触达既有用户可见链路，必须在 Phase 4 以 `BDD-Gate:` 重跑当前仍保留的 owner spec 场景（auth email / privacy_delete / target_import / report_generate / resume_parse / resume_tailor），并把证据归档到 [test-checklist](./test-checklist.md)。
- **替代验证 gate**:
  - **Contract test**：kernel `Runtime{Register,RunOnce,ReapOnce,Start,Shutdown}` 接口语义 unit test；`BackoffPolicy.Next` table 测试。
  - **Integration test**：outbox dispatcher 5s scan + FOR UPDATE SKIP LOCKED + retry / dead-letter / metrics 在真 PG 上覆盖；`async.queueWeights` typed config 注入测试。
  - **Regression rerun**：`backend/internal/targetjob/pipeline_test.go` / `targetjob/e2e_scenario_test.go` / `privacy/runner/delete_handler_test.go` / `resume/jobs/*_test.go` / `cmd/api/resume_parse_runner_scenario_test.go` / `cmd/api/resume_tailor_runner_scenario_test.go` / `cmd/api/reports_http_scenario_test.go` 必须 PASS；各 owner spec BDD 场景在 P4 rerun。
  - **Out-of-scope negative search**：`make lint-runner-out-of-scope`（新增 lint script）扫描 spec D-12 列出的 out-of-scope entry point；本 plan 自身 zero-reference gate 在 P4 收口。
  - **Doc reconcile**：spec D-* 决策落实后必须同步修订 `backend-runtime-topology` § 模块边界、`backend-review` D-13 / D-16、`backend-targetjob` D-5、`backend-resume` § 模块边界、`backend-auth` D-* / `email_dispatch` 章节、`event-and-outbox-contract` § 模块边界、`secrets-and-config` 新增 `async.*` typed config 节点。doc reconcile gate 以 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + 针对各 owner spec D-* 边界条款的精确 grep 为准（本仓库未维护独立 `scripts/check_docs/` 工具集）。
  - **Drift gate**：`cd backend && go build ./...`、`cd backend && go vet ./...`、Go race test、`make codegen-check`（如新增 generated 资源）、`validate_context.py`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`git diff --check`。

### 3.1 Operation Matrix

本计划不新增 user-facing operation；以下当前仍可执行 job_type 的 API envelope、fixture、frontend consumer 与 scenario 入口均不变，matrix 标注当前 handler、kernel registration、persistence 与 scenario gate。B3 当前 8 个 canonical job_type 中，`privacy_export` 仅保留为 DB / OpenAPI 501 contract（`requestPrivacyExport` fixture），无 runner producer / handler。

| canonical job_type | 触发 operationId（OpenAPI） | fixture | frontend consumer | backend handler | kernel registration | persistence | AI dependency | scenario coverage |
|--------------------|---------------------------|---------|-------------------|----------------------|--------------------------|-------------|---------------|-------------------|
| `email_dispatch` | `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json` | email-code login UI（auth flow） | `backend/internal/auth/email_dispatch_handler.go` | `backend/internal/auth/email_dispatch_handler.go` 注册到 kernel；producer 改为 `INSERT INTO async_jobs(job_type='email_dispatch')` | `async_jobs(email_dispatch)` + `DeliveryWriter` sink | stub/fixture（dev `DevMailSink`） | BDD `E2E.P0.003`；smoke `TestAuthEmailEndToEnd`（1 个 scan 周期内 email code 可见） |
| `privacy_delete` | `deleteMe` (`DELETE /v1/me`) | `openapi/fixtures/Auth/deleteMe.json` | account / privacy 设置面 | `backend/internal/privacy/runner/delete_handler.go` | `PrivacyDeleteHandler` 直接注册到 `runner.Runtime` | `privacy_requests` + `async_jobs(privacy_delete)` + 用户级资源级联删除 | none | BDD `E2E.P0.003` + `E2E.P0.033`；smoke `DELETE /api/v1/me` → `privacy_requests.status='completed'` |
| `report_generate` | source event：`completePracticeSession`（同事务写 pending `feedback_reports` row + `async_jobs(report_generate)`）；查询：`getFeedbackReport` / `listTargetJobReports` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` / `openapi/fixtures/Reports/getFeedbackReport.json` / `openapi/fixtures/Reports/listTargetJobReports.json` | 报告面 | `backend/internal/review/runner.go` + `reaper.go` + `lease.go`（report-only polling worker） | 新建 `backend/internal/review/generate_handler.go` 实现 `runner.Handler`；删除 `review.Runner` / `Reaper` / `ComputeReportFailureBackoff` | `feedback_reports` + `async_jobs(report_generate)` + `outbox_events('report.generated')` | A3/F3 评审 AI profile | BDD `E2E.P0.041` + `E2E.P0.052` / `053` / `054` / `055`（Go HTTP BDD tests）；regression `review/runner_test.go`（重写到 kernel）+ `cmd/api/reports_http_scenario_test.go` |
| `target_import` | `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` | 目标导入面 | `targetjob.ParseExecutor` | 直接注册到 `runner.Runtime` | `target_jobs` + `async_jobs(target_import)` + `outbox_events('target.import.requested')` | A3 解析 profile | BDD `E2E.P0.010` / `011` / `012` / `013`；regression `targetjob/pipeline_test.go` + `e2e_scenario_test.go` |
| `source_refresh` | N/A（internal-only，由调度或目标更新触发，无独立 API surface） | N/A | N/A（无 frontend） | `targetjob.SourceRefreshHandler` | 直接注册到 `runner.Runtime` | `source_records` + `async_jobs(source_refresh)` | A3 解析 profile（同 target_import） | regression `targetjob/pipeline_test.go::TestSourceRefreshHandler_MarksStale` |
| `resume_parse` | `registerResume` / `confirmResumeStructuredMaster` | `openapi/fixtures/Resumes/registerResume.json` / `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` | 简历上传面 | `resumejobs.ParseHandler` | 直接注册到 `runner.Runtime` | `resume_assets` + `async_jobs(resume_parse)` | A3 简历解析 profile | BDD `E2E.P0.034` / `035`；regression `resume/jobs/parse_test.go` + `cmd/api/resume_parse_runner_scenario_test.go` |
| `resume_tailor` | `requestResumeTailor`（查询 `getResumeTailorRun` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion`） | `openapi/fixtures/ResumeTailor/requestResumeTailor.json` / `openapi/fixtures/ResumeTailor/getResumeTailorRun.json` / `openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json` / `openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json` | 简历适配面 | `resumejobs.TailorHandler` | 直接注册到 `runner.Runtime`；退避走 kernel `BackoffPolicy` | `resume_tailor_runs` + `async_jobs(resume_tailor)` | A3 简历适配 profile | BDD `E2E.P0.077` / `078` / `080`；regression `resume/jobs/tailor_test.go` + `cmd/api/resume_tailor_runner_scenario_test.go` |

矩阵字段 `frontend consumer` 仅列出受影响入口的语义所属，无前端实现修改；本计划属于 backend-internal 重构，所有仍保留 user-facing operation 的 OpenAPI envelope / fixture / scenario 行为不变，由 P4 `BDD-Gate:` owner scenario rerun + P2 handler test rerun 提供 evidence handoff。D-22 裁剪后，范围外模块的 job/event 不再作为本计划当前正向矩阵行。

## 4 实施步骤

### Phase 1: Kernel 基础设施

#### 1.1 新建 `backend/internal/runner/` package 骨架

`runtime.go`（Runtime 与 Registry）、`lease.go`（lease SQL contract）、`reaper.go`（lease timeout 回收）、`backoff.go`（统一退避表）、`handler.go`（`Handler` / `JobHandlerFunc` / `ClaimedJob` / `JobOutcome` 类型）和 `doc.go` 构成 kernel；所有公共结构配套 godoc。

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

`targetjob.SourceRefreshHandler` / `ParseExecutor` 直接实现 `runner.Handler`，由 `cmd/api/main.go.buildTargetJobRuntime` 返回给统一 `runner.Runtime` 注册。

#### 2.2 `privacy_delete` 迁移

`privacyrunner.NewPrivacyDeleteHandler(...)` 直接实现 `runner.Handler` 并注册到 kernel。

#### 2.3 Out-of-scope module reconciliation

D-22 后复盘模块范围外，当前不再注册对应 runner handler；本 plan 的当前收口只要求范围外模块不能作为正向 target surface 回流。

#### 2.4 `resume_parse` + `resume_tailor` 迁移

`resumejobs.NewParseHandler(...)` + `resumejobs.NewTailorHandler(...)` 直接实现 `runner.Handler` 并注册到 kernel；lease/finalize SQL 与 retry backoff 只由 kernel 持有。

#### 2.5 `report_generate` 迁移；删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff`

新建 `review.GenerateHandler` 实现 `runner.Handler`（迁移 `review/runner.go` 中 `LeaseAsyncJob → service.GenerateReport → UpdateAsyncJobSucceeded/Failed` 逻辑）；删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff` / `review.DefaultReportFailureBackoff`；保留 `review.Service` / `review.LeaseAsyncJob`（如有业务态变更需求，迁到 kernel-side store）。

#### 2.6 Out-of-scope jobs-recommendations reconciliation

D-22 后 Jobs Recommendations / JD Match 模块范围外；当前 runner 不注册对应 scan/search handler。`make lint-runner-out-of-scope` 拦截局部 runtime 与重复 contract 回流。

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

修订 `auth.EmailCodeService` enqueue 路径，把 auth challenge 发信请求改为 `INSERT INTO async_jobs(job_type='email_dispatch', payload, ...)` 同事务 enqueue；kernel 注册 `EmailDispatchHandler{writer DeliveryWriter}` 通过既有 `DevMailSink` / 生产 sink 发邮件；保留 `EmailDispatchPayload` validator；`cmd/api` 的发件生命周期由 runner kernel 统一管理。

### Phase 4: 收口 / 文档同步 / lint negative gate

#### 4.1 `cmd/api/main.go` 单点 shutdown

`main()` 持有单一 `runtime *runner.Runtime`；所有当前保留 handler 在 `buildTargetJobRuntime` / `buildResumeRuntime` / `buildReportRuntime` / `buildAuth*` 中 build 完成后注册到 runtime；signal context 触发 → `runtime.Shutdown(ctx)` 单调 drain。

#### 4.2 删除 out-of-scope 形态

kernel package 承接租约、退避、回收、shutdown 与 outbox 测试；业务域只保留 handler 状态机与持久化测试。`cmd/api` runtime struct 只暴露路由、idempotency、AI runtime 和 `map[string]runner.Handler`，不持有局部执行循环。

#### 4.3 Out-of-scope negative lint

新增 `scripts/lint/runner_out_of_scope.py`（+ `runner_out_of_scope_test.py`，与既有 `backend_review_out_of_scope.py` / `backend_practice_out_of_scope.py` 同风格的 Python lint），扫描 [spec D-12](../../spec.md#31-已锁定决策) 列出的 out-of-scope entry point；接入 `make lint-runner-out-of-scope`；audit / tests / lint 自身允许保留证据。

#### 4.4 同步 owner spec 边界条款

- `backend-runtime-topology` §5 模块边界 / §6 验收：把「backend internal runner 实现」从 future shorthand 改为 active owner `backend-async-runner`。
- `backend-review` D-13 / D-16：标注 `review.Runner` / `review.Reaper` 已由 `backend-async-runner/001` 接管；保留 history 证据。
- `backend-targetjob` D-5 / §模块边界：kernel 持有 runtime，targetjob 保留原生 handler 业务实现。
- `backend-resume` §模块边界：handler 注册改为 kernel。
- `backend-auth` D-* / `email_dispatch`：`BackgroundMailDispatcher` 范围外；producer 通过 `async_jobs(email_dispatch)` enqueue。
- `event-and-outbox-contract` §5 模块边界：「backend internal runner 实现」改为 `backend-async-runner` owner。
- `secrets-and-config` D-9 / §config dictionary：additive 新增 `async.leaseTimeoutSeconds` / `async.shutdownGraceSeconds` / `async.reaperIntervalSeconds` / `async.scanIntervalSeconds` typed config 节点。
- `engineering-roadmap` §5.2 / §6.3 S2：确认 `backend-async-runner` 已作为 active subject 引用，负向 gate 拦截 `backend-async-runner.*未创建` / `未创建.*backend-async-runner` 回流。

#### 4.5 Owner BDD 场景回归

依次 rerun 各当前仍保留 owner spec BDD suite（target_import / report_generate / privacy_delete / resume_parse / resume_tailor / auth email），证据归档到 [test-checklist](./test-checklist.md)；任一场景失败 → 回到对应 phase 修复。

#### 4.6 Spec / plan 状态收尾

完成全部 acceptance criteria 后，把本 plan 状态从 `active` 改为 `completed`；spec.md / history.md 已在创建时即为 `active`，本步骤不再涉及 `draft → active` 过渡；同步 spec INDEX + plans INDEX。

#### 4.7 L2 scheduler/backoff remediation

针对 code review 暴露的 scheduling 与 retry-finalization 缺口补齐 runtime hardening：`Runtime.Start` 的生产 lease loop 按 registered job_type 独立运行，防止 long-running `report_generate` / `resume_parse` 阻塞 low-priority `email_dispatch`；`Runtime.dispatch` 的 retry `available_at` 与 terminal `completed_at` 使用 handler 返回后的 fresh timestamp；`review.GenerateHandler` 将 failure outcome 归一化给 kernel finalize，`review.Repository.PersistReportFailure` 只维护 `feedback_reports` / outbox / audit 域状态，不再更新 `async_jobs` 或复用 out-of-scope review-store backoff。

#### 4.8 BUG-0106 privacy identity cleanup remediation

修复真实 provider manual UAT 暴露的 `privacy_delete` 完成语义缺口：`DELETE /api/v1/me` 受理时必须同步软删 `users.deleted_at` / `users.status='deleted'` 并撤销该用户所有 session；`privacy_delete` runner 在 upload/profile/JD Match 等 domain cleanup 全部成功后执行用户行最终 hard delete，确保 request/job completed 后不能再通过原邮箱查询到 UAT account identity。执行顺序必须保持失败可重试：任一 domain cleanup 失败不得 hard delete 用户行，成功路径需要 focused handler/store tests 和 cmd/api privacy integration regression 锁定。

#### 4.9 Runtime builder dead dependency cleanup

`buildReportRuntime` / `buildResumeRuntime` 不记录日志，也不向子组件传递 logger；删除两个未导出 builder 的 `*slog.Logger` 参数、nil default 赋值及所有调用点实参。`buildTargetJobRuntime` 仍保留 logger，因为 AI runtime reload warning 真实消费该依赖。以 cmd/api builder/full-funnel tests 和 package `staticcheck` 作为验证 gate。

#### 4.10 Canonical runner contract cleanup

生产入口已经只使用 `runner.Runtime`，因此删除仅由测试使用的 `targetjob.Drainer`、`targetjob.ClaimedJob` / `targetjob.JobOutcome` / `targetjob.JobHandler`、targetjob store 中重复的 claim/finalize SQL，以及 `runner.FromTargetjobHandler` adapter。`target_import` / `source_refresh` / `privacy_delete` / `resume_parse` / `resume_tailor` handler 直接实现 `runner.Handler`；相关 cmd/api 场景通过 `runner.Runtime.RunOnce` 验证 lease、dispatch 与 finalize，场景文件和测试名不再保留 drainer 标签。`scripts/lint/runner_out_of_scope.py` 必须把上述旧形态作为 zero-reference gate。

## 5 验收标准

- 本计划列出的实现 / 测试项全部通过（覆盖 [spec C-1~C-22](../../spec.md#6-验收标准)，含 C-13a missing-consumer safety 与 BUG-0106 privacy identity cleanup）。
- 替代验证 gate 全部 PASS：contract / integration / regression rerun / out-of-scope negative lint / doc reconcile / drift gate。
- 不存在新增的用户可见行为缺口；既有 owner spec BDD 场景 rerun 通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| review.Runner 删除后报告生成回归 | Phase 2.5 完成时 rerun `review/runner_test.go` + `cmd/api/reports_http_scenario_test.go`；任何失败必须先修复再进入下一 phase |
| email_dispatch 切到 async_jobs 后 email code 投递延迟 | 把 `email_dispatch` 列入 `low` priority bucket，且 production `Runtime.Start` 对每个 registered job_type 启动独立 lease loop；P3/P4 收口阶段 smoke 验证 auth email start → DevMailSink delivery 延迟 ≤ 1 个 scan 周期，并用 scheduler regression 证明 long-running critical/default handler 不会阻塞 email loop |
| Outbox dispatcher 上线后 consumer 缺失导致 outbox 行被误确认或长期积压 | runtime 缺少 consumer 时不得置 `published`；dry-run consumer 仅允许测试显式注入；缺少 consumer 的 event 走 retry/dead-letter 并暴露 `outbox_publish_failures_total`，P3 完成前与 F2 / 各 owner 明确启用边界 |
| 范围外模块的 runner 名称回流 | P2.3 / P2.6 作为 negative reconcile；`make lint-runner-out-of-scope` 和 active-doc grep 继续拦截局部 runtime / deleted module positive surface |
| 多 owner spec D-* 边界条款同步遗漏 | Phase 4.4 用 checklist 逐项打勾；P4 收尾必须运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` |
| 退避收口期间 in-flight job 行为变化导致已部署环境异常 | 本仓库无线上环境，P0 不需要兼容 layer；本 plan 仅需保证 dev / test scenario 通过 |
| typed config 节点新增导致 A4 owner spec 需要 additive 修订 | Phase 1.6 把 A4 修订作为前置 checklist item；若实施期发现 A4 owner spec 已有冲突决策，停止进入 plan-review / design 修订，不以 kernel default 常量绕过 |
