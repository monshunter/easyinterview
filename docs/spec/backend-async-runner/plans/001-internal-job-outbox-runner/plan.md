# Internal Job and Outbox Runner

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-07-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 Test Checklist**: [test-checklist](./test-checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

保持单一 `backend/internal/runner/` kernel：统一承接 6 个可执行 handler（`email_dispatch` / `privacy_delete` / `report_generate` / `target_import` / `resume_parse` / `resume_tailor`）的 lease / retry / dead-letter / reaper / shutdown 与 B3 outbox dispatcher 协议。所有业务 handler 直接实现 `runner.Handler`，业务域不持有重复 async job contract/SQL 或测试专用 runtime；`privacy_export` 仅保留 DB / OpenAPI 501 contract，无 runner handler。

## 2 背景

[backend-async-runner spec](../../spec.md) §1 / §3 是当前 runtime 真理源。`cmd/api` 只持有一个 `runner.Runtime`，6 个可执行 handler 直接注册；kernel 单点拥有 lease/finalize/reaper SQL，并明确分离business-job与outbox/infra backoff；`runner.OutboxDispatcher` 消费业务事务写入的 outbox；场景通过 `Runtime.RunOnce` 同步驱动真实 kernel。计划维护以 [spec v1.16](../../spec.md)、[checklist](./checklist.md) 和 [test-checklist](./test-checklist.md) 的当前 gate 为准。

## 2.1 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-13 | 1.16 | Reopen Phase 7 for OPENAPI-002 TargetJob refresh handler/job/queue removal while preserving independent source-record persistence. |
| 2026-07-13 | 1.15 | Reopen Phase 6 to separate report action-local initial+3 retries/10s-20s-40s from infrastructure lease attempts; supersede report job max4, durable reservation and crash/replay global-cap semantics without rewriting historical evidence. |
| 2026-07-13 | 1.14 | Close Phase 5 report goal: split business/infra backoff, cap report job/provider execution at four attempts, fence all kernel finalizes plus current report/`resume_tailor` direct async-job transactions, and remove duplicate review-store lease/reaper ownership with a structure gate. |
| 2026-07-13 | 1.13 | L2：report job explicit max_attempts4；claim attempts is lease generation，kernel finalize and current report direct transactions fence on running+claimed generation. |
| 2026-07-13 | 1.12 | Reopen in place：business async uses10s/20s/40s/80s cap80，outbox/infra retains30s/2m/10m/1h/6h，and report max4 provider calls use only10s/20s/40s. |
| 2026-07-10 | 1.11 | Remove the test-only targetjob drainer, duplicate async job contracts and SQL, and the targetjob-to-kernel adapter; run retained handler integration tests directly through runner.Runtime. |
| 2026-07-10 | 1.10 | Remove unused logger dependencies from report and resume runtime builders and all composition call sites. |
| 2026-07-10 | 1.9 | 技术债口径清理：把 `report_generate` 的初始 `feedback_reports` 行描述为 pending row，不再使用 placeholder 口径。 |
| 2026-07-10 | 1.8 | 技术债口径清理：把 review runner 描述改为实施前基线与当前 kernel owner 事实，不再使用旧交接口径。 |
| 2026-07-07 | 1.7 | Wording cleanup：收敛 runner out-of-scope gate 与 auth email smoke 为当前 out-of-scope / email-code 口径，不改变 runner 可执行契约。 |
| 2026-07-06 | 1.6 | D-22 后裁剪复查：当前 runner 正向范围收敛为 7 个可执行 handler + `privacy_export` contract-only；`debrief_generate` / `jd_match_agent_scan` / `jd_match_search` 不再作为当前 plan/context/test 的正向 target surface。 |

## 3 质量门禁分类

- **Plan 类型**: `code-internal` + `contract` + `tooling` + `docs`。涉及 backend Go 代码 / SQL / 配置 / lint 脚本 / 多个 owner spec 边界条款同步。
- **TDD 策略**: Code plan requires TDD。通过 `/implement backend-async-runner/001-internal-job-outbox-runner backend` → `/tdd` 执行；每个代码 / 契约 checklist item 先写或调整 focused test（kernel unit / domain handler integration / outbox integration / cmd/api lifecycle test），再实现最小变更；详见 [test-plan](./test-plan.md) 跨 phase 测试映射。
- **BDD 策略**: 不适用。本计划是 backend-internal runner/SQL/runtime 迁移，不新增用户可见 UI、API envelope 或业务入口；不得把 Go HTTP、package test 或 integration test 包装成 E2E。
- **替代验证 gate**:
  - **Contract test**：kernel `Runtime{Register,RunOnce,ReapOnce,Start,Shutdown}` 接口语义 unit test；`BackoffPolicy.Next` table 测试。
  - **Integration test**：outbox dispatcher 5s scan + FOR UPDATE SKIP LOCKED + retry / dead-letter / metrics 在真 PG 上覆盖；`async.queueWeights` typed config 注入测试。
  - **Regression rerun**：开发中按需运行 runner/domain/cmd-api focused tests；阶段收口统一从仓库根执行 `make test`，不得由 E2E shell 再次编排。
  - **Out-of-scope negative search**：`make lint-runner-out-of-scope`（新增 lint script）扫描 spec D-12 列出的 out-of-scope entry point；本 plan 自身 zero-reference gate 在 P4 收口。
  - **Doc reconcile**：spec D-* 决策落实后必须同步修订 `backend-runtime-topology` § 模块边界、`backend-review` D-13 / D-16、`backend-targetjob` D-5、`backend-resume` § 模块边界、`backend-auth` D-* / `email_dispatch` 章节、`event-and-outbox-contract` § 模块边界、`secrets-and-config` 新增 `async.*` typed config 节点。doc reconcile gate 以 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` + 针对各 owner spec D-* 边界条款的精确 grep 为准（本仓库未维护独立 `scripts/check_docs/` 工具集）。
  - **Drift gate**：`cd backend && go build ./...`、`cd backend && go vet ./...`、Go race test、`make codegen-check`（如新增 generated 资源）、`validate_context.py`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`git diff --check`。

### 3.1 Operation Matrix

本计划不新增 user-facing operation；以下当前仍可执行 job_type 的 API envelope、fixture 与 frontend consumer 均不变，matrix 标注当前 handler、kernel registration、persistence 与代码回归 gate。B3 当前 7 个 canonical job_type 中，`privacy_export` 仅保留为 DB / OpenAPI 501 contract（`requestPrivacyExport` fixture），无 runner producer / handler。

| canonical job_type | 触发 operationId（OpenAPI） | fixture | frontend consumer | backend handler | kernel registration | persistence | AI dependency | code regression |
|--------------------|---------------------------|---------|-------------------|----------------------|--------------------------|-------------|---------------|-------------------|
| `email_dispatch` | `startAuthEmailChallenge` | `openapi/fixtures/Auth/startAuthEmailChallenge.json` | email-code login UI（auth flow） | `backend/internal/auth/email_dispatch_handler.go` | handler 注册到 kernel | `async_jobs(email_dispatch)` + `DeliveryWriter` sink | stub/fixture（dev `DevMailSink`） | auth/runner Go tests + root `make test` |
| `privacy_delete` | `deleteMe` | `openapi/fixtures/Auth/deleteMe.json` | account / privacy 设置面 | `backend/internal/privacy/runner/delete_handler.go` | `PrivacyDeleteHandler` 注册到 kernel | `privacy_requests` + `async_jobs(privacy_delete)` + 用户级资源级联删除 | none | privacy/runner Go tests + root `make test` |
| `report_generate` | `completePracticeSession` / `getFeedbackReport` / `listTargetJobReports` | corresponding PracticeSessions/Reports fixtures | 报告面 | `backend/internal/review/generate_handler.go` | `GenerateHandler` 注册到 kernel | `feedback_reports` + `async_jobs(report_generate)` + `outbox_events` | A3/F3 report profile | review/runner Go tests + root `make test` |
| `target_import` | `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` | 目标导入面 | `targetjob.ParseExecutor` | handler 注册到 kernel | `target_jobs` + `async_jobs(target_import)` + `outbox_events` | A3 parse profile | targetjob/runner Go tests + root `make test` |
| `resume_parse` | `registerResume` / `confirmResumeStructuredMaster` | corresponding Resumes fixtures | 简历上传面 | `resumejobs.ParseHandler` | handler 注册到 kernel | `resume_assets` + `async_jobs(resume_parse)` | A3 resume parse profile | resume/runner Go tests + root `make test` |
| `resume_tailor` | `requestResumeTailor` | corresponding ResumeTailor fixtures | 简历适配面 | `resumejobs.TailorHandler` | handler 注册到 kernel | `resume_tailor_runs` + `async_jobs(resume_tailor)` | A3 resume tailor profile | resume/runner Go tests + root `make test` |

矩阵字段 `frontend consumer` 仅列出受影响入口的语义所属，无前端实现修改；本计划属于 backend-internal 重构，所有仍保留 user-facing operation 的 OpenAPI envelope / fixture 行为不变。D-22 裁剪后，范围外模块的 job/event 不再作为本计划当前正向矩阵行。

## 4 实施步骤

### Phase 1: Kernel 基础设施

#### 1.1 新建 `backend/internal/runner/` package 骨架

`runtime.go`（Runtime 与 Registry）、`lease.go`（lease SQL contract）、`reaper.go`（lease timeout 回收）、`backoff.go`（统一退避表）、`handler.go`（`Handler` / `JobHandlerFunc` / `ClaimedJob` / `JobOutcome` 类型）和 `doc.go` 构成 kernel；所有公共结构配套 godoc。

#### 1.2 实现 Lease / Finalize SQL contract

按 [spec §4.2](../../spec.md#42-sql-约束) 实现 `LeaseAsyncJob` / `FinalizeAsyncJob` / `ReclaimExpiredLeases`；查询条件、排序、`FOR UPDATE SKIP LOCKED` 必须与 spec 完全一致；列名固定为 `locked_at` / `attempts` / `available_at` / `status`；retryable exhausted 统一 finalize 为 `dead`，non-retryable failure 统一 finalize 为 `failed`。

#### 1.3 实现 BackoffPolicy

`DefaultBackoffPolicy()` 返回business infrastructure schedule `[10s,20s,40s,80s]`，attempts超界cap80；`DefaultOutboxBackoffPolicy()`返回infra delivery schedule `[30s,2m,10m,1h,6h]`。`BackoffPolicy.Next`只做schedule映射；runtime与outbox必须显式选择正确default。`MaxAttempts=5`仍是普通async job默认终态上限，具体row可用`max_attempts`收紧；这些值不编码report产品retry。Report的10s/20s/40s由一次`GenerateReport`动作内部的waiter持有。

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

#### 2.1 HISTORICAL-SUPERSEDED: TargetJob import + refresh handler 迁移

本项记录此前两个 TargetJob handler 进入统一 kernel 的历史基础；refresh handler 当前由 Phase 7 删除，只有 `ParseExecutor` 继续作为 `target_import` handler。

#### 2.2 `privacy_delete` 迁移

`privacyrunner.NewPrivacyDeleteHandler(...)` 直接实现 `runner.Handler` 并注册到 kernel。

#### 2.3 Out-of-scope module reconciliation

D-22 后复盘模块范围外，当前不再注册对应 runner handler；本 plan 的当前收口只要求范围外模块不能作为正向 target surface 回流。

#### 2.4 `resume_parse` + `resume_tailor` 迁移

`resumejobs.NewParseHandler(...)` + `resumejobs.NewTailorHandler(...)` 直接实现 `runner.Handler` 并注册到 kernel；lease/finalize SQL 与 retry backoff 只由 kernel 持有。

#### 2.5 `report_generate` 迁移；删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff`

新建 `review.GenerateHandler` 实现 `runner.Handler`（迁移 `review/runner.go` 中 `LeaseAsyncJob → service.GenerateReport → UpdateAsyncJobSucceeded/Failed` 逻辑）；删除 `review.Runner` / `review.Reaper` / `review.ComputeReportFailureBackoff` / `review.DefaultReportFailureBackoff`，并删除`backend/internal/store/review/{lease_async_job,reaper}.go`重复lease/reaper owner。保留`review.Service`业务编排；lease/finalize/reaper只由kernel持有，structure negative test锁定重复文件、Repository方法与claim/reaper SQL零回流。

#### 2.6 Out-of-scope jobs-recommendations reconciliation

D-22 后 Jobs Recommendations / JD Match 模块范围外；当前 runner 不注册对应 scan/search handler。`make lint-runner-out-of-scope` 拦截局部 runtime 与重复 contract 回流。

#### 2.7 退避收口

确认 P2 完成后所有需要job重排的业务 handler走kernel business `DefaultBackoffPolicy()`；删除各域hard-coded基础设施退避。Focused tests覆盖10s/20s/40s/80s+cap80；report产品retry不复用该policy，而在同一动作内独立证明10s/20s/40s；outbox单独证明30s/2m/10m/1h/6h且不得混用。

### Phase 3: Outbox dispatcher + email_dispatch

#### 3.1 实现 `runner.OutboxDispatcher`

新建 `backend/internal/runner/outbox/` (或并入 runner package)：`Dispatcher.RunOnce(ctx)` 按 [spec §4.2](../../spec.md#42-sql-约束) 拉取 pending 行；5s scan + `FOR UPDATE SKIP LOCKED` + batch ≤ 100 + sort by `next_attempt_at asc, created_at asc`；已注册 consumer ack 成功后才置 `published`；临时失败或缺少 runtime consumer 时只按`DefaultOutboxBackoffPolicy()`的30s/2m/10m/1h/6h后移`next_attempt_at`并记录redacted错误，attempts≥5 → `failed`。

### Phase 5: HISTORICAL-EVIDENCE policy split (report coupling superseded)

#### 5.1 RED-GREEN policy factories

以table tests锁定`DefaultBackoffPolicy()`=10s/20s/40s/80s cap80、`DefaultOutboxBackoffPolicy()`=30s/2m/10m/1h/6h；runtime默认注入前者，outbox默认注入后者，空policy fallback只能回business policy且不得让outbox隐式fallback。

#### 5.2 HISTORICAL-SUPERSEDED: Report durable-cap integration

`report_generate` retryable failure由kernel按10s/20s/40s重排，但backend-review在每次provider call前先占用`llm_attempt_count`，attempt4 terminal，禁止80s和第5次调用。普通business job仍可在其`max_attempts`允许时使用80s及cap80。

#### 5.3 HISTORICAL-EVIDENCE gates

运行runner unit/integration、report retry budget、outbox dead-letter、fresh timestamp、registered-handler shared-business-policy与负向策略混用测试；同步owner contexts/docs/index。Phase 5证据全部通过后恢复completed状态。

#### 5.4 HISTORICAL-SUPERSEDED: Report job max-attempt ceiling

`report_generate` producer必须显式写`async_jobs.max_attempts=4`，不能继承DDL默认5。Reservation/CAS失败虽未消耗provider attempt，仍消耗job lease attempt；attempt4 retryable直接dead，不得进入business80s或第5次lease。用producer SQL、lease/finalize及reservation-failure regression做负向证明。

#### 5.5 HISTORICAL-SUPERSEDED (report reservation only): Lease-generation fencing

Claimed `attempts`是lease generation。Report provider pre-call reservation 与 `FinalizeAsyncJob` 都接收 job ID + claimed attempts；reservation 在同一事务先锁 `id + status='running' + attempts=claimed` 再增加 report budget，finalize 以相同 predicate 更新恰好1行；0行返回typed stale-lease。Reaper/reclaim后，旧worker不能再占用 report budget、调用 provider 或提交 success/retry/failure。Phase 5同步封闭report reserve/status/success/failure与直接写async-job result/outbox的`resume_tailor`事务；验收范围保持这些报告路径、tailor直接事务及kernel finalize。

> 上述5.2/5.4及5.5中的report pre-call reservation是历史已执行证据，已被Phase 6取代；kernel finalize与直接业务副作用的running+claimed-attempt fencing继续有效。

### Phase 6: Report action-local retry ownership correction

#### 6.1 Separate product retry from job lease attempts

删除`report_generate`显式`max_attempts=4`作为产品上限、durable provider reservation及`llm_attempt_count`依赖。`async_jobs.attempts/max_attempts`继续只服务lease generation、retryable job重排和terminal finalize，不得换算为provider调用数或10s/20s/40s产品等待。

#### 6.2 Preserve fencing without product reservation

保留kernel finalize与report status/success/failure、`resume_tailor` result/outbox的`jobID + claimed attempts` fencing；stale worker仍必须对report/outbox/audit/job零副作用，但不得再在provider调用前写入持久化产品attempt计数。

#### 6.3 Current correction gates

用RED/GREEN证明：一次`GenerateReport`动作initial+最多3次retry、等待10s/20s/40s；动作返回销毁retry context，新动作从0开始；改变或重放`async_jobs.attempts`不会恢复/继承产品retry context。该证明属于 Go/integration test，不生成 E2E marker；同时对`llm_attempt_count`、pre-call reserve、report job显式max4、crash/replay全局cap执行当前范围负向搜索。

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

#### 4.5 Owner 代码层回归

依次 rerun target_import / report_generate / privacy_delete / resume_parse / resume_tailor / auth email 的 owner package 与必要 integration tests，证据归档到 [test-checklist](./test-checklist.md)；阶段完成统一从仓库根执行 `make test`。本内部迁移不创建 BDD/E2E 场景。

#### 4.6 Spec / plan 状态收尾

完成全部 acceptance criteria 后，把本 plan 状态从 `active` 改为 `completed`；spec.md / history.md 已在创建时即为 `active`，本步骤不再涉及 `draft → active` 过渡；同步 spec INDEX + plans INDEX。

#### 4.7 L2 scheduler/backoff remediation

针对 code review 暴露的 scheduling 与 retry-finalization 缺口补齐 runtime hardening：`Runtime.Start` 的生产 lease loop 按 registered job_type 独立运行，防止 long-running `report_generate` / `resume_parse` 阻塞 low-priority `email_dispatch`；`Runtime.dispatch` 的 retry `available_at` 与 terminal `completed_at` 使用 handler 返回后的 fresh timestamp；`review.GenerateHandler` 将 failure outcome 归一化给 kernel finalize，`review.Repository.PersistReportFailure` 只维护 `feedback_reports` / outbox / audit 域状态，不再更新 `async_jobs` 或复用 out-of-scope review-store backoff。

#### 4.8 BUG-0106 privacy identity cleanup remediation

修复真实 provider manual UAT 暴露的 `privacy_delete` 完成语义缺口：`DELETE /api/v1/me` 受理时必须同步软删 `users.deleted_at` / `users.status='deleted'` 并撤销该用户所有 session；`privacy_delete` runner 在 upload/profile/JD Match 等 domain cleanup 全部成功后执行用户行最终 hard delete，确保 request/job completed 后不能再通过原邮箱查询到 UAT account identity。执行顺序必须保持失败可重试：任一 domain cleanup 失败不得 hard delete 用户行，成功路径需要 focused handler/store tests 和 cmd/api privacy integration regression 锁定。

#### 4.9 Runtime builder dead dependency cleanup

`buildReportRuntime` / `buildResumeRuntime` 不记录日志，也不向子组件传递 logger；删除两个未导出 builder 的 `*slog.Logger` 参数、nil default 赋值及所有调用点实参。`buildTargetJobRuntime` 仍保留 logger，因为 AI runtime reload warning 真实消费该依赖。以 cmd/api builder/full-funnel tests 和 package `staticcheck` 作为验证 gate。

#### 4.10 Canonical runner contract cleanup

生产入口已经只使用 `runner.Runtime`，因此删除仅由测试使用的 `targetjob.Drainer`、`targetjob.ClaimedJob` / `targetjob.JobOutcome` / `targetjob.JobHandler`、targetjob store 中重复的 claim/finalize SQL，以及 `runner.FromTargetjobHandler` adapter。该历史 phase 当时迁移的 refresh handler 由 Phase 7 删除；其余保留 handler 继续直接实现 `runner.Handler`。相关 cmd/api integration tests 通过 `runner.Runtime.RunOnce` 验证 lease、dispatch 与 finalize，测试文件和用例名不再保留 drainer 标签。

## 5 验收标准

- 本计划列出的当前实现 / 测试项全部通过（覆盖 [spec C-1~C-24](../../spec.md#6-验收标准)，含 C-13a missing-consumer safety、BUG-0106 privacy identity cleanup、report action-local retry ownership与当前lease-generation fencing范围）。
- 替代验证 gate 全部 PASS：contract / integration / regression rerun / out-of-scope negative lint / doc reconcile / drift gate。
- 不存在新增的用户可见行为缺口；owner package/integration regression 与根 `make test` 通过。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| review.Runner 删除后报告生成回归 | Phase 2.5 完成时 rerun kernel、`GenerateHandler`、report contract 与 cmd/api composition tests；任何失败必须先修复再进入下一 phase |
| email_dispatch 切到 async_jobs 后 email code 投递延迟 | 把 `email_dispatch` 列入 `low` priority bucket，且 production `Runtime.Start` 对每个 registered job_type 启动独立 lease loop；P3/P4 收口阶段 smoke 验证 auth email start → DevMailSink delivery 延迟 ≤ 1 个 scan 周期，并用 scheduler regression 证明 long-running critical/default handler 不会阻塞 email loop |
| Outbox dispatcher 上线后 consumer 缺失导致 outbox 行被误确认或长期积压 | runtime 缺少 consumer 时不得置 `published`；dry-run consumer 仅允许测试显式注入；缺少 consumer 的 event 走 retry/dead-letter 并暴露 `outbox_publish_failures_total`，P3 完成前与 F2 / 各 owner 明确启用边界 |
| 范围外模块的 runner 名称回流 | P2.3 / P2.6 作为 negative reconcile；`make lint-runner-out-of-scope` 和 active-doc grep 继续拦截局部 runtime / deleted module positive surface |
| 多 owner spec D-* 边界条款同步遗漏 | Phase 4.4 用 checklist 逐项打勾；P4 收尾必须运行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` |
| 退避收口期间 in-flight job 行为变化导致已部署环境异常 | 本仓库无线上环境，P0 不需要兼容 layer；本 plan 以 owner package/integration tests 与根 `make test` 为完成 gate |
| typed config 节点新增导致 A4 owner spec 需要 additive 修订 | Phase 1.6 把 A4 修订作为前置 checklist item；若实施期发现 A4 owner spec 已有冲突决策，停止进入 plan-review / design 修订，不以 kernel default 常量绕过 |
| product、business job与infra schedule误复用 | runner/outbox两个显式factory + injection tests；report使用动作内waiter与reset测试，负向断言runner attempts/max_attempts不拥有产品10s/20s/40s或调用上限 |

## Phase 7: OPENAPI-002 TargetJob refresh runner contraction

本批次依赖顺序固定为：统一 RED → B1/B3/OpenAPI 真理源与生成物 → A4/B4/F3/backend-upload/backend-async-runner 各自 owner surface → backend-targetjob Phase 18 集成 → 全局 zero-reference。Runner 只在 B3 Phase 9 的 12-event/7-job/6-API-facing handoff 可消费后进入 GREEN；任一上游 handoff 未完成时不得宣称本 Phase 完成。

Consume B3 2.15 regenerated 12-event/7-job manifests before runtime edits. RED runner/registry/builder/queue tests must fail while the old refresh job can be registered, leased, handled, assigned to low priority, or emitted as a metric label. GREEN deletes its handler, registration, cmd/api wiring, queue assignment and targeted tests while keeping `target_import` operational and `source_records` table/model/query behavior intact.

Generated/baseline ownership stays with B3 Phase 9, but this phase must compile and test against those artifacts and reject stale generated symbols. BDD is not applicable because no new user behavior is introduced; replacement evidence is B3 contract tests plus runner registry, target-import, scheduler, reaper, cmd/api builder and zero-reference tests. Current runtime/code/generated-consumer surfaces must have zero positive refresh job/handler/dotted task/queue references, excluding history and explicit negative fixtures; `source_records` is explicitly outside that search.
