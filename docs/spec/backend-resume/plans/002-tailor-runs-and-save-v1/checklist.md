# Backend Resume Tailor Runs and Save v1 Checklist

> **版本**: 1.13
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 0: current contract preflight

- [x] 0.1 读取 backend-resume spec、B2 OpenAPI/fixtures/generated artifacts、backend handler/store/job/runtime 和 P0.074-P0.080 场景，确认当前合同为 flat `resumes` + 10 个 Resume / ResumeTailor operation（验证：`make lint-openapi` PASS，当前 inventory 为 10 tags / 37 operations；`openapi/openapi.yaml` 仅包含 current Resume operationIds）
- [x] 0.2 context manifest 指向当前 backend owner surface，不再以 removed package path 作为主入口（验证：`python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/context.yaml --target backend` PASS）
- [x] 0.3 front/back contract pre-read 已覆盖 `docs/development.md` §2、`backend/README.md`、`openapi/README.md`、`test/scenarios/README.md` 与相关 scenario README（验证：本 plan 后续 gates 只使用 current OpenAPI/generated/fixture/runtime/scenario truth sources）

## Phase 1: flat API and removed-route boundary

- [x] 1.1 `getResume` / `listResumes` 使用 current flat `resumeId` contract，fixture parity、cursor order、cross-user 404 与 privacy negative 均由 focused tests 覆盖（验证：P0.074 trigger/verify 覆盖 `make validate-fixtures`、flat read handler/service/store tests 和 privacy checks）
- [x] 1.2 removed route family 不出现在 generated route catalog，runtime inputs 返回 404（验证：`go test ./backend/cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations' -count=1` PASS）
- [x] 1.3 BDD-Gate: `E2E.P0.074` flat resume read API 场景保持 Ready（验证：`test/scenarios/e2e/p0-074-resume-flat-read-api/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）

## Phase 2: `updateResume`

- [x] 2.1 `PATCH /api/v1/resumes/{resumeId}` 覆盖 `structuredProfile` / `displayName`，拒绝 server-owned fields，cross-user/missing row 返回 404（验证：`go test ./backend/internal/resume/... -run 'TestUpdateResume|TestUpdateResumeFixtureParity' -count=1` PASS）
- [x] 2.2 IK replay/mismatch 复用统一 idempotency middleware，响应与 B2 fixture 字节一致（验证：P0.075 trigger/verify 覆盖 `updateResume` fixture parity、IK gate、server-owned field 422）
- [x] 2.3 BDD-Gate: `E2E.P0.075` flat resume update and IK 场景保持 Ready（验证：`test/scenarios/e2e/p0-075-resume-update-flat-fields-and-ik/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）

## Phase 3: `duplicateResume`

- [x] 3.1 `POST /api/v1/resumes/{resumeId}/duplicate` 复制只读来源快照并应用 editable overlay，源 resume 不变（验证：`go test ./backend/internal/resume/... -run 'TestDuplicateResume|TestDuplicateResumeFixtureParity' -count=1` PASS）
- [x] 3.2 rollback、IK replay、invalid input、cross-user source 404 均有 focused tests 与 fixture parity（验证：P0.076 trigger/verify 覆盖 duplicate route gate、service/store tests、fixture parity 和 privacy checks）
- [x] 3.3 BDD-Gate: `E2E.P0.076` flat resume duplicate save-as-new 场景保持 Ready（验证：`test/scenarios/e2e/p0-076-resume-duplicate-save-as-new/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）

## Phase 4: `requestResumeTailor` / `getResumeTailorRun`

- [x] 4.1 `requestResumeTailor` 以 `resumeId` 创建 `async_jobs(job_type='resume_tailor')`，payload 只包含当前所需 ID/mode 字段（验证：`go test ./backend/internal/resume/... ./backend/cmd/api -run 'TestRequestResumeTailor|TestResumeTailorEndpointsHTTPScenario|TestResumeTailorFixtureParity' -count=1` PASS）
- [x] 4.2 `getResumeTailorRun` 从 async job status/result 返回 queued/generating/ready/failed；suggestions 来自 task output（验证：tailor handler/store tests + P0.077/P0.078 trigger/verify PASS）
- [x] 4.3 BDD-Gate: `E2E.P0.077` flat resume tailor async dispatch and ready 场景保持 Ready（验证：`test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）
- [x] 4.4 BDD-Gate: `E2E.P0.078` resume tailor failure and retry 场景保持 Ready（验证：`test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）

## Phase 5: resume.tailor async job and outbox

- [x] 5.1 tailor job 通过 A3 AIClient + F3 feature_key 路由，覆盖 success、timeout、output_invalid、retry-to-ready（验证：`go test ./backend/internal/resume/jobs -run 'TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox|TestTailorHandlerModeRoutingAndFailurePaths' -count=1` PASS）
- [x] 5.2 success 写 typed `ai_task_runs` 和 ready-only `resume.tailor.completed` outbox；payload allowlist 不含 prompt/raw resume/match summary/suggested bullet 文本（验证：`go test ./backend/internal/resume/store ./backend/cmd/api -run 'TestCompleteTailorRunSuccessWritesResultAndOutbox|TestResumeTailorRunnerHTTPScenario|TestResumeTailorRunnerFailureScenario' -count=1` PASS）
- [x] 5.3 BDD-Gate: `E2E.P0.080` tailor privacy and out-of-scope negative 场景保持 Ready（验证：`test/scenarios/e2e/p0-080-resume-tailor-privacy-negative/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）

## Phase 6: flat save fixture parity and read-only detail boundary

- [x] 6.1 flat save fixtures 保持 current backend contract，同时前端详情不再暴露 Rewrites/Edit 二次操作（验证：P0.079 trigger/verify 覆盖 read-only detail negative Vitest、flat save fixture parity 和 route boundary checks）
- [x] 6.2 BDD-Gate: `E2E.P0.079` flat save fixture parity + read-only detail boundary 场景保持 Ready（验证：`test/scenarios/e2e/p0-079-resume-rewrites-accept-only-save/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）
- [x] 6.3 docs/index/context 收口（验证：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS；`make docs-check` PASS；`git diff --check` PASS）
- [x] 6.4 product-scope pruning owner 记录本 phase 6.100 evidence，聚合 residual scan 不再把本 owner plan/checklist 识别为旧叙事热点（验证：`make lint-core-loop-pruning-surface` PASS，`real_residuals=0`）

## Phase 7: tailor provenance conversion simplification

- [x] 7.1 `VersionProvenance` 与私有 persisted wire type 的写入/读回使用显式同构转换，删除双向逐字段复制（验证：store/package tests、scoped/full backend `staticcheck`、owner context/docs gates）
  <!-- verified: 2026-07-10 method=resume-tailor-provenance-conversion-simplification evidence="The final backend S1016 red identified the write-side field copy. Both persisted write and readback mappings now use explicit conversions while retaining the private wire type; the readback test asserts all seven fields. Focused/store/full Resume and backend-wide Go tests PASS; scoped/backend-wide staticcheck and top-level make lint PASS; owner/product contexts, completed-state sync-doc-index/docs-check, diff and pruning gates PASS real_residuals=0." -->

## Phase 8: tailor scenario negative-gate precision

- [x] 8.1 `scenario_script_contract_test.py` 拒绝 P0.075-P0.080 使用裸 `inline|mirror` 搜索；六份 verify 统一使用 contextual tailor/mode 正则并排除 `*_test.go`，行为相关场景按变更批次串行通过。
  <!-- verified: 2026-07-10 method=tailor-scenario-negative-gate-precision evidence="RED: P0.077 verify rejected legal Content-Disposition inline; new contract test failed against P0.077/P0.078 broad grep. GREEN: all three verify scripts use contextual production regex and exclude *_test.go; contract tests 4 passed; P0.077/P0.078/P0.080 full scenario lifecycles PASS." -->
  <!-- verified: 2026-07-10 method=resume-mode-negative-gate-coverage evidence="Phase 10 exposed the same false positive in P0.075/P0.076. The contract now covers all six P0.075-P0.080 verify scripts; P0.075/P0.076 full lifecycles and the four-test script contract pass." -->

## Phase 9: current OpenAPI inventory wording

- [x] 9.1 Align the preflight inventory with the current 37-operation B2 contract while preserving the 10-operation Resume/ResumeTailor subset; verify OpenAPI inventory, fixtures, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=current-openapi-inventory-wording evidence="OpenAPI inventory and fixture validation report 10 tags, 37 operations and 37 fixtures; Resume/ResumeTailor remains 10 operations. Current owner plan/checklist search has no 35/36 inventory and backend context passes." -->

## Phase 10: flat Resume mutation handler pipeline consolidation

- [x] 10.1 RED: scoped `dupl` identifies `UpdateResume` / `DuplicateResume` as the only clone group in `internal/resume/handler`; the duplicated block covers user resolution through success response.
  <!-- red: 2026-07-10 method=resume-mutation-handler-dupl evidence="go run github.com/golangci/dupl@v0.0.0-20250308024227-f665c8d69b32 -t 100 backend/internal/resume/handler reports one production clone group between update.go and duplicate.go." -->
- [x] 10.2 Introduce one private typed mutation helper while retaining operation-specific service assertions, validators and success statuses; require scoped `dupl` zero clone groups and focused/full Resume tests.
  <!-- verified: 2026-07-10 method=resume-mutation-handler-helper-green evidence="handleResumeMutation centralizes the shared HTTP pipeline; both public handlers retain service assertions, typed validators and 200/201 statuses. Scoped handler dupl reports 0 groups and focused UpdateResume/DuplicateResume handler tests pass." -->
- [x] 10.3 BDD-Gate: run complete `E2E.P0.075` and `E2E.P0.076` lifecycles plus backend-wide Go, vet/staticcheck, owner/product context and docs/pruning gates.
  <!-- verified: 2026-07-10 method=resume-mutation-handler-closeout evidence="P0.075/P0.076 full serial lifecycles pass after repairing their overbroad mode negative gates; full Resume/cmd-api and backend Go tests, vet, staticcheck, both context validators, scenario contract, docs/index/diff and pruning gates pass. Scenario cleanup reports no owned long-lived resources." -->

## Phase 11: shared Resume mode negative gate

- [x] 11.1 RED: P0.075-P0.080 verify and P0.080 trigger contain seven copies of the same contextual regex, glob exclusion and failure branch; this duplication already allowed P0.075/P0.076 to drift.
  <!-- red: 2026-07-10 method=resume-mode-gate-copy-inventory evidence="rg over P0.075-P0.080 scripts reports six verify copies plus one P0.080 trigger copy of the same tailor/mode contextual regex." -->
- [x] 11.2 Add one `_shared` executable gate, replace all seven inline blocks, and make the scenario contract reject caller-owned regex copies.
  <!-- verified: 2026-07-10 method=shared-resume-mode-gate-green evidence="One executable helper owns the contextual regex, test-file exclusion, failure message and rg-error propagation. Seven callers contain zero inline regex copies; the four-test scenario contract, direct gate execution and bash -n across all touched scripts pass." -->
- [x] 11.3 BDD-Gate: run `bash -n`, scenario contract tests and serial P0.075-P0.080 setup/trigger/verify/cleanup lifecycles, then owner/product context and docs/pruning gates.
  <!-- verified: 2026-07-10 method=shared-resume-mode-gate-closeout evidence="All six P0.075-P0.080 setup/trigger/verify/cleanup lifecycles pass serially; cleanup succeeds for each. Scenario script/environment contracts pass 24 tests, touched shell parses, both contexts resolve and docs/index/diff/pruning gates pass with real_residuals=0." -->

## Phase 12: unified Resume runtime negative gate

- [x] 12.1 RED: six verify scripts and P0.080 trigger still copy the module vocabulary search; P0.075/P0.076 omit the production-only test exclusion used by the other callers.
  <!-- red: 2026-07-10 method=resume-module-gate-copy-and-drift evidence="rg reports seven inline mistakes|growth|drill|inline-debrief-record branches; only five of seven exclude *_test.go." -->
- [x] 12.2 Rename and broaden the shared helper to own both scans through one error-aware function; delete the old helper name and all caller-owned module regex blocks.
  <!-- verified: 2026-07-10 method=unified-resume-runtime-gate-green evidence="resume-runtime-negative-gate.sh owns both patterns through run_negative_scan; seven consumers have no inline module regex and no old-helper invocation, and the old file is absent. Contract tests, direct execution and bash -n pass." -->
- [x] 12.3 BDD-Gate: run source contract/negative checks, `bash -n`, serial P0.075-P0.080 lifecycles, contexts and docs/pruning gates.
  <!-- verified: 2026-07-10 method=unified-resume-runtime-gate-closeout evidence="Both shared scans pass directly; source negatives and bash -n pass; scenario contracts pass 24 tests. All six P0.075-P0.080 lifecycles pass serially with both P0.080 evidence markers, every cleanup succeeds, both contexts resolve and docs/index/diff/pruning gates pass with real_residuals=0." -->
