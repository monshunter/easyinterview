# TargetJob Import and Parse Bootstrap Checklist

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-08

**关联计划**: [plan](./plan.md)

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
- [x] 2.3 实现 `getTargetJob`；验证: handler tests 覆盖按 `(user_id, target_job_id)` 读取、requirements 数组按 `display_order` 排序、`summary` / `fitSummary` / `provenance` 字段非空校验、`latestReportId` 暂为 nil 占位、越权 / 软删返回 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND`
- [x] 2.4 实现 `updateTargetJob`；验证: handler tests 覆盖状态机合法迁移（`draft → preparing → applied → interviewing → offer | rejected → archived` + `archived` 兜底）、非法迁移返回 B1 `TARGET_INVALID_STATE_TRANSITION`、`Idempotency-Key` 跨用户隔离、仅写入非空字段、不修改 analysis_status
- [x] 2.5 Remediation: 实现 `updateTargetJob` 按 `(user_id, idempotency_key)` 去重；验证: service / store tests 覆盖同一用户同 key 重复 PATCH 不重复执行 mutation、不同用户同 key 不复用记录、重复 key 返回同一 `targetJobId`，且不修改 `analysis_status`

## Phase 3: Source ingestion (URL / file / manual)

- [x] 3.1 实现 `manual_text` / `manual_form` 写入；验证: tests 覆盖 manual_text 在 `target_job_sources.source_type='manual_text'` 与 `target_jobs.raw_jd_text` 同源写入；manual_form 同步 `analysis_status='ready'` + 至少 1 条草稿 `must_have` requirement，返回 terminal `target_import/succeeded` job，且不派发 runner job、不发出 `target.import.requested` / `target.parsed`
- [x] 3.2 实现 `file` 引用；验证: tests 覆盖 `file_objects.purpose='target_job_attachment'` + `(user_id, file_object_id)` 校验、缺失 / 越权 / purpose 不符返回 B1 `TARGET_JOB_NOT_FOUND` 或 `TARGET_IMPORT_SOURCE_INVALID`（按是否泄露存在性决策）、`target_jobs.source_file_object_id` 与 `target_job_sources.file_object_id` 写入正确，本 plan 下 `raw_jd_text` 允许暂留空待异步阶段或 manual_text 兜底
- [x] 3.3 实现 `url` 抓取守护；验证: SSRF 测试矩阵覆盖 scheme 非 https 拒绝 / 私网 (RFC1918 / 169.254 / `::1` / `fc00::/7`) / 元数据服务 / cross-origin redirect 进入私网 / body cap 1 MiB / timeout 10s / UA 标记；非法/超长/空白映射 B1 `TARGET_IMPORT_SOURCE_INVALID`，上游暂时不可达映射 `TARGET_IMPORT_SOURCE_UNAVAILABLE`；`target_job_sources.url` 为 sanitized URL，`snapshot_text` 为抓取文本且不含 query secret
- [x] 3.4 Remediation: URL source 去除 query secret 并持久化抓取 snapshot；验证: service / urlfetch / pipeline tests 覆盖 `source_url` 与 `target_job_sources.url` 不含 query / fragment / userinfo，合法 URL fetch body 被写入 `snapshot_text` 并作为 parse 输入

## Phase 4: Async parse pipeline

- [x] 4.1 实现 `target_import` drainer；验证: drainer focused tests 覆盖 handler 入队后立即返回 202、worker pool 并发上限可观测、graceful shutdown drain timeout、pending job 在重启后仍能被 drain；不启动独立 worker 进程也能完成 BDD 与本地验证
- [x] 4.2 调用 F3 Resolve + A3 Complete；验证: tests 覆盖 `RegistryClient.Resolve("target.import.parse", language)` 三元组传递、payload 含 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`、F3 disabled / unsupported profile 触发失败路径、A3 缺 secret fail-closed、prompt body / response body 不入库 / 不入日志 / 不入 metric label
- [x] 4.3 写入解析结果与发出 `target.parsed`；验证: tests 覆盖事务内 upsert `target_job_requirements`（按 `(target_job_id, kind, label)` 去重，display_order 累加）、`target_jobs.summary` / `fit_summary` / `analysis_status='ready'` 同事务更新、outbox `target.parsed` payload 仅含 `targetJobId / userId / analysisStatus / requirementCount / coreThemes`
- [x] 4.4 实现失败路径与 retryable 语义；验证: tests 覆盖 A3 / source 错误到 retryable=true / false 的映射矩阵（`AI_PROVIDER_TIMEOUT` / `AI_FALLBACK_EXHAUSTED` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` retryable；`AI_OUTPUT_INVALID` / `AI_UNSUPPORTED_CAPABILITY` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID` / `TARGET_IMPORT_SOURCE_INVALID` non-retryable）、事务内 `target.analysis.failed` outbox 写入、失败不删除 `target_job_sources` 记录
- [x] 4.5 占位 `source_refresh` 触发入口；验证: tests 覆盖 `target.parsed` 触发后写入 internal-only `async_jobs(job_type=source_refresh)`（B3 dotted task `source.refresh`）、payload 不含 source URL 完整路径、drainer 端空 handler 标记 `target_job_sources.freshness_status='stale'`，并标注待 future plan 接管
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
- [x] 6.6 Active-scope 负向搜索通过；验证: `backend/internal/targetjob`、`backend/cmd/api`、`docs/spec/backend-targetjob`、`test/scenarios/e2e/p0-010..013-*` active code/docs 不引入 `mistake.*` / `growth.*` / 独立 `voice` route / 独立 `report` 一级 route / 旧 `feature_key` 别名（如 `jd.parse` / `target.parse`）/ embedding / rerank capability / 独立 worker 进程前置依赖 / 旧 `interview_round` 独立模块；允许命中仅限本 gate 文本、negative test token、test fake 未实现方法和 handler service-not-configured guard

## Phase 7: L2 remediation and reopened BDD gate

- [x] 7.1 Remediation: URL fetch dial 路径绑定已校验 public IP，覆盖 DNS rebinding / TOCTOU；验证: `cd backend && go test ./internal/targetjob/urlfetch -run 'TestFetch_RejectsDNSRebindOnDial|TestDialContextRejectsPrivateResolvedAddress' -count=1`
- [x] 7.2 Remediation: `updateTargetJob` 状态机校验移入 store 事务并锁定 target row；验证: `cd backend && go test ./internal/targetjob -run 'TestSQLStore_UpdateTargetJobLifecycle_IdempotentRejectsStaleStatusTransition|TestService_UpdateTargetJob_DelegatesStatusTransitionValidationToStore' -count=1`
- [x] 7.3 Remediation: BDD 场景脚本与索引不再把包级 focused tests 标记为真实场景通过；验证: p0-010..013 `trigger.sh` 执行 `cmd/api` HTTP scenario tests，`verify.sh` 输出 `method=cmd-api-http` / `validBddEvidence=true`，`test/scenarios/e2e/INDEX.md` 标记四个场景为 Ready
- [x] 7.4 Remediation: `cmd/api` 接入 `target_import` / `source_refresh` drainer、`ParseExecutor`、A3 runtime client、F3 static contract bridge 与 `urlfetch`；验证: `cd backend && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresDrainerAndAIClient|TestBuildAPIHandlerMountsTargetJobRoutesBehindSessionMiddleware' -count=1`
- [x] 7.5 Remediation: TargetJob handler 错误响应统一 generated `ApiErrorResponse`，删除 legacy `{"errors":[...]}` envelope；验证: `cd backend && go test ./internal/targetjob -run 'TestHandler_ImportTargetJob_RejectsMissingIdempotencyKey|TestHandler_ImportTargetJob_RejectsMissingSession|TestHandler_ErrorResponsesUseGeneratedEnvelope' -count=1` 通过，断言 `error.code/message/requestId/retryable` 存在且无 `errors` key
- [x] 7.6 Remediation: `listTargetJobs` 设置实际生效的 `pageInfo.pageSize`；验证: `cd backend && go test ./internal/targetjob -run 'TestService_ListTargetJobs_PassesFiltersAndShapesPaginated|TestService_ListTargetJobs_PageInfoReportsEffectivePageSize' -count=1` 通过，覆盖默认 page size、clamp 到 100、非法值兜底与空列表 envelope
- [x] 7.7 Remediation: `ParseExecutor` 从 F3 contract bridge / prompt metadata 组装 A3 payload，`APP_ENV=test` 的 `cmd/api` runtime 注入只作用于 `target.import.parse` 的 deterministic JSON parse fixture client；验证: `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_UsesPromptMessagesFromRegistryResolution|TestDeterministicParseAIClient_OnlyInterceptsTargetImportParse' -count=1 && go test ./cmd/api -run 'TestBuildTargetJobRuntimeWiresDrainerAndAIClient' -count=1` 通过
- [x] 7.8 Remediation: AI parse 输出无有效 requirement 或含非法 kind / label / evidence level 时走 `AI_OUTPUT_INVALID`；验证: `cd backend && go test ./internal/targetjob -run 'TestParseExecutor_AIOutputInvalid_WhenRequirementsAreSemanticallyInvalid' -count=1` 通过，覆盖 all-invalid requirements、空 label、非法 kind、非法 evidence level 均写 `target.analysis.failed` 且不把 TargetJob 标记为 `ready`
- [x] 7.9 Remediation: BDD handoff 文案与场景状态对齐，包级 focused tests 不再被记录为真实 BDD PASS；验证: `backend/internal/targetjob/doc.go`、`test/scenarios/e2e/INDEX.md` 与 p0-010..013 README / verify outputs 标注 `cmd-api-http` / `validBddEvidence=true`，`bash -n test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/trigger.sh test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/scripts/verify.sh test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/trigger.sh test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/scripts/verify.sh test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/trigger.sh test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/scripts/verify.sh test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/trigger.sh test/scenarios/e2e/p0-013-targetjob-manual-form-ready/scripts/verify.sh` 通过，主 checklist 6.1-6.4 已用 `.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.010..013/result.json` 证据闭合
- [x] 7.10 将 E2E.P0.010 / 011 / 012 / 013 迁移为 auth -> HTTP API -> cmd/api drainer 的真实场景；验证: 新增 `backend/cmd/api` HTTP scenario harness，p0-010..013 `trigger.sh` 执行 `go test -v ./cmd/api -run 'TestE2EP0010HTTPTextImportParseReady|TestE2EP0011HTTPURLImportFetchAndParse|TestE2EP0012HTTPParseFailureRetryableAndNonRetryable|TestE2EP0013HTTPManualFormReady'` 对应场景，`verify.sh` 输出 `status=passed` / `method=cmd-api-http` / `validBddEvidence=true`，证据位于 `.test-output/runs/targetjob-http-20260508/e2e/E2E.P0.010..013/result.json`
