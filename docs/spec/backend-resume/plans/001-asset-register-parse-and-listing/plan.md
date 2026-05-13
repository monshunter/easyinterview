# Backend Resume Asset Register Parse and Listing

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-05-13

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [backend-resume spec](../../spec.md) §6 C-1..C-8 + C-13 落到 backend Go handler + store + AI 编排：

- 实现 `POST /api/v1/resumes` (registerResume) handler，含 sourceType 三路分支（`upload` / `paste` / `guided`）+ IK + cross-user 隔离 + 调用 [backend-upload `RegisterFileObject`](../../../backend-upload/spec.md) internal API 校验 `purpose=resume`、object exists 与实际 size；
- 实现 `GET /api/v1/resumes/{resumeAssetId}` (getResume) handler，cross-user 返回 404；
- 实现 `GET /api/v1/resumes` (listResumes) handler，cursor pagination + `updated_at DESC, id DESC` 唯一稳定序；**直接解除 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) Phase 3.3 `listResumes` disabled-list 阻塞**；
- 实现 `resume_assets` store layer：`CreateWithParseJob(pending + async_jobs resume_parse)` / `MarkParsing` / `MarkReady(parsedSummary, parsedTextSnapshot)` / `MarkFailed(errorCode)` / `Get` / `List(cursor, pageSize)` / `DeleteForUser`；
- 实现 `resume.parse` async job handler（按 backend-targetjob 同款 `cmd/api` in-process drainer 注册，不引入独立 worker）：通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md) → 解析 JSON parse draft → 写 `resume_assets` + outbox `resume.parse.completed`；
- 接 [B3 events `resume.parse.completed`](../../../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)：只有最终 ready 成功路径通过 outbox 写入 envelope 字段集（`resumeAssetId / userId / parseStatus`）+ PII 边界（不含 raw text / guided answers / parsed_summary）；失败路径不发 completed event；
- 在 `cmd/api` 挂载 `registerResume` / `getResume` / `listResumes` route，验证 session middleware、IK middleware、path params 与 in-process `resume_parse` drainer wiring 都走真实 runtime；
- 明确本 plan 只落地 `ResumeAsset` source 登记、解析草稿与列表读取，不在用户 Preview Confirm 前创建正式 `structured_master` `ResumeVersion`；保存 v1 与版本读写由 backend-resume/002 承接；
- 通过 spec §6 C-1..C-8 + C-13 验收 + 新增 E2E.P0.034 / E2E.P0.035 两个 BDD 场景；
- 不实现 versions / suggestions / tailor / branch / export 流程（归 plan 002 / 003）；真实 PDF 导出按 spec D-6 的 P0 `501 RESUME_EXPORT_NOT_AVAILABLE` / P1 plan 003 处理，本 plan 不实现。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 `backend-resume` (C7) 为 Resume Workshop 阶段 1 第 2 个 subspec：必须在 `frontend-resume-workshop` mock-first 路径切真前完成。本 plan 是 backend-resume 第一批 plan，承担 P0 用户路径 "登记原始简历 source → 等待解析 → 看到 source 列表 / parse 状态" 的 backend 端到端。`docs/ui-design/resume-onboarding.md` 要求的 "Preview Confirm → 保存 v1 主版本" 不属于本 plan，必须在后续 backend-resume/002 + frontend-resume-workshop/002 中闭合，避免未确认解析草稿成为正式版本。

`listResumes` operation 同时是 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) Phase 3.3 ResumePickerModal "disabled-list 模式" 的解锁前置：本 plan 落地后 workspace 001 owner 可启动原地修订（spec 1.2 → 1.3 / plan checklist active-list 改造），不创建 sibling。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 register / get handler skeleton；Phase 2 起来就有 store layer；Phase 3 起来就有 resume.parse async job + AIClient 集成；Phase 4 起来就有 list + pagination + cross-user 隔离；Phase 5 收口 + BDD + 解锁 workspace 001。

执行本 plan 前必须确认：

- [B2 D-18](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) Phase 1-5 已完成（B1 vocabulary、OpenAPI schema/operation、fixtures、inventory lint、generated server/client artifacts 全部就位；`registerResume` sourceType 扩展与 `listResumes` fixtures 可被真实 handler 字节比对）。
- [B3 D-14](../../../event-and-outbox-contract/plans/002-resume-tailor-mode-drift-fix/plan.md) Phase 1-4 已完成（ResumeTailorMode enum、baseline manifest、generated 类型、negative grep 与 B3 spec 描述全部对齐）；本 plan 直接消费 `resume.parse.completed` envelope，不消费 `resume.tailor.completed`，但 events drift gate 必须 PASS。
- [B4 002 resume_versions / resume_version_suggestions / resume_assets 字段补充](../../../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) 已完成（resume_assets 已有 `source_type` / `original_text` / `guided_answers` / `parsed_text_snapshot` 字段）。
- [backend-upload/001](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) 是完成条件（createUploadPresign + Register internal API 可用）。截至 2026-05-13 backend-upload/001 completed，`createUploadPresign` handler、`RegisterFileObject` internal service、privacy delete baseline、Register-time object `Stat` + actual size mismatch rejection、live roundtrip no-op/skip guard 均已可用；backend-resume/001 必须消费这个当前契约，upload path 不得只检查 fileObject row 存在。
- [F3 001 baseline](../../../prompt-rubric-registry/plans/001-baseline/plan.md) 已 ready（`resume.parse` feature_key + prompt / rubric / model profile 就位）。
- [A3 003](../../../ai-provider-and-model-routing/plans/003-provider-registry-and-capability-profiles/plan.md) 已 ready（AIClient + provider registry + Capability Model Profile）。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior + contract`。本 plan 实现 backend handler / store / async job / AI 调用；用户可见 HTTP API 行为。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. handler unit test：sourceType 三路参数校验 + IK + 422 / 404 / 跨用户隔离；
  2. store integration test：CRUD + parse_status 状态机 + cross-user 隔离 + cursor pagination 边界；
  3. resume.parse job unit test（stub AIClient provider）：成功路径 / 解析 JSON 失败 / AI provider timeout retryable / output_invalid;
  4. outbox event unit test：envelope 字段集 + PII 红线（不含 raw text）；
  5. listResumes integration test：≥ 25 行 + cursor 第二页 + `hasMore=false` + cross-user 不可见；
  6. `cmd/api` route/runtime test：session middleware、IK middleware、route path params、resume_parse in-process drainer wiring 与 shutdown。
  执行入口：`/implement backend-resume/001-asset-register-parse-and-listing` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.034 register-and-list + E2E.P0.035 parse-async-job-lifecycle。详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/resume/handler/... -run TestRegisterSourceType -count=1`
  - `cd backend && go test ./internal/resume/store/... -tags=integration -count=1`
  - `cd backend && go test ./internal/resume/jobs/... -run TestResumeParseJob -count=1`（stub AIClient）
  - `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseDrainerHTTPScenario' -count=1`
  - smoke：`curl -X POST /api/v1/resumes` 与 mock-server fixture 字节比对
  - grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume drainer/outbox payload tests（C-13 negative）
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `registerResume` | `openapi/fixtures/Resumes/registerResume.json` `default` / `paste-text` / `guided-answers`; validation / IK mismatch cases are handler tests unless B2 adds explicit error fixtures | `frontend-resume-workshop/002-create-flow-and-onboarding` (future), backend scenario harness in E2E.P0.034 | `backend/internal/resume/handler/register.go` real handler + `cmd/api` `POST /api/v1/resumes` route with session + IK middleware | `resume_assets` + `file_objects` reference for upload + `async_jobs` resume_parse in the same transaction; no `structured_master` `resume_versions` row before Preview Confirm | `resume.parse.default` through A3/F3 stub in tests | E2E.P0.034 + E2E.P0.035 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found` | `frontend-resume-workshop` adapter/detail flows (future real switch) | `backend/internal/resume/handler/get.go` real handler + `cmd/api` `GET /api/v1/resumes/{resumeAssetId}` route | `resume_assets` | none | E2E.P0.034 |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated` | `frontend-resume-workshop/001` list view and `frontend-workspace-and-practice` ResumePicker unblock | `backend/internal/resume/handler/list.go` real handler + `cmd/api` `GET /api/v1/resumes` route | `resume_assets` cursor pagination | none | E2E.P0.034 |

## 4 实施步骤

### Phase 1: register / get handler skeleton + sourceType 三路

#### 1.1 实现 `internal/resume/handler/register.go`
- 实现 generated server interface `RegisterResume`
- sourceType 三路校验（upload 必带 fileObjectId / paste 必带 rawText / guided 必带 guidedAnswers）
- upload 路径：调 [backend-upload `RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)`](../../../backend-upload/spec.md) internal API；该调用必须以对象存储 `Stat` 证明 object exists 且实际 size 与 `file_objects.byte_size` 一致后，才允许把 `resume_assets.file_object_id` 写入本 subject store
- IK + 24h TTL（B1 idempotency 工具）
- 在同一事务内创建 `resume_assets` queued row + `async_jobs(job_type=resume_parse, resource_type=resume_asset)` row；返回 202 + `ResumeAssetWithJob{resumeAssetId, job(jobType=resume_parse, status=queued)}`，与 [B2 fixture `registerResume.json`](../../../mock-contract-suite/spec.md) `default` / `paste-text` / `guided-answers` 三个 scenario 字节一致

#### 1.2 实现 `internal/resume/handler/get.go`
- 实现 generated server interface `GetResume`
- cross-user 返回 404（不暴露存在）
- 返回 `ResumeAsset` 字段（按 B2 schema）

#### 1.3 unit test
- `register_test.go`: 三 sourceType + IK replay + IK mismatch + 422 + 跨用户
- `get_test.go`: 200 / 404 cross-user / 404 not exist

### Phase 2: resume_assets store + state machine

#### 2.1 实现 `internal/resume/store/assets.go`
- Repository：`CreateWithParseJob / Get / List(cursor, pageSize) / MarkParsing / MarkReady(parsedSummary, parsedTextSnapshot) / MarkFailed / DeleteForUser`
- `CreateWithParseJob` 必须以事务提交 `resume_assets` 与 `async_jobs`，并支持 user-scoped IK replay 返回首次 `resumeAssetId` / `job`；outbox 或 job 写入失败时不得留下 orphan asset
- parse_status state machine：`queued → processing → ready | failed`
- cursor pagination：按 `updated_at DESC, id DESC` 唯一稳定序

#### 2.2 integration test
- `assets_integration_test.go`：CRUD + state transition + cross-user isolation + cursor 边界（empty / single page / multiple pages / `hasMore=false`）

### Phase 3: resume.parse async job + AIClient 集成

#### 3.1 实现 `internal/resume/jobs/parse.go`
- 注册到 `cmd/api` in-process drainer / runtime composition（job_type=resume_parse, dotted=resume.parse）
- 从 resume_assets 读 `file_object_id`（upload）或 `original_text`（paste）或 `guided_answers` jsonb（guided）作为 prompt input
- 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md)
- 解析 LLM JSON 输出 → 写 `parsed_summary` + `parsed_text_snapshot` + `parse_status='ready'`
- 用户 Preview Confirm 前不得创建正式 `structured_master` `resume_versions` 行；parse output 只是草稿，保存 v1 由 backend-resume/002 承接
- 失败路径：写 `parse_status='failed'` + `error_code`；retryability 由 `async_jobs` attempt / retry metadata 表达，不向 `resume_assets.parse_status` 私加 `failed_retryable`
- 写入 `ai_task_runs` typed columns（model_profile_name / version / prompt_version / rubric_version 等）

#### 3.2 outbox event `resume.parse.completed`
- envelope 字段集（[B3 §3.1.4](../../../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）：`resumeAssetId / userId / parseStatus`
- 只在最终 `parse_status='ready'` 时写入；AI output invalid / provider timeout / retryable exhausted 等失败路径不发 `resume.parse.completed`
- PII 边界：不含 raw text / guided answers / parsed_summary

#### 3.3 resume_parse in-process drainer wiring
- 沿用 [backend-targetjob](../../../backend-targetjob/spec.md) 的 in-process drainer 口径：`cmd/api` 进程内 claim `async_jobs(job_type=resume_parse)` 并调用 `backend/internal/resume/jobs/parse.go`
- 提供 `RunOnce` 或等价 deterministic stepping，方便 BDD / `cmd/api` scenario test 在无 timer race 的情况下验证 queued → ready / failed / retry
- `Start(ctx)` / `Shutdown(ctx)` 必须随 `cmd/api` lifecycle 管理；不得新增独立 worker binary、`WORKER_*` config 或 `backend-async-runtime` 旧 shorthand

#### 3.4 unit test
- `parse_test.go`（stub AIClient）：成功 / parse JSON 失败 / AI timeout retryable / output_invalid
- `drainer_test.go`：`Handles(resume_parse)`、`RunOnce` 成功处理、shutdown 不泄漏 goroutine、未知 job type 不被本 drainer claim

### Phase 4: listResumes handler

#### 4.1 实现 `internal/resume/handler/list.go`
- 实现 generated server interface `ListResumes`
- cursor pagination（按 `updated_at DESC, id DESC`）
- 返回 `PaginatedResumeAsset{items, pageInfo{nextCursor, pageSize, hasMore}}`
- cross-user 过滤（仅返回 `user_id = current_user_id`）

#### 4.2 integration test
- `list_integration_test.go`: empty / 25 行 + cursor 第二页 / cross-user 不可见 / cursor invalid 拒绝

#### 4.3 `cmd/api` route/runtime wiring
- 新增 `buildResumeRuntime`（或等价 composition helper），组合 resume store / upload service / prompt registry / AIClient / idempotency middleware / resume_parse drainer
- 挂载：
  - `POST /api/v1/resumes` → session middleware + IK middleware + `RegisterResume`
  - `GET /api/v1/resumes` → session middleware + `ListResumes`
  - `GET /api/v1/resumes/{resumeAssetId}` → session middleware + path param adapter + `GetResume`
- `APP_ENV=test` 可使用 deterministic resume.parse fixture AIClient，但只能拦截 `resume.parse`；真实 dev / Kind / staging / prod 必须走 A3/F3 profile fail-fast 规则
- `cmd/api` tests 断言 route 存在、缺 session 返回 auth error、缺 IK 返回 generated error envelope、同 IK replay 不重复创建 `resume_assets` / `async_jobs` / outbox side effect

### Phase 5: 收口 + BDD + 解锁 workspace 001

#### 5.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `cd backend && go test ./...` PASS
- `cd backend && go test ./internal/resume/...` PASS
- `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseDrainerHTTPScenario' -count=1` PASS
- mock-first 对齐：handler 真实响应与 [B2 fixtures](../../../mock-contract-suite/spec.md) `registerResume.json` (`default` / `paste-text` / `guided-answers`)、`getResume.json` (`default` / `not-found`)、`listResumes.json` (`default` / `empty` / `paginated`) 字节比对 PASS
- grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume drainer/outbox payload tests：0 命中（C-13 negative）
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- `make docs-check` PASS

#### 5.2 BDD 场景验证

- 执行 `test/scenarios/e2e/p0-034-resume-register-and-list/` 全 PASS
- 执行 `test/scenarios/e2e/p0-035-resume-parse-async-job-lifecycle/` 全 PASS
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.034 / P0.035 行

#### 5.3 解锁 workspace 001

通知 [frontend-workspace-and-practice/001-workspace-and-interview-context](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) owner：
- `listResumes` operation 已就位（real backend + fixture parity）
- workspace 001 spec §3.2 待确认事项 1 已具备解除条件
- workspace plan Phase 3.3 disabled-list 模式可启动原地修订到 active-list

本 plan 不直接修订 workspace 文件，只在 5.3 完成 "可解锁" 信号传递；workspace owner 独立完成修订。

#### 5.4 spec / history / INDEX 同步

- backend-resume spec.md 本次 L1 修订后保持 1.1 active；实施完成时再追加完成行
- backend-resume history.md 已记录 2026-05-12 既有 L1 修订；本轮若改变 spec 版本、日期或历史语义，收尾阶段再追加 history 行；plan 001 落地后追加新行（如完成）
- 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-resume` 状态从 "未创建" 改为 "active"（roadmap spec 3.11 → 3.12 if not already）

### Phase 6: L2 remediation - handler errors, parse retry state, and gate hardening

#### 6.1 修复 register/list 业务校验 HTTP 映射
- `RegisterResume` 对 service 层 `ErrValidationFailed` 返回 `422 + VALIDATION_FAILED`，覆盖 backend-upload missing object / size mismatch 不创建 asset 的真实错误面；
- `ListResumes` 对 invalid cursor 返回 `422 + VALIDATION_FAILED`，不得把用户输入错误升级为 500；
- 补 handler unit test，证明错误 envelope 与状态码。

#### 6.2 修复 resume.parse retryable failure 状态语义
- AI timeout / retryable failure 每次失败都写 `resume_assets.parse_status='failed' + error_code`；
- retryable 信息只通过 `async_jobs` retry metadata 表达，不新增 `failed_retryable` parse_status；
- 后续重试允许同一 asset 从 `failed` 重新进入 `processing`，最终 ready 后只发一次 `resume.parse.completed`。

#### 6.3 加固 cmd/api 与 BDD gate
- `cmd/api` 场景补齐 handler validation mapping、invalid cursor、retryable failure → retry success 的可执行断言；
- E2E.P0.034 / E2E.P0.035 trigger/verify 必须检查新增测试名，拒绝只靠 happy path 或测反的 unit test 通过；
- 收口后重新执行 focused Go tests、两个场景脚本、docs/index/diff gate。

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-8 + C-13 全部 PASS（C-3 与 C-4 涉及 resume.parse async 完成 / 失败，必须 stub AIClient 验证）
- `cmd/api` route/runtime gate PASS：session middleware、IK middleware、register/get/list route、resume_parse drainer start/shutdown 与 deterministic `RunOnce` 均有测试证据
- BDD E2E.P0.034 + E2E.P0.035 PASS
- `frontend-workspace-and-practice/001` owner 已收到 `listResumes` 解锁信号
- engineering-roadmap §5.2 `backend-resume` 状态已升级到 active

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: resume.parse AI 输出 JSON 不稳定（schema 漂移） | F3 prompt 设计含 output schema example + [B2 §4.6 GenerationProvenance](../../../openapi-v1-contract/spec.md#46-ai-生成结果-provenance-约束) 强制 + parse 失败 retryable + `output_schema_version` typed column 追踪 |
| R2: `resume_assets.source_type` 字段为 NULL（baseline 数据） | Phase 2 store 实现兼容 NULL（未设置时不强制三路；新写必带）；migration 不回填历史行 |
| R3: cross-user isolation 漏洞导致越权 | handler 层 + store 层双层 `user_id` 过滤；integration test 强制覆盖 cross-user case |
| R4: backend-upload 未完成时本 plan 启动 | Plan 2 背景写明前置依赖；本 plan 不在 backend-upload/001 完成前启动 |
| R5: workspace 001 修订时序 | 本 plan Phase 5.3 仅发信号，不直接修订；workspace owner 在收到信号后启动 plan 1.2 → 1.3 原地修订，不创建 sibling |
| R6: B2/B3/B4 阶段 0 plan 未完成时启动本 plan | 本 plan §2 背景写明 4 个前置依赖（B2 D-18 / B3 D-14 / B4 D-17 / backend-upload 001）；任一未完成时 `/implement` 拒绝启动 |
| R7: handler 包测试通过但真实 API / drainer 未挂载 | Phase 4.3 / checklist 4.4-4.5 强制 `cmd/api` route/runtime wiring；BDD 场景必须输出 `method=cmd-api-http` 或等价 live runtime evidence，并拒绝 no-op / skip 作为 PASS |
