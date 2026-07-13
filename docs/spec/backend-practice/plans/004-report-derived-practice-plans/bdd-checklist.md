# 004 — Report-derived Practice Plans BDD Checklist

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.070 — Report-derived Plan Create/Read/Replay

- [x] 准备 Go HTTP scenario data：report source、target job、resume
- [x] 实现 `scripts/trigger.sh`：执行 retry / next-round `createPracticePlan` + `getPracticePlan` + idempotency replay Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，并拒绝 `FAIL` / `no tests to run`
- [x] 当前 focused Go gate 执行 `TestE2EP0070PracticeDerivedPlanCreateReadReplay`
- [x] 记录通过证据
  <!-- verified: 2026-07-06 method=focused-go-gate evidence="cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1 PASS; scenario directory `test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay` points to TestE2EP0070PracticeDerivedPlanCreateReadReplay." -->
- [x] Phase 3 refresh requires `REPORT_GENERIC_RETRY_PASS`, `REPORT_DERIVED_FOCUS_PASS`, `REPORT_DERIVED_SEMANTIC_PROMPT_PASS`, `REPORT_NEXT_EMPTY_FOCUS_PASS`, `REPORT_DERIVED_IDEMPOTENCY_PASS`, F3 `PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS` and real PostgreSQL markers; F3 activation remains separately owned by 002.
  <!-- verified: 2026-07-12 method=scenario evidence="P0.070 setup/trigger/verify/cleanup ran against disposable schema v19 PostgreSQL. API IK, active v0.2 registry and tagged store tests all had exact RUN/PASS/package-ok evidence; six REPORT_* markers and verified F3 marker consumption passed." -->

## E2E.P0.072 — Report Source Validation / Isolation / Privacy

- [x] 准备 missing / cross-user / wrong-target source matrix
- [x] 实现 `scripts/trigger.sh`：执行 invalid create attempts Go HTTP scenario，并保留真实 exit code
- [x] 实现 `scripts/verify.sh`：校验 runner marker、目标测试 `=== RUN` / `--- PASS`、package `ok`，拒绝 no-op，并确认 privacy marker 不进入 runner evidence
- [x] 当前 focused Go gate 执行 `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy`
- [x] 记录通过证据
  <!-- verified: 2026-07-06 method=focused-go-gate evidence="cd backend && go test ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api -run 'Derived|Source|TestE2EP0070|TestE2EP0072' -count=1 PASS; scenario directory `test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy` points to TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy." -->
- [x] Phase 3 refresh covers every frozen/source mismatch plus unsupported non-empty focus, requires `REPORT_DERIVED_ISOLATION_PASS` and `REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS`, proves final active runtime does not consume v0.1 legacy focus tokens, proves zero insert/leak and cleans scenario data.
  <!-- verified: 2026-07-12 method=scenario evidence="P0.072 setup/trigger/verify/cleanup ran against disposable schema v19 PostgreSQL; 12 source/frozen/focus failure cases proved zero insert and generic isolation, privacy and PostgreSQL markers. Final runtime/generated/OpenAPI/JSON fixture/scenario search emitted REPORT_DERIVED_LEGACY_IDENTIFIER_NEGATIVE_PASS; only explicit PROTOTYPE_MAPPING negative documentation is allowlisted." -->
