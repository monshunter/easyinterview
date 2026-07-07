# 002 — Practice Text Event Loop Plan

> **版本**: 1.10
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本 plan 固化 `practice` 文本面试 event loop 的当前合同：

- `PracticeScreen` 源级复刻 `ui-design/src/screen-practice.jsx` 文本分支，保留 TopBar、SessionMap、QuestionCard、Transcript、InputBar、RightPanel、HintBanner 和 FinishCta 的 DOM anchor / a11y / responsive 行为。
- 只通过 generated client 消费 `getPracticeSession`、`appendSessionEvent`、`completePracticeSession`；`appendSessionEvent` 使用 `clientEventId` 且不带 `Idempotency-Key`，`completePracticeSession` 使用 `Idempotency-Key` 且 body 只包含 `clientCompletedAt`。
- `practiceMode` 只表达 `assisted / strict` 辅助度；`practiceGoal` 只表达 `baseline / retry_current_round / next_round` 数据来源，不能改变辅助度显隐。
- `completePracticeSession` 返回 `ReportWithJob` 后只 handoff 到 `generating`，路由参数携带稳定 `InterviewContext` ID 与 `PracticeDisplayContext`；当前简历绑定字段为 `resumeId`。
- voice turn 由 `practice-voice-mvp` owner 接管；本 plan 的 text-loop gate 只证明 `createPracticeVoiceTurn` 不散落到 text event hook、completion handoff 或 report polling。

## 2 当前实现面

| Surface | 当前文件 | 合同 |
|---------|----------|------|
| Screen | `frontend/src/app/screens/practice/PracticeScreen.tsx` | `practice` route 渲染正式 screen；缺 session 进入 `PracticeSessionLostState`；`resumeId` 从 route / InterviewContext 传递 |
| Session load | `hooks/usePracticeSessionLoader.ts` | `getPracticeSession(sessionId)`，覆盖 loading / data / missing / error / refresh |
| Event loop | `hooks/usePracticeEvents.ts` | `answer_submitted / hint_requested / turn_skipped / session_paused / session_resumed` 单 endpoint；retry 复用 `clientEventId` |
| Display policy | `hooks/usePracticeAssistance.ts` | strict 隐藏 hint / live notes / experience cards；practiceGoal 不参与显隐计算 |
| Completion | `hooks/useCompletePracticeSession.ts` | `completePracticeSession(sessionId,{clientCompletedAt},Idempotency-Key)`；replay、防抖、409/5xx error mapping |
| Handoff | `utils/practiceHandoffParams.ts` | 输出 `planId / targetJobId / jdId / resumeId / roundId / sessionId / reportId` + display context；禁止 raw text / prompt / model provenance |
| Voice boundary | `hooks/usePracticeVoiceTurn.ts` | 唯一允许调用 `createPracticeVoiceTurn` 的 practice runtime hook |

## 3 Coverage Matrix

| 行为 | 测试 / Gate | 证据 |
|------|-------------|------|
| PracticeScreen DOM / visual source parity | `PracticeScreen.test.tsx`、`PracticeScreenIntegration.test.tsx`、`frontend/tests/pixel-parity/practice.spec.ts` | ≥20 `practice-*` anchors、desktop/mobile theme smoke、no prototype runtime data |
| `getPracticeSession` refresh / missing-session | `hooks/usePracticeSessionLoader.test.tsx`、`practiceSessionLost.test.tsx` | 404 渲染 lost state；workspace CTA 保留 `resumeId` |
| `appendSessionEvent` body / retry / idempotency boundary | `hooks/usePracticeEvents.test.tsx`、`appendSessionEventBody.test.tsx`、`idempotencyContract.test.tsx` | 5 event kind body parity；append 不带 `Idempotency-Key` |
| AssistantAction rendering | `components/AssistantActionRenderer.test.tsx`、`PracticeScreenIntegration.test.tsx` | `ask_question / ask_follow_up / show_hint / session_wait / session_completed` 映射当前 UI |
| strict / assisted policy | `hooks/usePracticeAssistance.test.ts`、`practiceGoalParity.test.tsx`、`practiceHints.test.tsx`、`practiceStrictToggleLocked.test.tsx` | `baseline / retry_current_round / next_round` 对显隐无副作用 |
| Pause / skip / session map | `practicePauseResume.test.tsx`、`practiceSkip.test.tsx`、`SessionMap.test.tsx` | pause/resume disables controls；turn map 展示 done/active/pending/skipped/follow-up states |
| Completion handoff | `hooks/useCompletePracticeSession.test.tsx`、`completePracticeSessionBody.test.tsx`、`practiceCompletion.test.tsx`、`utils/practiceHandoffParams.test.ts` | body 只含 `clientCompletedAt`；handoff 参数使用 `resumeId` |
| Privacy / current boundary | `practicePrivacy.test.tsx`、`nonCurrentNegative.test.ts`、P0.044/P0.047 verify scripts | `getFeedbackReport` 不在 practice runtime；voice turn 只在 voice owner hook；raw answer/question/hint 不泄漏 |
| Scenario behavior | `test/scenarios/e2e/p0-044` 至 `p0-047` | assisted happy path、mode policy、failure recovery、complete + generating handoff |

## 4 Operation Matrix

| operationId | Frontend consumer | Fixture | Boundary |
|-------------|-------------------|---------|----------|
| `getPracticeSession` | `usePracticeSessionLoader` | `openapi/fixtures/PracticeSessions/getPracticeSession.json` | route `sessionId` 必填；404 进入 lost state |
| `appendSessionEvent` | `usePracticeEvents` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `clientEventId` in body；无 `Idempotency-Key` header |
| `completePracticeSession` | `useCompletePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` | `Idempotency-Key` required；body 仅 `clientCompletedAt` |

## 5 BDD-Gate

- `E2E.P0.044`：文本面试 assisted happy path，覆盖 mount、answer、follow-up、Question advance、DOM anchor 和 runtime negative grep。
- `E2E.P0.045`：assisted / strict × `baseline / retry_current_round / next_round` 显隐矩阵，覆盖 hint、skip、pause/resume、strict lock 和 forbidden input negative gate。
- `E2E.P0.046`：AI timeout、404 lost state、409 mismatch、strict conflict、retry 复用。
- `E2E.P0.047`：complete 202、idempotency replay、`generating` handoff、privacy redline。

## 6 收口证据索引

当前 owner 完成以最新 gate 为准，不引用旧 PASS 状态：

- `validate_context.py frontend-workspace-and-practice/002 frontend`
- `pnpm --filter @easyinterview/frontend test` focused practice suite
- `pnpm --filter @easyinterview/frontend exec tsc --noEmit`
- `make validate-fixtures`
- `test/scenarios/e2e/p0-044.../scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-045.../scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-046.../scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-047.../scripts/{setup,trigger,verify,cleanup}.sh`
- `sync-doc-index --check`
- `make docs-check`
- `git diff --check`

## 7 修订记录

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-07-07 | 1.10 | Compress owner docs to current text event loop, `resumeId` handoff, generated-client operations, voice owner boundary, BDD gates and executable evidence index. |
