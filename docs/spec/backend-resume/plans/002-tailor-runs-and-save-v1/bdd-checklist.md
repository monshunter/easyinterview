# 002 BDD Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-10

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
- [x] `scripts/trigger.sh` 覆盖 `make validate-fixtures`、`TestResumeTailorEndpointsHTTPScenario`、`TestResumeTailorFixtureParity`、service/store tailor tests、`TestResumeTailorRunnerHTTPScenario`、ready job handler test
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 request/get tailor fixture parity、queued async job dispatch、ready suggestions in task output、typed `ai_task_runs` 和 ready-only outbox
- [x] 场景语义为 current `async_jobs` + `ai_task_runs` task output，不声明专属 suggestions persistence
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.077 Ready 行

## E2E.P0.078 resume tailor failure and retry

- [x] 场景目录 `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/` 存在，含标准七件套
- [x] `scripts/trigger.sh` 覆盖 `TestResumeTailorRunnerFailureScenario`、`TestTailorHandlerModeRoutingAndFailurePaths`、`TestCompleteTailorRunSuccessWritesResultAndOutbox`
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
- [x] `scripts/trigger.sh` 覆盖 job privacy tests、`TestCompleteTailorRunSuccessWritesResultAndOutbox`、cmd/api runner kernel ready/failure gates、runtime vocabulary negative grep
- [x] `scripts/verify.sh` 拒绝 skip/no-op，检查 outbox payload allowlist、`ai_task_runs`/audit privacy、runtime negative 和 private marker absence
- [x] 场景脚本使用当前测试名 `TestCompleteTailorRunSuccessWritesResultAndOutbox`
- [x] 在 `test/scenarios/e2e/INDEX.md` 保留 P0.080 Ready 行
- [x] P0.075-P0.080 verify 的 Resume mode negative gate 使用 contextual production regex 并排除 `*_test.go`；合法 `Content-Disposition: inline` 不误报，六份脚本由 contract test 固化

## Phase 10 mutation pipeline regression

- [x] `E2E.P0.075` setup / trigger / verify / cleanup serial lifecycle passes after the update handler delegates to the shared mutation pipeline.
  <!-- verified: 2026-07-10 method=p0-075-full-lifecycle evidence="The initial verify exposed the pre-existing bare inline|mirror false positive. A contract RED now covers P0.075-P0.080; the established contextual mode regex and *_test.go exclusion restore 4 contract passes and the rerun lifecycle passes through cleanup." -->
- [x] `E2E.P0.076` setup / trigger / verify / cleanup serial lifecycle passes after the duplicate handler delegates to the shared mutation pipeline.
  <!-- verified: 2026-07-10 method=p0-076-full-lifecycle evidence="Setup, trigger, verify and cleanup pass with fixture, route, handler, service, store, rollback and privacy evidence after the contextual mode negative gate fix." -->

## Phase 11 shared gate regression

- [x] P0.075-P0.080 verify scripts and the P0.080 trigger call the same `_shared` Resume mode gate; caller scripts contain no contextual regex copy.
  <!-- verified: 2026-07-10 method=resume-mode-gate-consumer-contract evidence="Contract test enumerates six verify consumers plus the P0.080 trigger, requires the shared invocation, rejects inline regex copies and asserts the shared regex/test exclusion." -->
- [x] `E2E.P0.075` through `E2E.P0.080` complete setup / trigger / verify / cleanup serially after the extraction.
  <!-- verified: 2026-07-10 method=p0-075-through-p0-080-shared-gate-regression evidence="All six scenario lifecycles pass in order; each verify invokes the shared gate and each cleanup completes without shared-environment operations." -->

## Phase 12 unified runtime gate regression

- [x] Six verify scripts and the P0.080 trigger call `resume-runtime-negative-gate.sh`; no caller regex or `resume-mode-negative-gate.sh` remains.
  <!-- verified: 2026-07-10 method=resume-runtime-gate-consumer-contract evidence="The contract enumerates seven consumers, requires the runtime helper, rejects both inline patterns and verifies the old helper file is absent." -->
- [x] `E2E.P0.075` through `E2E.P0.080` complete setup / trigger / verify / cleanup serially with both evidence markers preserved.
  <!-- verified: 2026-07-10 method=p0-075-through-p0-080-unified-runtime-gate evidence="All six lifecycles pass in order; P0.080 trigger and verify preserve the mode/module evidence pair and every cleanup completes." -->
