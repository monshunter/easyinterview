# Backend Resume Register Parse and Listing Checklist

> **版本**: 3.4
> **状态**: completed
> **更新日期**: 2026-07-14

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
- [x] 4.5 `cmd/api` handler integration test 验证 register/get/list、auth 404/401、IK replay 和零重复 side effect；该 Go test 属于代码层，focused run 只作开发反馈，阶段完成由仓库根 `make test` 承接。
- [x] 4.6 字节比对 [B2 fixture `listResumes.json`](../../../mock-contract-suite/spec.md) `default` / `empty` / `paginated` 三 variant（验证：fixture parity test）

## Phase 5: 收口 + BDD + 解锁 workspace 001

- [x] 5.1 仓库根 `make test` 统一完成 backend 与 frontend 全量单元测试回归；`TestResumeRegisterListHTTPContract` / `TestResumeParseRunnerIntegration` 等 focused 代码层测试仅作开发反馈，不作为 E2E 或阶段完成证据。
- [x] 5.2 code-level contract：`registerResume`、`getResume`、`listResumes` 的 handler response 与 fixture 保持一致；fixture parity 不属于 E2E，阶段单测完成由仓库根 `make test` 承接。
- [x] 5.3 grep `inline|rewrite|mirror` in `backend/internal/resume/` + resume runner kernel/outbox payload：0 命中（C-13 negative）（验证：`git grep` 输出 + payload assertion）<!-- verified: 2026-05-13 method=rg+unit-test -->
- [x] 5.7 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-resume` 状态从 "未创建" 改为 "active"（与 backend-upload 同步行）（验证：`sync-doc-index --check`）<!-- verified: 2026-05-13 roadmap already active -->
- [x] 5.8 通知 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) owner：`listResumes` 已就位，可启动 disabled-list → active-list 原地修订（验证：cross-plan 引用 commit + workspace 001 plan checklist 追加 unblock 引用）<!-- verified: 2026-05-13 commit=1d1f69c -->

## Phase 6: L2 remediation - handler errors, parse retry state, and gate hardening

- [x] 6.1 `RegisterResume` / `ListResumes` 将 service/store validation 错误映射为 `422 + VALIDATION_FAILED`，覆盖 upload missing object / size mismatch 与 invalid cursor（验证：handler unit test + cmd/api scenario PASS）<!-- verified: 2026-05-13 method=go-test tests=TestRegisterResumeValidationErrorsReturnUnprocessableEntity,TestListResumesInvalidCursorReturnsUnprocessableEntity,TestResumeRegisterListHTTPValidationScenario -->
- [x] 6.2 `resume.parse` retryable failure 每次写 `parse_status='failed' + error_code`，retry metadata 仍由 `async_jobs` 表达，并允许 failed asset 重试回 processing 后 ready（验证：job/store/cmd-api retry tests PASS）<!-- verified: 2026-05-13 method=go-test tests=TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox,TestParseHandlerRetriesFailedAssetBackToProcessing,TestParseStatusTransition,TestResumeParseRunnerRetryableFailureScenario -->
- [x] 6.4 收口验证：仓库根 `make test` 完成前后端全量单测回归；`make docs-check`、`sync-doc-index --check` 与 `git diff --check` 作为独立 gates；focused Go tests 仅作开发反馈。

## Phase 7: D-20 简历扁平化适配（resumes / resumeId / structured_profile）

> product-scope D-20 / backend-resume D-13。依赖 B4 002 Phase 6 + B2 004 Phase 7。Red 优先。

- [x] 7.1 store 文件名与表名对齐为 `resumes.go`：读写 `resumes`、`structured_profile` / `display_name`；`source_type` 当前收敛 {`upload`,`paste`}，范围外输入只作为 validation negative（验证：`cd backend && go test ./internal/resume/store -run 'TestCreateWithParseJob|TestCompleteParse|TestParseStatusTransition' -count=1` PASS）
- [x] 7.2 handler register/get/list 使用 generated `Resume`、`ResumeWithJob`、`resumeId`、`PaginatedResume`；register 仅支持 `upload` / `paste`，unsupported input 返回 validation error（验证：handler unit test + `cmd/api` wiring test PASS）
- [x] 7.3 parse job 写 `resumes.structured_profile`（无 master 确认），`resume.parse.completed` envelope 使用 `resumeId`（验证：parse job unit test + outbox envelope test PASS）

## Phase 8: LLM-derived display_name for ready resumes

- [x] 8.1 `backend/internal/resume/jobs/parse.go` 从 LLM structured output 派生可识别 `display_name`，过滤通用上传 / 粘贴标题（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1` PASS）
- [x] 8.2 `CompleteParseSuccess` 在 ready 事务中写入派生 `display_name`，无法派生时保留空值，不回退到注册 title 或 raw resume 第一行（验证：`cd backend && go test ./internal/resume/store -run 'TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）

## Phase 9: Upload file readable text snapshot

- [x] 9.1 `backend/internal/resume/jobs/parse.go` 对 upload source 提取 PDF / Markdown / text 可读正文，AI prompt input 与 `parsed_text_snapshot` 使用同一正文，不能使用文件名或二进制 bytes；DOCX 不再属于当前上传白名单（验证：`cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerExtractsReadableUploadText|TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' -count=1` PASS）
- [x] 9.2 `CreateWithParseJob` 创建 queued resume 时只保存来源 `title`，不写 `display_name`；ready 后只由 parse success 写入 LLM-derived `display_name`（验证：`cd backend && go test ./internal/resume/store -run 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady|TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' -count=1` PASS）
- [x] 9.3 `resume.parse` bounded reader 读取完整合法对象后提取正文，不因头部截断丢失 PDF 尾部结构，并拒绝 literal/binary 乱码；focused test 使用小型合成 PDF 与小型 injected overflow，不以真实文件或默认尺寸作为门禁。
- [x] 9.4 `resume.parse` 在已抽取正文后遇到 AI provider / AI output 失败，仍写入 `parsed_text_snapshot` 供只读详情显示，且不发 completed event（验证：`cd backend && go test ./internal/resume/jobs -run TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox -count=1` PASS；`cd backend && go test ./internal/resume/store -run TestCompleteParseFailureCanPersistExtractedTextSnapshot -count=1` PASS；local UAT 真实 PDF `parse_status=failed` / `AI_OUTPUT_INVALID` 时 snapshot_len=3083）

## Phase 10: Display name robustness and prompt contract hardening

- [x] 10.1 `resume.parse` prompt schema / prompt body 显式要求 required `displayName`，并更新 prompt hash；验证: `make lint-prompts`。<!-- verified: 2026-07-07 method=lint command="make lint-prompts" -->
- [x] 10.2 `decodeResumeParseResponse` 优先使用 AI `displayName`，并拒绝通用上传/粘贴标题、上传文件名和 raw 第一行直出；验证: `cd backend && go test ./internal/resume/jobs -run TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox -count=1`。
- [x] 10.3 AI provider / output 失败但已有可读正文时，`CompleteParseFailure` 同时写入 fallback `display_name`，确保 failed-with-snapshot 详情不再长期显示“名称生成中”；验证: `cd backend && go test ./internal/resume/jobs -run TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox -count=1` 与 `cd backend && go test ./internal/resume/store -run TestCompleteParseFailureCanPersistExtractedTextSnapshot -count=1`。
- [x] 10.4 `ResumeDetailView` 对 `failed` 或已有正文的上传详情停止轮询 `getResume`，避免同一详情 URL 重复请求；验证: `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx`。

## Phase 11: Markdown snapshot and active resume limits

- [x] 11.1 `resume.parse` prompt schema / prompt body required `markdownText`，要求 LLM 保持原简历结构和事实并输出 Markdown；验证: `make lint-prompts`。<!-- verified: 2026-07-07 method=prompt-lint command="python3 scripts/lint/prompt_lint.py --prompts-dir config/prompts --migrations-dir migrations" -->
- [x] 11.2 `decodeResumeParseResponse` 校验并返回 `markdownText`，parse success 将 `ParsedTextSnapshot` 写为 Markdown；验证: `cd backend && go test ./internal/resume/jobs -run 'TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox|TestParseHandlerRequiresMarkdownTextInAIResponse' -count=1`。
- [x] 11.3 `RegisterResume` / `CreateWithParseJob` 使用注入 active limit；small-limit focused tests 覆盖新 IK 拒绝、零 resume/job、same-IK replay 与并发原子性；默认值归 A4。
- [x] 11.4 HISTORICAL-SUPERSEDED: 删除 backend-resume 对 upload default/RuntimeConfig 传播的重复测试；A4 owns config，backend-upload/frontend owners use focused guards。
- [x] 11.5 AI provider / output validation 失败但已抽取 PDF 可读正文时，失败态 `parsed_text_snapshot` 保存 Markdown fallback（标题、章节、技能 bullet），不保存原始折叠文本；验证: `cd backend && go test ./internal/resume/jobs -run TestParseHandlerMarkdownFallbackSurvivesPDFAIOutputFailure -count=1`。

## Phase 12: Source-format preview and DOCX exclusion

- [x] 12.1 `createUploadPresign` / upload register 对 `purpose=resume` 仅允许 PDF / Markdown / TXT，DOCX 在 presign/register 前返回 validation error；验证: upload handler/service focused tests。
- [x] 12.2 `resume.parse` 删除 DOCX 解包路径，历史 DOCX object 误入 parse 时返回 unsupported source text error，不进入 AI prompt；验证: parse focused tests。
- [x] 12.3 `getResumeSource` 按 `user_id + resumeId` scoped lookup upload-backed PDF 原件并返回 inline `application/pdf`；paste / Markdown / TXT / missing / archived / cross-user 返回 404；验证: store/service/handler/cmd-api focused tests。

## Phase 13: Resume parse output-capacity handoff

- [x] 13.1 OWNER-GATE: A3 catalog/code-default contract 保证 active profile 至少 16K；backend-resume 不复制 profile 默认数值测试。
- [x] 13.2 FOCUSED-GATE: backend-resume 只用小型 stub response 验证 structured output 截断与 `finish_reason=length` fail-closed、deterministic snapshot 和 ready-only outbox；focused run 仅作开发反馈，阶段完成由根 `make test` 承接。

## Phase 14: Deterministic full-resume snapshot and truncation fail-closed

- [x] 14.1 RED：长简历输入末尾唯一 marker 必须进入 AI prompt；stub AI 返回不含 `markdownText` 的 structured-only JSON，当前 decoder 因旧 required 字段失败（验证：focused Go test RED）
- [x] 14.2 GREEN：成功/失败 `parsed_text_snapshot` 均由完整提取正文确定性构建；decoder、prompt、schema 删除 `markdownText` 回显合同，长输入尾 marker 在 prompt 与 snapshot 中均保留；focused Go tests 仅作开发反馈，阶段完成由根 `make test` 承接。
- [x] 14.3 RED/GREEN：`FinishReason="length"` 在 JSON decode 前映射 `AI_OUTPUT_INVALID`，保留含尾 marker 的完整 snapshot，不发 `resume.parse.completed`；focused Red/Green 仅作开发反馈，阶段完成由根 `make test` 承接。
- [x] 14.4 同步 prompt body/schema/hash、baseline seed migration、eval cases 与 `resolved-prompts.json`；`make lint-prompts` / `make eval-offline-resolve` PASS，当前合同负向 grep 不再要求 `markdownText`<!-- verified: 2026-07-12 method=prompt-lint+eval-offline result=24/24-pass seed-body=matched -->

## Phase 15: Closed ResumeSummary list projection

- [x] 15.1 前置 gate：B2 owner 已原地新增 `ResumeSummary`，保持现有 `PaginatedResume` 与 `pageInfo` 不变并仅将 `items` 改为 `ResumeSummary[]`；default/empty/paginated fixtures 与 generated Go/TS artifacts 同步；没有新增 pagination wrapper，OpenAPI inventory/codegen/fixture drift 为零。 <!-- verified: 2026-07-14 method=generated-contract+fixture-validation tests=backend/cmd/codegen/openapi,validate-fixtures -->
- [x] 15.2 RED store tests：当前 list SQL/row scanner 在 exact projection gate 下失败；断言显式 summary 列、stable cursor、user scope、active-only，并禁止选择/扫描 `original_text|parsed_text_snapshot|structured_profile|file_object_id|parsed_summary object|created_at|deleted_at` 或 `SELECT *`；NULL `source_type` scan 与非法 enum service mapping 必须 fail closed，create path 继续写合法非空 source。 <!-- verified: 2026-07-14 method=go-test-red tests=TestListUsesClosedResumeSummaryProjection,TestListRejectsNullSourceType,TestCreateWithParseJobRejectsInvalidSourceType,TestDuplicateResumeRejectsInvalidSourceType observed=full-detail-record-build-failure -->
- [x] 15.3 GREEN store：实现独立 summary row/query；`summaryHeadline` 只投影 trim 后非空 headline；`hasReadableContent` 仅在 trim 后 snapshot 非空、trim 后 original text 非空或 structured profile 为非空 object 时为 true，空白文本/空对象为 false，且不按 file/source/status 推断；原始正文/JSON/file object 不装入 list result；`Get` full-detail scanner 保持独立。 <!-- verified: 2026-07-14 method=go-test tests=internal/resume/store summary-projection+null-source+create-duplicate-source-integrity -->
- [x] 15.4 Service/domain：`List` 只映射 closed `ResumeSummary`，覆盖 exact fields、null headline、readable true/false、25+ rows、第二页、invalid cursor、cross-user；`Get` 仍返回 full `Resume` 详情。验证：service tests。 <!-- verified: 2026-07-14 method=go-test tests=TestGetAndListResumesMapStoreRecordsWithUserScope,TestListResumesMapsNullHeadlineAndUnreadableContent,TestListResumesRejectsInvalidSourceType,TestListCursorPagination -->
- [x] 15.5 Handler/fixture：generated `PaginatedResume` 外层与 `pageInfo` 保持不变，`items: ResumeSummary[]` exact JSON keys 与 default/empty/paginated fixture 字节一致；逐项断言详情字段 absent；get fixture parity 保持完整详情。Handler / `TestResumeRegisterListHTTPContract` 仅作开发反馈，阶段单测完成由根 `make test` 承接。
- [x] 15.6 负向 gate：list store/service/handler 不复用 full-detail mapper，不出现 `SELECT *`，不把 forbidden fields 写入 response；Go tests、OpenAPI codegen/inventory、frontend typecheck 与 fixture parity PASS。

## Phase 16: Injected Resume content guards

- [x] 16.1 OWNER-GATE: active/upload/paste/extracted missing/default/override/invalid 与跨字段约束只由 A4 typed contract 覆盖；本 owner 删除重复 loader/composition tests。
- [x] 16.2 FOCUSED-GATE: 使用小型 injected limits 验证 active-count 并发/replay、paste/extracted overflow 在 resume/job/provider 前零副作用；不构造默认大小文件或字符串。

## BDD Gate

- [x] BDD-Gate: `BDD.RESUME.ASSET.001` 由 [BDD checklist](./bdd-checklist.md) 关联 register/parse/list/detail owner behavior tests；不创建或声明真实 E2E PASS。
