# Backend Resume Tailor Runs and Save v1 Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 0: current contract preflight

- [x] 0.1 读取 backend-resume spec、B2 OpenAPI/fixtures/generated artifacts、backend handler/store/job/runtime 和 P0.074-P0.080 场景，确认当前合同为 flat `resumes` + 10 个 Resume / ResumeTailor operation（验证：`make lint-openapi` PASS，inventory 为 10 tags / 36 operations；`openapi/openapi.yaml` 仅包含 current Resume operationIds）
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
- [x] 5.2 success 写 typed `ai_task_runs` 和 ready-only `resume.tailor.completed` outbox；payload allowlist 不含 prompt/raw resume/match summary/suggested bullet 文本（验证：`go test ./backend/internal/resume/store ./backend/cmd/api -run 'TestCompleteTailorRunSuccessWritesResultAndOutbox|TestResumeTailorDrainerHTTPScenario|TestResumeTailorDrainerFailureScenario' -count=1` PASS）
- [x] 5.3 BDD-Gate: `E2E.P0.080` tailor privacy and non-current negative 场景保持 Ready（验证：`test/scenarios/e2e/p0-080-resume-tailor-privacy-negative/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）

## Phase 6: flat save fixture parity and read-only detail boundary

- [x] 6.1 flat save fixtures 保持 current backend contract，同时前端详情不再暴露 Rewrites/Edit 二次操作（验证：P0.079 trigger/verify 覆盖 read-only detail negative Vitest、flat save fixture parity 和 route boundary checks）
- [x] 6.2 BDD-Gate: `E2E.P0.079` flat save fixture parity + read-only detail boundary 场景保持 Ready（验证：`test/scenarios/e2e/p0-079-resume-rewrites-accept-only-save/scripts/setup.sh && .../trigger.sh && .../verify.sh && .../cleanup.sh` PASS）
- [x] 6.3 docs/index/context 收口（验证：`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS；`make docs-check` PASS；`git diff --check` PASS）
- [x] 6.4 product-scope pruning owner 记录本 phase 6.100 evidence，聚合 residual scan 不再把本 owner plan/checklist 识别为旧叙事热点（验证：`make lint-core-loop-pruning-surface` PASS，`real_residuals=0`）
