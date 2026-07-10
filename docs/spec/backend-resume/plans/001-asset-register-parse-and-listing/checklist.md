# Backend Resume Register Parse and Listing Checklist

> **版本**: 2.7
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: register / get handler skeleton + sourceType 双路

- [x] 1.1 实现 `backend/internal/resume/handler/register.go`，generated server interface `RegisterResume`（验证：编译 PASS + `go vet` PASS）
- [x] 1.2 sourceType 双路参数校验：`upload` 必带 fileObjectId / `paste` 必带 rawText；unsupported sourceType 和其他错误组合返回 422（验证：unit test `TestRegisterSourceType` 双路 + 错误组合 PASS）
- [x] 1.3 upload 路径：调用 [backend-upload `RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)`](../../../backend-upload/spec.md) internal API，确认对象存储 object exists 且实际 size 与 `file_objects.byte_size` 一致后，写入 `resumes.file_object_id` FK；missing object / size mismatch 不创建 resume（验证：integration test verify FK 建立 + mismatch rejects）
- [x] 1.4 IK 校验（缺失 / 24h replay 返回首次 resumeId / mismatch 422）（验证：unit test `TestRegisterIdempotency` PASS）
- [x] 1.5 同一事务内创建 `resumes` queued row + `async_jobs(job_type=resume_parse, resource_type=resume_asset)` row，返回 202 + `ResumeWithJob{resumeId, job(jobType=resume_parse, status=queued)}`，与 [B2 fixture `registerResume.json`](../../../mock-contract-suite/spec.md) `default` / `paste-text` 两个 scenario 字节一致（验证：fixture parity test + orphan resume rollback test）
- [x] 1.6 实现 `backend/internal/resume/handler/get.go`，generated server interface `GetResume`（验证：编译 PASS）
- [x] 1.7 getResume：cross-user 返回 404 + 不暴露存在（验证：integration test cross-user PASS）

## Phase 2: resumes store + state machine

- [x] 2.1 实现 `backend/internal/resume/store/resumes.go` Repository：`CreateWithParseJob / Get / List(cursor, pageSize) / MarkParsing / MarkReady / MarkFailed / DeleteForUser`（验证：编译 PASS）
- [x] 2.2 parse_status state machine：`queued → processing → ready | failed`；非法转换拒绝（验证：unit test `TestParseStatusTransition` PASS）
- [x] 2.3 cursor pagination 实现：按 `updated_at DESC, id DESC` 唯一稳定序（验证：integration test 25 行 + 第二页 cursor PASS）
- [x] 2.4 integration test：CRUD + state transition + cross-user isolation + FK 约束 + `resumes` / `async_jobs` 原子提交与 rollback（验证：`cd backend && go test ./internal/resume/store/... -tags=integration -count=1` PASS）

## Phase 3: resume.parse async job + AIClient 集成

- [x] 3.1 实现 `backend/internal/resume/jobs/parse.go` 与 resume_parse in-process runner kernel，注册 `job_type=resume_parse` / dotted `resume.parse`；不得新增独立 worker binary 或 `WORKER_*` config（验证：runner registry / topology negative 测试）
- [x] 3.2 从 `resumes` 读 file_object（upload）/ original_text（paste）作为 prompt input（验证：unit test verify 双路 input 路径）
- [x] 3.3 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md)；不 hardcode prompt 正文（验证：unit test stub AIClient verify profile / feature_key 路由）
- [x] 3.4 解析 LLM JSON 输出 → 写 `parsed_summary` + `parsed_text_snapshot` + `parse_status='ready'`（验证：unit test `TestResumeParseHappyPath`）
- [x] 3.5 失败路径：AI timeout / output_invalid → `parse_status='failed'` + `error_code`；retryable 信息落在 `async_jobs` retry metadata，不向 `parse_status` 私加 `failed_retryable`（验证：unit test `TestResumeParseFailureRetryable`）
- [x] 3.6 写入 `ai_task_runs` typed columns：model_profile_name / model_profile_version / prompt_version / rubric_version / route / validation_status（验证：integration test verify `ai_task_runs` 行）
- [x] 3.7 outbox `resume.parse.completed`：仅最终 ready 成功路径将 envelope 字段（resumeId / userId / parseStatus）写入 outbox_events；AI output_invalid / provider timeout / retry exhausted 等失败路径不发 completed event；PII 边界断言不含 raw text / parsed_summary（验证：unit test + payload assertion）
- [x] 3.8 PII leak negative：log / audit / outbox payload 写入路径不序列化 raw resume content、prompt body 或 model raw response；允许 SQL/store 层出现必要列名（如 `original_text` / `parsed_summary`），禁止把列值写入日志或事件（验证：专用 lint / unit test 覆盖 log sink 与 outbox payload）
- [x] 3.9 parse success 直接写当前 `resumes.structured_profile` / `display_name` / `parsed_text_snapshot`；不得创建 `structured_master` `resume_versions` 行（验证：parse job / store integration tests）

## Phase 4: listResumes handler

- [x] 4.1 实现 `backend/internal/resume/handler/list.go`，generated server interface `ListResumes`（验证：编译 PASS）
- [x] 4.2 cursor pagination 实现 + 返回 `PaginatedResume{items, pageInfo}`（验证：integration test 25+ 行 + 第二页 PASS）
- [x] 4.3 cross-user 过滤：仅返回 `user_id = current_user_id` 行（验证：integration test cross-user PASS）
- [x] 4.4 `cmd/api` route wiring：挂载 `POST /api/v1/resumes`（session + IK middleware）、`GET /api/v1/resumes`、`GET /api/v1/resumes/{resumeId}`，并把 resume_parse runner kernel 纳入 `Start(ctx)` / `Shutdown(ctx)` lifecycle（验证：`cd backend && go test ./cmd/api -run TestBuildResumeRuntime -count=1`）
- [x] 4.5 `cmd/api` HTTP scenario：通过真实 route 验证 register/get/list、auth 404/401、IK replay、不重复创建 resume/job/outbox（验证：`cd backend && go test ./cmd/api -run TestResumeRegisterListHTTPScenario -count=1`）
- [x] 4.6 字节比对 [B2 fixture `listResumes.json`](../../../mock-contract-suite/spec.md) `default` / `empty` / `paginated` 三 variant（验证：fixture parity test）

## Phase 5: 收口 + BDD + 解锁 workspace 001

- [x] 5.1 跑 `cd backend && go test ./...` + `cd backend && go test ./internal/resume/...` + `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseRunnerHTTPScenario' -count=1` 全 PASS（验证：exit 0）<!-- verified: 2026-05-13 method=go-test -->
- [x] 5.2 mock-first 对齐：`registerResume` (`default` / `paste-text`)、`getResume` (`default` / `not-found`)、`listResumes` (`default` / `empty` / `paginated`) 通过 `cmd/api` 真实 route 的响应与对应 fixture 字节比对 PASS <!-- verified: 2026-05-13 method=scenario+handler-fixture-parity -->
- [x] 5.3 grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume runner kernel/outbox payload：0 命中（C-13 negative）（验证：`git grep` 输出 + payload assertion）<!-- verified: 2026-05-13 method=rg+unit-test -->
- [x] 5.4 BDD-Gate: E2E.P0.034 resume-register-and-list PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）<!-- verified: 2026-05-13 method=scenario -->
- [x] 5.5 BDD-Gate: E2E.P0.035 resume-parse-async-job-lifecycle PASS（含 stub AIClient + outbox event 验证）<!-- verified: 2026-05-13 method=scenario -->
- [x] 5.6 在 `test/scenarios/e2e/INDEX.md` 追加 P0.034 + P0.035 行（关联需求 `backend-resume C-1..C-8, C-13`）
- [x] 5.7 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-resume` 状态从 "未创建" 改为 "active"（与 backend-upload 同步行）（验证：`sync-doc-index --check`）<!-- verified: 2026-05-13 roadmap already active -->
- [x] 5.8 通知 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) owner：`listResumes` 已就位，可启动 disabled-list → active-list 原地修订（验证：cross-plan 引用 commit + workspace 001 plan checklist 追加 unblock 引用）<!-- verified: 2026-05-13 commit=1d1f69c -->

## Phase 6: L2 remediation - handler errors, parse retry state, and gate hardening

- [x] 6.1 `RegisterResume` / `ListResumes` 将 service/store validation 错误映射为 `422 + VALIDATION_FAILED`，覆盖 upload missing object / size mismatch 与 invalid cursor（验证：handler unit test + cmd/api scenario PASS）<!-- verified: 2026-05-13 method=go-test tests=TestRegisterResumeValidationErrorsReturnUnprocessableEntity,TestListResumesInvalidCursorReturnsUnprocessableEntity,TestResumeRegisterListHTTPValidationScenario -->
- [x] 6.2 `resume.parse` retryable failure 每次写 `parse_status='failed' + error_code`，retry metadata 仍由 `async_jobs` 表达，并允许 failed asset 重试回 processing 后 ready（验证：job/store/cmd-api retry tests PASS）<!-- verified: 2026-05-13 method=go-test tests=TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox,TestParseHandlerRetriesFailedAssetBackToProcessing,TestParseStatusTransition,TestResumeParseRunnerRetryableFailureScenario -->
- [x] 6.3 加固 E2E.P0.034 / E2E.P0.035 trigger/verify，检查新增 validation/retry 测试名且拒绝 no-op/skip（验证：两个 scenario `setup -> trigger -> verify -> cleanup` PASS）<!-- verified: 2026-05-13 method=scenario logs=.test-output/e2e/p0-034-resume-register-and-list/trigger.log,.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log -->
- [x] 6.4 收口验证：focused Go tests、`go test ./internal/resume/... ./cmd/api`、`make docs-check`、`sync-doc-index --check`、`git diff --check` 全 PASS <!-- verified: 2026-05-13 method=go-test+scenario+docs-check+sync-doc-index+diff-check -->

## Phase 7: D-20 简历扁平化适配（resumes / resumeId / structured_profile）

> product-scope D-20 / backend-resume D-13。依赖 B4 002 Phase 6 + B2 004 Phase 7。Red 优先。

- [x] 7.1 store 文件名与表名对齐为 `resumes.go`：读写 `resumes`、`structured_profile` / `display_name`；`source_type` 当前收敛 {`upload`,`paste`}，范围外输入只作为 validation negative（验证：`cd backend && go test ./internal/resume/store -run 'TestCreateWithParseJob|TestCompleteParse|TestParseStatusTransition' -count=1` PASS）
- [x] 7.2 handler register/get/list 使用 generated `Resume`、`ResumeWithJob`、`resumeId`、`PaginatedResume`；register 仅支持 `upload` / `paste`，unsupported input 返回 validation error（验证：handler unit test + `cmd/api` wiring test PASS）
- [x] 7.3 parse job 写 `resumes.structured_profile`（无 master 确认），`resume.parse.completed` envelope 使用 `resumeId`（验证：parse job unit test + outbox envelope test PASS）
- [x] 7.4 收口：`cd backend && go test ./internal/resume/... ./cmd/api`；owner 文档与 P0.034 场景通过 stale-token grep，`async_jobs.resource_type=resume_asset` 作为当前内部 job resource 值保留；`sync-doc-index --check`（验证：全 gate PASS + 负向 grep 0 命中）

## Phase 8: LLM-derived display_name for ready resumes

- [x] 8.1 `backend/internal/resume/jobs/parse.go` 从 LLM structured output 派生可识别 `display_name`，过滤通用上传 / 粘贴标题（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1` PASS）<!-- verified: 2026-07-07 method=go-test -->
- [x] 8.2 `CompleteParseSuccess` 在 ready 事务中写入派生 `display_name`，无法派生时保留空值，不回退到注册 title 或 raw resume 第一行（验证：`cd backend && go test ./internal/resume/store -run 'TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）<!-- verified: 2026-07-07 method=go-test+scenario -->
- [x] 8.3 cmd/api resume_parse runner kernel ready / retry-to-ready 场景断言 stored resume 使用 LLM-derived `displayName`（验证：`cd backend && go test ./cmd/api -run 'TestResumeParseRunnerHTTPScenario|TestResumeParseRunnerRetryableFailureScenario' -count=1` PASS；P0.035 trigger/verify 检查当前测试名）<!-- verified: 2026-07-07 method=go-test+scenario -->

## Phase 9: Upload file readable text snapshot

- [x] 9.1 `backend/internal/resume/jobs/parse.go` 对 upload source 提取 PDF / Markdown / text 可读正文，AI prompt input 与 `parsed_text_snapshot` 使用同一正文，不能使用文件名或二进制 bytes；DOCX 不再属于当前上传白名单（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1` PASS）<!-- verified: 2026-07-07 method=go-test+scenario -->
- [x] 9.2 `CreateWithParseJob` 创建 queued resume 时只保存来源 `title`，不写 `display_name`；ready 后只由 parse success 写入 LLM-derived `display_name`（验证：`cd backend && go test ./internal/resume/store -run 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）<!-- verified: 2026-07-07 method=go-test+scenario -->
- [x] 9.3 `resume.parse` upload 对象读取预算覆盖真实浏览器生成 PDF，不因 256KiB 截断导致 `parsed_text_snapshot` 为空，并拒绝 PDF literal / binary 乱码正文（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerRejectsUnreadablePDFText|TestParseHandlerExtractsReadableUploadText' -count=1` PASS，assert read budget >= 554631 bytes；local UAT 真实 PDF snapshot 以中文正文开头）<!-- verified: 2026-07-07 method=go-test+local-uat -->
- [x] 9.4 `resume.parse` 在已抽取正文后遇到 AI provider / AI output 失败，仍写入 `parsed_text_snapshot` 供只读详情显示，且不发 completed event（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox -count=1` PASS；`cd backend && go test ./internal/resume/store -run TestCompleteParseFailureCanPersistExtractedTextSnapshot -count=1` PASS；local UAT 真实 PDF `parse_status=failed` / `AI_OUTPUT_INVALID` 时 snapshot_len=3083）<!-- verified: 2026-07-07 method=go-test+local-uat -->

## Phase 10: Display name robustness and prompt contract hardening

- [x] 10.1 `resume.parse` prompt schema / prompt body 显式要求 required `displayName`，并更新 prompt hash；验证: `make lint-prompts`。<!-- verified: 2026-07-07 method=lint command="make lint-prompts" -->
- [x] 10.2 `decodeResumeParseResponse` 优先使用 AI `displayName`，并拒绝通用上传/粘贴标题、上传文件名和 raw 第一行直出；验证: `cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1`。<!-- verified: 2026-07-07 method=go-test -->
- [x] 10.3 AI provider / output 失败但已有可读正文时，`CompleteParseFailure` 同时写入 fallback `display_name`，确保 failed-with-snapshot 详情不再长期显示“名称生成中”；验证: `cd backend && go test ./internal/resume/jobs -run TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox -count=1` 与 `cd backend && go test ./internal/resume/store -run TestCompleteParseFailureCanPersistExtractedTextSnapshot -count=1`。<!-- verified: 2026-07-07 method=go-test -->
- [x] 10.4 `ResumeDetailView` 对 `failed` 或已有正文的上传详情停止轮询 `getResume`，避免同一详情 URL 重复请求；验证: `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx`。<!-- verified: 2026-07-07 method=vitest -->

## Phase 11: Markdown snapshot and active resume limits

- [x] 11.1 `resume.parse` prompt schema / prompt body required `markdownText`，要求 LLM 保持原简历结构和事实并输出 Markdown；验证: `make lint-prompts`。<!-- verified: 2026-07-07 method=prompt-lint command="python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations" -->
- [x] 11.2 `decodeResumeParseResponse` 校验并返回 `markdownText`，parse success 将 `ParsedTextSnapshot` 写为 Markdown；验证: `cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox|TestParseHandlerRequiresMarkdownTextInAIResponse' -count=1`。<!-- verified: 2026-07-07 method=go-test -->
- [x] 11.3 `resume.maxActive` 默认 10 并由 `RegisterResume` / `CreateWithParseJob` 强制；达到上限的新 IK 返回 validation error，不创建 resume/job；相同 IK replay 不误拒；验证: focused service/store tests。<!-- verified: 2026-07-07 method=go-test command="cd backend && go test ./internal/resume/... ./internal/platform/config ./cmd/api -count=1" -->
- [x] 11.4 `upload.maxBytes.resume` 默认改为 2MiB，配置校验与前端本地校验一致；验证: config/cmd-api focused tests。<!-- verified: 2026-07-07 method=go-test+vitest -->
- [x] 11.5 AI provider / output validation 失败但已抽取 PDF 可读正文时，失败态 `parsed_text_snapshot` 保存 Markdown fallback（标题、章节、技能 bullet），不保存原始折叠文本；验证: `cd backend && go test ./internal/resume/jobs -run TestParseHandlerMarkdownFallbackSurvivesPDFAIOutputFailure -count=1`。<!-- verified: 2026-07-07 method=go-test -->

## Phase 12: Source-format preview and DOCX exclusion

- [x] 12.1 `createUploadPresign` / upload register 对 `purpose=resume` 仅允许 PDF / Markdown / TXT，DOCX 在 presign/register 前返回 validation error；验证: upload handler/service focused tests。<!-- verified: 2026-07-07 method=go-test packages=./internal/upload/service,./internal/upload/handler tests=TestCreateUploadPresignRejectsResumeDOCX,TestCreateUploadPresignRejectsResumeDOCXBeforePresign -->
- [x] 12.2 `resume.parse` 删除 DOCX 解包路径，历史 DOCX object 误入 parse 时返回 unsupported source text error，不进入 AI prompt；验证: parse focused tests。<!-- verified: 2026-07-07 method=go-test package=./internal/resume/jobs tests=TestParseHandlerRejectsDOCXUploadText,TestParseHandlerExtractsReadableUploadText -->
- [x] 12.3 `getResumeSource` 按 `user_id + resumeId` scoped lookup upload-backed PDF 原件并返回 inline `application/pdf`；paste / Markdown / TXT / missing / archived / cross-user 返回 404；验证: store/service/handler/cmd-api focused tests。<!-- verified: 2026-07-07 method=go-test packages=./internal/resume,./internal/resume/handler,./internal/resume/store,./cmd/api tests=TestGetResumeSource,TestGetSourceFile,TestGeneratedRouteCatalogHasNoResumeVersionOperations -->
