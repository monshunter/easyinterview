# 004 — Derived Plans and Debrief Seeding BDD Plan

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-17

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 使用当前 `test/scenarios/e2e` 编号规则。`backend-debrief/001` 已占 `E2E.P0.060-064`，`frontend-debrief/001` 已占 `E2E.P0.065-069`，因此本 plan 分配 `E2E.P0.070-073`。

- 套件: `e2e`
- 阶段: `P0`
- 执行入口: `cd backend && go test ./cmd/api -run 'TestE2EP0070|TestE2EP0071|TestE2EP0072|TestE2EP0073' -count=1`
- 外部 shell 场景资产: 本 plan 初始落在 `backend/cmd/api/practice_http_scenario_test.go` 的 HTTP scenario tests；未来如需提升为 `test/scenarios/e2e/p0-070-*` local runner shell assets，保持同一 Given / When / Then，不重编号。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Phase | 覆盖 |
|---------|------|------|------------|------|
| `E2E.P0.070` | Derived practice plan create/read + idempotency replay | primary + cross-layer contract | Phase 1 | C-2, C-3, D-14, D-24 |
| `E2E.P0.071` | Debrief practice session starts from confirmed debrief question | primary | Phase 2 | C-3, D-14, D-5 |
| `E2E.P0.072` | Source validation, isolation, and privacy | failure/recovery + privacy | Phase 1 + Phase 3 | C-13, D-11, D-14 |
| `E2E.P0.073` | Debrief goal mode regression and legacy-negative | alternate + regression | Phase 2 + Phase 3 | C-8b, D-5, D-21 |

## 2 Scenarios

| 场景 ID | Given | When | Then | 验证入口 |
|---------|-------|------|------|----------|
| `E2E.P0.070` | 用户 A 有 ready feedback report、completed debrief、target job、resume asset；source 都属于同一 target job | 用户 A 分别调用 `createPracticePlan(goal='retry_current_round', sourceReportId=...)`、`createPracticePlan(goal='next_round', sourceReportId=...)`、`createPracticePlan(goal='debrief', sourceDebriefId=...)` 并重放同一 Idempotency-Key | 三个 plan 均 201；`getPracticePlan` 返回 source ids；same key replay 返回同 response；DB 写入对应 source 列且互斥；audit 仅含 ids/counts | `test/scenarios/e2e/p0-070-practice-derived-plan-create-read-replay/scripts/trigger.sh` -> `verify.sh` (`TestE2EP0070PracticeDerivedPlanCreateReadReplay`) |
| `E2E.P0.071` | 用户 A 有 `goal='debrief'` ready plan，source debrief raw_questions[0].questionText = `__DEBRIEF_FIRST_QUESTION__`；fake F3/A3 first_question client 可计数 | 用户 A 调用 `startPracticeSession` | 返回 201 + running session + currentTurn.questionText = `__DEBRIEF_FIRST_QUESTION__`；first_question registry/AI 调用次数为 0；outbox `practice.session.started` payload 只含 ids/status/mode/goal | `test/scenarios/e2e/p0-071-practice-debrief-start-source-question/scripts/trigger.sh` -> `verify.sh` (`TestE2EP0071PracticeDebriefStartUsesSourceQuestion`) |
| `E2E.P0.072` | 用户 A/B 各有 source records；另有 draft debrief、empty raw_questions debrief、wrong-target source；注入 raw marker `__PRIVATE_DEBRIEF_TEXT__` 到 debrief notes/answer/reaction | 用户 A 用缺失、跨用户、wrong-target、draft、empty source 调用 create/start；运行 privacy assertions | 错误 envelope 不泄露跨用户 source 内容；invalid source 返回 422/404 等既有 canonical envelope；audit/outbox/log/metric/idempotency response 不含 raw marker | `test/scenarios/e2e/p0-072-practice-derived-source-isolation-privacy/scripts/trigger.sh` -> `verify.sh` (`TestE2EP0072PracticeDerivedSourceValidationIsolationPrivacy`) |
| `E2E.P0.073` | 用户 A 有两个 debrief-derived plans：mode=assisted 与 mode=strict；另有非法 body `mode='debrief'` | 用户 A 对两个合法 plan 分别 start session；再提交非法 mode create request；运行 scoped legacy grep | assisted/strict 都 start 成功且 goal=debrief；非法 mode 返回 422；runtime/generated/fixtures 不出现 active `PracticeModeDebrief` / `mode='debrief'` / `legacy debrief replay value` | `test/scenarios/e2e/p0-073-practice-debrief-mode-regression/scripts/trigger.sh` -> `verify.sh` (`TestE2EP0073PracticeDebriefAssistedStrictAndLegacyNegative`) |

## 3 数据隔离与污染恢复

- 每个 scenario 使用独立 user / target job / report / debrief / plan / session / idempotency key。
- cleanup 顺序遵循 `test/scenarios/README.md`：scenario 自身数据优先，其次共享组件，最后才重建环境。
- 不预设 Helm chart、Kind namespace 或外部平台名称；当前执行入口为 Go HTTP scenario。
