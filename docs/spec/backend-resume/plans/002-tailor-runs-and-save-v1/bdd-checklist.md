# 002 BDD Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.074 flat resume reads and removed route boundary

- [x] 场景目录 `test/scenarios/e2e/p0-074-resume-flat-read-api/` 存在，含 `README.md`、`data/seed-input.md`、`data/expected-outcome.md`、四段脚本
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、removed route/catalog tests、flat read handler/service/store focused tests
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 route/catalog boundary、flat fixture parity、privacy negative
- [x] 场景语义为当前 flat read + removed route/catalog boundary
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.074 Ready 行

## E2E.P0.075 flat resume update and IK

- [x] 场景目录 `test/scenarios/e2e/p0-075-resume-update-flat-fields-and-ik/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、`TestResumeRegisterListHTTPScenario` runtime gate、`TestUpdateResumeFixtureParity`、update service/store tests
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 `updateResume` fixture parity、IK gate、server-owned field 422、cross-user/not-found evidence
- [x] 场景语义为当前 `updateResume` flat field update
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.075 Ready 行

## E2E.P0.076 flat resume duplicate save-as-new

- [x] 场景目录 `test/scenarios/e2e/p0-076-resume-duplicate-save-as-new/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、runtime route gate、`TestDuplicateResumeFixtureParity`、duplicate service/store tests
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 source snapshot copy、structuredProfile overlay、rollback、fixture parity 和 privacy negative
- [x] 场景语义为当前 `duplicateResume` save-as-new
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.076 Ready 行

## E2E.P0.077 flat resume tailor async dispatch and ready

- [x] 场景目录 `test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、`TestResumeTailorEndpointsHTTPScenario`、`TestResumeTailorFixtureParity`、service/store tailor tests、`TestResumeTailorDrainerHTTPScenario`、ready job handler test
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 request/get tailor fixture parity、queued async job dispatch、ready suggestions in task output、typed `ai_task_runs` 和 ready-only outbox
- [x] 场景语义为 current `async_jobs` + `ai_task_runs` task output，不声明专属 suggestions persistence
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.077 Ready 行

## E2E.P0.078 resume tailor failure and retry

- [x] 场景目录 `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `TestResumeTailorDrainerFailureScenario`、`TestTailorHandlerModeRoutingAndFailurePaths`、`TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 timeout retryable、output_invalid terminal、retry-to-ready、ready-only outbox 和 privacy negative
- [x] 场景脚本使用当前测试名 `TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.078 Ready 行

## E2E.P0.079 flat save fixture parity and read-only detail boundary

- [x] 场景目录 `test/scenarios/e2e/p0-079-resume-rewrites-accept-only-save/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、removed route/catalog tests、flat save fixture parity、frontend read-only detail negative Vitest
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 removed route inputs absent、`updateResume` / `duplicateResume` / `requestResumeTailor` fixture parity、read-only detail boundary evidence
- [x] 场景语义为当前 flat save fixture parity + read-only detail boundary
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.079 Ready 行

## E2E.P0.080 tailor privacy and runtime vocabulary negative

- [x] 场景目录 `test/scenarios/e2e/p0-080-resume-tailor-privacy-negative/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 job privacy tests、`TestCompleteTailorRunSuccessWritesResultAndOutbox`、cmd/api drainer ready/failure gates、runtime vocabulary negative grep
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 outbox payload allowlist、`ai_task_runs`/audit privacy、runtime negative 和 private marker absence
- [x] 场景脚本使用当前测试名 `TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.080 Ready 行
