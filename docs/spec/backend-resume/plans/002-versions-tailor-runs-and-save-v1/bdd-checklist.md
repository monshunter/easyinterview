# 002 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-17

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.074 resume confirm master and version reads

- [x] 创建场景目录 `test/scenarios/e2e/p0-074-resume-confirm-master-and-version-reads/`，含 `README.md`（baseline + 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 B2 fixture scenario + 测试数据：`confirmResumeStructuredMaster.json` 全 4 scenario、`getResumeVersion.json` `default` / `not-found-404`、`listResumeVersions.json` `default` / `empty` / `paginated`；`cmd/api` focused HTTP scenario 覆盖 session / IK / confirm / get / list route；handler / service / store tests 覆盖用户 A / B、blank displayName、parse_not_ready、cross-user 404、invalid cursor；store integration 直接 seed ready / processing asset 与 25 行版本来覆盖 partial UNIQUE、soft-delete replacement 与 pagination（Phase 5 branch endpoint 尚未落地，P0.074 只验证 Phase 2 + 3 已实现行为）
- [x] 实现 `scripts/setup.sh`（准备 `.test-output` 与 seed/expected evidence）/ `scripts/trigger.sh`（运行 `make validate-fixtures`、`cd backend && go test ./cmd/api -run 'TestResumeConfirmStructuredMasterHTTPScenario|TestResumeVersionReadHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v`、handler fixture parity、service/store focused tests、`DATABASE_URL=... go test ./internal/resume/store -tags=integration -run 'TestStructuredMasterUnique|TestResumeVersionListPagination' -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` + no-op / skip 不可 PASS + DB state + partial UNIQUE INDEX 兜底 + 字节比对 fixture + cross-user 404 + cursor 序稳定性 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`（记录 cleanup，Go tests 自清理数据库行）
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-074-resume-confirm-master-and-version-reads/trigger.log` + `cmd/api` HTTP scenario log + verify 输出 + DB 行 dump + partial UNIQUE INDEX 兜底证据 + fixture byte diff 0 + cross-user 404 + cursor 序证据 + 隐私 grep 0 命中 + `method=cmd-api-http` 或等价 live runtime evidence + no no-op / no skip 证据
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.074 行（关联需求 `backend-resume C-6, C-14, C-15`，状态 Ready，automated）

## E2E.P0.075 resume update version merge and IK

- [x] 创建场景目录 `test/scenarios/e2e/p0-075-resume-update-version-merge-and-ik/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 B2 fixture + 测试数据：`updateResumeVersion.json` `default` / `validation-error-422`，并由本 plan Phase 4.7 补齐 `idempotency-replay` fixture；`cmd/api` focused HTTP scenario 覆盖 session / IK / update route / replay / mismatch / server-owned 422 / not-found；handler / service / store tests 覆盖用户 A / B、partial merge、client provenance 剥离、cross-user 404、deleted row 404；store integration 直接 seed ready asset、active version 与 soft-deleted version
- [x] 实现 `scripts/setup.sh`（准备 `.test-output` 与 seed/expected evidence）/ `scripts/trigger.sh`（运行 `make validate-fixtures`、`cd backend && go test ./cmd/api -run 'TestResumeUpdateVersionHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v`、handler fixture parity、service/store focused tests、`DATABASE_URL=... go test ./internal/resume/store -tags=integration -run 'TestResumeVersionUpdatePatch|TestStructuredMasterUnique|TestResumeVersionListPagination' -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` + no-op / skip 不可 PASS + DB merge semantic + server-reset provenance + 字节比对 fixture + cross-user / deleted 404 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-075-resume-update-version-merge-and-ik/trigger.log` + verify 输出 + DB merge 轨迹 + fixture byte diff 0 + cross-user 404 + deleted 404 + 隐私 grep 0 命中 + `method=cmd-api-http` live runtime evidence + no no-op / no skip 证据
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.075 行（关联需求 `backend-resume C-14`，状态 Ready，automated）

## E2E.P0.076 resume branch version sync paths

- [x] 创建场景目录 `test/scenarios/e2e/p0-076-resume-branch-version-sync-paths/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 B2 fixture + 测试数据：`branchResumeVersion.json` `copy-master-sync` + `blank-sync` + `default` / `idempotent-replay` / `validation-error-422` / `ai-select-202-with-job`；`cmd/api` focused HTTP scenario 覆盖 session / IK / copy_master / ai_select accepted / replay / mismatch / invalid seed / not found；handler / service / store tests 覆盖用户 A / B、copy provenance reset、blank profile、cross-user parent 404、foreign target 404、ai_select rollback
- [x] 实现 `scripts/setup.sh`（准备 `.test-output` 与 seed/expected evidence）/ `scripts/trigger.sh`（运行 `make validate-fixtures`、`cd backend && go test ./cmd/api -run 'TestResumeBranchVersionHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v`、handler fixture parity、service/store focused tests、`DATABASE_URL=... go test ./internal/resume/store -tags=integration -run TestBranchVersion -count=1 -v`）/ `scripts/verify.sh`（DB structured_profile 拷贝断言 + blank 空字段 + 不入队 async_jobs 断言 + 字节比对 fixture + cross-user 404 + `method=cmd-api-http` + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-076-resume-branch-version-sync-paths/trigger.log` + verify 输出 + DB resume_versions / async_jobs 行断言 + fixture byte diff 0 + 隐私 grep 0 命中 + `method=cmd-api-http` live runtime evidence + no no-op / no skip 证据
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.076 行（关联需求 `backend-resume C-10`，状态 Ready，automated）

## E2E.P0.077 resume tailor async dispatch and ready

- [x] 创建场景目录 `test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] Phase 5 dispatch fixture + 测试数据：`branchResumeVersion.json` `ai-select-202-with-job`；`cmd/api` focused HTTP scenario 覆盖 `seedStrategy=ai_select` 202 accepted + queued job；handler / service / store tests 覆盖 provisional version、queued `resume_tailor_runs`、queued `async_jobs(resume_tailor)`、rollback、cross-user target isolation
- [x] Phase 5 dispatch scripts：`scripts/setup.sh`（准备 `.test-output` 与 seed/expected evidence）/ `scripts/trigger.sh`（运行 `make validate-fixtures`、`cd backend && go test ./cmd/api -run TestResumeBranchVersionHTTPScenario -count=1 -v`、handler fixture parity、service branch gate、live DB integration）/ `scripts/verify.sh`（skip/no-op negative、fixture parity、privacy grep、retired-vocabulary grep、`method=cmd-api-http`）/ `scripts/cleanup.sh`
- [x] Phase 5 dispatch subset 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录 Phase 5 dispatch 验证证据：`.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready/trigger.log` + verify 输出 + `branchResumeVersion` `ai-select-202-with-job` fixture parity + DB provisional version / queued run / queued async job / rollback 断言 + privacy grep 0 命中 + live runtime evidence + no no-op / no skip
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.077 行（关联需求 `backend-resume C-10, C-16`，状态 Ready，automated；描述明确为 Phase 5 dispatch slice，ready/drainer extension pending Phase 7）
- [x] Phase 6 扩展 requestTailor / getResumeTailorRun：补齐 `requestResumeTailor.json` `default`（含 `Idempotency-Key` 请求头）/ `idempotency-replay` + `getResumeTailorRun.json` `default` / `queued` / `generating` / `failed`，并扩展脚本覆盖 requestTailor + getTailorRun 部分；执行 `setup → trigger → verify → cleanup` PASS，证据位于 `.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready/trigger.log` / `verify.log`
- [x] Phase 7 扩展 drainer ready path：注入 A3 AIClient stub success JSON（matchSummary + suggestions[3]），覆盖 `RunOnce(resume_tailor)`、`resume_version_suggestions`、`ai_task_runs`、ready-only outbox 与 outbox payload PII 边界（验证：`TestResumeTailorDrainerHTTPScenario`、`TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox`、live `TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox` PASS）
- [x] 完整 `E2E.P0.077` 执行 `setup → trigger → verify → cleanup` 全 PASS（branch ai_select + requestTailor + getTailorRun queued/ready + drainer + outbox）（验证：`.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready/trigger.log` / `verify.log`；verify 输出 `method=cmd-api-http`、ready suggestions、ready-only completed outbox、privacy grep 0 命中；shell `LC_ALL=C.UTF-8` locale warning 非阻塞）

## E2E.P0.078 resume tailor failure and retry

- [x] 创建场景目录 `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 stub provider 三 variant（timeout / output_invalid / retry 成功）+ 测试数据：用户 A 3 个 queued resume_tailor_runs 对应三 variant；F3 feature_key ready（验证：`TestResumeTailorDrainerFailureScenario` + `TestTailorHandlerModeRoutingAndFailurePaths`）
- [x] 实现 `scripts/setup.sh`（准备 `.test-output` 与 seed/expected evidence）/ `scripts/trigger.sh`（A drainer `RunOnce` 处理 timeout + B invalid output + C retry success，运行 `cd backend && go test ./cmd/api -run TestResumeTailorDrainerFailureScenario -count=1 -v`）/ `scripts/verify.sh`（error_code + ai_task_runs 多行 + ready-only outbox + 旧口径 grep + 隐私 grep + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-078-resume-tailor-failure-and-retry/trigger.log` + timeout retryable / output_invalid terminal / retry-to-ready + `ai_task_runs` 多行断言 + outbox ready-only 断言 + 隐私 grep 0 命中 + live runtime evidence
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.078 行（关联需求 `backend-resume C-16`，状态 Ready，automated）

## E2E.P0.079 resume suggestion accept reject terminal

- [x] 创建场景目录 `test/scenarios/e2e/p0-079-resume-suggestion-accept-reject-terminal/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture + 测试数据：`acceptResumeTailorSuggestion.json` / `rejectResumeTailorSuggestion.json` `default` / `idempotency-replay` / `already-decided-409`（缺 variant 由本 plan Phase 8.7 补齐 fixture，并将当前 `conflict-409` + `TARGET_INVALID_STATE_TRANSITION` 漂移替换为 `VALIDATION_FAILED` + `detail.reason='SUGGESTION_ALREADY_DECIDED'`）；用户 A 拥有 ready targeted version + pending suggestions（带初始 structured_profile 状态供 D-12 断言）；用户 B 拥有自己 suggestion（验证：`make validate-fixtures` PASS；`TestResumeSuggestionDecisionCASIsolationAndProfileStability` live integration PASS）
- [x] 实现 `scripts/setup.sh`（准备 `.test-output` 与 seed/expected evidence）/ `scripts/trigger.sh`（运行 `make validate-fixtures`、`cd backend && go test ./cmd/api -run TestResumeSuggestionAcceptRejectHTTPScenario -count=1 -v`、handler fixture parity、service decision tests、live store CAS integration）/ `scripts/verify.sh`（suggestions status 状态机 + decided_at + **`resume_versions.structured_profile` 不被改动**（D-12）+ IK middleware replay + 字节比对 fixture + cross-user 404 + 隐私 grep + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS（shell `LC_ALL=C.UTF-8` locale warning 非阻塞）
- [x] 记录验证证据：`.test-output/e2e/p0-079-resume-suggestion-accept-reject-terminal/trigger.log` + DB suggestion 状态机轨迹 + `resume_versions.structured_profile` 不变断言 + IK replay 证据 + cross-user 404 + 隐私 grep 0 命中 + live runtime evidence
- [x] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.079 行（关联需求 `backend-resume C-16`，状态 Ready，automated）

## E2E.P0.080 resume versions privacy and legacy negative

- [ ] 创建场景目录 `test/scenarios/e2e/p0-080-resume-versions-privacy-legacy/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备前置：E2E.P0.074-079 均已 PASS（依次写入 resume_versions / resume_tailor_runs / resume_version_suggestions / ai_task_runs / outbox_events / audit_events 行）；本场景只做 regression / privacy negative 反查
- [ ] 实现 `scripts/setup.sh`（dev stack 拉起 + 重放前序场景数据 fixture 或共享 DB state）/ `scripts/trigger.sh`（grep 检查：A `git grep -nE 'inline|rewrite|mirror' backend/internal/resume/` + B `git grep -nE 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume/` + C `cd backend && go test ./internal/resume/... -run 'TestOutboxPrivacy|TestAuditPrivacy|TestAiTaskRunsPrivacy' -count=1 -v` + D outbox payload assertion unit test 重跑；当前 plan / BDD prose 与历史 out-of-scope 文档不纳入 zero-reference gate，避免说明文字自匹配）/ `scripts/verify.sh`（断言 grep 0 命中 + outbox payload 字段集严格匹配 + ai_task_runs 不含 raw text / suggested bullet + audit_events 不含 prompt body + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-080-resume-versions-privacy-legacy/trigger.log` + grep 输出 0 命中 + outbox payload 字段集合断言 + ai_task_runs / audit_events PII grep 输出 + live runtime evidence
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.080 行（关联需求 `backend-resume C-13`，状态 Ready，automated）
