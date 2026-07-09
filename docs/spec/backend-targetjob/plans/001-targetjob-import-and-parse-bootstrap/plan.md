# TargetJob Import and Parse Bootstrap

> **版本**: 1.12
> **状态**: active
> **更新日期**: 2026-07-09

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

落地 P0 后端 TargetJob 域：把 [B2 OpenAPI](../../../openapi-v1-contract/spec.md) 已定义的 `importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` / `archiveTargetJob` 接到真实 handler / service / store，把 4 类导入源（`url` / `manual_text` / `file` / `manual_form`）写入 [B4 baseline](../../../db-migrations-baseline/spec.md) 的 `target_jobs` / `target_job_requirements` / `target_job_sources` 三张表，把 [B3](../../../event-and-outbox-contract/spec.md) 的 `target_import` 异步 job 通过 backend-internal goroutine drainer drain，并在事务内调用 [F3 RegistryClient](../../../prompt-rubric-registry/spec.md) `Resolve("target.import.parse", language)` + [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 完成 JD 解析。解析成功发出 `target.parsed` 并保留 ready TargetJob；解析失败只发出 `target.analysis.failed` 诊断事件并删除失败 TargetJob 资产，不允许失败 JD 作为可继续规划持久化；用户删除面试列表卡片时通过 `archiveTargetJob` 持久软归档，避免刷新后回灌。

本次 v1.10 原地修订修复解析失败准入缺口：`target_import` 失败事务必须删除 `target_jobs`，级联清理 source / requirements，保留 async job/outbox 失败证据；`listTargetJobs` / `getTargetJob` 不得再返回 `analysisStatus=failed` 的脏规划。
本次 v1.11 原地修订新增 TargetJob 持久归档合同：`POST /targets/{targetJobId}/archive` 与 generated `archiveTargetJob` 必须设置 `status='archived'` 和 `deleted_at`，并让 `listTargetJobs` / `getTargetJob` 继续通过 `deleted_at is null` 隐藏已归档卡片。
本次 v1.12 原地修订补齐归档与异步解析的边界：若归档发生在 `target_import` job 排队或重试期间，parse worker 读取到 TargetJob 已不可见时必须终结该 async job，不得再走失败清理并制造 retry storm。

本计划闭环后，[`frontend-home-job-picks-and-parse`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 的 parse 屏可从 mock fixture 切到真实 backend，剩余 Job Picks 推荐与 practice plan 创建归后续 plan / 后续 subspec。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 把 `Home / Job Picks / Parse` workstream 的 backend owner 锁定为 `backend-targetjob`，但当前仓库尚无任何 backend handler / service 实现：[`backend/internal/`](../../../../../backend/internal) 只覆盖 auth / migrations / 共享脚手架。OpenAPI v1 已定义 5 个 TargetJob operation，B4 baseline 已包含三张表，B3 已冻结 `target.import.requested` / `target.parsed` / `target.analysis.failed` 与 `target_import` job，F3 已锁定 `target.import.parse` feature_key。本计划是把这些已就位的契约缝合成一个可上线的后端域。

[backend-auth/001-email-code-session-bootstrap](../../../backend-auth/plans/001-email-code-session-bootstrap/plan.md) 已经为本计划建立了重要先例：

- 在 `cmd/api` 进程内用 backend-internal goroutine drainer 完成异步派发，不引入独立 worker。
- 通过 generated `BuildEmailDispatchPayload` 强制 redact PII；本计划同样要使用 generated outbox payload helper 守 `target.*` 事件红线。
- 隐私 / observability 通过 grep / privacy test 验证；本计划沿用同款 gate。

[ai-provider-and-model-routing/004](../../../ai-provider-and-model-routing/spec.md) 已落地 provider registry / capability profile catalog / observability decorator / fail-closed 语义，本计划只消费已暴露的 `AIClient` 接口与 `target.import.default` profile，不修改 A3 真理源。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior` + `backend` + `contract` + `async-pipeline`。
- **TDD 策略**: 通过 `/implement backend-targetjob/001-targetjob-import-and-parse-bootstrap backend` → `/tdd` 执行；Phase 0 先以 contract tests / codegen drift tests 修订 B1/B2/B3/F1 owner truth source，后续每个 backend checklist item 先写 focused Go test（handler contract test、service test、store test、URL fetcher SSRF test、drainer test、outbox emit test、privacy grep test）再实现最小代码；checklist 的 `验证:` 后必须列出测试断言。
- **BDD 策略**: 需要 BDD。本计划引入用户可见 API 行为（5 个 operation + 异步解析观测 + manual_form 同步兜底 + workspace 归档删除），必须维护 `bdd-plan.md` / `bdd-checklist.md`，并在主 checklist 用 `BDD-Gate:` 引用 `E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013` / `E2E.P0.018`。2026-05-08 L2 code review 重新打开该 gate 后，p0-010..013 已迁移为 `auth -> HTTP API -> cmd/api runtime drainer -> F3 prompt registry + A3 AI client + urlfetch -> store/outbox` 的 HTTP scenario harness，verify 输出 `method=cmd-api-http` / `validBddEvidence=true`。
- **替代验证 gate**: 不适用；BDD gate 是行为入口。补充 gate 包括 focused Go tests、OpenAPI generated contract tests、B1/B2 codegen drift check（`make codegen-conventions && make codegen-openapi && make codegen-check`）、fixtures validation（`make validate-fixtures`）、events.yaml / jobs.yaml drift check（`make codegen-events && make lint-events`）、`migrations_lint`、F1 metric registry tests、privacy grep（`raw_jd_text` / `Authorization` / `prompt body`）、URL fetch SSRF unit test 矩阵、drainer drain / shutdown test、F3 Resolve fail-closed test、`make docs-check`、`make lint-config`。

### 3.1 Operation Matrix

本计划走 [docs/development.md §2.3](../../../../development.md#23-backend-first-path) backend-first path。下面矩阵是 `/implement` / `/tdd` 前的 contract handoff proof，防止 OpenAPI fixture、frontend mock、backend handler 与 BDD 状态混淆。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `importTargetJob` | `openapi/fixtures/TargetJobs/importTargetJob.json` (`default`, Phase 0 add `manual-text-primary`, `manual-form-ready-terminal-job`, `url-invalid-source`, `url-source-unavailable`) | `frontend-home-job-picks-and-parse` home / parse flow via generated client, including selected `resumeId` | `backend/internal/api` generated `ServerInterface` adapter → `backend/internal/targetjob` handler / service / store | `url/manual_text/file`: `target_jobs.resume_id` + `target_jobs` + `target_job_sources` + `async_jobs(target_import)` + outbox `target.import.requested`（`manual_text` event `sourceType=text`）；`manual_form`: 同步写 `target_jobs.resume_id` + `target_jobs` + `target_job_requirements` 草稿，返回 terminal `Job(type=target_import,status=succeeded)`，不派发 runner job / import requested event | `target.import.default` only for async parse sources; `manual_form` is `none` | `E2E.P0.010` manual_text primary + idempotency；`E2E.P0.011` URL source；`E2E.P0.012` parse failure；`E2E.P0.013` manual_form ready |
| `listTargetJobs` | `openapi/fixtures/TargetJobs/listTargetJobs.json` (`default`, `prototype-baseline`) | TargetJob list / workspace pickers via generated client | `backend/internal/targetjob` list handler / store cursor query | read `target_jobs.resume_id` + `target_jobs` + optional latest ready `practice_plans.currentPracticePlanId`; soft-deleted rows filtered | none | `E2E.P0.010` verifies imported job is visible in list; `E2E.P0.018` verifies workspace list re-entry carries resume binding |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json` (`default`, `prototype-baseline`) | parse confirmation / workspace detail via generated client | `backend/internal/targetjob` get handler / store detail query | read `target_jobs.resume_id` + `target_jobs` + `target_job_requirements` + summary / fit JSON + optional latest ready practice plan; user-scoped 404 on missing / cross-user / soft-delete | none after parse completion; provenance is persisted output | `E2E.P0.010`, `E2E.P0.012`, `E2E.P0.018`; focused handler tests cover cross-user 404, soft-delete and resume binding recovery |
| `updateTargetJob` | `openapi/fixtures/TargetJobs/updateTargetJob.json` (`default`, Phase 0 add `invalid-state-transition`, `cross-user-hidden-not-found`) | workspace lifecycle / notes edits via generated client | `backend/internal/targetjob` update handler / idempotency service / store update；实施前状态为 `not-yet-implemented` | update `target_jobs.status` / `location_text` / `notes` / hints scoped by `(user_id, id)` and `(user_id, idempotency_key)` | none | `E2E.P0.010` verifies minimal status / notes update after parse; focused handler tests cover `TARGET_INVALID_STATE_TRANSITION` and cross-user idempotency |
| `archiveTargetJob` | `openapi/fixtures/TargetJobs/archiveTargetJob.json` | Workspace delete icon via generated client | `backend/internal/targetjob` archive handler / idempotency service / store archive; `ParseExecutor` treats post-archive target invisibility as terminal `TARGET_JOB_NOT_FOUND` | update `target_jobs.status='archived'`, `deleted_at=now`, `updated_at=now` scoped by `(user_id, id)` and `(user_id, idempotency_key)`; list/detail/parse reads keep filtering `deleted_at is null` | none | `E2E.P0.018` verifies workspace delete persists across refresh; focused handler/store tests cover already-archived conflict and cross-user 404; `TestParseExecutor_MissingTargetIsTerminalWithoutFailureCleanup` covers queued/retrying import after archive |

## 4 实施步骤

### Phase 0: Owner contract remediation

#### 0.1 修订 B1/B2 TargetJob API 场景契约

在实现 backend 代码前先修订 B1/B2 可执行契约：B1 `shared/conventions.yaml` 追加 `TARGET_JOB_NOT_FOUND`、`TARGET_IMPORT_SOURCE_INVALID`、`TARGET_IMPORT_SOURCE_UNAVAILABLE`、`TARGET_INVALID_STATE_TRANSITION` 四个错误码并重生成 Go/TS/OpenAPI 错误码；B2 `openapi/openapi.yaml` additive 扩展 `TargetJobRequirement.kind` 到 `must_have` / `nice_to_have` / `hidden_signal` / `interview_focus`，保持 `ImportTargetJobRequest` 四类 source variant，明确 `manual_form` 返回 terminal `Job(status=succeeded)`。TargetJobs fixtures 必须补 `manual-text-primary`、`manual-form-ready-terminal-job`、`url-invalid-source`、`url-source-unavailable`、`cross-user-hidden-not-found`、`invalid-state-transition` scenarios。

#### 0.2 修订 B3 import source event 语义

保持 `target.import.requested.sourceType` 为 B3 event-local `url` / `text` / `file`：B2 `manual_text` 在事件中映射为 `text`，B2 `manual_form` 因不进入异步 runner 而不发 `target.import.requested`。若未来 analytics 需要 exact API source variant，只能新增 optional payload 字段或 eventVersion，不得把 `manual_form` 塞进当前 v1 `sourceType`。

#### 0.3 修订 F1 TargetJob metrics registry

F1 metrics dictionary 必须登记 `target_job_imports_total`、`target_job_parse_duration_seconds`、`target_job_parse_failures_total`，并把有界 `error_code` / `source_type` 纳入 allowed labels；`error_code` 只能来自 B1，`source_type` 只能来自 B2/B3，禁止 URL、target id、user id、prompt version 或自由文本进入 label。

### Phase 1: Storage / config / generated surface boundaries

#### 1.1 锁定 store 接口与 SQL 实现

复用 B4 baseline 的 `target_jobs` / `target_job_requirements` / `target_job_sources` 三张表与现有索引。store 接口必须覆盖：插入 target_job（含初始 `analysis_status='queued'` 或 `'ready'` 和创建时绑定的 `resume_id`）、读取按 `(user_id, id)`、列表按 `(user_id, status, analysis_status, q, cursor, page_size)`、更新生命周期字段、写入 requirements 批量、更新 `analysis_status` + `summary` + `fit_summary`、写入 / 更新 `target_job_sources`，并提供解析成功 / 失败的原子提交方法。所有方法必须按 `user_id` 过滤；越权返回 `sql.ErrNoRows` 等价语义，handler 层映射为 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND`。本次 v1.8 明确允许新增 migration 为 `target_jobs.resume_id` 建立当前产品绑定字段，并要求真实 Postgres gate 覆盖迁移后的 import/list/get 行为。

#### 1.2 锁定 config / secret 边界

从 [A4 secrets/config](../../../secrets-and-config/spec.md) 读取 provider secret / feature flag 等已存在 app-level 配置；URL 抓取 timeout 与 UA 标记由本域代码常量提供。统一出网代理不属于 easyinterview 后端项目配置，若部署环境需要代理，由 nginx / 接入层 / 平台透明处理，本 plan 不要求 A4 提供 app-level proxy key。缺必需 secret 必须 fail-fast。`APP_ENV=test` 允许 stub provider；非测试本地 app run 或未来部署选中真实 `target.import.default` profile 时 A3 / F3 缺 secret 必须 fail-closed，不得静默回退 stub（D-10 / spec C-10）。任何新增 app-level 配置 key 先停止并修订 A4。

#### 1.3 锁定 generated handler / outbox / job surface

使用 [B2 generated `ServerInterface`](../../../openapi-v1-contract/spec.md) 注册 `importTargetJob` / `listTargetJobs` / `getTargetJob` / `updateTargetJob` / `archiveTargetJob` 五个 handler；确认 generated request / response types 与本计划一致，差异时回到 B2 修订。outbox 写入与 `target_import` job 派发必须使用 [B3 generated payload / event helper](../../../event-and-outbox-contract/spec.md)；事件 payload 不得携带 `raw_jd_text` / `source_url` / 文件 URL / prompt / response 明文（PiiBoundary 已锁定）。

### Phase 2: Synchronous TargetJob CRUD

#### 2.1 实现 `importTargetJob` 同步阶段

接收 `ImportTargetJobRequest` 与必填 `Idempotency-Key`，按 `(user_id, idempotency_key)` 去重；新建场景在事务内插入 `target_jobs`（初始 `analysis_status='queued'`，`manual_form` 为 `'ready'`）+ `target_job_sources`（除 `manual_form` 外）。`url` / `manual_text` / `file` 写 outbox `target.import.requested` 事件并派发 `target_import` job；`manual_text` 的 event `sourceType=text`；`manual_form` 不派发 runner job、不发 `target.import.requested` / `target.parsed`，但响应仍返回 generated `TargetJobWithJob` 与 terminal generated `Job`（`jobType=target_import`、`status=succeeded`）。

#### 2.2 实现 `listTargetJobs`

按 `(user_id, status, analysis_status, q, cursor, page_size)` 过滤；q 走 `idx_target_jobs_user_status_updated` / `idx_target_jobs_fts` 索引；`pageSize` clamp 到 `[1,100]`；游标使用 `(updated_at, id)` 复合编码并 base64url；空结果返回 generated `PaginatedTargetJob` 空 envelope。

#### 2.3 实现 `getTargetJob`

按 `(user_id, target_job_id)` 读取；包含 requirements 数组（按 `display_order` 排序）、`summary` / `fitSummary` / `provenance`、`latestReportId`（当前为 nil 占位）、`openQuestionIssueCount`；越权或软删返回 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND`。

#### 2.4 实现 `updateTargetJob`

接收 `UpdateTargetJobRequest` 与 `Idempotency-Key`；按 `(user_id, idempotency_key)` 去重；验证 `status` 状态机合法迁移（`draft → preparing → applied → interviewing → offer | rejected → archived`，可后退至 `archived`），非法迁移返回 B1 `TARGET_INVALID_STATE_TRANSITION`；`locationText` / `notes` / `titleHint` / `companyNameHint` 仅写非空字段。

### Phase 3: Source ingestion (URL / file / manual)

#### 3.1 实现 `manual_text` 与 `manual_form` 写入

`manual_text`：`target_job_sources.source_type='manual_text'`，`snapshot_text=request.rawText`，`raw_jd_text=request.rawText`；不抓取、不调用 AI 之外的 IO。`manual_form`：直接写入 `target_jobs.title` / `company_name` / `raw_jd_text=rawDescription`，`analysis_status='ready'`，并把 `rawDescription` 拆为草稿 `target_job_requirements`（最低 1 条 `must_have`，由后续 plan 增强；本 phase 落最小值即可）。

#### 3.2 实现 `file` 引用

接收 `fileObjectId`；read-side 校验 `file_objects.purpose='target_job_attachment'` 与 `(user_id, file_object_id)` 隔离；缺失或越权返回 B1 `TARGET_JOB_NOT_FOUND`，purpose 不符返回 `TARGET_IMPORT_SOURCE_INVALID`；将 `target_jobs.source_file_object_id` 与 `target_job_sources.file_object_id` 设置为该引用；`raw_jd_text` 暂留空，等异步阶段从文件抓取（接入 [`backend-upload`](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 后由后续 plan 处理；本 plan 下 file 类型解析输入仍可由 file_object metadata + 调用方提供的 `titleHint` / `companyNameHint` + manual_text 兜底，避免阻塞 P0）。

#### 3.3 实现 `url` 抓取守护

实现 `urlfetch` 内部包：scheme 仅 `https`、UA `EasyInterview JD-Crawler/<version> (+https://easyinterview.local/crawler)`、timeout 10s、body cap 1 MiB、不跟随 cross-origin redirect 进入私网、解析后 IP 校验阻止 RFC1918 / 169.254 / 元数据服务；非法 scheme、私网、metadata、超长、空白文本映射 B1 `TARGET_IMPORT_SOURCE_INVALID`，上游超时 / 临时不可达映射 `TARGET_IMPORT_SOURCE_UNAVAILABLE`。`target_job_sources.url` 写 sanitized URL（去除 query secret），`snapshot_text` 写抓取到的 JD 文本摘要 / 正文，不得混放 URL。

### Phase 4: Async parse pipeline

#### 4.1 实现 `target_import` drainer

复用 [backend-auth](../../../backend-auth/spec.md) 同款 backend-internal goroutine drainer 模式：handler 入队后立即返回 202；后台 worker pool 按 `target_import` job 调度，支持 graceful shutdown / drain timeout / pending job retry on restart；并发上限作为代码常量，与 backend-auth 同款记录在包级 doc。drainer 必须有 drain / shutdown / pending-on-restart focused tests。

#### 4.2 调用 F3 Resolve + A3 Complete

业务侧调用 `RegistryClient.Resolve("target.import.parse", targetJob.targetLanguage)` 获取 `(prompt_version, rubric_version, model_profile_name)`；F3 返回 unsupported / disabled profile 必须直接走失败路径（写 `target.analysis.failed.retryable=false`，error code 取自 F3 / B1 共享错误码），不静默回退 stub。随后调用 `AIClient.Complete(model_profile_name, payload)`，payload 必须携带 `feature_key + prompt_version + rubric_version + model_profile_name + language + data_source_version`；调用结果只取结构化字段写入 DB，原始 prompt / response body 不入库 / 不入日志 / 不入 metric label。

#### 4.3 写入解析结果与发出 `target.parsed`

事务内 upsert `target_job_requirements`（按 `(target_job_id, kind, label)` 去重，`display_order` 累加）；更新 `target_jobs.summary` / `fit_summary` / `analysis_status='ready'`；事务内 outbox 写入 `target.parsed` 事件（payload 仅 `targetJobId / userId / analysisStatus / requirementCount / coreThemes`），并写入 `source_refresh` 占位 async job。事务外不再读取 AI response 明文。

#### 4.4 实现失败路径与 retryable 语义

A3 / source 错误映射：`AI_PROVIDER_TIMEOUT` / `AI_FALLBACK_EXHAUSTED` / `TARGET_IMPORT_SOURCE_UNAVAILABLE` → `retryable=true`；`AI_OUTPUT_INVALID` / `AI_UNSUPPORTED_CAPABILITY` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID` / `TARGET_IMPORT_SOURCE_INVALID` → `retryable=false`。失败事务内写入 `target.analysis.failed` outbox 事件（`targetJobId / errorCode / retryable`），随后删除失败 `target_jobs` 行并级联清理 `target_job_sources` / `target_job_requirements`，保证失败 JD 不作为可继续规划或列表资产持久化；用户重试必须重新 import。

#### 4.5 占位 `source_refresh` 触发入口

`target.parsed` 事件触发 internal-only `async_jobs(job_type=source_refresh)` 占位（B3 dotted task name `source.refresh`，但 DB canonical 使用 `source_refresh`）；本 plan 不实现真实抓取刷新，仅保证 job 写入路径可观测、payload 不泄露 source URL 完整路径，并在 drainer 端用空 handler 标记 `freshness_status='stale'` 等待后续 plan 接管。

### Phase 8: JD identity and current-plan binding remediation

- `target.import.parse` AI output must include canonical `title` and `companyName`; parse success persists non-empty values to `target_jobs` in the same success transaction as summary / fitSummary / requirements / outbox.
- `listTargetJobs` / `getTargetJob` expose optional current practice-plan binding projection (`currentPracticePlanId`, `resumeId`) derived from the latest ready `practice_plans` row so frontend plan-list cards can open the bound-resume detail path without synthetic route ids.
- Verified focused backend suites, prompt lint, OpenAPI lint and fixture validation on 2026-07-08.

### Phase 9: TargetJob-level resume binding remediation

- `ImportTargetJobRequest.resumeId` is required for all source variants and must reference a non-archived resume owned by the current user.
- `target_jobs.resume_id` stores the JD-level resume binding selected at Home import time. `listTargetJobs` / `getTargetJob` expose `TargetJob.resumeId` from this column, while `currentPracticePlanId` remains derived from the latest ready `practice_plans` row.
- If a ready practice plan exists, its `resume_id` must match the target job binding for the plan card path; practice plan creation uses the route/context `resumeId` and the backend response continues to echo `PracticePlan.resumeId`.
- Runtime smoke must prove the local DB row created from `importTargetJob` has `target_jobs.resume_id`, and `GET /targets` returns that value before any practice plan row exists.

### Phase 10: Real-provider parse robustness remediation

- `target.import.parse` success requires a non-empty `title`, but `companyName` is a display field that may be absent from valid JDs. Empty `companyName` must coalesce to a language-specific placeholder (`未提供` for `zh-*`, `Unknown company` otherwise) instead of failing the whole parse as `AI_OUTPUT_INVALID`.
- A3 output-schema validation and TargetJob response decoding may normalize one provider deviation: a single complete JSON value wrapped in a markdown code fence. Prose before/after the fence or multiple JSON values remain invalid to preserve strict BUG-0095 behavior.
- `cmd/api` TargetJob runtime must wrap parse AI with A3 observability and write `ai_task_runs` for `jd_parse`, including provider/model/status/validation status. Real-provider parse failures must have task-run evidence for diagnosis.
- Runtime smoke must prove the screenshot class of JD imports reaches `analysisStatus='ready'`, persists fallback `companyName`, renders the parse route without `JD 解析失败`, and records a successful `ai_task_runs` row.

### Phase 5: Privacy / observability / idempotency redlines

#### 5.1 隐私 grep / payload negative tests

privacy grep 必须覆盖：log / metric label / audit / outbox payload / async job payload 不含 `raw_jd_text`、`source_url` 完整 URL 或 query 串、文件 object URL、AI prompt body、AI response body、provider secret、`Authorization:` header。负向测试用 sentinel 字符串验证 grep 与 generated payload helper 拒绝违规字段。

#### 5.2 F1 metric registry preflight

实施前先确认 `target_job_imports_total`、`target_job_parse_duration_seconds`、`target_job_parse_failures_total` 已登记到 [F1 baseline](../../../observability-stack/spec.md) metrics 字典，label 仅使用 F1 allowed labels（`service` / `operation` / `job_type` / `source_type` / `language` / `result` / `error_code`）；未登记则停止并修订 F1，不在本域私造 metric。

#### 5.3 Idempotency 跨 user 隔离

`(user_id, idempotency_key)` 唯一索引或等价 store 行为；不同用户即使复用同一 `Idempotency-Key` 也必须返回各自记录，禁止跨用户复用 active `target_import` job；同一用户同 key 重复请求返回同一 `targetJobId` 与同一 active job，DB / outbox 不出现重复 row（沿用 [backend-auth](../../../backend-auth/spec.md) `DELETE /me` 已建立的 user-scoped dedupe 模式）。

### Phase 6: BDD and handoff

#### 6.1 BDD-Gate `E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013`

按 `bdd-plan.md` / `bdd-checklist.md` 在 `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/`、`p0-011-targetjob-url-import-fetch-and-parse/`、`p0-012-targetjob-parse-failure-retryable/`、`p0-013-targetjob-manual-form-ready/` 四个目录落地 setup / trigger / verify / cleanup，并通过 `test/scenarios/e2e/INDEX.md` 接入。

#### 6.2 Handoff 给 frontend-home-job-picks-and-parse

在 `backend/README.md` 或 `backend/internal/targetjob/doc.go` 写明 4 个 operation 的同步 / 异步语义、错误码、idempotency 行为、URL fetch 守护规则、隐私红线、可观测 metric 名与 BDD 入口；前端切换 mock → real 时只需要关注 generated client + B2 fixture parity，不需要进入 backend 内部。

#### 6.3 Active-scope 负向搜索

代码 / 文档 active scope 不得引入：`mistake.*` / `growth.*` / 独立 `voice` route / 独立 `report` 一级 route / 旧 `feature_key` 命名（如 `jd.parse` / `target.parse` 旧别名）/ embedding / rerank capability / 独立 worker 进程依赖 / 旧 `interview_round` 独立模块假设。除本 gate 文本、negative test token、test fake 未实现方法和 handler service-not-configured guard 外，grep 命中即视为 plan 失败。

### Phase 7: L2 remediation and reopened BDD gate

#### 7.1 URL fetch DNS rebinding remediation

`urlfetch` 的实际 dial 路径必须绑定到已校验的 public IP，避免 `checkURLPolicy` 与 `http.Client.Do` 之间发生 DNS rebinding / TOCTOU。验证必须覆盖解析阶段返回 public IP、dial 阶段返回 private / metadata / loopback IP 时请求被拒绝且不发起连接。

#### 7.2 `updateTargetJob` transactional state-machine remediation

`updateTargetJob` 的状态机校验必须进入 store 事务，在持有目标 row lock 后基于当前 DB 状态重新校验；service 层不得用事务外 preflight 结果作为最终迁移依据。验证必须覆盖不同 `Idempotency-Key` 并发或 stale 迁移不能越过当前 DB 状态。

#### 7.3 BDD evidence honesty remediation

`E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013` 的 evidence 必须来自真实 HTTP scenario harness。脚本不得回退为包级 focused tests；verify output 必须保留 `method=cmd-api-http` 与 `validBddEvidence=true`，防止未来 L2 review 把辅助 TDD 证据误读为 BDD 通过。

#### 7.4 `cmd/api` runtime drainer wiring remediation

`cmd/api` 必须在同一进程中组装 `target_import` / `source_refresh` drainer、`ParseExecutor`、A3 AI runtime client、`urlfetch` 与 F3 `target.import.parse` contract bridge。当前 backend truth source 仍未提供可消费的 F3 prompt/rubric runtime package，因此 runtime wiring 可用静态 contract bridge 先闭合启动路径；真实 HTTP BDD 仍必须等 F3 runtime package 或场景注入能力到位后再完成。

#### 7.5 Generated error envelope remediation

TargetJob handler 必须返回 B1/B2 generated `ApiErrorResponse` envelope：所有错误响应统一为 `{ "error": { "code", "message", "requestId", "retryable" } }`，不得再返回 non-current `{"errors":[...]}`。focused tests 必须覆盖 import / list / get / update 的鉴权、idempotency、not-found、invalid-source、source-unavailable 与 invalid-transition 错误，且断言没有 non-current `errors` key。

#### 7.6 List pagination envelope remediation

`listTargetJobs` 返回的 `pageInfo.pageSize` 必须等于实际生效的 clamp 后 page size，满足 generated `PageInfo` 契约与 TargetJobs fixture；默认值、过大值、非法值与空列表都必须有 focused service / handler tests。

#### 7.7 F3 prompt ownership and test-runtime parse fixture remediation

`ParseExecutor` 不得直接硬编码 prompt body；必须从 `RegistryClient.Resolve("target.import.parse", language)` 返回的 contract bridge / prompt metadata 组装 A3 `CompletePayload`，并在 `APP_ENV=test` 的 `cmd/api` runtime 中提供只作用于 `target.import.parse` 的 deterministic JSON parse fixture client，以便 HTTP drainer 成功路径可以不依赖真实 provider。该 fixture 只能在 test env 选中，不得影响 dev / staging / prod 的 fail-closed 语义。

#### 7.8 AI output validation remediation

AI parse 输出必须在写 DB 前严格校验：无有效 requirement、非法 requirement kind、空 label、非法 evidence level 或无法转换为 B4 / B2 契约时必须走 `AI_OUTPUT_INVALID` 失败路径，不得以 `requirementCount=0` 标记 `ready` 成功。

#### 7.9 BDD handoff evidence alignment remediation

`backend/internal/targetjob/doc.go`、`test/scenarios/e2e/INDEX.md` 与 `p0-010..013` 场景脚本必须一致标注当前 HTTP scenario harness 入口；包级 `go test` 结果可以作为 TDD 辅助证据，但不得替代 `cmd/api` HTTP BDD PASS。真实 BDD evidence 必须可追溯到 p0-010..013 `result.json`，并保持 `validBddEvidence=true`。

#### 7.10 HTTP scenario migration remediation

`E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013` 必须通过 `cmd/api` HTTP scenario harness 执行，覆盖 auth middleware、HTTP API、TargetJob handler/service、in-process drainer 与 parse executor。验证输出必须保留 `status=passed` / `method=cmd-api-http` / `validBddEvidence=true`，不得退化成包级 focused test 代理证据。

#### 7.11 Retired `profile_id` schema-drift remediation

TargetJob store / service / handler / drainer 所有 active SQL 必须与当前 B4 `target_jobs` 表结构一致，不得继续 select / insert / scan 已退役 profile 模块的 `profile_id` 列。验证必须覆盖 sqlmock 列集合与真实 Postgres integration gate；解析失败资产由 Phase 11 删除，`GET /targets/{id}` 应返回 404 + `TARGET_JOB_NOT_FOUND`，而不是暴露 failed TargetJob 或旧列漂移 500。

#### 7.12 Required integration gate skip-proof remediation

BUG-0142 的真实 Postgres gate 是强制 schema-drift 防线，不得在缺少 `DATABASE_URL`、DB 不可达或 focused test 未执行时以 `SKIP` / `no tests to run` 形式整体 PASS。该 gate 已随 Phase 11 准入语义改为 `TestSQLStoreIntegration_CompleteParseFailureDeletesTargetAndSources`：既要 fail fast 暴露环境缺口，也要在真实 DB 下证明失败解析写出 `target.analysis.failed` 后删除 `target_jobs` 并级联 source / requirements，防止后续把可选 integration smoke 当成强制 gate。

### Phase 11: Parse failure admission remediation

#### 11.1 Failure commit deletes failed TargetJob assets

`CompleteParseFailure` 必须在同一事务中写入 `target.analysis.failed` outbox 后删除失败 `target_jobs` 行，依赖现有 FK cascade 删除 `target_job_sources` / `target_job_requirements`。失败证据只保留在 async job/outbox 中，用户重试必须通过新的 import 创建新的 `targetJobId`。验证必须覆盖 SQL store 事务、ParseExecutor 失败路径、cmd/api HTTP 场景中失败后详情 404 / 列表不可见。

#### 11.2 Read-side and frontend defense alignment

后端 read-side 不允许返回失败解析资产；前端 `WorkspacePlanList` / `PlanSwitcherModal` 仍必须请求 `analysisStatus=ready` 并过滤空标题作为防御，避免历史脏数据或非当前环境回灌到面试列表。验证必须覆盖 `listTargetJobs` query filter、failed / blank title negative card、TopBar 空参数回到列表而不是复用 stale context。

### Phase 12: TargetJob archive/delete integration

#### 12.1 B2 OpenAPI and generated archive surface

Additive 修订 `openapi/openapi.yaml`、TargetJobs fixture 与 operation inventory，新增 `POST /targets/{targetJobId}/archive` / `archiveTargetJob`。该 endpoint 必须要求 `Idempotency-Key`，返回 generated `TargetJob`，错误继续使用 generated `ApiErrorResponse`。`make codegen-openapi`、`make lint-openapi`、`make validate-fixtures` 必须证明 generated Go server / TS client / fixture 统一包含 `archiveTargetJob`。

#### 12.2 Store / service / handler persistent archive

`backend/internal/targetjob` 必须新增 user-scoped archive store/service/handler。成功路径写 `target_jobs.status='archived'`、`deleted_at=now`、`updated_at=now`，且 read-side 继续通过 `deleted_at is null` 隐藏归档记录；越权/不存在返回 `TARGET_JOB_NOT_FOUND`，已归档重复请求返回 `TARGET_INVALID_STATE_TRANSITION` conflict，同一用户同一 idempotency key 可返回同一 archived TargetJob。验证必须覆盖 handler signature、缺 `Idempotency-Key`、success、already archived conflict、cross-user 404、list/get 不再返回归档记录。

#### 12.3 Frontend generated-client delete integration

`WorkspacePlanList` 删除图标必须调用 generated `archiveTargetJob`，携带 `Idempotency-Key`，成功后从当前列表移除卡片；失败时不执行导航、不乐观删除并显示可恢复错误。Home 最近模拟面试复用卡片主体但不展示删除按钮。验证必须覆盖 mock transport / real generated client path，禁止继续把删除实现为纯本地隐藏。

#### 12.4 BDD and screenshot acceptance

`E2E.P0.018` / local browser smoke 必须在 real backend 模式下证明：workspace 卡片点击主体仍进入 `parse`，`立即面试` 仍启动 practice，删除图标调用后卡片从 UI 消失，刷新后仍不回灌。验收产物必须包含浏览器截图和后端 API/DB 证据。

## 5 验收标准

- Phase 0 owner contract gates 通过：B1/B2 codegen drift clean，TargetJobs fixture scenarios schema-valid，B3 sourceType mapping lint clean，F1 TargetJob metrics registry tests 通过。
- 5 个 TargetJob operation 的 handler / service / store focused Go tests 全部通过；URL fetcher SSRF / length / timeout / redirect / metadata-IP 矩阵全部覆盖；drainer drain / shutdown / pending-on-restart tests 通过。
- 异步解析成功 / 失败（retryable / non-retryable）路径事务一致；outbox 在事务内写入 `target.import.requested` / `target.parsed` / `target.analysis.failed`，成功路径写入下游 internal-only `source_refresh` 占位 job；失败路径删除失败 TargetJob 资产，`listTargetJobs` / `getTargetJob` 均不可返回失败解析产物。
- F3 Resolve fail-closed 与 A3 缺 secret fail-closed 路径覆盖；除 `APP_ENV=test` 外不静默回退 stub。
- TargetJob handler 错误响应符合 generated `ApiErrorResponse`，`listTargetJobs.pageInfo.pageSize` 符合 B1/B2 envelope，AI parse 输出无有效 requirement 或非法字段时走 `AI_OUTPUT_INVALID`。
- privacy grep 0 命中 `raw_jd_text` / `source_url` 完整 URL / 文件 URL / prompt / response / `Authorization:` 等敏感模式。
- F1 metric registry preflight 通过；`make codegen-events` / `make codegen-conventions` / `make codegen-openapi` / `make validate-fixtures` / `make migrations_lint` / `make lint-config` / `make lint-events` / `make docs-check` 全绿。
- BDD-Gate `E2E.P0.010` / `E2E.P0.011` / `E2E.P0.012` / `E2E.P0.013` 全部通过，verify 输出可追溯证据；包级 `go test` 代理证据不得作为该 gate 的完成依据。
- TargetJob active SQL 与当前 B4 schema 对齐，`rg "profile_id|ProfileID|profileID" backend/internal/targetjob ...` 无命中，真实 Postgres integration gate 覆盖 ready TargetJob 详情与失败后 404 行为，缺 DB 时 fail fast 而不是 `SKIP` 假绿，本地 host-run backend 上解析失败后的 `GET /targets/{id}` 不再返回可见 failed 资产。
- TargetJob import/list/get persist and expose the JD-level resume binding: `ImportTargetJobRequest.resumeId` is required, `target_jobs.resume_id` is non-null for created rows, and list/detail responses keep `TargetJob.resumeId` present even when no `practice_plans` row exists.
- TargetJob archive persists user delete from workspace: `archiveTargetJob` sets `status='archived'` and `deleted_at`, list/detail hide the row after refresh, repeated archive is conflict-scoped, and frontend delete uses generated client rather than local-only hiding.
- Valid JDs that omit company names parse successfully with a language-specific fallback company display value, strict fenced-JSON-only normalization, `ai_task_runs` evidence, and real browser proof that `/parse` no longer renders `JD 解析失败`.
- Active-scope 负向搜索 0 命中已丢弃模块 / route / capability。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| F3 `target.import.parse` baseline prompt / rubric 尚未由 F3 001 plan 落地 | 业务侧只调用 `RegistryClient.Resolve` 抽象，不在 `ParseExecutor` hardcode prompt；`APP_ENV=test` 仅通过 deterministic JSON parse fixture 闭合场景成功路径；dev / staging / prod 中 F3 未 ready 时 `Resolve` 返回 disabled / unsupported，按 D-10 走失败路径而不是绕过 F3 |
| 独立 worker / Asynq 还未落地 | 沿用 backend-auth 同款 in-process drainer + B3 payload 红线；`backend-async-runner` 未来替换时 0 改动 outbox / payload helper |
| URL fetch 引发 SSRF / 内网泄露 | Phase 3.3 SSRF 测试矩阵覆盖私网 / 链路本地 / 元数据 / redirect / oversize；UA 与 timeout 显式可测试 |
| AI prompt / response 明文进入日志 / 事件 | Phase 5.1 privacy grep + generated outbox payload helper + observability decorator hash-only 摘要 |
| Idempotency 被跨用户复用 | Phase 5.3 store-level user-scoped dedupe + handler 层 user_id 过滤 + 越权返回 HTTP 404 + B1 `TARGET_JOB_NOT_FOUND` |
| 旧 `feature_key` / 旧 voice / 旧 mistake 模块在 review 中悄悄回潮 | Phase 6.3 active-scope 负向搜索 + plan-code-review L2 检查 |
| `manual_form` 草稿 requirements 质量低 | 仅作为 P0 兜底；后续若产品要求细化，由独立 plan 处理，不在本 plan 引入 AI 重解析 |
| BDD 场景未来可能回退成只跑包级 focused tests，导致 evidence 再次被误读 | Phase 7.3 / 7.9 / 7.10 已把 p0-010..013 固定为 `cmd/api` HTTP scenario harness；verify output 必须保留 `method=cmd-api-http` / `validBddEvidence=true`，并由 6.1-6.4 BDD gate 注释记录 result.json 证据 |
