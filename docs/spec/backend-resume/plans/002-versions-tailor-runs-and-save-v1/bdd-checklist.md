# 002 BDD Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-06-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.074 flat resume reads and retired version routes

- [x] 场景目录 `test/scenarios/e2e/p0-074-resume-confirm-master-and-version-reads/` 存在，含 `README.md`、`data/seed-input.md`、`data/expected-outcome.md`、四段脚本
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、retired route/catalog tests、flat read handler/service/store focused tests
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 `TestResumeVersionRoutesAreGonePerD20`、`TestGeneratedRouteCatalogHasNoResumeVersionOperations`、flat fixture parity、privacy/legacy negative
- [x] 场景语义已从旧 confirm master / version read 更新为 D-20 flat read + retired route/catalog negative
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.074 Ready 行

## E2E.P0.075 flat resume update and IK

- [x] 场景目录 `test/scenarios/e2e/p0-075-resume-update-version-merge-and-ik/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、`TestResumeRegisterListHTTPScenario` runtime gate、`TestUpdateResumeFixtureParity`、update service/store tests
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 `updateResume` fixture parity、IK gate、server-owned field 422、cross-user/not-found evidence
- [x] 场景语义已从旧 `updateResumeVersion` 更新为 D-20 `updateResume`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.075 Ready 行

## E2E.P0.076 flat resume duplicate sync paths

- [x] 场景目录 `test/scenarios/e2e/p0-076-resume-branch-version-sync-paths/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、runtime route gate、`TestDuplicateResumeFixtureParity`、duplicate service/store tests
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 source snapshot copy、structuredProfile overlay、rollback、fixture parity 和 privacy/legacy negative
- [x] 场景语义已从旧 `branchResumeVersion` 更新为 D-20 `duplicateResume`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.076 Ready 行

## E2E.P0.077 flat resume tailor async dispatch and ready

- [x] 场景目录 `test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、`TestResumeTailorEndpointsHTTPScenario`、`TestResumeTailorFixtureParity`、service/store tailor tests、`TestResumeTailorDrainerHTTPScenario`、ready job handler test
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 request/get tailor fixture parity、queued async job dispatch、ready suggestions in task output、typed `ai_task_runs` 和 ready-only outbox
- [x] 场景语义已收敛为 D-20 `async_jobs` + `ai_task_runs`，不再声明专属 `resume_tailor_runs` 表
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.077 Ready 行

## E2E.P0.078 resume tailor failure and retry

- [x] 场景目录 `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `TestResumeTailorDrainerFailureScenario`、`TestTailorHandlerModeRoutingAndFailurePaths`、`TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 timeout retryable、output_invalid terminal、retry-to-ready、ready-only outbox 和 privacy/legacy negative
- [x] 场景脚本已修正为当前测试名 `TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.078 Ready 行

## E2E.P0.079 retired suggestion routes and accept-only save flow

- [x] 场景目录 `test/scenarios/e2e/p0-079-resume-suggestion-accept-reject-terminal/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、retired route/catalog tests、flat save fixture parity、frontend Rewrites/Detail Vitest
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 old accept/reject routes gone、`updateResume` / `duplicateResume` / `requestResumeTailor` fixture parity、accept-only save flow evidence
- [x] 场景语义已从服务端 accept/reject 状态机更新为 D-20 retired routes + frontend accept-only save
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.079 Ready 行

## E2E.P0.080 tailor privacy and legacy negative

- [x] 场景目录 `test/scenarios/e2e/p0-080-resume-versions-privacy-legacy/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 job privacy tests、`TestCompleteTailorRunSuccessWritesResultAndOutbox`、cmd/api drainer ready/failure gates、retired vocabulary negative grep
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 outbox payload allowlist、`ai_task_runs`/audit privacy、legacy negative 和 private marker absence
- [x] 场景脚本已修正为当前测试名 `TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.080 Ready 行
