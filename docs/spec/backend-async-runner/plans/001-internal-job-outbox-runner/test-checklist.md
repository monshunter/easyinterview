# Internal Job and Outbox Runner Test Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Test Plan**: [test-plan](./test-plan.md)
**关联计划**: [plan](./plan.md)

- [x] Phase 1 本计划定义的单元测试项全部通过：`cd backend && go test ./internal/runner/...`（含 `runtime_test.go` / `lease_test.go` / `backoff_test.go` / `reaper_test.go` / `config_test.go`）
- [x] Phase 1 integration 测试通过：`backend/internal/runner/lease_integration_test.go` 真 PG 覆盖 lease / finalize / reclaim 列名与排序
- [x] Phase 1 failure path 断言通过：`TestFinalizeAsyncJob_PermanentFailureAtMax` / `TestReaper_DoesNotIncrementAttempts` / `TestRuntime_ShutdownTimeoutPropagates`
- [x] Phase 1 boundary 断言通过：`TestBackoffPolicy_BoundaryAttempts` / `TestLeaseAsyncJob_RespectsPriorityBuckets`
- [x] Phase 1 typed config 注入通过：`backend/internal/platform/config/loader_test.go::TestAsyncSection` + `backend/internal/runner/config_test.go`（覆盖 `async.queueWeights` / `leaseTimeoutSeconds` / `shutdownGraceSeconds` / `reaperIntervalSeconds` / `scanIntervalSeconds`）
- [x] Phase 1 runtime handler trace 透传通过：`backend/internal/runner/runtime_trace_test.go::TestRuntime_HandlerInheritsTraceparent` + `TestRuntime_HandlerLogsTraceIdField`

- [x] Phase 2 `target_import` / `source_refresh` regression 通过：`cd backend && go test ./internal/targetjob/...`
- [x] Phase 2 `privacy_delete` regression 通过：`cd backend && go test ./internal/privacy/runner/...` + cmd/api smoke `DELETE /api/v1/me`
- [x] Phase 2 out-of-scope debrief runner surface negative guard 通过：当前正向 package/test list 不再包含 deleted debrief handler。
- [x] Phase 2 `resume_parse` / `resume_tailor` regression 通过：`cd backend && go test ./internal/resume/jobs/... ./cmd/api -run 'TestResume(Parse|Tailor)Runner' -count=1`
- [x] Phase 2 `report_generate` regression 通过：`cd backend && go test ./internal/review/... ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1`（含 kernel 重写后的 generate_handler_test）
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
- [x] Phase 3 end-to-end auth email 通过：`backend/cmd/api/main_test.go::TestAuthEmailEndToEnd`（email code 在 1 个 scan 周期内可见）

- [x] Phase 4 cmd/api 单点 shutdown + outbox startup wiring 通过：`cd backend && go test ./cmd/api -run '^(TestMainRunnerKernelDrivesOutboxDispatcher|TestMain_SingleRuntimeShutdown)$' -count=1 -v`
- [x] Phase 4 out-of-scope 形态文件删除断言：`backend/internal/review/structure_test.go::TestNoOutOfScopeRunnerFiles`（或等价 grep）
- [x] Phase 4 `BackgroundMailDispatcher` 引用 0 命中：`backend/internal/auth/mail_test.go::TestNoBackgroundDispatcher`
- [x] Phase 4 `make lint-runner-out-of-scope` PASS（spec D-12 zero-reference；脚本路径 `scripts/lint/runner_out_of_scope.py` + `runner_out_of_scope_test.py`；覆盖局部 runtime 注册路径）
- [x] Phase 4 全局 `cd backend && go build ./...` / `cd backend && go vet ./...` / `cd backend && go test ./...` PASS
- [x] Phase 4 `validate_context.py` PASS（target=backend）
- [x] Phase 4 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- [x] Phase 4 doc reconcile：`backend-runtime-topology` / `backend-review` / `backend-targetjob` / `backend-resume` / `backend-auth` / `event-and-outbox-contract` / `secrets-and-config` / `engineering-roadmap` D-* 边界条款已同步且 0 范围外口径回流；owner spec 负向 grep `grep -n "未来 .backend-async-runner" docs/spec/{backend-runtime-topology,backend-review,backend-targetjob,backend-resume,backend-auth,event-and-outbox-contract}/spec.md` 期望 0 命中；roadmap 负向 grep `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner" docs/spec/engineering-roadmap/spec.md` 期望 0 命中。
- [x] Phase 4 BDD-Gate owner rerun 全 PASS（当前正向 scope：`E2E.P0.003` / `010` / `011` / `012` / `013` / `034` / `035` / `077` / `078` / `080`，`E2E.P0.033` live repeatability，Go HTTP BDD `E2E.P0.041` / `052` / `053` / `054` / `055` via `cmd/api/reports_http_scenario_test.go`）。D-22 删除的 out-of-scope 场景不再是当前 gate。
- [x] Phase 4 p0-033 live repeatability regression 通过：dev-stack `make dev-doctor` OK；`DATABASE_URL` + `OBJECT_STORAGE_*` live env 下 `scripts/setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` PASS，覆盖真实 MinIO PUT、`RegisterFileObject`、`DELETE /api/v1/me`、privacy runner succeeded、audit tombstone integration 且重复运行不撞固定 audit id
- [x] Phase 4 scheduler/backoff L2 review regression 通过：`cd backend && go test ./internal/runner ./internal/review ./cmd/api -run '^(TestRuntime_FinalizeUsesTimestampAfterHandlerReturns|TestRuntime_StartDoesNotLetCriticalJobStarveEmailDispatch|TestGenerateHandler_NormalizesFinalizedRetryableFailureThroughKernel|TestE2EP0052ReportGenerationHappyPath|TestE2EP0054ReportAIFailureAndRetry)$' -count=1` PASS，覆盖 fresh finalize timestamp、`email_dispatch` 防饥饿、`report_generate` failure 走 kernel finalize。
- [x] Phase 4 review store integration gate 可发现但本机跳过：`cd backend && go test -tags=integration ./internal/store/review -run '^TestPersistReportFailure' -count=1 -v` PASS with SKIP because `DATABASE_URL is not set`；非 integration 包级回归由 `cd backend && go test ./internal/store/review -count=1` 覆盖。
- [x] Phase 4 backend all-packages regression 通过：`cd backend && go test ./... -count=1` PASS；`make lint-runner-out-of-scope` PASS。
- [x] Phase 4 `git diff --check` PASS
- [x] Phase 4 BUG-0106 privacy identity cleanup regression 通过：`go test ./backend/internal/auth ./backend/internal/privacy/runner ./backend/cmd/api -run '^(TestDeleteMeSoftDeletesUserRevokesAllSessionsAndCreatesPrivacyHandoff|TestSQLStorePrivacyDeleteHandoffSoftDeletesUserAndRevokesSessions|TestSQLStoreMarkDeleteRequestCompletedDeletesAccountIdentityAndPreservesRequestTombstone|TestPrivacyDeleteHandlerHardDeletesAccountIdentityAfterDomainCleanup|TestPrivacyDeleteHandlerDomainFailureKeepsAccountIdentityForRetry|TestPrivacyDeleteRemovesAccountIdentityAfterJobCompletion)$' -count=1` PASS；`go test ./backend/internal/auth ./backend/internal/privacy/runner -count=1` PASS；`DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-check` PASS；`make docs-check` / `make lint-runner-out-of-scope` / `git diff --check` PASS。
- [x] Phase 4 canonical runner cleanup：结构 lint 拒绝 `targetjob.Drainer`、targetjob async job 类型/SQL 和 `FromTargetjobHandler`；五个业务 handler 直接满足 `runner.Handler`；targetjob/resume cmd/api 场景经 `runner.Runtime.RunOnce` 通过；全 backend test/staticcheck 通过。
  <!-- verified: 2026-07-10 evidence="runner_out_of_scope_test.py 8 passed; make lint-runner-out-of-scope PASS; affected five packages and go test ./... PASS; staticcheck ./..., go build ./..., go vet ./... PASS; scenario_script_contract_test.py 4 passed; P0.035/P0.077/P0.078/P0.080 setup-trigger-verify-cleanup PASS." -->
