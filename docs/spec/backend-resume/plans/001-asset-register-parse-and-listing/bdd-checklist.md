# 001 BDD Checklist

> **版本**: 1.15
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.034 resume register + get + list

- [x] 创建场景目录 `test/scenarios/e2e/p0-034-resume-register-and-list/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 B2 fixture scenario + 测试数据：`registerResume` `default` / `paste-text`、`getResume` `default` / `not-found`、`listResumes` `default` / `empty` / `paginated`；2 个测试用户（A / B）；用户 A 通过 backend-upload createUploadPresign (`purpose=resume`) + PUT 取得 fileObjectId；backend-upload `RegisterFileObject` 已具备 object `Stat` + actual size mismatch rejection；准备 25 个 resume 用于 pagination（混合 sourceType）
- [x] 实现 `scripts/setup.sh`（A2 dev stack + backend-upload 拉起 + `cmd/api` resume route + session / IK middleware + 用户登录 + presign upload + 批量 register）/ `scripts/trigger.sh`（依序触发 A1/A2/A3/B1/C1/C2/D1/D2，并运行 `cd backend && go test ./cmd/api -run TestResumeRegisterListHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` 或等价 live runtime evidence + no-op / skip 不可 PASS + DB state + resume/job 同事务 + missing object / size mismatch 不创建 resume + cross-user 404 + cursor 分页边界 + 精确 fixture scenario 字节比对 + validation 直接断言 + 隐私 grep + 当前范围负向 grep）/ `scripts/cleanup.sh`（清理 file_objects + resumes + async_jobs + 用户）
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS <!-- verified: 2026-05-13 method=scenario log=.test-output/e2e/p0-034-resume-register-and-list/trigger.log -->
- [x] 记录验证证据：`.test-output/e2e/p0-034-resume-register-and-list/trigger.log` + `cmd/api` HTTP scenario log + verify 输出 + DB state machine 轨迹 + resume/job 同事务证据 + missing object / size mismatch rejection 证据 + register/get/list fixture byte diff 0 + validation error direct assertion + 隐私 grep 0 命中 + cross-user 404 验证 + cursor 序稳定性 + `method=cmd-api-http` 或等价 live runtime evidence + no no-op / no skip 证据
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.034 行（关联需求 `backend-resume C-1, C-2, C-5, C-6, C-7, C-8`，状态 Ready，automated）
- [x] L2 remediation：trigger/verify 检查 `TestRegisterResumeValidationErrorsReturnUnprocessableEntity`、`TestListResumesInvalidCursorReturnsUnprocessableEntity`，证明 upload validation / invalid cursor 不会被 500 掩盖 <!-- verified: 2026-05-13 method=scenario log=.test-output/e2e/p0-034-resume-register-and-list/trigger.log -->
- [x] D-16 remediation：trigger/verify 或 focused substitute gate 检查 `resume.maxActive` 达到默认 10 时新 register 返回 422、不创建 resume/job，且同 IK replay 不误拒。<!-- verified: 2026-07-07 method=focused-substitute tests=TestRegisterResumePassesConfiguredActiveLimitToStore,TestCreateWithParseJobRejectsNewResumeWhenActiveLimitReached,TestCreateWithParseJobAllowsIdempotentReplayAtActiveLimit -->

## Phase 15 closed summary and cross-owner consumers

- [x] P0.034 seed 为 list rows 填充正文、snapshot、structured profile、file object、parsed summary 与审计时间，确保 forbidden-field absence 不是因为源数据为空。 <!-- verified: 2026-07-14 method=scenario-seed result=PASS -->
- [x] P0.034 trigger 执行 store summary projection、service mapper、handler JSON exact-key、full get detail 与真实 cmd/api register/get/list tests；verify 检查所有 exact PASS marker 并拒绝 skip/no-op。 <!-- verified: 2026-07-14 method=scenario result=PASS evidence="method=cmd-api-http; projection/service/handler/full-get markers present" -->
- [x] P0.034 verify 断言每个 summary exact keys 为 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`，逐项 forbidden detail fields absent，分页与 cross-user 仍通过。 <!-- verified: 2026-07-14 method=scenario-verify result=PASS evidence="exact keys=9; scalar SQL only; pagination and cross-user PASS" -->
- [x] P0.036 fixture/client 保持 `PaginatedResume` 外层，仅将 `items` 改为 `ResumeSummary[]`；list/Home selector 不读取正文/structured profile；`summaryHeadline` / `hasReadableContent` 覆盖可见列表需求。 <!-- verified: 2026-07-14 method=scenario result=PASS evidence="5/5; summaryFields=9; list transport=1" -->
- [x] P0.037 证明 list payload 不预取或透传 full detail，点击 row 后才由 `getResume` 提供正文与结构化详情。 <!-- verified: 2026-07-14 method=scenario+browser result=PASS evidence="8/8; get before open=0, after open=1; ready detail initial=1 maxInFlight=1" -->
- [x] 执行 P0.034、P0.036、P0.037 `setup → trigger → verify → cleanup` 全 PASS，并记录 fixture key diff、store projection SQL evidence、full get parity 与真实 scenario logs。 <!-- verified: 2026-07-14 method=scenario-lifecycle result=PASS evidence="all three fresh on current source; P0.034 full lifecycle and P0.036/P0.037 wrappers PASS" -->

## E2E.P0.035 resume.parse async job lifecycle

- [x] 创建场景目录 `test/scenarios/e2e/p0-035-resume-parse-async-job-lifecycle/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture + stub provider：A3 AIClient stub provider 3 response variant（success JSON / output_invalid / timeout）；`cmd/api` in-process resume runner kernel 启动并注册 resume.parse handler；F3 `resume.parse` feature_key prompt / rubric / profile 就位；缺少 live env、runner kernel no-op 或 focused gate skip 时本场景必须 fail
- [x] 实现 `scripts/setup.sh`（`cmd/api` runtime + in-process runner kernel + stub provider variant 注入 + 用户登录 + register 3 个 resume_asset 对应三 variant + upload PDF/Markdown/text seed）/ `scripts/trigger.sh`（runner kernel `RunOnce` / 等价 deterministic stepping + 等待解析完成 / 失败 / 重试 + 运行 `cd backend && go test ./cmd/api -run TestResumeParseRunnerHTTPScenario -count=1 -v` 和 upload extraction/store focused tests）/ `scripts/verify.sh`（断言 `method=cmd-api-http` 或等价 live runtime evidence + no-op / skip 不可 PASS + parse_status 状态机 + queued `display_name` 为空 + LLM-derived `display_name` 不回退到 title/raw 第一行 + upload readable `parsed_text_snapshot` + DOCX 前置拒绝 + ai_task_runs 多行 + ready-only outbox event 写入 + failure no completed event + shutdown 无 goroutine leak + 隐私 grep + 范围外输入 grep）/ `scripts/cleanup.sh` <!-- verified: 2026-07-07 method=scenario scenario=E2E.P0.035 -->
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS <!-- verified: 2026-05-13 method=scenario log=.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log -->
- [x] 记录验证证据：`.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log` + `cmd/api` runner kernel scenario log + verify 输出 + DB parse_status 转换轨迹 + ai_task_runs 行 dump + outbox_events 行 dump（ready-only completed event，failure no completed event，PII grep 0 命中）+ stub provider call log + shutdown / no goroutine leak 证据 + DOCX 前置拒绝证据 + `method=cmd-api-http` 或等价 live runtime evidence + no no-op / no skip 证据
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.035 行（关联需求 `backend-resume C-3, C-4, C-13`，状态 Ready，automated）
- [x] L2 remediation：trigger/verify 检查 `TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox`、`TestParseHandlerRetriesFailedAssetBackToProcessing`、`TestResumeParseRunnerRetryableFailureScenario`，证明 retryable timeout 先落 failed/error_code 再重试成功 <!-- verified: 2026-05-13 method=scenario log=.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log -->
- [x] D-14/D-15 remediation：trigger/verify 检查 `TestParseHandlerExtractsReadableUploadText`、`TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox`、`TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady`、`TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically`、`TestCompleteParseFailureCanPersistExtractedTextSnapshot`、`TestResumeParseRunnerHTTPScenario` 和 `TestResumeParseRunnerRetryableFailureScenario`，证明 queued 不预填 `displayName`，ready / retry-to-ready 后写入 LLM-derived `displayName` 而非通用上传 / 粘贴标题或 raw 第一行，failed-with-snapshot 失败态写入正文派生 fallback `display_name`，且 upload `parsed_text_snapshot` 来自 PDF/Markdown/text 可读正文。 <!-- verified: 2026-07-07 method=scenario scenario=E2E.P0.035 -->
- [x] D-17 remediation：trigger/verify 或 focused substitute gate 检查 prompt schema required `markdownText`、decode 缺失拒绝、success `parsed_text_snapshot` 使用 AI Markdown 输出。<!-- verified: 2026-07-07 method=focused-substitute tests=TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox,TestParseHandlerRequiresMarkdownTextInAIResponse,prompt_lint -->
- [x] D-15/D-18 remediation：trigger/verify 或 focused substitute gate 检查 Resume upload DOCX 前置拒绝、parse fallback 拒绝 DOCX、`getResumeSource` 仅对当前用户 upload-backed PDF 返回 inline PDF，非 PDF/跨用户/missing 返回 404。<!-- verified: 2026-07-07 method=focused-substitute tests=upload handler/service,resume jobs,resume service/handler/store/cmd-api -->
- [x] D-19 long-resume budget：新增 `TestCatalogKeepsResumeParseOutputBudget`，从 repo-tracked profile catalog 断言 `resume.parse.default.max_tokens >= 8192`<!-- verified: 2026-07-12 method=go-test-red-green -->
- [x] D-19 runner evidence：P0.035 trigger/verify 执行并检查 budget regression 测试名、runner marker、PASS，拒绝 no-op / skip<!-- verified: 2026-07-12 method=scenario trigger_log=.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log -->
- [x] D-19 execution：执行 P0.035 `setup → trigger → verify → cleanup` 并记录 long-resume budget gate 证据<!-- verified: 2026-07-12 method=scenario result=PASS cleanup=PASS -->
- [x] D-17/D-21 contract：stub success 改为 structured-only JSON；长输入尾部唯一 marker 同时进入 AI prompt 与 deterministic `parsed_text_snapshot`，当前 prompt/schema 不再要求 `markdownText`<!-- verified: 2026-07-12 method=go-test+prompt-lint -->
- [x] D-21 truncation：stub `FinishReason="length"` 必须在 decode 前落 `AI_OUTPUT_INVALID`，完整 snapshot 保留且 completed outbox 为零<!-- verified: 2026-07-12 method=mutation-red-green -->
- [x] D-17/D-21 runner evidence：P0.035 trigger/verify 检查 long-input tail-marker、structured-only response、finish-reason fail-closed 测试名与 PASS，拒绝 no-op / skip<!-- verified: 2026-07-12 method=scenario -->
- [x] D-17/D-21 execution：执行 P0.035 `setup → trigger → verify → cleanup` 并记录完整输入、deterministic snapshot 与 output truncation terminality 证据<!-- verified: 2026-07-12 method=scenario result=PASS cleanup=PASS -->

## Phase 16 configured Resume content boundaries

- [x] P0.034 proves configured active/paste exact/limit+1 business boundaries with zero resume/job partial state; P0.081 locks the public 10MiB/384KiB defaults.
- [x] P0.035 proves configured extracted text reaches AI and +1 is rejected before AI while deterministic snapshot/truncation gates remain intact.
- [x] P0.081 consumes current runtime-config upload/paste limits; backend remains authoritative and no browser storage is used.
  <!-- verified: 2026-07-14 evidence="Fresh P0.034/P0.035/P0.081 current-source runs pass without committed large input files." -->
