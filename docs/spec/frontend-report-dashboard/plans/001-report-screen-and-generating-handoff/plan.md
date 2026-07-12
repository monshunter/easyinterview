# 001 — Conversation Report Screen and Handoff

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

将 generating/report UI 原地改为 conversation-level report：三项 summary metrics、四个无 tab 内容区（dimensions/highlights/issues/next actions）与 competency-focused replay，删除 QuestionsTab、逐题 summary、hint/phone/mode context。

## 2 Operation Matrix

| operationId | fixture | consumer | backend | persistence | AI | scenario |
|-------------|---------|----------|---------|-------------|----|----------|
| `getFeedbackReport` | queued/ready/failed/new shape | generating/report | backend-review | feedback_reports | read none | P0.056/P0.058 |
| `createPracticePlan` | retry/next | replay handler | backend-practice | practice_plans | none | P0.057 |
| `startPracticeSession` | opening message | replay handler | backend-practice | session/messages | practice.session.chat | P0.057 |

## 3 质量门禁分类

- **Plan 类型**: user-visible UI + API consumer + contract migration。
- **TDD 策略**: prototype/source tests first; formal component/hook tests then consume generated session-level report.
- **BDD 策略**: P0.056 happy generating/report, P0.057 replay/next, P0.058 failure, P0.059 parity/negative, P0.099 real screenshot.
- **替代验证 gate**: i18n/typecheck/build/source/pixel parity/stale negative.

## 4 Coverage Matrix

| Source | Category | Phase | Verification | UI anchor | Negative |
|--------|----------|-------|--------------|-----------|----------|
| three metrics/four sections | source structure | 1-2 | prototype/formal tests | screen-report::ReportScreen | QuestionsTab |
| generating copy | UX | 1-2 | i18n/DOM tests | ReportGeneratingScreen | 逐题/题目回顾 |
| ready report | primary | 3 | P0.056 | dashboard | questionAssessments |
| replay competency | primary | 4 | P0.057 | Header CTA | retryFocusTurnIds |
| failure/missing | recovery | 3 | P0.058 | failure states | fake report |
| geometry/screenshot | visual | 5 | P0.059/P0.099 | updated prototype | old tab/bbox |

## 5 实施步骤

### Phase 1: UI truth source
- Rewrite report prototype/data/generating copy to readiness/dimensions/evidence/next.
- Delete perQuestion state, Questions tab/list/toggle and hint/phone context.

### Phase 2: Formal structure
- Delete QuestionsTab and question summary/card/body paths.
- Use three summary metrics and four always-visible content sections; simplify ContextStrip.
- Update i18n/a11y/responsive geometry.

### Phase 3: Data states
- Consume dimensionAssessments/retryFocusCompetencyCodes.
- Cover queued/generating/ready/failed/notFound/missing/empty evidence.

### Phase 4: Replay/next
- Retry plan uses competency codes; next round uses stable round context.
- Fresh session opens conversation with assistant message.

### Phase 5: Parity and real scenario
- Full frontend/source/pixel parity/negative gates.
- P0.099 real browser conversation → generating → report screenshots.

## 6 验收标准

- No Questions tab/card/list/count/toggle or turn-based replay.
- Four current report surfaces render server data and error/empty states.
- Replay/next create fresh conversation sessions.
- Desktop/mobile prototype/formal/real screenshots close.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Reopen for conversation-level report and competency replay. |
