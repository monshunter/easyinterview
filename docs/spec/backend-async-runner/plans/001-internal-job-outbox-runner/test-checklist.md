# Internal Job and Outbox Runner Test Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-22

**关联 Test Plan**: [test-plan](./test-plan.md)
**关联计划**: [plan](./plan.md)

- [x] Phase 1 本计划定义的单元测试项全部通过：`cd backend && go test ./internal/runner/...`（含 `runtime_test.go` / `lease_test.go` / `backoff_test.go` / `reaper_test.go` / `adapter_targetjob_test.go` / `config_test.go`）
- [x] Phase 1 integration 测试通过：`backend/internal/runner/lease_integration_test.go` 真 PG 覆盖 lease / finalize / reclaim 列名与排序
- [x] Phase 1 failure path 断言通过：`TestFinalizeAsyncJob_PermanentFailureAtMax` / `TestReaper_DoesNotIncrementAttempts` / `TestRuntime_ShutdownTimeoutPropagates`
- [x] Phase 1 boundary 断言通过：`TestBackoffPolicy_BoundaryAttempts` / `TestLeaseAsyncJob_RespectsPriorityBuckets`
- [x] Phase 1 typed config 注入通过：`backend/internal/platform/config/loader_test.go::TestAsyncSection` + `backend/internal/runner/config_test.go`（覆盖 `async.queueWeights` / `leaseTimeoutSeconds` / `shutdownGraceSeconds` / `reaperIntervalSeconds` / `scanIntervalSeconds`）
- [x] Phase 1 runtime handler trace 透传通过：`backend/internal/runner/runtime_trace_test.go::TestRuntime_HandlerInheritsTraceparent` + `TestRuntime_HandlerLogsTraceIdField`

- [x] Phase 2 `target_import` / `source_refresh` regression 通过：`cd backend && go test ./internal/targetjob/...`
- [x] Phase 2 `privacy_delete` regression 通过：`cd backend && go test ./internal/privacy/runner/...` + cmd/api smoke `DELETE /api/v1/me`
- [x] Phase 2 `debrief_generate` regression 通过：`cd backend && go test ./internal/debrief/...`
- [x] Phase 2 `resume_parse` / `resume_tailor` regression 通过：`cd backend && go test ./internal/resume/jobs/... ./cmd/api -run 'TestResume(Parse|Tailor)Drainer' -count=1`
- [x] Phase 2 `report_generate` regression 通过：`cd backend && go test ./internal/review/... ./cmd/api -run 'TestE2EP0052|TestE2EP0053|TestE2EP0054|TestE2EP0055' -count=1`（含 kernel 重写后的 generate_handler_test）
- [x] Phase 2 `jd_match_agent_scan` regression 通过：`cd backend && go test ./internal/jdmatch/... ./cmd/api -run 'TestJDMatchAgentScanDrainerScenario|TestBuildJDMatchRuntimeWiresRoutesDrainerAndLifecycle' -count=1`（迁移后 lifecycle 断言重写为 kernel handles `jd_match_agent_scan` 且不 handles `jd_match_search`）
- [x] Phase 2 退避收口：`backend/internal/runner/backoff_integration_test.go::TestAllHandlersUseSharedBackoff` 通过；旧 `ComputeReportFailureBackoff` / 固定 15s 在非 lint/history 路径 0 命中

## Phase 3: Outbox dispatcher + email_dispatch

- [ ] Phase 3 outbox primary 通过：`TestOutboxDispatcher_ClaimsPendingBatch` + `TestOutboxDispatcher_BatchSizeLimit`（真 PG）
- [ ] Phase 3 outbox failure 通过：`TestOutboxDispatcher_DeadLetterAtAttemptFive` + `TestOutboxDispatcher_RedactsLastError`
- [ ] Phase 3 outbox idempotency 通过：`TestOutboxDispatcher_DuplicateEventIdHandledIdempotently`
- [ ] Phase 3 `source_event_only` skip 通过：`TestOutboxDispatcher_SkipsSourceEventOnly`
- [ ] Phase 3 missing consumer safety 通过：`TestDispatcherMissingConsumerDoesNotAck` + `TestDispatcherDryRunConsumerRequiresExplicitRegistration`
- [ ] Phase 3 trace 透传通过：`TestOutboxDispatcher_PropagatesTraceParent` + `TestOutboxDispatcher_WarnsOnMissingTrace`
- [ ] Phase 3 指标暴露通过：`TestKernelMetrics_FamilyAndLabels` 断言 `outbox_events_pending` / `outbox_publish_duration_seconds` / `outbox_publish_failures_total` / `async_job_duration_seconds` / `async_jobs_processed_total` / `async_job_queue_depth` / `async_job_lag_seconds`
- [ ] Phase 3 `email_dispatch` producer 切换通过：`TestStartAuthEmailChallenge_EnqueuesEmailDispatchJob`
- [ ] Phase 3 `email_dispatch` handler 通过：`backend/internal/runner/email_dispatch_integration_test.go::TestEmailDispatchHandler_*` + `TestEmailDispatchHandler_PayloadRedaction`
- [ ] Phase 3 end-to-end auth email 通过：`backend/cmd/api/main_test.go::TestAuthEmailEndToEnd`（magic link 在 1 个 scan 周期内可见）

## Phase 4: 收口 / 文档同步 / lint negative gate

- [ ] Phase 4 cmd/api 单点 shutdown 通过：`backend/cmd/api/main_test.go::TestMain_SingleRuntimeShutdown`
- [ ] Phase 4 旧形态文件删除断言：`backend/internal/review/structure_test.go::TestNoLegacyRunnerFiles`（或等价 grep）
- [ ] Phase 4 `BackgroundMailDispatcher` 引用 0 命中：`backend/internal/auth/mail_test.go::TestNoBackgroundDispatcher`
- [ ] Phase 4 `make lint-runner-legacy` PASS（spec D-12 zero-reference；脚本路径 `scripts/lint/runner_legacy.py` + `runner_legacy_test.py`；覆盖 `jdmatchRuntime.Drainer` / `JobTypeJdMatchAgentScan` 旧注册路径）
- [ ] Phase 4 全局 `cd backend && go build ./...` / `cd backend && go vet ./...` / `cd backend && go test ./...` PASS
- [ ] Phase 4 `validate_context.py` PASS（target=backend）
- [ ] Phase 4 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- [ ] Phase 4 doc reconcile：`backend-runtime-topology` / `backend-review` / `backend-debrief` / `backend-targetjob` / `backend-resume` / `backend-auth` / `backend-jobs-recommendations` / `event-and-outbox-contract` / `secrets-and-config` / `engineering-roadmap` D-* 边界条款已同步且 0 旧口径回流；owner spec 负向 grep `grep -n "未来 .backend-async-runner" docs/spec/{backend-runtime-topology,backend-review,backend-debrief,backend-targetjob,backend-resume,backend-auth,event-and-outbox-contract}/spec.md` 期望 0 命中；roadmap 负向 grep `grep -n "backend-async-runner.*未创建\|未创建.*backend-async-runner" docs/spec/engineering-roadmap/spec.md` 期望 0 命中；backend-jobs 负向 grep `grep -n "jd_match_search.*注册.*runner\|jd_match_search.*kernel" docs/spec/backend-jobs-recommendations/spec.md` 期望 0 命中
- [ ] Phase 4 BDD-Gate owner rerun 全 PASS（auth email + privacy_delete `E2E.P0.003` / `033`；target_import `E2E.P0.010` / `012` / `013`；report_generate `E2E.P0.041` / `052` / `054` / `055`；debrief_generate `E2E.P0.060` / `062`；resume_parse `E2E.P0.034` / `035`；resume_tailor `E2E.P0.077` / `078` / `080`；jd_match_agent_scan / JD Match privacy `E2E.P0.094` / `095` / `096` / `097`）
- [ ] Phase 4 `git diff --check` PASS
