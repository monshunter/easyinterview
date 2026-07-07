# Backend Resume Asset Register Parse and Listing Checklist

> **版本**: 1.7
> **状态**: active
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: register / get handler skeleton + sourceType 三路

- [x] 1.1 实现 `backend/internal/resume/handler/register.go`，generated server interface `RegisterResume`（验证：编译 PASS + `go vet` PASS）
- [x] 1.2 sourceType 三路参数校验：`upload` 必带 fileObjectId / `paste` 必带 rawText / `guided` 必带 guidedAnswers；其他组合 422（验证：unit test `TestRegisterSourceType` 3 路 + 错误组合 PASS）
- [x] 1.3 upload 路径：调用 [backend-upload `RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)`](../../../backend-upload/spec.md) internal API，确认对象存储 object exists 且实际 size 与 `file_objects.byte_size` 一致后，写入 `resume_assets.file_object_id` FK；missing object / size mismatch 不创建 resume asset（验证：integration test verify FK 建立 + mismatch rejects）
- [x] 1.4 IK 校验（缺失 / 24h replay 返回首次 resumeAssetId / mismatch 422）（验证：unit test `TestRegisterIdempotency` PASS）
- [x] 1.5 同一事务内创建 `resume_assets` queued row + `async_jobs(job_type=resume_parse, resource_type=resume_asset)` row，返回 202 + `ResumeAssetWithJob{resumeAssetId, job(jobType=resume_parse, status=queued)}`，与 [B2 fixture `registerResume.json`](../../../mock-contract-suite/spec.md) `default` / `paste-text` / `guided-answers` 三个 scenario 字节一致（验证：fixture parity test + orphan asset rollback test）
- [x] 1.6 实现 `backend/internal/resume/handler/get.go`，generated server interface `GetResume`（验证：编译 PASS）
- [x] 1.7 getResume：cross-user 返回 404 + 不暴露存在（验证：integration test cross-user PASS）

## Phase 2: resume_assets store + state machine

- [x] 2.1 实现 `backend/internal/resume/store/assets.go` Repository：`CreateWithParseJob / Get / List(cursor, pageSize) / MarkParsing / MarkReady / MarkFailed / DeleteForUser`（验证：编译 PASS）
- [x] 2.2 parse_status state machine：`queued → processing → ready | failed`；非法转换拒绝（验证：unit test `TestParseStatusTransition` PASS）
- [x] 2.3 cursor pagination 实现：按 `updated_at DESC, id DESC` 唯一稳定序（验证：integration test 25 行 + 第二页 cursor PASS）
- [x] 2.4 integration test：CRUD + state transition + cross-user isolation + FK 约束 + `resume_assets` / `async_jobs` 原子提交与 rollback（验证：`cd backend && go test ./internal/resume/store/... -tags=integration -count=1` PASS）

## Phase 3: resume.parse async job + AIClient 集成

- [x] 3.1 实现 `backend/internal/resume/jobs/parse.go` 与 resume_parse in-process drainer，注册 `job_type=resume_parse` / dotted `resume.parse`；不得新增独立 worker binary 或 `WORKER_*` config（验证：runner registry / topology negative 测试）
- [x] 3.2 从 `resume_assets` 读 file_object（upload）/ original_text（paste）/ guided_answers jsonb（guided）作为 prompt input；guided 不从 `original_text` 反序列化（验证：unit test verify 三路 input 路径）
- [x] 3.3 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md)；不 hardcode prompt 正文（验证：unit test stub AIClient verify profile / feature_key 路由）
- [x] 3.4 解析 LLM JSON 输出 → 写 `parsed_summary` + `parsed_text_snapshot` + `parse_status='ready'`（验证：unit test `TestResumeParseHappyPath`）
- [x] 3.5 失败路径：AI timeout / output_invalid → `parse_status='failed'` + `error_code`；retryable 信息落在 `async_jobs` retry metadata，不向 `resume_assets.parse_status` 私加 `failed_retryable`（验证：unit test `TestResumeParseFailureRetryable`）
- [x] 3.6 写入 `ai_task_runs` typed columns：model_profile_name / model_profile_version / prompt_version / rubric_version / route / validation_status（验证：integration test verify `ai_task_runs` 行）
- [x] 3.7 outbox `resume.parse.completed`：仅最终 ready 成功路径将 envelope 字段（resumeAssetId / userId / parseStatus）写入 outbox_events；AI output_invalid / provider timeout / retry exhausted 等失败路径不发 completed event；PII 边界断言不含 raw text / guided answers / parsed_summary（验证：unit test + payload assertion）
- [x] 3.8 PII leak negative：log / audit / outbox payload 写入路径不序列化 raw resume content、`guided_answers` 内容、prompt body 或 model raw response；允许 SQL/store 层出现必要列名（如 `original_text` / `guided_answers` / `parsed_summary`），禁止把列值写入日志或事件（验证：专用 lint / unit test 覆盖 log sink 与 outbox payload）
- [x] 3.9 Preview Confirm 前不得创建正式 `structured_master` `resume_versions` 行；parse output 只作为草稿，保存 v1 交给 backend-resume/002（验证：integration test 断言 parse job 完成后 `resume_versions` count unchanged）

## Phase 4: listResumes handler

- [x] 4.1 实现 `backend/internal/resume/handler/list.go`，generated server interface `ListResumes`（验证：编译 PASS）
- [x] 4.2 cursor pagination 实现 + 返回 `PaginatedResumeAsset{items, pageInfo}`（验证：integration test 25+ 行 + 第二页 PASS）
- [x] 4.3 cross-user 过滤：仅返回 `user_id = current_user_id` 行（验证：integration test cross-user PASS）
- [x] 4.4 `cmd/api` route wiring：挂载 `POST /api/v1/resumes`（session + IK middleware）、`GET /api/v1/resumes`、`GET /api/v1/resumes/{resumeAssetId}`，并把 resume_parse drainer 纳入 `Start(ctx)` / `Shutdown(ctx)` lifecycle（验证：`cd backend && go test ./cmd/api -run TestBuildResumeRuntime -count=1`）
- [x] 4.5 `cmd/api` HTTP scenario：通过真实 route 验证 register/get/list、auth 404/401、IK replay、不重复创建 asset/job/outbox（验证：`cd backend && go test ./cmd/api -run TestResumeRegisterListHTTPScenario -count=1`）
- [x] 4.6 字节比对 [B2 fixture `listResumes.json`](../../../mock-contract-suite/spec.md) `default` / `empty` / `paginated` 三 variant（验证：fixture parity test）

## Phase 5: 收口 + BDD + 解锁 workspace 001

- [x] 5.1 跑 `cd backend && go test ./...` + `cd backend && go test ./internal/resume/...` + `cd backend && go test ./cmd/api -run 'TestBuildResumeRuntime|TestResumeRegisterListHTTPScenario|TestResumeParseDrainerHTTPScenario' -count=1` 全 PASS（验证：exit 0）<!-- verified: 2026-05-13 method=go-test -->
- [x] 5.2 mock-first 对齐：`registerResume` (`default` / `paste-text` / `guided-answers`)、`getResume` (`default` / `not-found`)、`listResumes` (`default` / `empty` / `paginated`) 通过 `cmd/api` 真实 route 的响应与对应 fixture 字节比对 PASS <!-- verified: 2026-05-13 method=scenario+handler-fixture-parity -->
- [x] 5.3 grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume drainer/outbox payload：0 命中（C-13 negative）（验证：`git grep` 输出 + payload assertion）<!-- verified: 2026-05-13 method=rg+unit-test -->
- [x] 5.4 BDD-Gate: E2E.P0.034 resume-register-and-list PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）<!-- verified: 2026-05-13 method=scenario -->
- [x] 5.5 BDD-Gate: E2E.P0.035 resume-parse-async-job-lifecycle PASS（含 stub AIClient + outbox event 验证）<!-- verified: 2026-05-13 method=scenario -->
- [x] 5.6 在 `test/scenarios/e2e/INDEX.md` 追加 P0.034 + P0.035 行（关联需求 `backend-resume C-1..C-8, C-13`）
- [x] 5.7 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-resume` 状态从 "未创建" 改为 "active"（与 backend-upload 同步行）（验证：`sync-doc-index --check`）<!-- verified: 2026-05-13 roadmap already active -->
- [x] 5.8 通知 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) owner：`listResumes` 已就位，可启动 disabled-list → active-list 原地修订（验证：cross-plan 引用 commit + workspace 001 plan checklist 追加 unblock 引用）<!-- verified: 2026-05-13 commit=1d1f69c -->

## Phase 6: L2 remediation - handler errors, parse retry state, and gate hardening

- [x] 6.1 `RegisterResume` / `ListResumes` 将 service/store validation 错误映射为 `422 + VALIDATION_FAILED`，覆盖 upload missing object / size mismatch 与 invalid cursor（验证：handler unit test + cmd/api scenario PASS）<!-- verified: 2026-05-13 method=go-test tests=TestRegisterResumeValidationErrorsReturnUnprocessableEntity,TestListResumesInvalidCursorReturnsUnprocessableEntity,TestResumeRegisterListHTTPValidationScenario -->
- [x] 6.2 `resume.parse` retryable failure 每次写 `parse_status='failed' + error_code`，retry metadata 仍由 `async_jobs` 表达，并允许 failed asset 重试回 processing 后 ready（验证：job/store/cmd-api retry tests PASS）<!-- verified: 2026-05-13 method=go-test tests=TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox,TestParseHandlerRetriesFailedAssetBackToProcessing,TestParseStatusTransition,TestResumeParseDrainerRetryableFailureScenario -->
- [x] 6.3 加固 E2E.P0.034 / E2E.P0.035 trigger/verify，检查新增 validation/retry 测试名且拒绝 no-op/skip（验证：两个 scenario `setup -> trigger -> verify -> cleanup` PASS）<!-- verified: 2026-05-13 method=scenario logs=.test-output/e2e/p0-034-resume-register-and-list/trigger.log,.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log -->
- [x] 6.4 收口验证：focused Go tests、`go test ./internal/resume/... ./cmd/api`、`make docs-check`、`sync-doc-index --check`、`git diff --check` 全 PASS <!-- verified: 2026-05-13 method=go-test+scenario+docs-check+sync-doc-index+diff-check -->

## Phase 7: D-20 简历扁平化适配（resumes / resumeId / structured_profile）

> product-scope D-20 / backend-resume D-13。依赖 B4 002 Phase 6 + B2 004 Phase 7。Red 优先。

- [ ] 7.1 store `assets.go`→`resumes.go`：表名 `resumes` + 读写 `structured_profile` / `display_name` 列；删 `guided_answers`；`source_type` 收敛 {`upload`,`paste`}（验证：`cd backend && go test ./internal/resume/store/...` PASS）
- [ ] 7.2 handler register/get/list 迁移：generated `ResumeAsset`→`Resume`、`ResumeAssetWithJob`→`ResumeWithJob`、`resumeAssetId`→`resumeId`、`PaginatedResumeAsset`→`PaginatedResume`；register 删 `guided` 422 分支（验证：handler unit test + `cmd/api` wiring test PASS）
- [ ] 7.3 parse job 写 `resumes.structured_profile`（无 master 确认），`resume.parse.completed` envelope 改 `resumeId`（验证：parse job unit test + outbox envelope test PASS）
- [ ] 7.4 收口：`cd backend && go test ./internal/resume/... ./cmd/api`；零 `resumeAssetId` / `resume_assets` / `ResumeAsset` 残留 grep（generated 由 B2 重生除外）；`sync-doc-index --check`（验证：全 gate PASS + 负向 grep 0 命中）

## Phase 8: LLM-derived display_name for ready resumes

- [x] 8.1 `backend/internal/resume/jobs/parse.go` 从 LLM structured output 派生可识别 `display_name`，过滤通用上传 / 粘贴标题（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1` PASS）<!-- verified: 2026-07-07 method=go-test -->
- [x] 8.2 `CompleteParseSuccess` 在 ready 事务中写入派生 `display_name`，无法派生时保留空值，不回退到注册 title 或 raw resume 第一行（验证：`cd backend && go test ./internal/resume/store -run 'TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）<!-- verified: 2026-07-07 method=go-test+scenario -->
- [x] 8.3 cmd/api resume_parse drainer ready / retry-to-ready 场景断言 stored resume 使用 LLM-derived `displayName`（验证：`cd backend && go test ./cmd/api -run 'TestResumeParseDrainerHTTPScenario|TestResumeParseDrainerRetryableFailureScenario' -count=1` PASS；P0.035 trigger/verify 检查当前测试名）<!-- verified: 2026-07-07 method=go-test+scenario -->

## Phase 9: Upload file readable text snapshot

- [x] 9.1 `backend/internal/resume/jobs/parse.go` 对 upload source 提取 PDF / DOCX / Markdown / text 可读正文，AI prompt input 与 `parsed_text_snapshot` 使用同一正文，不能使用文件名或二进制 bytes（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1` PASS）<!-- verified: 2026-07-07 method=go-test+scenario -->
- [x] 9.2 `CreateWithParseJob` 创建 queued resume 时只保存来源 `title`，不写 `display_name`；ready 后只由 parse success 写入 LLM-derived `display_name`（验证：`cd backend && go test ./internal/resume/store -run 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）<!-- verified: 2026-07-07 method=go-test+scenario -->
