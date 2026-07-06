# 004 — Report-derived Practice Plans BDD Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-07-06

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.070 — Report-derived Plan Create/Read/Replay

- [x] 准备 Go HTTP scenario data：report source、target job、resume
- [x] 实现 `scripts/trigger.sh`：执行 retry / next-round `createPracticePlan` + `getPracticePlan` + idempotency replay Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，并拒绝 `FAIL` / `no tests to run`
- [x] 当前 focused Go gate 执行 `TestE2EP0070PracticeDerivedPlanCreateReadReplay`
- [x] 记录通过证据
  <!-- verified: 2026-07-06 method=focused-go-gate evidence="cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1 PASS; scenario directory `test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay` points to TestE2EP0070PracticeDerivedPlanCreateReadReplay." -->

## E2E.P0.072 — Report Source Validation / Isolation / Privacy

- [x] 准备 missing / cross-user / wrong-target source matrix
- [x] 实现 `scripts/trigger.sh`：执行 invalid create attempts Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，拒绝 no-op，并确认 privacy marker 不进入 runner evidence
- [x] 当前 focused Go gate 执行 `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy`
- [x] 记录通过证据
  <!-- verified: 2026-07-06 method=focused-go-gate evidence="cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1 PASS; scenario directory `test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy` points to TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy." -->
