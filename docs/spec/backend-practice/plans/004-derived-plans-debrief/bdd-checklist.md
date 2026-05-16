# 004 — Derived Plans and Debrief Seeding BDD Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-16

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.070 — Derived Practice Plan Create/Read/Replay

- [x] 准备 Go HTTP scenario data：report source、debrief source、target job、resume asset
- [x] 实现 trigger：3 个 derived createPracticePlan + getPracticePlan + idempotency replay
- [x] 实现 verify：source 列互斥、response source ids、audit ids/counts、no raw text
- [x] 执行 `TestE2EP0070PracticeDerivedPlanCreateReadReplay`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0070PracticeDerivedPlanCreateReadReplay -count=1` -->

## E2E.P0.071 — Debrief Session Starts From Source Question

- [x] 准备 completed debrief raw_questions fixture 与 fake F3/A3 call counter
- [x] 实现 trigger：startPracticeSession(goal=debrief plan)
- [x] 实现 verify：currentTurn 来自 debrief raw_questions[0]、first_question AI 零调用、started outbox 无 raw text
- [x] 执行 `TestE2EP0071PracticeDebriefStartUsesSourceQuestion`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0071PracticeDebriefStartUsesSourceQuestion -count=1` -->

## E2E.P0.072 — Source Validation / Isolation / Privacy

- [x] 准备 missing/cross-user/wrong-target/draft/empty source matrix
- [x] 实现 trigger：invalid create/start attempts
- [x] 实现 verify：canonical error envelope、cross-user 不泄露、privacy marker 零出现
- [x] 执行 `TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy -count=1` -->

## E2E.P0.073 — Debrief Goal Mode Regression

- [x] 准备 assisted/strict debrief plans 与非法 mode body
- [x] 实现 trigger：两个合法 start + 一个非法 create
- [x] 实现 verify：合法 start 均成功、非法 mode 422、legacy-negative grep 通过
- [x] 执行 `TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative`
- [x] 记录通过证据
  <!-- verified: 2026-05-16 `cd backend && go test ./cmd/api -run TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative -count=1`; scoped legacy grep no runtime/generated/fixture matches -->
