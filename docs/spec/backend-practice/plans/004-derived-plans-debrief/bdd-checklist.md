# 004 — Derived Plans and Debrief Seeding BDD Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-17

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.070 — Derived Practice Plan Create/Read/Replay

- [x] 准备 Go HTTP scenario data：report source、debrief source、target job、resume asset
- [x] 实现 `scripts/trigger.sh`：执行 3 个 derived createPracticePlan + getPracticePlan + idempotency replay Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，并拒绝 `FAIL` / `no tests to run`
- [x] 按四段脚本执行 `TestE2EP0070PracticeDerivedPlanCreateReadReplay`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh -->

## E2E.P0.071 — Debrief Session Starts From Source Question

- [x] 准备 completed debrief raw_questions fixture 与 fake F3/A3 call counter
- [x] 实现 `scripts/trigger.sh`：执行 startPracticeSession(goal=debrief plan) Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，拒绝 no-op，并确认 source question marker 不进入 runner evidence
- [x] 按四段脚本执行 `TestE2EP0071PracticeDebriefStartUsesSourceQuestion`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-071-practice-debrief-start-source-question/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh -->

## E2E.P0.072 — Source Validation / Isolation / Privacy

- [x] 准备 missing/cross-user/wrong-target/draft/empty source matrix
- [x] 实现 `scripts/trigger.sh`：执行 invalid create/start attempts Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，拒绝 no-op，并确认 privacy marker 不进入 runner evidence
- [x] 按四段脚本执行 `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh -->

## E2E.P0.073 — Debrief Goal Mode Regression

- [x] 准备 assisted/strict debrief plans 与非法 mode body
- [x] 实现 `scripts/trigger.sh`：执行两个合法 start + 一个非法 create Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，拒绝 no-op，并运行 scoped legacy-negative grep
- [x] 按四段脚本执行 `TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 test/scenarios/e2e/p0-073-practice-debrief-mode-regression/scripts/setup.sh -> trigger.sh -> verify.sh -> cleanup.sh; scoped legacy grep no runtime/generated/fixture matches -->
