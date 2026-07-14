# Backend Resume Tailor Runs and Save v1 Checklist

> **版本**: 1.14
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 0: current contract preflight

- [x] 0.2 context manifest 指向当前 backend owner surface，不再以 removed package path 作为主入口（验证：`python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-resume/plans/002-tailor-runs-and-save-v1/context.yaml --target backend` PASS）
- [x] 0.3 front/back contract pre-read 已覆盖 `docs/development.md` §2、`backend/README.md`、`openapi/README.md`、`test/scenarios/README.md` 与相关 scenario README（验证：本 plan 后续 gates 只使用 current OpenAPI/generated/fixture/runtime/scenario truth sources）

## Phase 1: flat API and removed-route boundary

- [x] 1.2 removed route family 不出现在 generated route catalog，runtime inputs 返回 404（验证：`go test ./backend/cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations' -count=1` PASS）

## Phase 2: `updateResume`

- [x] 2.1 `PATCH /api/v1/resumes/{resumeId}` 覆盖 `structuredProfile` / `displayName`，拒绝 server-owned fields，cross-user/missing row 返回 404（验证：`go test ./backend/internal/resume/... -run 'TestUpdateResume|TestUpdateResumeFixtureParity' -count=1` PASS）

## Phase 3: `duplicateResume`

- [x] 3.1 `POST /api/v1/resumes/{resumeId}/duplicate` 复制只读来源快照并应用 editable overlay，源 resume 不变（验证：`go test ./backend/internal/resume/... -run 'TestDuplicateResume|TestDuplicateResumeFixtureParity' -count=1` PASS）

## Phase 4: `requestResumeTailor` / `getResumeTailorRun`

- [x] 4.1 `requestResumeTailor` 以 `resumeId` 创建 `async_jobs(job_type='resume_tailor')`，payload 只包含当前所需 ID/mode 字段；`TestRequestResumeTailor` / `TestResumeTailorEndpointsHTTPContract` / fixture parity 仅作开发反馈，阶段单测完成由根 `make test` 承接。

## Phase 5: resume.tailor async job and outbox

- [x] 5.1 tailor job 通过 A3 AIClient + F3 feature_key 路由，覆盖 success、timeout、output_invalid、retry-to-ready（验证：`go test ./backend/internal/resume/jobs -run 'TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox|TestTailorHandlerModeRoutingAndFailurePaths' -count=1` PASS）
- [x] 5.2 success 写 typed `ai_task_runs` 和 ready-only `resume.tailor.completed` outbox；payload allowlist 不含 prompt/raw resume/match summary/suggested bullet 文本。Runner integration tests 独立验证真实组件组合，不包装为 E2E；阶段单测完成由根 `make test` 承接。

## Phase 6: flat save fixture parity and read-only detail boundary

- [x] 6.3 docs/index/context 收口（验证：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS；`make docs-check` PASS；`git diff --check` PASS）
- [x] 6.4 product-scope pruning owner 记录本 phase 6.100 evidence，聚合 residual scan 不再把本 owner plan/checklist 识别为旧叙事热点（验证：`make lint-core-loop-pruning-surface` PASS，`real_residuals=0`）

## Phase 7: tailor provenance conversion simplification

- [x] 7.1 `VersionProvenance` 与私有 persisted wire type 的写入/读回使用显式同构转换，删除双向逐字段复制（验证：store/package tests、scoped/full backend `staticcheck`、owner context/docs gates）



## Phase 9: current OpenAPI inventory wording

- [x] 9.1 Align the preflight inventory with the current 37-operation B2 contract while preserving the 10-operation Resume/ResumeTailor subset; verify OpenAPI inventory, fixtures, owner contexts and docs/diff/pruning gates.
  <!-- verified: 2026-07-10 method=current-openapi-inventory-wording evidence="OpenAPI inventory and fixture validation report 10 tags, 37 operations and 37 fixtures; Resume/ResumeTailor remains 10 operations. Current owner plan/checklist search has no 35/36 inventory and backend context passes." -->

## Phase 10: flat Resume mutation handler pipeline consolidation

- [x] 10.1 RED: scoped `dupl` identifies `UpdateResume` / `DuplicateResume` as the only clone group in `internal/resume/handler`; the duplicated block covers user resolution through success response.
  <!-- red: 2026-07-10 method=resume-mutation-handler-dupl evidence="go run github.com/golangci/dupl@v0.0.0-20250308024227-f665c8d69b32 -t 100 backend/internal/resume/handler reports one production clone group between update.go and duplicate.go." -->
- [x] 10.2 Introduce one private typed mutation helper while retaining operation-specific service assertions, validators and success statuses; require scoped `dupl` zero clone groups；focused tests 仅作开发反馈，阶段单测完成由仓库根 `make test` 承接。

## Phase 11: shared Resume mode negative gate

- [x] 11.2 Add one `_shared` executable gate, replace all seven inline blocks, and make the scenario contract reject caller-owned regex copies.
  <!-- verified: 2026-07-10 method=shared-resume-mode-gate-green evidence="One executable helper owns the contextual regex, test-file exclusion, failure message and rg-error propagation. Seven callers contain zero inline regex copies; the four-test scenario contract, direct gate execution and bash -n across all touched scripts pass." -->

## Phase 12: unified Resume runtime negative gate

  <!-- red: 2026-07-10 method=resume-module-gate-copy-and-drift evidence="rg reports seven inline mistakes|growth|drill|inline-debrief-record branches; only five of seven exclude *_test.go." -->
- [x] 12.2 Rename and broaden the shared helper to own both scans through one error-aware function; delete the old helper name and all caller-owned module regex blocks.
  <!-- verified: 2026-07-10 method=unified-resume-runtime-gate-green evidence="resume-runtime-negative-gate.sh owns both patterns through run_negative_scan; seven consumers have no inline module regex and no old-helper invocation, and the old file is absent. Contract tests, direct execution and bash -n pass." -->

## BDD Gate

- [x] BDD-Gate: `BDD.RESUME.TAILOR.001` 由 [BDD checklist](./bdd-checklist.md) 关联 tailor/run/save owner behavior tests；不创建或声明真实 E2E PASS。
