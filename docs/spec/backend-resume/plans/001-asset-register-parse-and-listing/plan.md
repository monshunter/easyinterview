# Backend Resume Register Parse and Listing

> **版本**: 3.3
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [backend-resume spec](../../spec.md) §6 C-1..C-8 + C-13 落到 backend Go handler + store + AI 编排：

- 实现 `POST /api/v1/resumes` (registerResume) handler，含 sourceType 双路分支（`upload` / `paste`）+ unsupported input validation + IK + cross-user 隔离 + 调用 [backend-upload `RegisterFileObject`](../../../backend-upload/spec.md) internal API 校验 `purpose=resume`、object exists 与实际 size；
- 实现 `GET /api/v1/resumes/{resumeId}` (getResume) handler，cross-user 返回 404；
- 实现 `GET /api/v1/resumes` (listResumes) handler，cursor pagination + `updated_at DESC, id DESC` 唯一稳定序；列表只查询并返回 closed `ResumeSummary`，完整 `Resume` 只由 `getResume` 返回；**直接解除 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) Phase 3.3 `listResumes` disabled-list 阻塞**；
- 实现 `resumes` store layer：`CreateWithParseJob(pending + async_jobs resume_parse)` / `MarkParsing` / `MarkReady(parsedSummary, parsedTextSnapshot)` / `MarkFailed(errorCode)` / `Get` / `List(cursor, pageSize)` / `DeleteForUser`；
- 实现 `resume.parse` async job handler（按 backend-targetjob 同款 `cmd/api` backend-internal runner 注册，不引入独立后台执行进程）：通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md) → 解析 JSON parse draft → 写 `resumes` + outbox `resume.parse.completed`；
- D-20 flat Resume 完成态下，`resume.parse` 成功还必须从 LLM `displayName` / structured output 派生可识别 `display_name`，不得把“上传的简历 / 粘贴的简历”等通用标题、上传文件名或 raw resume 第一行作为 ready 简历最终名称；若 LLM 输出失败但已抽取可读正文，失败路径也要写入非通用 fallback `display_name`，避免详情长期停留在“名称生成中”；
- upload source 的 prompt input 与 `parsed_text_snapshot` 必须来自当前支持文件的可读正文提取（PDF / Markdown / text），不得使用文件名、截断文件片段、PDF literal 乱码或二进制 bytes 直转 string；DOCX 不属于当前 Resume 上传支持范围，必须在 presign/register 前拒绝；PDF 读取预算必须覆盖真实浏览器生成简历文件所需的 xref / 字体映射；
- `GET /api/v1/resumes/{resumeId}/source` 必须只服务当前用户 upload-backed PDF 原件，供前端详情 PDF preview object 使用；paste、Markdown、TXT、DOCX、缺失对象、归档和跨用户访问返回 404；
- `resume.parse` 成功路径必须把完整抽取正文发送给 LLM，但模型只返回 `displayName` 与结构化字段，不再回显整份 `markdownText`；`parsed_text_snapshot` 由后端从同一份完整正文确定性构建，成功与失败路径都不依赖模型复述；
- `resume.parse.default` 必须为结构化 strict JSON 提供至少 8192 tokens 输出预算；同时用长输入尾部唯一 marker 证明业务代码未截断 1M-context 模型可承载的简历正文，并在 `finish_reason=length` 时 fail closed 为 `AI_OUTPUT_INVALID`；
- 若 LLM provider / output validation 失败但 upload / paste 已抽取出可读正文，失败路径必须保存 Markdown fallback 快照，而不是把 PDF 抽取文本原样折叠成一段；fallback 只用于失败态展示，不发送 `resume.parse.completed`，也不伪装为 LLM 成功结果；
- `registerResume` 必须强制 active resume 数量上限，默认 `resume.maxActive=10`，并继续强制 upload 文件大小上限，默认 `upload.maxBytes.resume=2097152`；
- 接 [B3 events `resume.parse.completed`](../../../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)：只有最终 ready 成功路径通过 outbox 写入 envelope 字段集（`resumeId / userId / parseStatus`）+ PII 边界（不含 raw text / parsed_summary）；失败路径不发 completed event；
- 在 `cmd/api` 挂载 `registerResume` / `getResume` / `listResumes` route，验证 session middleware、IK middleware、path params 与 backend-internal `resume_parse` runner wiring 都走真实 runtime；
- 明确本 plan 落地 flat `Resume` source 登记、解析、Markdown 快照与列表读取，不创建 `structured_master` `ResumeVersion`；
- 通过 spec §6 C-1..C-8 + C-13 验收 + 新增 E2E.P0.034 / E2E.P0.035 两个 BDD 场景；
- 不实现 update / duplicate / tailor / export 流程（归 plan 002 / 003）；真实 PDF 导出按 spec D-6 的 P0 `501 RESUME_EXPORT_NOT_AVAILABLE` / P1 plan 003 处理，本 plan 不实现。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 标记 `backend-resume` (C7) 为 Resume Workshop 阶段 1 第 2 个 subspec：必须在 `frontend-resume-workshop` mock-first 路径切真前完成。本 plan 是 backend-resume 第一批 plan，承担 P0 用户路径 "登记原始简历 source → 等待解析 → 看到 source 列表 / parse 状态" 的 backend 端到端。当前 flat Resume 合同中，`registerResume` + `resume.parse` 直接写入同一 `resumes` 行的正文快照、结构化内容和可识别名称。

`listResumes` operation 同时是 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) Phase 3.3 ResumePickerModal "disabled-list 模式" 的解锁前置：本 plan 落地后 workspace 001 owner 可启动原地修订（spec 1.2 → 1.3 / plan checklist active-list 改造），不创建 sibling。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 register / get handler skeleton；Phase 2 起来就有 store layer；Phase 3 起来就有 resume.parse async job + AIClient 集成；Phase 4 起来就有 list + pagination + cross-user 隔离；Phase 5 收口 + BDD + 解锁 workspace 001。

执行本 plan 前必须确认：

- [B2 D-18](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) Phase 1-5 已完成（B1 vocabulary、OpenAPI schema/operation、fixtures、inventory lint、generated server/client artifacts 全部就位；`registerResume` sourceType 扩展与 `listResumes` fixtures 可被真实 handler 字节比对）。
- [B3 D-14](../../../event-and-outbox-contract/plans/002-resume-tailor-mode-drift-fix/plan.md) Phase 1-4 已完成（ResumeTailorMode enum、baseline manifest、generated 类型、negative grep 与 B3 spec 描述全部对齐）；本 plan 直接消费 `resume.parse.completed` envelope，不消费 `resume.tailor.completed`，但 events drift gate 必须 PASS。
- [B4 002 flat Resume migration](../../../db-migrations-baseline/plans/002-flat-resume-migration/plan.md) 已完成，当前 schema 使用 `resumes` 与 `practice_plans.resume_id`。
- [backend-upload/001](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) 是完成条件（createUploadPresign + Register internal API 可用）。截至 2026-05-13 backend-upload/001 completed，`createUploadPresign` handler、`RegisterFileObject` internal service、privacy delete baseline、Register-time object `Stat` + actual size mismatch rejection、live roundtrip no-op/skip guard 均已可用；backend-resume/001 必须消费这个当前契约，upload path 不得只检查 fileObject row 存在。
- [F3 001 baseline](../../../prompt-rubric-registry/plans/001-baseline/plan.md) 已 ready（`resume.parse` feature_key + prompt / rubric / model profile 就位）。
- [A3 003](../../../ai-provider-and-model-routing/plans/003-provider-registry-and-capability-profiles/plan.md) 已 ready（AIClient + provider registry + Capability Model Profile）。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior + contract`。本 plan 实现 backend handler / store / async job / AI 调用；用户可见 HTTP API 行为。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. handler unit test：sourceType 双路参数校验 + unsupported input validation + IK + 422 / 404 / 跨用户隔离；
  2. store integration test：CRUD + parse_status 状态机 + cross-user 隔离 + cursor pagination 边界；
  3. resume.parse job unit test（stub AIClient provider）：成功路径 / 解析 JSON 失败 / AI provider timeout retryable / output_invalid;
  4. outbox event unit test：envelope 字段集 + PII 红线（不含 raw text）；
  5. listResumes store/service/handler tests：≥ 25 行 + cursor 第二页 + `hasMore=false` + cross-user 不可见 + exact `ResumeSummary` keys + forbidden detail fields absent；
  6. `cmd/api` route/runtime test：session middleware、IK middleware、route path params、resume_parse backend-internal runner wiring 与 shutdown。
  7. profile catalog regression：从 `config/ai-profiles.yaml` 读取 `resume.parse.default`，Red 断言输出预算不低于 8192 且 profile version 随预算变更；
  8. long-resume integrity regression：构造长输入及末尾唯一 marker，Red 断言完整 prompt 与 deterministic snapshot 都保留 marker，模型响应无需 `markdownText`；
  9. truncation terminality regression：stub AI 返回 `finish_reason=length`，Red 断言 parse 在 JSON decode 前以 `AI_OUTPUT_INVALID` 失败、保留完整快照且不发 completed outbox。
  执行入口：`/implement backend-resume/001-asset-register-parse-and-listing` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.034 register-and-list + E2E.P0.035 parse-async-job-lifecycle。详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/resume/handler/... -run TestRegisterSourceType -count=1`
  - `cd backend && go test ./internal/resume/store/... -tags=integration -count=1`
  - `cd backend && go test ./internal/resume/jobs/... -run TestResumeParseJob -count=1`（stub AIClient）
  - `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseRunnerHTTPScenario' -count=1`
  - smoke：`curl -X POST /api/v1/resumes` 与 mock-server fixture 字节比对
  - grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume runner kernel/outbox payload tests（C-13 negative）
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `registerResume` | `openapi/fixtures/Resumes/registerResume.json` `default` / `paste-text`; validation / IK mismatch cases are handler tests unless B2 adds explicit error fixtures | `frontend-resume-workshop/002-create-flow`, backend scenario harness in E2E.P0.034 | `backend/internal/resume/handler/register.go` real handler + `cmd/api` `POST /api/v1/resumes` route with session + IK middleware | `resumes` + `file_objects` reference for upload + `async_jobs` resume_parse in the same transaction | `resume.parse.default` through A3/F3 stub in tests | E2E.P0.034 + E2E.P0.035 |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` `default` / `not-found` | `frontend-resume-workshop` full-detail adapter/detail flows only | `backend/internal/resume/handler/get.go` real handler + `cmd/api` `GET /api/v1/resumes/{resumeId}` route | `resumes` full-detail projection | none | E2E.P0.034 + E2E.P0.037 |
| `getResumeSource` | `openapi/fixtures/Resumes/getResumeSource.json` `default` / `not-found` | `frontend-resume-workshop/001` PDF detail preview object | `backend/internal/resume/handler/get.go` `GetResumeSource` + `cmd/api` `GET /api/v1/resumes/{resumeId}/source` route | `resumes.file_object_id` + `file_objects.object_key` + object storage bytes | none | E2E.P0.037 focused substitute |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated`；外层保持 `PaginatedResume`，Phase 15 由 B2 owner 仅把 `items` 改为 `ResumeSummary[]` exact keys | `frontend-resume-workshop/001` summary-only list view / Home selector and `frontend-workspace-and-practice` ResumePicker unblock | `backend/internal/resume/handler/list.go` summary handler + `cmd/api` `GET /api/v1/resumes` route | `resumes` explicit summary projection; no full-detail row load | none | E2E.P0.034 + E2E.P0.036 |

## 4 实施步骤

### Phase 1: register / get handler skeleton + sourceType 双路

#### 1.1 实现 `internal/resume/handler/register.go`
- 实现 generated server interface `RegisterResume`
- sourceType 双路校验（upload 必带 fileObjectId / paste 必带 rawText；guided 等 unsupported sourceType 返回 422）
- upload 路径：调 [backend-upload `RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)`](../../../backend-upload/spec.md) internal API；该调用必须以对象存储 `Stat` 证明 object exists 且实际 size 与 `file_objects.byte_size` 一致后，才允许把 `resumes.file_object_id` 写入本 subject store
- IK + 24h TTL（B1 idempotency 工具）
- 在同一事务内创建 `resumes` queued row + `async_jobs(job_type=resume_parse, resource_type=resume_asset)` row；返回 202 + `ResumeWithJob{resumeId, job(jobType=resume_parse, status=queued)}`，与 [B2 fixture `registerResume.json`](../../../mock-contract-suite/spec.md) `default` / `paste-text` 两个 scenario 字节一致

#### 1.2 实现 `internal/resume/handler/get.go`
- 实现 generated server interface `GetResume`
- cross-user 返回 404（不暴露存在）
- 返回 `Resume` 字段（按 B2 schema）

#### 1.3 unit test
- `register_test.go`: 双 sourceType + unsupported/invalid input + IK replay + IK mismatch + 422 + 跨用户
- `get_test.go`: 200 / 404 cross-user / 404 not exist

### Phase 2: resumes store + state machine

#### 2.1 实现 `internal/resume/store/resumes.go`
- Repository：`CreateWithParseJob / Get / List(cursor, pageSize) / MarkParsing / MarkReady(parsedSummary, parsedTextSnapshot) / MarkFailed / DeleteForUser`
- `CreateWithParseJob` 必须以事务提交 `resumes` 与 `async_jobs`，并支持 user-scoped IK replay 返回首次 `resumeId` / `job`；outbox 或 job 写入失败时不得留下 orphan resume
- parse_status state machine：`queued → processing → ready | failed`
- cursor pagination：按 `updated_at DESC, id DESC` 唯一稳定序

#### 2.2 integration test
- `resumes_integration_test.go`：CRUD + state transition + cross-user isolation + cursor 边界（empty / single page / multiple pages / `hasMore=false`）

### Phase 3: resume.parse async job + AIClient 集成

#### 3.1 实现 `internal/resume/jobs/parse.go`
- 注册到 `cmd/api` backend-internal runner / runtime composition（job_type=resume_parse, dotted=resume.parse）
- 从 `resumes` 读 `file_object_id`（upload）或 `original_text`（paste）作为 prompt input
- 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md)
- 解析 LLM JSON 输出 → 写 `parsed_summary` + `parsed_text_snapshot` + `parse_status='ready'`
- parse success 直接写当前 `resumes.structured_profile` / `display_name` / `parsed_text_snapshot`；不得创建 `structured_master` `resume_versions` 行
- 失败路径：写 `parse_status='failed'` + `error_code`；retryability 由 `async_jobs` attempt / retry metadata 表达，不向 `parse_status` 私加 `failed_retryable`
- 写入 `ai_task_runs` typed columns（model_profile_name / version / prompt_version / rubric_version 等）

#### 3.2 outbox event `resume.parse.completed`
- envelope 字段集（[B3 §3.1.4](../../../event-and-outbox-contract/spec.md#314-v1-payload-schema-inventory)）：`resumeId / userId / parseStatus`
- 只在最终 `parse_status='ready'` 时写入；AI output invalid / provider timeout / retryable exhausted 等失败路径不发 `resume.parse.completed`
- PII 边界：不含 raw text / parsed_summary

#### 3.3 resume_parse backend-internal runner wiring
- 沿用 [backend-targetjob](../../../backend-targetjob/spec.md) 的 backend-internal runner 口径：`cmd/api` 进程内 claim `async_jobs(job_type=resume_parse)` 并调用 `backend/internal/resume/jobs/parse.go`
- 提供 `RunOnce` 或等价 deterministic stepping，方便 BDD / `cmd/api` scenario test 在无 timer race 的情况下验证 queued → ready / failed / retry
- `Start(ctx)` / `Shutdown(ctx)` 必须随 `cmd/api` lifecycle 管理；不得新增独立后台执行 binary、后台执行专用 config 或 `backend-async-runner` 之外的范围外 shorthand

#### 3.4 unit test
- `parse_test.go`（stub AIClient）：成功 / parse JSON 失败 / AI timeout retryable / output_invalid
- `cmd/api/resume_parse_runner_scenario_test.go`：`Runtime.RunOnce` 成功/重试处理；kernel shutdown、未知 job type 与 lease/finalize 由 `backend/internal/runner` tests 覆盖

### Phase 4: listResumes handler

#### 4.1 实现 `internal/resume/handler/list.go`
- 实现 generated server interface `ListResumes`
- cursor pagination（按 `updated_at DESC, id DESC`）
- 返回 `PaginatedResume{items, pageInfo{nextCursor, pageSize, hasMore}}`
- cross-user 过滤（仅返回 `user_id = current_user_id`）

#### 4.2 integration test
- `list_integration_test.go`: empty / 25 行 + cursor 第二页 / cross-user 不可见 / cursor invalid 拒绝

#### 4.3 `cmd/api` route/runtime wiring
- 新增 `buildResumeRuntime`（或等价 composition helper），组合 resume store / upload service / prompt registry / AIClient / idempotency middleware / resume_parse runner kernel
- 挂载：
  - `POST /api/v1/resumes` → session middleware + IK middleware + `RegisterResume`
  - `GET /api/v1/resumes` → session middleware + `ListResumes`
  - `GET /api/v1/resumes/{resumeId}` → session middleware + path param adapter + `GetResume`
- `APP_ENV=test` 可使用 deterministic resume.parse fixture AIClient，但只能拦截 `resume.parse`；真实 dev / Kind / staging / prod 必须走 A3/F3 profile fail-fast 规则
- `cmd/api` tests 断言 route 存在、缺 session 返回 auth error、缺 IK 返回 generated error envelope、同 IK replay 不重复创建 `resumes` / `async_jobs` / outbox side effect

### Phase 5: 收口 + BDD + 解锁 workspace 001

#### 5.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `cd backend && go test ./...` PASS
- `cd backend && go test ./internal/resume/...` PASS
- `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseRunnerHTTPScenario' -count=1` PASS
- mock-first 对齐：handler 真实响应与 [B2 fixtures](../../../mock-contract-suite/spec.md) `registerResume.json` (`default` / `paste-text`)、`getResume.json` (`default` / `not-found`)、`listResumes.json` (`default` / `empty` / `paginated`) 字节比对 PASS
- grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume runner kernel/outbox payload tests：0 命中（C-13 negative）
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
- backend-resume subject chronology 已记录 2026-05-12 既有 L1 修订；本轮若改变 spec 版本、日期或既有语义，收尾阶段再追加记录行；plan 001 落地后追加新行（如完成）
- 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-resume` 状态从 "未创建" 改为 "active"（roadmap spec 3.11 → 3.12 if not already）

### Phase 6: L2 remediation - handler errors, parse retry state, and gate hardening

#### 6.1 修复 register/list 业务校验 HTTP 映射
- `RegisterResume` 对 service 层 `ErrValidationFailed` 返回 `422 + VALIDATION_FAILED`，覆盖 backend-upload missing object / size mismatch 不创建 asset 的真实错误面；
- `ListResumes` 对 invalid cursor 返回 `422 + VALIDATION_FAILED`，不得把用户输入错误升级为 500；
- 补 handler unit test，证明错误 envelope 与状态码。

#### 6.2 修复 resume.parse retryable failure 状态语义
- AI timeout / retryable failure 每次失败都写 `parse_status='failed' + error_code`；
- retryable 信息只通过 `async_jobs` retry metadata 表达，不新增 `failed_retryable` parse_status；
- 后续重试允许同一 asset 从 `failed` 重新进入 `processing`，最终 ready 后只发一次 `resume.parse.completed`。

#### 6.3 加固 cmd/api 与 BDD gate
- `cmd/api` 场景补齐 handler validation mapping、invalid cursor、retryable failure → retry success 的可执行断言；
- E2E.P0.034 / E2E.P0.035 trigger/verify 必须检查新增测试名，拒绝只靠 happy path 或测反的 unit test 通过；
- 收口后重新执行 focused Go tests、两个场景脚本、docs/index/diff gate。

### Phase 7: D-20 简历扁平化适配（resumes / resumeId / structured_profile）

> product-scope D-20 / backend-resume D-13。把 register / get / list / parse 迁移到扁平 `resumes` 单表口径。依赖 [B4 002 flat Resume migration](../../../db-migrations-baseline/plans/002-flat-resume-migration/plan.md) + [B2 004 Phase 7](../../../openapi-v1-contract/plans/004-resume-additive-coverage/plan.md) contract collapse。Red 优先。

#### 7.1 store filename / table alignment

`internal/resume/store/resumes.go` 使用 `resumes` 表并读写 `structured_profile` / `display_name` 列；`source_type` 当前收敛为 {`upload`, `paste`}，范围外输入只保留为 validation negative。

（验证：`cd backend && go test ./internal/resume/store/...` PASS）

#### 7.2 handler register/get/list 迁移 resumeId

`register.go` / `get.go` / `list.go` 使用 generated 类型 `Resume`、`ResumeWithJob`、path param `resumeId`、`PaginatedResume`；register `sourceType` 仅支持 `upload` / `paste`，unsupported input 统一返回 validation error。

（验证：handler unit test + `cmd/api` wiring test PASS）

#### 7.3 parse job 写 structured_profile

`jobs/parse.go`：parse 完成直接写 `resumes.structured_profile`（无 master 确认步骤）；`resume.parse.completed` envelope 改 `resumeId`。

（验证：parse job unit test + outbox envelope test PASS）

#### 7.4 收口

`cd backend && go test ./internal/resume/... ./cmd/api`；owner 文档与场景用例通过 stale-token grep，`async_jobs.resource_type=resume_asset` 作为当前内部 job resource 值保留；`sync-doc-index --check`。

（验证：全 gate PASS + 负向 grep 0 命中）

### Phase 8: LLM-derived display_name for ready resumes

> product-scope D-20 / backend-resume D-14。创建入口只保存 source title，不写可见 `display_name`；parse 成功后必须根据 LLM 结构化结果派生完成态 `display_name`。

#### 8.1 parse job 派生 display_name

`jobs/parse.go`：解析 LLM JSON 后，从 `basics.name`、`basics.headline` / `title`、首个 experience title 或首个 project name 等字段中选取可读名称，过滤通用上传 / 粘贴标题，并在成功路径传给 store。

（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1` PASS）

#### 8.2 store 完成态写入 display_name

`store/resumes.go`：`CompleteParseSuccess` 在 `parse_status='ready'` 的同一事务内写入非空 `display_name`；无法可靠派生时保留空 `display_name`，不得回退到注册 title 或 raw resume 第一行。

（验证：`cd backend && go test ./internal/resume/store -run 'TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）

#### 8.3 cmd/api runner kernel scenario

`cmd/api` resume_parse runner kernel 场景必须断言 ready body / stored resume 使用 LLM-derived `displayName`，retry-to-ready 后同样生效。

（验证：`cd backend && go test ./cmd/api -run 'TestResumeParseRunnerHTTPScenario|TestResumeParseRunnerRetryableFailureScenario' -count=1` PASS；E2E.P0.035 trigger/verify 检查当前测试名）

### Phase 9: Upload file readable text snapshot

#### 9.1 upload parse extracts readable text

`jobs/parse.go`：upload source 按 object key / content type 对 PDF / Markdown / text 进行可读正文提取；PDF 至少解析 text objects，Markdown / text 保持 UTF-8 正文。DOCX 不再进入解析路径，上传注册侧必须拒绝。AI prompt input 与 `parsed_text_snapshot` 使用同一可读正文，且不得包含文件名或原始二进制片段。

（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1` PASS）

#### 9.2 source title no longer seeds display_name

`store/resumes.go`：`CreateWithParseJob` 创建 queued resume 时只保存 `title` 作为来源标题，不写 `display_name`；ready 后 `display_name` 只由 parse success 写入。无法从 LLM 派生时保留空 `display_name`，由前端显示中性 pending label。

（验证：`cd backend && go test ./internal/resume/store -run 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）

#### 9.3 PDF read budget covers real upload files

`jobs/parse.go`：upload object 读取预算必须覆盖真实浏览器生成的简历 PDF，避免只读 256KiB 头部导致尾部 xref / 字体映射缺失；PDF 抽取优先使用 `pdftotext -layout - -`，Go parser / literal fallback 只有通过可读性 gate 才能返回正文，不能把 PDF binary / literal 乱码写入 snapshot。focused test 使用 554631 bytes 真实失败样本大小作为读取预算下限 gate，并用 unreadable PDF fixture 断言乱码不会进入 AI prompt / snapshot。

（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerRejectsUnreadablePDFText|TestParseHandlerExtractsReadableUploadText' -count=1` PASS；local UAT 对 554631 bytes PDF 重排后 `parsed_text_snapshot` 以中文正文开头）

#### 9.4 extracted text survives LLM failure

`jobs/parse.go` / `store/resumes.go`：一旦 upload / paste 正文已成功抽取，后续 prompt registry、AI provider 或 AI output validation 失败时，`CompleteParseFailure` 也必须写入 `parsed_text_snapshot`，保证只读详情仍能展示原始内容；失败路径仍不发 `resume.parse.completed`。

（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox -count=1` PASS；`cd backend && go test ./internal/resume/store -run TestCompleteParseFailureCanPersistExtractedTextSnapshot -count=1` PASS）

### Phase 10: Display name robustness and prompt contract hardening

#### 10.1 prompt schema requires displayName

`config/prompts/resume.parse/v0.1.0.schema.json` / `.md`：在 schema-derived output contract 中加入 required `displayName`，要求模型返回根据候选人姓名、岗位定位或核心技术生成的短名称；不得返回“上传的简历 / 粘贴的简历”、上传文件名或直接照抄 raw 第一行。更新 prompt hash，并运行 prompt lint。

（验证：`make lint-prompts` PASS）

#### 10.2 parse job validates and consumes displayName

`backend/internal/resume/jobs/parse.go`：`decodeResumeParseResponse` 优先读取 AI `displayName`，并复用后端过滤规则拒绝通用标题、文件名和 raw 第一行直出；缺失或无效时回退到 structured fields 派生。

（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1` PASS）

#### 10.3 failure path writes extracted-text fallback display_name

`jobs/parse.go` / `store/resumes.go`：AI provider / output 失败但 `parsed_text_snapshot` 已存在时，失败事务同时写入一个非通用 fallback `display_name`；fallback 只能从可读正文中组合候选人姓名 + headline / 技术定位，不能使用文件名或 raw 第一行单独直出。

（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox -count=1` PASS；`cd backend && go test ./internal/resume/store -run TestCompleteParseFailureCanPersistExtractedTextSnapshot -count=1` PASS）

#### 10.4 frontend detail polling stops on terminal / readable states

`ResumeDetailView` 的 `getResume` 轻量轮询只允许 `sourceType=upload && parseStatus in queued|processing && no parsedTextSnapshot/originalText`；`failed`、`ready` 或任一可读正文已存在时立即停止。

（验证：`corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx` PASS）

### Phase 11: Markdown snapshot and active resume limits

#### 11.1 prompt schema requires markdownText

`config/prompts/resume.parse/v0.1.0.schema.json` / prompt body：新增 required `markdownText`，要求模型在不改变简历行文结构、段落顺序和事实内容的情况下把抽取正文转成 Markdown。

（验证：`make lint-prompts` PASS；parse decode focused test）

#### 11.2 parse success persists Markdown snapshot

`backend/internal/resume/jobs/parse.go`：`decodeResumeParseResponse` 校验并返回 `markdownText`，成功路径写入 `CompleteParseSuccess.ParsedTextSnapshot=markdownText`；失败路径仍保留已抽取正文，但必须先规范为 Markdown fallback，供详情失败态/兜底展示。

（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox|TestDecodeResumeParseResponseRequiresMarkdownText' -count=1` PASS）

#### 11.3 active resume limit

`config/config.yaml` 新增 `resume.maxActive: 10`，`backend/internal/resume/store/resumes.go` 在 IK replay miss 后、insert 前按 user active resumes 计数；达到上限返回 service validation error，不创建 resume / async job。`archiveResume` 的 `deleted_at` 行不计入 active 数量。

（验证：`cd backend && go test ./internal/resume/... -run 'TestRegisterResumePassesConfiguredActiveLimit|TestCreateWithParseJobRejectsNewResumeWhenActiveLimitReached|TestCreateWithParseJobAllowsIdempotentReplayAtActiveLimit' -count=1` PASS）

#### 11.4 upload default size limit

`upload.maxBytes.resume` 默认改为 `2097152`，配置校验继续要求正数；前端本地校验与后端默认一致。

（验证：`cd backend && go test ./internal/platform/config/... ./cmd/api -run 'TestRepoLocalConfigLoadsPublicDefaults|TestBuildUploadRoutesAlignsIdempotencyTTLWithPresignTTL' -count=1` PASS）

#### 11.5 PDF / AI failure Markdown fallback

当 PDF / Markdown / text 已抽取可读正文但 AI provider 或 output validation 失败时，`CompleteParseFailure` 写入 Markdown fallback：标题、章节、技能项和工作经历至少形成 Markdown heading / list / paragraph，而不是保存原始 PDF 行文本。该路径保持 `parse_status='failed'`，不写 success outbox。DOCX 作为不支持格式在上传注册前被拒绝，不产生失败态 snapshot。

（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerMarkdownFallbackSurvivesPDFAIOutputFailure -count=1` PASS）

### Phase 12: Source-format preview and DOCX exclusion

#### 12.1 upload whitelist narrows to PDF / Markdown / text

`backend/internal/upload/service/register.go` 与 `backend/internal/upload/handler/presign.go` 对 `purpose=resume` 只允许 PDF、Markdown 和 text；DOCX 在 presign/register 前返回 validation error，不创建 file object 或 resume parse job。

（验证：`cd backend && go test ./internal/upload/service ./internal/upload/handler -run 'TestCreateUploadPresignRejectsResumeDOCX|TestCreateUploadPresignRejectsResumeDOCXBeforePresign' -count=1` PASS）

#### 12.2 parse job rejects DOCX fallback input

`backend/internal/resume/jobs/parse.go` 删除 DOCX 解包和 XML 文本提取逻辑；即使历史对象误入 parse job，也必须返回 unsupported source text error，而不是读取 ZIP/XML 内容进入 prompt。

（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerRejectsDOCXUploadText|TestParseHandlerExtractsReadableUploadText' -count=1` PASS）

#### 12.3 PDF source endpoint

`backend/internal/resume/store/resumes.go`、`service.go`、`handler/get.go` 与 `cmd/api/main.go` 实现 `getResumeSource`：按 `user_id + resumeId` scoped lookup upload PDF 对象，返回 inline PDF bytes；paste、Markdown、TXT、缺失对象、归档和跨用户访问返回 404；响应不暴露 object key。

（验证：`cd backend && go test ./internal/resume ./internal/resume/handler ./internal/resume/store ./cmd/api -run 'TestGetResumeSource|TestGetSourceFile' -count=1` PASS）

### Phase 13: Real-resume output budget regression

#### 13.1 RED: lock the long-resume output budget

在 `backend/internal/ai/aiclient/profile/catalog_test.go` 增加 focused regression，从 repo-tracked `config/ai-profiles.yaml` 加载 `resume.parse.default`，断言 `max_tokens >= 8192`。当前 2048 配置必须先失败，作为真实 3170 字符简历 strict JSON 在 token cap 处截断的复现门禁。

（验证：`cd backend && go test ./internal/ai/aiclient/profile -run TestCatalogKeepsResumeParseOutputBudget -count=1` RED）

#### 13.2 GREEN: raise and version the profile budget

将 `resume.parse.default.max_tokens` 提升到 8192，并递增 profile version。不得修改 provider、model、temperature、timeout 或其它 profile，避免把本次截断修复扩大成模型路由调整。

（验证：13.1 focused test + `cd backend && go test ./internal/ai/aiclient/profile -count=1` PASS）

#### 13.3 Harden E2E.P0.035 evidence

扩展 E2E.P0.035 trigger/verify，执行并核验 `TestCatalogKeepsResumeParseOutputBudget` 的 runner marker、测试名与 PASS；继续保留 output-invalid 失败态 snapshot 与 privacy 红线断言。

（验证：P0.035 `setup → trigger → verify → cleanup` PASS）

#### 13.4 BDD gate

`BDD-Gate: E2E.P0.035` 证明 repo-tracked resume.parse runtime profile 可承载真实长简历输出，并保持失败态 snapshot 与 ready-only outbox 语义。

### Phase 14: Deterministic full-resume snapshot and truncation fail-closed

> 本 Phase 原地替换 Phase 11 的旧 `markdownText` 回显合同；不保留双轨输出。8192-token profile 预算继续作为结构化 JSON 的安全余量，而不是完整正文保存机制。

#### 14.1 RED: long input tail marker and structured-only response

在 `backend/internal/resume/jobs` 增加长输入回归：输入正文超过常规简历长度且末尾包含唯一 marker；stub AI 返回不含 `markdownText` 的完整结构化 JSON。测试必须先证明当前 decoder 因 required `markdownText` 失败，并断言发送给 AI 的 prompt 包含末尾 marker。

（验证：focused Go test RED，错误来自 `markdownText` 旧合同而非 fixture/build 失败）

#### 14.2 GREEN: persist deterministic complete snapshot

`backend/internal/resume/jobs/parse.go` 用完整提取正文确定性构建 `parsed_text_snapshot`，成功和失败路径复用同一快照；`decodeResumeParseResponse` 只解析 `displayName` 与结构化字段。`config/prompts/resume.parse` prompt/schema 删除 `markdownText`，不得在业务代码中新增字符或 token 切片。

（验证：14.1 focused test PASS；现有 PDF / Markdown / text 提取、displayName、failure snapshot 与 ready-only outbox tests PASS）

#### 14.3 RED/GREEN: finish_reason=length fails closed

stub AI 返回看似可解码或被截断的 JSON 且 `FinishReason="length"`。parse handler 必须在 decode 前按 `AI_OUTPUT_INVALID` 完成失败事务，持久化含尾部 marker 的完整 snapshot，不写 `resume.parse.completed`；不得把 length completion 当作 ready。

（验证：focused Go test 先 RED 后 GREEN）

#### 14.4 Sync prompt-owned artifacts

同步 `resume.parse` prompt body、schema、template hash、baseline seed migration 与 offline resolved prompt；更新 eval cases 为 structured-only output。`make lint-prompts`、`make eval-offline-resolve` 后 drift 为零，且负向 grep 证明当前 prompt/schema/eval/seed 不再要求 `markdownText`。

#### 14.5 Harden and execute E2E.P0.035

E2E.P0.035 trigger/verify 必须执行并核验 long-input tail marker、structured-only response、`finish_reason=length` fail-closed 测试名与 PASS；继续拒绝 no-op/skip，保留 profile budget、snapshot privacy、ready-only outbox 与 runner 证据。

#### 14.6 BDD gate

`BDD-Gate: E2E.P0.035` 证明完整简历正文进入 prompt 并由程序持久化、模型不回显正文、输出被截断时不产生 ready 假成功。

### Phase 15: Closed ResumeSummary list projection

> 本 Phase 依赖 B2 OpenAPI owner 原地新增 `ResumeSummary`，保持现有 `PaginatedResume` wrapper 与 `pageInfo` 不变，仅将 `items` 改为 `ResumeSummary[]`，并同步 `listResumes` fixtures 和 generated server/client artifacts；不得新增平行 pagination wrapper，当前 backend owner 也不以手写响应 struct 绕过 generated contract。依赖落地后通过 `/implement backend-resume/001-asset-register-parse-and-listing backend` → `/tdd` 顺序执行。

#### 15.1 RED: store projection rejects full-detail loading

新增 store SQL/mock/integration tests，要求 list query 使用显式列清单并扫描到独立 `ResumeSummary` row；测试必须在当前完整 `Resume` scan / `SELECT *` 实现下失败，并对 `original_text`、`parsed_text_snapshot`、`structured_profile`、`file_object_id`、完整 `parsed_summary`、`created_at`、`deleted_at` 建立禁止选择/禁止映射断言。保留 `updated_at DESC, id DESC` cursor 稳定序、active-only 与 user scope。

#### 15.2 GREEN: summary store and derived scalars

实现 list-only projection：`id,title,display_name,language,source_type,parse_status` 直接读取；`summaryHeadline` 按 `parsed_summary.headline` → `parsed_summary.basics.headline` → `structured_profile.headline` → `structured_profile.basics.headline` 取首个 trim 后非空 string scalar；`hasReadableContent` 的 SQL boolean expression 精确为“trim 后 `parsed_text_snapshot` 非空 OR trim 后 `original_text` 非空 OR `structured_profile` 为非空 JSON object”，不得按 `fileObjectId`、`sourceType` 或 `parseStatus` 推断；`updated_at` 继续服务排序和展示。不得把正文或 JSON payload 装入 list row。`Get` 的 full-detail store path 保持独立且行为不变。

`ResumeSummaryRecord.SourceType` 使用非空 string 直接扫描并在 service 校验 `upload | paste`；NULL 或非法 legacy value 必须使 list fail closed。不得用 `file_object_id` 或默认值修补，也不得过滤掉坏行伪装成功。当前 create path 的非空写入回归测试作为替代 gate；本 Phase 不增加兼容 migration。

#### 15.3 Service/domain separation

新增 `ResumeSummary` domain/mapper，service `List` 只接收 summary rows，并继续使用现有 `PaginatedResume` 外层返回 `ResumeSummary[]` items；禁止新增 wrapper，也禁止先构造完整 `Resume` 再丢字段。单元测试覆盖 exact mapping、null/blank headline，以及 `hasReadableContent` 的 trim snapshot、trim original text、非空 structured object、空白文本/空对象反例，不允许 file/source/status 推断；同时覆盖 cursor encode/decode、第二页和 cross-user；`Get` 测试继续覆盖 full `Resume` 详情。

#### 15.4 Handler, generated contract and fixture parity

在 B2 schema/fixture/codegen 前置完成后，handler 与 `cmd/api` 使用 generated `PaginatedResume`（`items: ResumeSummary[]`）。JSON exact-key tests 逐项断言允许字段，并逐项禁止 `originalText`、`parsedTextSnapshot`、`structuredProfile`、`fileObjectId`、`parsedSummary`、`createdAt`、`deletedAt`；覆盖 default/empty/paginated fixture、invalid cursor 422、cross-user 隔离和既有 `pageInfo`。`getResume` fixture parity 必须证明完整详情未被误删。

#### 15.5 Operation and negative matrix gates

生成物、fixture、backend handler、frontend consumer 必须与 §3.1 matrix 同步；负向搜索拒绝 list handler/service/store 复用 full-detail mapper、`SELECT *` 或把 forbidden fields 写入 list response。OpenAPI inventory/codegen drift、Go compile、frontend typecheck 与 fixture parity 均为必过 gate。

#### 15.6 BDD gates

- `BDD-Gate: E2E.P0.034`：真实 `cmd/api` register/get/list 链路证明 summary pagination exact keys、forbidden details absent、get full detail 和 cross-user 隔离。
- `BDD-Gate: E2E.P0.036`：frontend list/Home consumer 只依赖 summary fields，不从正文/structured profile 推断列表状态。
- `BDD-Gate: E2E.P0.037`：打开 row 后才通过 `getResume` 获取完整详情，list payload 不承担详情预取。

### Phase 16: Configured Resume content boundaries

#### 16.1 RED/GREEN typed limits

Add missing/default/override/invalid tests for `resume.maxActive=10`, `upload.maxBytes.resume=10485760`, `resume.maxExtractedTextBytes=393216` and `resume.maxPasteTextBytes=393216`. Replace the historical 2MiB upload and 8MiB parse-reader production constants with injected config; explicit zero/negative or paste>extracted fails startup.

#### 16.2 Registration and parse boundaries

Upload 10MiB and paste 384KiB boundaries succeed; each +1 fails before resume/job creation. Extracted PDF/Markdown/TXT text is read through a bounded path that accepts 384KiB and rejects +1 before AI, while preserving deterministic snapshot and retry semantics. All text checks use UTF-8 bytes with no silent truncation.

#### 16.3 RuntimeConfig and BDD

Expose `resumeUploadBytes` and `resumePasteTextBytes` only; extracted-text and active-count remain internal. Frontend P0.081 consumes public limits, while P0.034/P0.035 prove backend registration/parse authority and zero partial state. Existing list/detail contracts remain unchanged.

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-8 + C-13 全部 PASS（C-3 与 C-4 涉及 resume.parse async 完成 / 失败，必须 stub AIClient 验证）
- `cmd/api` route/runtime gate PASS：session middleware、IK middleware、register/get/list route、resume_parse runner kernel start/shutdown 与 deterministic `RunOnce` 均有测试证据
- BDD E2E.P0.034 + E2E.P0.035 PASS
- D-14 display_name gates PASS：prompt schema、parse job、store create / complete success / failure、cmd/api runner kernel ready/retry scenario 均断言 ready 或 failed-with-snapshot resume 不保留通用上传 / 粘贴名称、上传文件名，也不把 raw resume 第一行作为名称
- D-15 upload text snapshot gates PASS：upload PDF / Markdown / text 的 `parsed_text_snapshot` 与 AI prompt input 来自可读正文，不是文件名、截断文件片段、PDF literal 乱码或二进制 bytes；DOCX 被 presign/register 和 parse fallback 双层拒绝；已抽取正文在 LLM 失败时仍持久化
- D-18 PDF source preview gates PASS：`getResumeSource` 只对当前用户 upload-backed PDF 返回 inline PDF，paste / Markdown / TXT / missing / archived / cross-user 返回 404
- D-16/D-17 limits and deterministic snapshot gates PASS：`resume.maxActive=10`、upload 10MiB、paste/extracted 384KiB 的 code default/config/limit/+1 可测；成功/失败态 `parsed_text_snapshot` 均由完整且未截断的提取正文确定性构建，模型输出不再包含 `markdownText`
- D-19 output budget gate PASS：`resume.parse.default.max_tokens >= 8192`，profile version 已递增，E2E.P0.035 verify 检查 focused regression runner 证据
- D-21 context/truncation gates PASS：长输入末尾 marker 同时存在于 AI prompt 与 snapshot；`finish_reason=length` 映射 `AI_OUTPUT_INVALID`、保留完整快照且不发 completed outbox
- D-22 summary gates PASS：`listResumes` store/service/handler/generated fixture 只承接 closed `ResumeSummary` exact fields，所有 forbidden detail fields absent；`getResume` 仍返回完整详情
- BDD E2E.P0.034 + E2E.P0.036 + E2E.P0.037 的 Phase 15 gate PASS
- `frontend-workspace-and-practice/001` owner 已收到 `listResumes` 解锁信号
- engineering-roadmap §5.2 `backend-resume` 状态已升级到 active

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: resume.parse AI 输出 JSON 不稳定（schema 漂移） | F3 prompt 设计含 output schema example + [B2 §4.6 GenerationProvenance](../../../openapi-v1-contract/spec.md#46-ai-生成结果-provenance-约束) 强制 + parse 失败 retryable + `output_schema_version` typed column 追踪 |
| R2: `source_type` 输入与当前 upload / paste 合同漂移 | Phase 1 handler tests 和 P0.034 scenario 固化 upload / paste 正向路径，并用 unsupported sourceType negative 锁定错误映射 |
| R3: cross-user isolation 漏洞导致越权 | handler 层 + store 层双层 `user_id` 过滤；integration test 强制覆盖 cross-user case |
| R4: backend-upload 未完成时本 plan 启动 | Plan 2 背景写明前置依赖；本 plan 不在 backend-upload/001 完成前启动 |
| R5: workspace 001 修订时序 | 本 plan Phase 5.3 仅发信号，不直接修订；workspace owner 在收到信号后启动 plan 1.2 → 1.3 原地修订，不创建 sibling |
| R6: B2/B3/B4 阶段 0 plan 未完成时启动本 plan | 本 plan §2 背景写明 4 个前置依赖（B2 D-18 / B3 D-14 / B4 D-17 / backend-upload 001）；任一未完成时 `/implement` 拒绝启动 |
| R7: handler 包测试通过但真实 API / runner kernel 未挂载 | Phase 4.3 / checklist 4.4-4.5 强制 `cmd/api` route/runtime wiring；BDD 场景必须输出 `method=cmd-api-http` 或等价 live runtime evidence，并拒绝 no-op / skip 作为 PASS |
| R8: AI 输出失败导致名称永久停留在生成中状态 | Phase 10 同时硬化 prompt `displayName` 合同和失败态 fallback `display_name` 写入；前端只在 truly pending 状态展示“名称生成中” |
| R9: 模型回显改变简历事实、结构或被输出 cap 截断 | Phase 14 删除 `markdownText` 输出合同，快照由完整提取正文确定性构建；长输入尾部 marker 锁住 prompt/snapshot 完整性 |
| R10: 数量限制破坏 IK replay | Phase 11 在 `CreateWithParseJob` dedupe hit 后再执行 active count gate；focused tests 覆盖达到上限时新 IK 拒绝、相同 IK replay 不误拒 |
| R11: AI 失败态把 PDF 抽取正文折叠成普通段落 | Phase 11.5 将 failure snapshot 规范为 Markdown fallback，并以 PDF upload + AI invalid output focused test 锁定标题、章节和技能 bullet |
| R12: DOCX 继续进入 prompt 或 UI 白名单 | Phase 12 在 upload handler/service 与 parse fallback 双层拒绝 DOCX，并用 focused tests 锁定前置拒绝和解析拒绝 |
| R13: PDF 预览泄漏对象存储 key 或跨用户原件 | Phase 12 的 source endpoint 只返回 user-scoped inline PDF bytes，store/service/handler tests 覆盖 missing、paste 和 cross-user 404 |
| R14: 结构化输出仍触达 token cap | Phase 13 固化 8192-token 下限，Phase 14 在 `finish_reason=length` 时 decode 前 fail closed；不通过放宽 JSON 校验伪装成功 |
| R15: handler 丢字段但数据库仍装载完整详情 | Phase 15 store query/row scanner gate 强制显式 summary projection，禁止 `SELECT *` 和 full-detail mapper |
| R16: summary 过窄迫使 frontend 重新请求或读取正文推断 | `summaryHeadline` / `hasReadableContent` 作为服务端标量投影，P0.036 验证列表所需信息闭环 |
| R17: list 合同修订误删 get 详情字段 | `ResumeSummary` 与 `Resume` 使用独立 generated/domain 类型，P0.034/P0.037 和 get fixture parity 锁定 full detail |

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 3.3 | Reopen Phase 16 for 10MiB upload and 384KiB paste/extracted typed limits, replacing 2MiB/8MiB hardcodes. |
| 2026-07-14 | 3.2 | Reopen Phase 15 to return a closed ResumeSummary projection from listResumes while preserving full getResume detail, with store/service/handler and P0.034/P0.036/P0.037 gates. |
| 2026-07-12 | 3.1 | Replace full-resume model echo with deterministic source snapshots; add long-input tail-marker and finish-reason truncation gates. |
| 2026-07-12 | 3.0 | Reopen the owner after a real 3170-character resume hit the 2048-token output cap; add the 8192-token profile regression and P0.035 gate. |
| 2026-07-10 | 2.9 | Run resume parse scenarios through runner.Runtime and update canonical handler/runtime ownership wording. |
| 2026-07-10 | 2.8 | 将 parse/list sourceType 与 tailor-mode 负向 gate 统一为 out-of-scope / 范围外口径；行为不变。 |
| 2026-07-10 | 2.7 | 收敛 backend-resume 001 / P0.034 文档到当前 `resumes`、`ResumeWithJob`、`resumeId` 与 upload/paste sourceType 口径。 |
| 2026-07-10 | 2.6 | 将 backend resume store 文件名和 owner 文档引用从 assets 收敛到当前 resumes 表口径。 |
| 2026-07-10 | 2.5 | 收敛 P0.035 / Phase 3 / Phase 12 文档到当前 flat Resume parse、DOCX 前置拒绝与 no-confirm flow 口径。 |
| 2026-07-10 | 2.4 | 将失败 / pending display name 文档口径收敛为 pending label / 生成中状态，不改变 resume.parse 合同。 |
| 2026-07-10 | 2.3 | 清理 DOCX 上传范围外标签口径；当前 Resume 上传支持范围仍仅为 PDF / Markdown / TXT 文本提取链路。 |
| 2026-07-07 | 2.2 | 本轮讨论决策：新增 PDF source preview endpoint，Resume 上传移出 DOCX 当前支持范围，仅保留 PDF / Markdown / TXT 文本提取链路。 |
