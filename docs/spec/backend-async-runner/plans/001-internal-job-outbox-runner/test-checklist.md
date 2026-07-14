# Internal Job and Outbox Runner Test Checklist

> **版本**: 1.13
> **状态**: active
> **更新日期**: 2026-07-13

**关联 Test Plan**: [test-plan](./test-plan.md)
**关联计划**: [plan](./plan.md)

- [x] Phase 1 本计划定义的单元测试项全部通过：`cd backend && go test ./internal/runner/...`（含 `runtime_test.go` / `lease_test.go` / `backoff_test.go` / `reaper_test.go` / `config_test.go`）
- [x] Phase 1 integration 测试通过：`backend/internal/runner/lease_integration_test.go` 真 PG 覆盖 lease / finalize / reclaim 列名与排序
- [x] Phase 1 failure path 断言通过：`TestFinalizeAsyncJob_PermanentFailureAtMax` / `TestReaper_DoesNotIncrementAttempts` / `TestRuntime_ShutdownTimeoutPropagates`
- [x] Phase 1 boundary 断言通过：`TestBackoffPolicy_BoundaryAttempts` / `TestLeaseAsyncJob_RespectsPriorityBuckets`
- [x] Phase 1 typed config 注入通过：`backend/internal/platform/config/loader_test.go::TestAsyncSection` + `backend/internal/runner/config_test.go`（覆盖 `async.queueWeights` / `leaseTimeoutSeconds` / `shutdownGraceSeconds` / `reaperIntervalSeconds` / `scanIntervalSeconds`）
- [x] Phase 1 runtime handler trace 透传通过：`backend/internal/runner/runtime_trace_test.go::TestRuntime_HandlerInheritsTraceparent` + `TestRuntime_HandlerLogsTraceIdField`

- [x] Phase 2 HISTORICAL-SUPERSEDED：当时的 `target_import` / `source_refresh` regression 通过；Phase 7 current contract 不再把 refresh 作为正向能力。
- [x] Phase 2 `privacy_delete` regression 通过：`cd backend && go test ./internal/privacy/runner/...` + cmd/api smoke `DELETE /api/v1/me`
- [x] Phase 2 out-of-scope debrief runner surface negative guard 通过：当前正向 package/test list 不再包含 deleted debrief handler。
- [x] Phase 2 `resume_parse` / `resume_tailor` regression 通过：`cd backend && go test ./internal/resume/jobs/... ./cmd/api -run 'TestResume(Parse|Tailor)Runner' -count=1`
- [x] Phase 2 `report_generate` regression 通过：`cd backend && go test ./internal/review/... ./cmd/api -run 'TestGenerate|TestBuildReportRuntime|TestMainReportRuntime' -count=1`（含 kernel 重写后的 `generate_handler_test.go`）
- [x] Phase 2 out-of-scope Jobs Recommendations / JD Match runner surface negative guard 通过：当前正向 package/test list 不包含 scan/search handler；局部 runtime 形态由 lint 负向断言覆盖。
- [x] Phase 2 退避收口：`backend/internal/runner/backoff_integration_test.go::TestAllHandlersUseSharedBackoff` 通过；out-of-scope `ComputeReportFailureBackoff` / 固定 15s 在非 lint/audit 路径 0 命中

- [x] Phase 3 outbox primary 通过：`TestOutboxDispatcher_ClaimsPendingBatch` + `TestOutboxDispatcher_BatchSizeLimit`（真 PG）
- [x] Phase 3 outbox failure 通过：`TestOutboxDispatcher_DeadLetterAtAttemptFive` + `TestOutboxDispatcher_RedactsLastError`
- [x] Phase 3 outbox idempotency 通过：`TestOutboxDispatcher_DuplicateEventIdHandledIdempotently`
- [x] Phase 3 `source_event_only` skip 通过：`TestOutboxDispatcher_SkipsSourceEventOnly`
- [x] Phase 3 missing consumer safety 通过：`TestDispatcherMissingConsumerDoesNotAck` + `TestDispatcherDryRunConsumerRequiresExplicitRegistration`
- [x] Phase 3 trace 透传通过：`TestOutboxDispatcher_PropagatesTraceParent` + `TestOutboxDispatcher_WarnsOnMissingTrace`
- [x] Phase 3 指标暴露通过：`TestKernelMetrics_FamilyAndLabels` 断言 `outbox_events_pending` / `outbox_publish_duration_seconds` / `outbox_publish_failures_total` / `async_job_duration_seconds` / `async_jobs_processed_total` / `async_job_queue_depth` / `async_job_lag_seconds`
- [x] Phase 3 `email_dispatch` producer 切换通过：`TestStartAuthEmailChallenge_EnqueuesEmailDispatchJob`
- [x] Phase 3 `email_dispatch` handler 通过：`backend/internal/runner/email_dispatch_integration_test.go::TestEmailDispatchHandler_*` + `TestEmailDispatchHandler_PayloadRedaction`
- [x] Phase 3 auth email composition/integration 通过：cmd/api auth integration test 证明 email code 在 1 个 scan 周期内可见。

- [x] Phase 4 cmd/api 单点 shutdown + outbox startup wiring 通过：`cd backend && go test ./cmd/api -run '^(TestMainRunnerKernelDrivesOutboxDispatcher|TestMain_SingleRuntimeShutdown)$' -count=1 -v`
- [x] Phase 4 out-of-scope 形态文件删除断言：`backend/internal/review/structure_test.go::TestNoOutOfScopeRunnerFiles`（或等价 grep）
- [x] Phase 4 `BackgroundMailDispatcher` 引用 0 命中：`backend/internal/auth/mail_test.go::TestNoBackgroundDispatcher`
- [x] Phase 4 `make lint-runner-out-of-scope` PASS（spec D-12 zero-reference；脚本路径 `scripts/lint/runner_out_of_scope.py` + `runner_out_of_scope_test.py`；覆盖局部 runtime 注册路径）
- [x] Phase 4 全局 `cd backend && go build ./...` / `cd backend && go vet ./...` / `cd backend && go test ./...` PASS
- [x] Phase 4 `validate_context.py` PASS（target=backend）
- [x] Phase 4 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- [x] Phase 4 doc reconcile：`backend-runtime-topology` / `backend-review` / `backend-targetjob` / `backend-resume` / `backend-auth` / `event-and-outbox-contract` / `secrets-and-config` / `engineering-roadmap` D-* 边界条款已同步且 0 范围外口径回流；owner spec 负向 grep `grep -n "未来 .backend-async-runner" docs/spec/{backend-runtime-topology,backend-review,backend-targetjob,backend-resume,backend-auth,event-and-outbox-contract}/spec.md` 期望 0 命中；roadmap 负向 grep `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner" docs/spec/engineering-roadmap/spec.md` 期望 0 命中。
- [x] Phase 4 owner regression 由 runner/domain/cmd-api 代码测试承接；代码测试不得进入 `test/scenarios/e2e/` 或作为 E2E 证据。
- [x] Phase 4 MinIO/privacy/audit repeatability 由真实依赖 integration tests 承接，不作为 E2E 证据。
- [x] Phase 4 scheduler/backoff L2 review regression 通过：runner runtime、review GenerateHandler 与 cmd/api composition focused tests 覆盖 fresh finalize timestamp、`email_dispatch` 防饥饿、`report_generate` failure 走 kernel finalize。
- [x] Phase 4 review store integration gate 可发现但本机跳过：`cd backend && go test -tags=integration ./internal/store/review -run '^TestPersistReportFailure' -count=1 -v` PASS with SKIP because `DATABASE_URL is not set`；非 integration 包级回归由 `cd backend && go test ./internal/store/review -count=1` 覆盖。
- [x] Phase 4 backend all-packages regression 通过：`cd backend && go test ./... -count=1` PASS；`make lint-runner-out-of-scope` PASS。
- [x] Phase 4 `git diff --check` PASS
- [x] Phase 4 BUG-0106 privacy identity cleanup regression 通过：`go test ./backend/internal/auth ./backend/internal/privacy/runner ./backend/cmd/api -run '^(TestDeleteMeSoftDeletesUserRevokesAllSessionsAndCreatesPrivacyHandoff|TestSQLStorePrivacyDeleteHandoffSoftDeletesUserAndRevokesSessions|TestSQLStoreMarkDeleteRequestCompletedDeletesAccountIdentityAndPreservesRequestTombstone|TestPrivacyDeleteHandlerHardDeletesAccountIdentityAfterDomainCleanup|TestPrivacyDeleteHandlerDomainFailureKeepsAccountIdentityForRetry|TestPrivacyDeleteRemovesAccountIdentityAfterJobCompletion)$' -count=1` PASS；`go test ./backend/internal/auth ./backend/internal/privacy/runner -count=1` PASS；`DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-check` PASS；`make docs-check` / `make lint-runner-out-of-scope` / `git diff --check` PASS。
- [x] Phase 4 canonical runner cleanup：结构 lint 拒绝 `targetjob.Drainer`、targetjob async job 类型/SQL 和 `FromTargetjobHandler`；五个业务 handler 直接满足 `runner.Handler`；targetjob/resume cmd/api integration tests 经 `runner.Runtime.RunOnce` 通过；全 backend test/staticcheck 通过。
  <!-- verified: 2026-07-10 evidence="runner out-of-scope lint, affected packages, full backend tests, staticcheck, build and vet passed." -->

- [x] Phase 5 HISTORICAL-EVIDENCE business policy tests prove10s/20s/40s/80s cap80 and Runtime default injection. <!-- verified: 2026-07-13 evidence="exact business policy and runtime default tests PASS" -->
- [x] Phase 5 HISTORICAL-EVIDENCE outbox policy tests prove30s/2m/10m/1h/6h, attempt5 failed and no business fallback/cross-injection. <!-- verified: 2026-07-13 evidence="outbox exact policy, terminal and injection tests PASS" -->
- [x] Phase 5 HISTORICAL-SUPERSEDED report integration proved runner10s/20s/40s only，durable attempt4 terminal，no80s/noattempt5/crash-replay overrun. <!-- verified: 2026-07-13 evidence="historical contract only; superseded by Phase 6" -->
- [x] Phase 5 HISTORICAL-SUPERSEDED report enqueue/reservation-failure tests proved explicit max_attempts4 and attempt4 dead/no80s/no fifth lease. <!-- verified: 2026-07-13 evidence="historical contract only; superseded by Phase 6" -->
- [x] Phase 5 HISTORICAL-EVIDENCE DB integration proves reap/takeover fencing for kernel success/retry/failure using running+claimed attempts and typed stale0-row outcome. <!-- verified: 2026-07-13 evidence="kernel stale success/retry/failure PostgreSQL matrix PASS" -->
- [x] Phase 5 HISTORICAL-SUPERSEDED (report reservation only) direct transaction integration proved stale report pre-call reserve could not increment `llm_attempt_count`/admit provider；the still-current result/failure and `resume_tailor` fencing evidence remains valid. <!-- verified: 2026-07-13 evidence="historical product reservation superseded; direct result fencing retained" -->
- [x] Phase 5 HISTORICAL-EVIDENCE clean-break structure gate rejects duplicate review-store lease/reaper files、Repository methods and claim/reaper SQL. <!-- verified: 2026-07-13 evidence="structure test RED before deletion and GREEN after deletion" -->
- [x] Phase 5 HISTORICAL-EVIDENCE focused/full runner+review、lint、contexts、docs/index/diff gates passed under the former report contract. <!-- verified: 2026-07-13 evidence="current product retry requires Phase 6" -->

- [x] Phase 6 action-local retry tests prove initial+3, exact waits10s/20s/40s, context cancellation, return-time destruction and independent next-action reset.
  <!-- verified: 2026-07-13 evidence="backend full/race tests proved calls4/waits10-20-40/destroyed/second1" -->
- [x] Phase 6 runner/store integration proves async_jobs attempts/max_attempts are infrastructure-only and cannot restore/inherit/schedule product retry state.
  <!-- verified: 2026-07-13 evidence="dev PostgreSQL completion integration PASS with generic producer max_attempts=5 and async coupling=false" -->
- [x] Phase 6 fencing tests preserve stale report/outbox/audit/job zero-side-effect behavior without pre-call durable product reservation.
  <!-- verified: 2026-07-13 evidence="review/store/runner/practice race and PostgreSQL fencing regressions PASS" -->
- [x] Phase 6 action retry and async-attempt separation remain code-level assertions and contain no crash/replay global-cap claim.
- [x] Phase 6 negative searches reject active `llm_attempt_count`, report explicit max_attempts4, pre-call reserve and runner-owned product10s/20s/40s semantics.
  <!-- verified: 2026-07-13 evidence="out-of-scope and migration lint PASS; no active coupling found" -->
- [x] Phase 6 focused/full tests, context validation, docs/index and diff gates pass before completion.
  <!-- verified: 2026-07-13 evidence="focused/full runner+review tests and race gates PASS; contexts valid; docs/index zero drift; git diff --check clean" -->

## Phase 7: OPENAPI-002 TargetJob refresh runner contraction

- [ ] 7.1 CONTRACT-ORDER: 统一 RED 已记录，且先消费 B3 Phase 9 的 12-event/7-job/6-API-facing generated handoff；上游未完成时不进入 Runner GREEN。
- [ ] 7.2 SIX-HANDLER-GATE: kernel registry/cmd-api composition 只保留 `email_dispatch`、`privacy_delete`、`report_generate`、`target_import`、`resume_parse`、`resume_tailor` 六个可执行 handler，scheduler/reaper/finalize 回归通过。
- [ ] 7.3 TARGETJOB-REGRESSION: focused targetjob/cmd-api tests prove paste import and failure recovery；阶段收口执行根 `make test`，代码测试不得进入 `test/scenarios/e2e/`。
- [ ] 7.4 ZERO-REF/RETAIN: refresh job/handler/dotted task/queue/metric positive surface 零命中；正向 probe 证明独立 `source_records` table/model/query 保留，再运行 context/diff gate。
