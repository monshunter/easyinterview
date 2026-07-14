# TargetJob Import and Parse Bootstrap Checklist

> **版本**: 1.28
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

> Phase 0–17 仅保留历史执行证据；其中 URL/file/manual-form/source-refresh 由 Phase 18 supersede，TargetJob latest-report pointer 由 Phase 19 supersede。当前完成判定以 Phase 18-19 与关联 active spec 为准。

## Phase 0: Owner contract remediation

- [x] 0.1 修订 B1/B2 TargetJob API 场景契约；验证: `shared/conventions.yaml` 增加 `TARGET_JOB_NOT_FOUND` / `TARGET_IMPORT_SOURCE_INVALID` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` / `TARGET_INVALID_STATE_TRANSITION` 并重生成 Go/TS/OpenAPI 错误码；B2 additive 扩展 `TargetJobRequirement.kind` 到 `must_have` / `nice_to_have` / `hidden_signal` / `interview_focus`；TargetJobs fixtures 增加 manual_text、manual_form terminal job、URL invalid/unavailable、cross-user hidden 404、invalid transition scenarios；`make codegen-conventions && make codegen-openapi && make codegen-check && make validate-fixtures` 通过
- [x] 0.2 修订 B3 import source event 语义；验证: B3 spec / `shared/events.yaml` / generated docs 明确 `manual_text -> sourceType=text`，`manual_form` 不发 `target.import.requested`；`make codegen-events && make lint-events` 通过，禁止业务包把 `manual_form` 写入当前 v1 event sourceType
- [x] 0.3 修订 F1 TargetJob metrics registry；验证: F1 metrics 字典登记 `target_job_imports_total` / `target_job_parse_duration_seconds` / `target_job_parse_failures_total`，allowed labels 包含有界 `error_code` / `source_type`；新增 metric registry tests 断言不含 URL、target id、user id、prompt version 或自由文本 label

## Phase 1: Storage / config / generated surface boundaries

- [x] 1.1 锁定 store 接口与 SQL 实现；验证: store tests 覆盖 `target_jobs` / `target_job_requirements` / `target_job_sources` 三表的 insert / get / list (含 status / analysisStatus / q / cursor / pageSize) / update / parse-result upsert / 软删过滤；所有 read / write 必须按 `user_id` scope，越权返回 `sql.ErrNoRows` 并由 handler 映射 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND`；不新增 migration
- [x] 1.2 锁定 config / secret 边界；验证: config tests 覆盖 URL fetch timeout / UA 标记由本域代码常量提供；统一出网代理不作为 app-level 配置，代码和文档不得新增等价 proxy key；A3 / F3 缺 provider secret 或 disabled / unsupported profile 时除 `APP_ENV=test` 外 fail-closed；新增 app-level 配置 key 触发 panic 并提示先修订 A4
- [x] 1.3 锁定 generated handler / outbox / job surface；验证: compile / contract tests 断言 4 个 TargetJob handler 经由 B2 generated `ServerInterface` 注册；outbox 写入与 `target_import` 派发使用 B3 generated payload helper，redacted fields negative tests 覆盖 `raw_jd_text` / `source_url` / 文件 URL / prompt / response / provider secret 等违规
- [x] 1.4 Remediation: 修复 `cmd/api` 编译与 TargetJob runtime registration；验证: `backend/internal/platform/secrets` 恢复 A4 `EnvSecretSource` 实现，`go test ./cmd/api` 通过，`cmd/api` route tests 证明 `/api/v1/targets`、`/api/v1/targets/import`、`/api/v1/targets/{targetJobId}` 经 auth middleware 挂载到 `targetjob.Handler`

## Phase 2: Synchronous TargetJob CRUD

- [x] 2.1 实现 `importTargetJob` 同步阶段；验证: handler / service tests 覆盖 4 类 source 各自路径、`Idempotency-Key` 必填；`url` / `manual_text` / `file` 新建场景在事务内插入 `target_jobs` + `target_job_sources` + outbox `target.import.requested`（manual_text event `sourceType=text`）+ 派发 `target_import` job，202 响应包含 generated `TargetJobWithJob` + `Job(jobType=target_import,status=queued)`；`manual_form` 同步 ready 并返回 `Job(jobType=target_import,status=succeeded)`，不发 import requested / parsed 事件
- [x] 2.2 实现 `listTargetJobs`；验证: tests 覆盖 status / analysisStatus / q / cursor / pageSize 过滤与索引使用、cursor base64url 编码 / decode、pageSize clamp 到 [1,100]、跨用户行不返回、软删行不返回、空结果返回 generated `PaginatedTargetJob` 空 envelope
- [x] 2.3 实现 `getTargetJob`；历史验证覆盖按 `(user_id, target_job_id)` 读取、requirements 排序、summary/fit/provenance 与 404 隔离；其中历史报告指针已由 Phase 19 supersede，不是当前正向合同。
- [x] 2.4 实现 `updateTargetJob`；验证: handler tests 覆盖状态机合法迁移（`draft → preparing → applied → interviewing → offer | rejected → archived` + `archived` 兜底）、非法迁移返回 B1 `TARGET_INVALID_STATE_TRANSITION`、`Idempotency-Key` 跨用户隔离、仅写入非空字段、不修改 analysis_status
- [x] 2.5 Remediation: 实现 `updateTargetJob` 按 `(user_id, idempotency_key)` 去重；验证: service / store tests 覆盖同一用户同 key 重复 PATCH 不重复执行 mutation、不同用户同 key 不复用记录、重复 key 返回同一 `targetJobId`，且不修改 `analysis_status`

## Phase 3: Source ingestion (URL / file / manual)

- [x] 3.1 实现 `manual_text` / `manual_form` 写入；验证: tests 覆盖 manual_text 在 `target_job_sources.source_type='manual_text'` 与 `target_jobs.raw_jd_text` 同源写入；manual_form 同步 `analysis_status='ready'` + 至少 1 条草稿 `must_have` requirement，返回 terminal `target_import/succeeded` job，且不派发 runner job、不发出 `target.import.requested` / `target.parsed`
- [x] 3.2 实现 `file` 引用；验证: tests 覆盖 `file_objects.purpose='target_job_attachment'` + `(user_id, file_object_id)` 校验、缺失 / 越权 / purpose 不符返回 B1 `TARGET_JOB_NOT_FOUND` 或 `TARGET_IMPORT_SOURCE_INVALID`（按是否泄露存在性决策）、`target_jobs.source_file_object_id` 与 `target_job_sources.file_object_id` 写入正确，本 plan 下 `raw_jd_text` 允许暂留空待异步阶段或 manual_text 兜底
- [x] 3.3 实现 `url` 抓取守护；验证: SSRF 测试矩阵覆盖 scheme 非 https 拒绝 / 私网 (RFC1918 / 169.254 / `::1` / `fc00::/7`) / 元数据服务 / cross-origin redirect 进入私网 / body cap 1 MiB / timeout 10s / UA 标记；非法/超长/空白映射 B1 `TARGET_IMPORT_SOURCE_INVALID`，上游暂时不可达映射 `TARGET_IMPORT_SOURCE_UNAVAILABLE`；`target_job_sources.url` 为 sanitized URL，`snapshot_text` 为抓取文本且不含 query secret
- [x] 3.4 Remediation: URL source 去除 query secret 并持久化抓取 snapshot；验证: service / urlfetch / pipeline tests 覆盖 `source_url` 与 `target_job_sources.url` 不含 query / fragment / userinfo，合法 URL fetch body 被写入 `snapshot_text` 并作为 parse 输入

## Phase 4: Async parse pipeline

- [x] 4.1 实现 `target_import` runner kernel；验证: runner kernel focused tests 覆盖 handler 入队后立即返回 202、worker pool 并发上限可观测、graceful shutdown drain timeout、pending job 在重启后仍能被 drain；不启动独立 worker 进程也能完成 BDD 与本地验证
- [x] 4.2 调用 F3 Resolve + A3 Complete；验证: tests 覆盖 `RegistryClient.Resolve("target.import.parse", language)` 三元组传递、payload 含 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`、F3 disabled / unsupported profile 触发失败路径、A3 缺 secret fail-closed、prompt body / response body 不入库 / 不入日志 / 不入 metric label
- [x] 4.3 写入解析结果与发出 `target.parsed`；验证: tests 覆盖事务内 upsert `target_job_requirements`（按 `(target_job_id, kind, label)` 去重，display_order 累加）、`target_jobs.summary` / `fit_summary` / `analysis_status='ready'` 同事务更新、outbox `target.parsed` payload 仅含 `targetJobId / userId / analysisStatus / requirementCount / coreThemes`
- [x] 4.4 实现失败路径与 retryable 语义；验证: tests 覆盖 A3 / source 错误到 retryable=true / false 的映射矩阵（`AI_PROVIDER_TIMEOUT` / `AI_FALLBACK_EXHAUSTED` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` retryable；`AI_OUTPUT_INVALID` / `AI_UNSUPPORTED_CAPABILITY` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID` / `TARGET_IMPORT_SOURCE_INVALID` non-retryable）、事务内 `target.analysis.failed` outbox 写入
- [x] 4.5 Internal-only `source_refresh` follow-up job；验证: tests 覆盖 `target.parsed` 触发后写入 internal-only `async_jobs(job_type=source_refresh)`（B3 dotted task `source.refresh`）、payload 不含 source URL 完整路径、runner kernel 端 `SourceRefreshHandler` 标记 `target_job_sources.freshness_status='stale'`
- [x] 4.6 Remediation: 解析成功 / 失败副作用必须原子提交且不引用 B4 `target_jobs` 未定义列；验证: parse executor 使用 `CompleteParseSuccess` / `CompleteParseFailure`，SQL store tests 覆盖 `analysis_status / summary / fit_summary`、`target.parsed` / `target.analysis.failed` outbox 与 `source_refresh` 在同一事务内提交，并覆盖 outbox 写失败 rollback

## Phase 5: Privacy / observability / idempotency redlines

- [x] 5.1 隐私 grep / payload negative tests；验证: privacy grep 0 命中 `raw_jd_text` / `source_url`(含完整 URL 与 query 串) / 文件 object URL / prompt body / response body / provider secret / `Authorization:` 模式；generated outbox / job payload helper 在 negative test 中拒绝任何 redacted field
- [x] 5.2 F1 metric registry preflight；验证: F1 baseline metrics 字典已登记 `target_job_imports_total` / `target_job_parse_duration_seconds` / `target_job_parse_failures_total` 与 allowed labels (`service` / `operation` / `job_type` / `source_type` / `language` / `result` / `error_code`)；metric tests 证明 label 不含 URL、target id、user id、prompt version 或自由文本
- [x] 5.3 Idempotency 跨 user 隔离；验证: store / handler tests 覆盖两个不同用户使用相同 `Idempotency-Key` 不会复用 active `target_import` job；同一用户同 key 重复请求返回同一 `targetJobId` 与同一 active job，DB / outbox 不出现重复 row 与重复事件
- [x] 5.4 Remediation: 强化 privacy / negative gates 覆盖本轮修复；验证: redline tests 覆盖 URL query secret 不进入 stored source URL / event / job payload，active-scope negative search 只允许测试自身声明 forbidden token

## Phase 6: BDD and handoff

- [x] 6.1 BDD-Gate: 验证 E2E.P0.010 通过（覆盖 `importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` 的 primary path）
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 bddChecklist=complete evidence=.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.010/result.json -->
- [x] 6.2 BDD-Gate: 验证 E2E.P0.011 通过
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 bddChecklist=complete evidence=.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.011/result.json -->
- [x] 6.3 BDD-Gate: 验证 E2E.P0.012 通过
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 bddChecklist=complete evidence=.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.012/result.json -->
- [x] 6.4 BDD-Gate: 验证 E2E.P0.013 通过
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 bddChecklist=complete evidence=.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.013/result.json -->
- [x] 6.5 Handoff 给 frontend-home-job-picks-and-parse；验证: `backend/README.md` 或 `backend/internal/targetjob/doc.go` 说明 4 个 operation 的同步 / 异步语义、错误码、idempotency 行为、URL fetch 守护规则、隐私红线、可观测 metric 名、BDD 入口与 mock → real 切换边界
- [x] 6.6 Active-scope 负向搜索通过；验证: `backend/internal/targetjob`、`backend/cmd/api`、`docs/spec/backend-targetjob`、`test/scenarios/e2e/p0-010..013-*` active code/docs 不引入 `mistake.*` / `growth.*` / 独立 `voice` route / 独立 `report` 一级 route / out-of-scope `feature_key` 别名（如 `jd.parse` / `target.parse`）/ embedding / rerank capability / 独立 worker 进程前置依赖 / out-of-scope `interview_round` 独立模块；允许命中仅限本 gate 文本、negative test token、test fake 未实现方法和 handler service-not-configured guard

## Phase 7: L2 remediation and reopened BDD gate

- [x] 7.1 Remediation: URL fetch dial 路径绑定已校验 public IP，覆盖 DNS rebinding / TOCTOU；验证: `cd backend && go test ./internal/targetjob/urlfetch -run 'TestFetch_RejectsDNSRebindOnDial|TestDialContextRejectsPrivateResolvedAddress' -count=1`
- [x] 7.2 Remediation: `updateTargetJob` 状态机校验移入 store 事务并锁定 target row；验证: `cd backend && go test ./internal/targetjob -run 'TestSQLStore_UpdateTargetJobLifecycle_IdempotentRejectsStaleStatusTransition|TestService_UpdateTargetJob_DelegatesStatusTransitionValidationToStore' -count=1`
- [x] 7.3 Remediation: BDD 场景脚本与索引不再把包级 focused tests 标记为真实场景通过；验证: p0-010..013 `trigger.sh` 执行 `cmd/api` HTTP scenario tests，`verify.sh` 输出 `method=cmd-api-http` / `validBddEvidence=true`，`test/scenarios/e2e/INDEX.md` 标记四个场景为 Ready
- [x] 7.4 Remediation: `cmd/api` 接入 `target_import` / `source_refresh` runner kernel、`ParseExecutor`、A3 runtime client、F3 static contract bridge 与 `urlfetch`；验证: `cd backend && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresRunnerAndAIClient|TestBuildAPIHandlerMountsTargetJobRoutesBehindSessionMiddleware' -count=1`
- [x] 7.5 Remediation: TargetJob handler 错误响应统一 generated `ApiErrorResponse`，删除 out-of-scope `{"errors":[...]}` envelope；验证: `cd backend && go test ./internal/targetjob -run 'TestHandler_ImportTargetJob_RejectsMissingIdempotencyKey|TestHandler_ImportTargetJob_RejectsMissingSession|TestHandler_ErrorResponsesUseGeneratedEnvelope' -count=1` 通过，断言 `error.code/message/requestId/retryable` 存在且无 `errors` key
- [x] 7.6 Remediation: `listTargetJobs` 设置实际生效的 `pageInfo.pageSize`；验证: `cd backend && go test ./internal/targetjob -run 'TestService_ListTargetJobs_PassesFiltersAndShapesPaginated|TestService_ListTargetJobs_PageInfoReportsEffectivePageSize' -count=1` 通过，覆盖默认 page size、clamp 到 100、非法值兜底与空列表 envelope
- [x] 7.7 Remediation: `ParseExecutor` 从 F3 contract bridge / prompt metadata 组装 A3 payload，`APP_ENV=test` 的 `cmd/api` runtime 注入只作用于 `target.import.parse` 的 deterministic JSON parse fixture client；验证: `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_UsesPromptMessagesFromRegistryResolution|TestDeterministicParseAIClient_OnlyInterceptsTargetImportParse' -count=1 && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresRunnerAndAIClient' -count=1` 通过
- [x] 7.8 Remediation: AI parse 输出无有效 requirement 或含非法 kind / label / evidence level 时走 `AI_OUTPUT_INVALID`；验证: `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_AIOutputInvalid_WhenRequirementsAreSemanticallyInvalid' -count=1` 通过，覆盖 all-invalid requirements、空 label、非法 kind、非法 evidence level 均写 `target.analysis.failed` 且不把 TargetJob 标记为 `ready`
- [x] 7.9 Remediation: BDD handoff 文案与场景状态对齐，包级 focused tests 不再被记录为真实 BDD PASS；验证: `backend/internal/targetjob/doc.go`、`test/scenarios/e2e/INDEX.md` 与 p0-010..013 README / verify outputs 标注 `cmd-api-http` / `validBddEvidence=true`，`bash -n test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/trigger.sh test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/verify.sh test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/trigger.sh test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/verify.sh test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/trigger.sh test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/verify.sh test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/trigger.sh test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/verify.sh` 通过，主 checklist 6.1-6.4 已用 `.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.010..013/result.json` 证据闭合
- [x] 7.10 将 E2E.P0.010 / 011 / 012 / 013 迁移为 auth -> HTTP API -> cmd/api runner kernel 的真实场景；验证: 新增 `backend/cmd/api` HTTP scenario harness，p0-010..013 `trigger.sh` 执行 `go test -v ./cmd/api -run 'TestE2EP0010HTTPTextImportParseReady|TestE2EP0011HTTPURLImportFetchAndParse|TestE2EP0012HTTPParseFailureRetryableAndNonRetryable|TestE2EP0013HTTPManualFormReady'` 对应场景，`verify.sh` 输出 `status=passed` / `method=cmd-api-http` / `validBddEvidence=true`，证据位于 `.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.010..013/result.json`
- [x] 7.11 Remediation: 删除 TargetJob active SQL 对已移除 `target_jobs.profile_id` 列的依赖；验证: `cd backend && go test ./internal/targetjob -run 'TestSQLStore_|TestService_GetTargetJob|TestHandler_ErrorResponsesUseGeneratedEnvelope' -count=1` 通过，`cd backend && DATABASE_URL=<local-dev-postgres-dsn> go test -tags=integration ./internal/targetjob -run TestSQLStoreIntegration_CompleteParseFailureDeletesTargetAndSources -count=1` 通过，`rg 'profile_id|ProfileID|profileID' backend/internal/targetjob backend/cmd/api openapi/fixtures/TargetJobs openapi/openapi.yaml migrations shared` 无 active 命中，host-run backend 上解析失败后的 `GET /api/v1/targets/{targetJobId}` 返回 404 而不是可见 failed 资产
- [x] 7.12 Remediation: BUG-0142 真实 Postgres integration gate 不得在缺少 DB 环境时 `SKIP` 假绿；验证: 未设置 `DATABASE_URL` 时 `cd backend && go test -tags=integration ./internal/targetjob -run TestSQLStoreIntegration_CompleteParseFailureDeletesTargetAndSources -count=1 -v` 非 0 且不输出 `--- SKIP`，设置真实 local-dev Postgres `DATABASE_URL` 后同一 focused gate 通过

## Phase 8: JD identity and current-plan binding remediation

- [x] 8.1 `target.import.parse` prompt/schema/parse executor require and persist `title` / `companyName` on parse success（验证：`cd backend && go test ./internal/targetjob -run 'TestParseExecutor|TestTargetImportPrompt|TestSQLStore_' -count=1`; `make lint-prompts` PASS）
- [x] 8.2 `TargetJob` list/detail responses expose persisted `resumeId`; Phase 17 supersedes latest-ready selection so `currentPracticePlanId` requires the exact current pair and bound resume without a mutable target_jobs progress column（验证：Phase 17 targetjob unit/real-Postgres projection gates）
- [x] 8.3 BDD-Gate: `E2E.P0.010` target import parse ready and `E2E.P0.018` workspace plan-card selection remain aligned with the additive response contract（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0010HTTPTextImportParseReady|TestE2EP0022PracticePlanBaselineCreateAndRead' -count=1`; focused frontend workspace suites PASS）

## Phase 9: TargetJob-level resume binding remediation

- [x] 9.1 OpenAPI, fixtures, generated Go/TS types and migration define required `ImportTargetJobRequest.resumeId` plus persisted `target_jobs.resume_id`（验证：`make codegen-openapi`; `make validate-fixtures`; local migration column check PASS）
- [x] 9.2 `importTargetJob` validates the resume belongs to the current user and persists `target_jobs.resume_id` for all source variants without leaking cross-user resume existence（验证：`cd backend && go test ./internal/targetjob -count=1` PASS）
- [x] 9.3 `listTargetJobs` / `getTargetJob` expose `TargetJob.resumeId` from `target_jobs.resume_id` even when no ready `practice_plans` row exists; `currentPracticePlanId` remains the only latest-plan projection（验证：targetjob store/service/handler tests PASS; local HTTP smoke imported a target with no practice plan and read back `resumeId` from DB/list/detail）
- [x] 9.4 BDD-Gate: `E2E.P0.010` import parse ready and `E2E.P0.018` workspace list re-entry stay aligned with target job-level resume binding（验证：focused backend cmd/api + frontend workspace tests + `E2E.P0.018` scenario wrapper + local API smoke PASS）

## Phase 10: Real-provider parse robustness remediation

- [x] 10.1 Accept valid AI output when the JD omits company name; verification: `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_HappyPathCoalescesMissingCompanyName|TestParseExecutor_AIOutputInvalid_WhenRequirementsAreSemanticallyInvalid' -count=1` PASS, with `zh-CN` fallback `未提供` and empty title still rejected.
- [x] 10.2 Keep JSON output validation strict while accepting a whole markdown fenced JSON value; verification: `cd backend && go test ./internal/ai/aiclient/outputschema -run TestValidate -count=1` PASS and `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_HappyPathAcceptsFencedJSON|TestParseExecutor_AIOutputInvalid_WhenRequirementsAreSemanticallyInvalid' -count=1` PASS, with leading/trailing prose still rejected.
- [x] 10.3 Wrap TargetJob parse AI in A3 observability at `cmd/api` runtime; verification: `cd backend && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresRunnerAndAIClient|TestBuildTargetJobRuntimeWrapsParseAIWithObservability' -count=1` PASS and final local import produced `ai_task_runs.task_type=jd_parse`, provider/model, `status=success`, `validation_status=ok`.
- [x] 10.4 BDD/real-browser smoke: redeploy backend and verify screenshot-class JD parse is ready; verification: `test/scenarios/env-redeploy.sh backend` PASS, `test/scenarios/env-verify.sh` PASS, local API import returned `analysisStatus='ready'`, title `AI 应用技术负责人`, company `未提供`, 14 requirements, and authenticated browser `/parse?...targetJobId=019f44a1-b43e-754f-ba0b-3cd9ed11ce1f` rendered `route-parse` with no `JD 解析失败`.

## Phase 11: parse failure admission remediation

- [x] 11.1 `CompleteParseFailure` writes `target.analysis.failed` and deletes the failed `target_jobs` row in one transaction, relying on FK cascade to remove source/requirement rows（验证：`cd backend && go test ./internal/targetjob -count=1` PASS；`DATABASE_URL=... go test -tags=integration ./internal/targetjob -run TestSQLStoreIntegration_CompleteParseFailureDeletesTargetAndSources -count=1 -v` PASS）
- [x] 11.2 `ParseExecutor` failure outcomes keep async job/outbox diagnostics but do not leave a readable TargetJob asset（验证：`cd backend && go test ./internal/targetjob -count=1` PASS）
- [x] 11.3 HTTP scenario proves parse failure is not admitted into TargetJob read/list assets（验证：`cd backend && go test ./cmd/api -run 'TestE2EP0010HTTPTextImportParseReady|TestE2EP0012HTTPParseFailureRetryableAndNonRetryable|TestE2EP0013HTTPManualFormReady|TestBuildTargetJobRuntimeWiresRunnerAndAIClient' -count=1` PASS）
- [x] 11.4 BDD-Gate: `E2E.P0.012` documents parse-failure deletion semantics and remains backed by cmd/api HTTP evidence（验证：`test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/setup.sh && .../trigger.sh && .../verify.sh` PASS）

## Phase 12: TargetJob archive/delete integration

- [x] 12.1 B2 additive contract adds `archiveTargetJob`; 验证: `openapi/openapi.yaml`、`openapi/fixtures/TargetJobs/archiveTargetJob.json`、operation inventory、generated Go server/TS client all include `POST /targets/{targetJobId}/archive`; `make codegen-openapi && make lint-openapi && make validate-fixtures` PASS
- [x] 12.2 Backend store/service/handler persist archive; 验证: `cd backend && go test ./internal/targetjob -run 'TestHandlerSignaturesMatchB2ServerInterface|TestHandler_ArchiveTargetJob|TestService_ArchiveTargetJob|TestSQLStore_ArchiveTargetJob|TestStoreSurfaceRequiresUserScopeOnReadsAndWrites' -count=1` PASS；`cd backend && go test ./internal/targetjob -count=1` PASS；覆盖 generated handler signature、缺 `Idempotency-Key`、success `status='archived' + deleted_at`、idempotent replay、already-archived conflict、cross-user 404 与 read-side soft-delete contract
- [x] 12.3 Frontend workspace delete calls generated `archiveTargetJob`; 验证: workspace tests prove delete sends `Idempotency-Key`, removes card only after success, does not navigate, reports failure, and no source path remains that implements delete as local-only hiding；`pnpm --filter @easyinterview/frontend test src/app/screens/home/MockInterviewCard.test.tsx src/app/screens/home/HomeRecentMocks.test.tsx src/app/screens/workspace/WorkspaceScreen.test.tsx src/app/screens/workspace/WorkspaceEmptyState.test.tsx` PASS
- [x] 12.4 BDD-Gate: `E2E.P0.018` and local screenshot acceptance prove persistent workspace archive; 验证: `test/scenarios/e2e/p0-018-workspace-default-render/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS；local real-backend browser smoke shows deleted card absent after refresh, DB readback `status='archived'` and `deleted_at is not null`, and screenshots capture top-right delete before archive plus post-delete workspace list
- [x] 12.5 Remediation: queued or retrying `target_import` jobs must terminate after their TargetJob is archived or otherwise no longer visible to parse reads; 验证: `go test ./backend/internal/targetjob -run TestParseExecutor_MissingTargetIsTerminalWithoutFailureCleanup -count=1` RED before fix then PASS after fix；`go test ./backend/internal/targetjob -run 'TestParseExecutor|TestSQLStore_ArchiveTargetJob|TestService_ArchiveTargetJob|TestHandler_ArchiveTargetJob|TestSQLStore_CompleteParseFailure' -count=1` PASS；`go test ./backend/internal/targetjob -count=1` PASS

## Phase 13: URL source-row invariant fail-closed cleanup

- [x] 13.1 `ParseExecutor` 在 URL fetch 前验证必备 source row；缺行时不发起 fetch/AI、不写 success/snapshot，以 non-retryable `TARGET_IMPORT_SOURCE_INVALID` 走失败清理事务；验证: focused red/green `TestParseExecutor_URLFetchWithoutSourceRowFailsBeforeFetch`，相关 `TestParseExecutor_URLFetch*` 与 `go test ./internal/targetjob -count=1` 通过。
  <!-- verified: 2026-07-10 method=tdd-url-source-row-invariant evidence="RED: focused test observed Succeeded:true with no source row. GREEN: missing-row test returns non-retryable TARGET_IMPORT_SOURCE_INVALID before fetch and asserts failure cleanup/no snapshot/no success; related URL fetch tests PASS after supplying their required source fixtures; go test ./internal/targetjob -count=1 PASS; go test ./cmd/api -run ^TestE2EP0011HTTPURLImportFetchAndParse$ -count=1 PASS; old silent-path wording search zero; owner contexts, sync-doc-index, docs-check, diff-check and pruning surface PASS real_residuals=0." -->

## Phase 14: TargetJob test dead initialization cleanup

- [x] 14.1 忽略 `newParseExecutorWithFakes` 在 `TestParseExecutorAITaskRuns` 中从未读取的初始 executor 返回值，保留包装 observability 的真实受测 executor；验证: `go test ./internal/targetjob -run '^TestParseExecutorAITaskRuns$' -count=1`、`staticcheck ./internal/targetjob/...` 与 package gate 通过。
  <!-- verified: 2026-07-10 method=targetjob-test-dead-initialization-cleanup evidence="RED: backend-wide staticcheck reported SA4006 at pipeline_test.go because the initial executor value was overwritten unread. GREEN: ignored that return value and declared the sole effective executor at observability assembly; focused TestParseExecutorAITaskRuns, staticcheck ./internal/targetjob/... and go test ./internal/targetjob -count=1 PASS; backend-targetjob/product contexts, sync-doc-index, docs-check, diff-check and pruning surface PASS real_residuals=0." -->

## Phase 15: App-config pseudo-tripwire cleanup

- [x] 15.1 删除零调用 panic API/self-test 与重复本地 getenv 文本扫描，保留 proxy-key 负向测试；验证：production `deadcode` RED/GREEN、TargetJob tests/staticcheck、`make lint-getenv-boundary`、`make lint-env-dict`、owner docs gates。
  <!-- verified: 2026-07-10 method=targetjob-app-config-pseudo-tripwire-cleanup evidence="Production deadcode RED identified MustNotIntroduceAppLevelConfigKey as test-only. Deleted the panic API/self-test and duplicate local getenv text scan; retained the domain proxy-key negative test. TargetJob tests, staticcheck, reachability/symbol inventory, lint-getenv-boundary and lint-env-dict PASS." -->

## Phase 16: Cmd/api cookie JSON harness consolidation

- [x] 16.1 Record scoped `cmd/api` `dupl` RED for the TargetJob and full-funnel harness request bodies.
  <!-- verified: 2026-07-10 method=cmd-api-cookie-json-harness-dupl evidence="The two receiver methods are cmd/api's only clone group at threshold 100 and differ only in receiver-owned handler/cookie plus the canonical header constant." -->
- [x] 16.2 Delegate the TargetJob receiver to one shared package test helper without changing P0.010-P0.013 requests or assertions.
  <!-- verified: 2026-07-10 method=cmd-api-cookie-json-helper evidence="TargetJob keeps its receiver and IdempotencyKeyHeader while delegating only shared request mechanics. P0.010-P0.013 all PASS, the cmd/api P0.098 handler harness passes, and cmd/api dupl reports zero clone groups; this is not live-browser evidence." -->
- [x] 16.3 Run P0.010-P0.013, P0.098, cmd/api/full backend/static and owner documentation gates.
  <!-- verified: 2026-07-10 method=cmd-api-cookie-json-harness-closeout evidence="P0.010-P0.013 and the cmd/api P0.098 handler harness PASS; cmd/api/full backend, vet/staticcheck, scoped dupl, backend-targetjob/e2e/product contexts and docs/index/diff/pruning gates PASS. This record does not claim a live browser." -->

## Phase 17: Backend-persisted practice progress projection

- [x] 17.1 RED: Get/List store tests require typed completion/ready-plan facts, `session_completed` event filtering, pair non-null filtering and one list query for multiple cards.<!-- verified: 2026-07-12 method=unit+P0.098 test=TestSQLStore_ListTargetJobsForUser_LoadsPageScopedPracticeLedgerFactsInOneQuery -->
- [x] 17.2 GREEN: add page-scoped no-N+1 SQL aggregation; remove global-latest-plan fallback; require completed/ready facts to match `target_jobs.resume_id`, while report/lifecycle status remains independent.<!-- verified: 2026-07-12 method=unit+real-postgres marker=wrong-resume-completion-ignored=PASS -->
- [x] 17.3 RED-GREEN: service projection requires complete provenance, lowercase allowlisted type and positive int32 strictly increasing/unique sequence; accepts `1,2,4` and selects existing successor `4`; covers zero/duplicate/out-of-order/wrong-resume/unknown/legacy facts, lifecycle-status independence, newer old-round retry, current-plan exact match and all-complete null current/plan.<!-- verified: 2026-07-12 method=unit+real-postgres markers="wrong-resume-completion-ignored,persisted-first-to-next,target-report-status-independent,out-of-order-gap-hidden,non-contiguous-round-1-2-4,get-list-first-next-final-parity" -->
- [x] 17.4 Handler/generated JSON tests prove Get/List wire parity and optional fail-closed behavior for invalid/unloaded summaries.<!-- verified: 2026-07-12 method=handler-tests -->
- [x] 17.5 BDD-Gate: P0.098 executes persisted first→next→final, duplicate completion, report-state independence, Get/List parity and no frontend business-storage assertions.<!-- verified: 2026-07-12 method=scenario-run result=PASS -->
- [x] 17.6 Run focused/full backend, OpenAPI/migration, query-count, context/docs/index/diff/privacy gates.<!-- verified: 2026-07-12 evidence="Get/List one-query projection; real DB first-next-final+wrong-resume+1-2-4; OpenAPI/migrate; make test; context/docs/index/diff" -->

## Phase 18: Paste-only JD import contract convergence

- [x] 18.1 RED: contract tests 必须证明旧 request source union、`TargetJob.sourceType/sourceUrl`、JD source 表/列、JD attachment purpose、URL fetch、同步手工表单、JD source refresh 与来源专属错误码当前仍可达；记录具体失败断言。
  <!-- verified: 2026-07-13 evidence="OpenAPI/schema/event/migration/config and backend compile RED exposed source union/response fields, target_job_sources columns/table, target_job_attachment, urlfetch/source_refresh branches and removed source-specific error codes before GREEN edits." -->
- [x] 18.2 GREEN: `importTargetJob` 只接受 `{rawText,targetLanguage,resumeId}`，空白返回 `VALIDATION_FAILED`；`target_jobs.raw_jd_text` 是唯一 JD 原文事实源，ParseExecutor 无 source fallback，所有有效请求只派发 queued `target_import`。
  <!-- verified: 2026-07-13 evidence="Handler rejects removed source wrapper before Store; blank rawText returns HTTP 422; service/store tests prove trimmed raw_jd_text, resume ownership, user-scoped idempotency, exact source-free event/job payload and queued-only import; ParseExecutor reads only RawJDText; full targetjob package PASS." -->
- [x] 18.3 GREEN: 删除 TargetJob URL/file/manual-form 分支、来源 response/event/metric 字段、来源表/列、JD attachment purpose、URL fetcher、JD source refresh job/event/handler/registration、来源专属错误码、`jd_source_url` prompt/config/seed 派生物；保留 resume/privacy upload 与独立 `source_records`。
  <!-- verified: 2026-07-13 method=contract+runtime+prompt evidence="URL/file/manual-form/source response/event/metric/persistence/config branches and source-specific errors are removed; urlfetch directory is absent; target.import.parse token set is exactly jd_text+language with hash 9ab316e...; resume/privacy upload and source_records gates remain positive." -->
- [x] 18.4 BDD-Gate: 修订并运行 `E2E.P0.010` direct rawText success/idempotency/ready 与 `E2E.P0.012` AI failure/recovery；删除 `E2E.P0.011` / `E2E.P0.013` 场景目录和 INDEX 行。
  <!-- verified: 2026-07-13 method=scenario-run evidence="E2E.P0.010 and E2E.P0.012 setup/trigger/verify/cleanup PASS against the current host-run stack; P0.011/P0.013 directories and active INDEX entries are deleted." -->
- [x] 18.5 Zero-ref: 对 production/OpenAPI/shared/config/migrations/prompts/generated/fixtures/scripts 运行精确负向搜索并通过；另以正向测试确认 `purpose=resume`、`privacy_export`、file_objects privacy cleanup 与独立 `source_records` 保持可用。
  <!-- verified: 2026-07-13 method=scoped-negative+positive-contract evidence="Current runtime/schema/fixture/prompt scans have zero obsolete TargetJob source branches; SQL inventory contains no removed source columns; upload register/privacy tests and migration probes preserve resume, privacy_export, file cleanup and independent source_records." -->
- [x] 18.6 HISTORICAL-SUPERSEDED: 上一轮最终汇总 gate 不作为 Phase 19 重开后的当前完成证据；当前 focused/full、zero-ref 与生命周期恢复仅由 19.5 承接。
  <!-- superseded: 2026-07-14 decision="用户批准方案 A；不在已演进的 OpenAPI/数据库契约上伪跑前置汇总，也不复用历史 PASS。" current-owner="19.5" -->

## Phase 19: Remove TargetJob latest-report pointer

- [x] 19.1 RED: OpenAPI/generated/fixture、TargetJob store 与 baseline SQL tests 至少各暴露一个仍可达的 report pointer，并在 GREEN 前失败。
  <!-- verified: 2026-07-14 method=tdd-red evidence="The baseline contract failed on latest_report_id; backend compilation failed where service still projected the field removed from generated TargetJob; frontend typecheck failed on the stale typed fixture. Current OpenAPI/generated/fixture negative contracts remained green and were not reverted." -->
- [x] 19.2 GREEN: 原地删除 public field、DB column、scan/insert/update、codegen 与 fixture 引用；不增加兼容字段、trigger 或 replacement pointer。
  <!-- verified: 2026-07-14 method=focused-go+vitest evidence="Baseline column, TargetJob record/SELECT/scan/service/sqlmock projections and the stale frontend fixture were removed in place. Migration focused, full TargetJob package and real-API fixture tests pass with no replacement pointer." -->
- [x] 19.3 REGRESSION-GATE: TargetJob Get/List canonical rounds、practice progress、resume binding、archive/isolation 与 real PostgreSQL tests 通过。
  <!-- verified: 2026-07-14 method=focused+real-postgres evidence="Full targetjob package passes with canonical Get/List progress, archive/isolation, resume binding and paste-only real PostgreSQL coverage after the clean environment rebuild." -->
- [x] 19.4 HANDOFF-GATE: backend-review `listTargetJobReports` 成为 current ready/latest attempt 唯一 API owner，frontend-report-dashboard ReportsScreen / P0.059 消费 overview；Parse/P0.016 不再消费列表，本 Phase 不复制 report selection logic。
  <!-- verified: 2026-07-14 method=consumer-negative+P0.016+P0.059 evidence="ReportsScreen is the sole frontend listTargetJobReports consumer; backend-review owns selection; Parse makes zero list requests and P0.016/P0.059 pass." -->
- [x] 19.5 ZERO-REF: production/generated/OpenAPI/fixtures/migrations 中旧字段/列精确零命中，历史/负向审计之外无正向表述。
  <!-- verified: 2026-07-14 decision="用户批准方案 A" method=aggregate-current-gates evidence="Current migration, OpenAPI, 37 fixtures, isolated-index codegen drift, full backend/frontend, context, docs, diff and active-surface negative gates pass; 18.6 remains historical-superseded." -->
