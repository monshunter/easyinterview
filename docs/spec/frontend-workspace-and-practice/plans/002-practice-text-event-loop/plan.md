# 002 — Practice Text Event Loop Plan

> **版本**: 1.13
> **状态**: active
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

本 plan 固化 `practice` 文本 / 电话模式 event loop 的当前合同：

- `PracticeScreen` 源级复刻 `ui-design/src/screen-practice.jsx` 当前真实面试分支，保留 TopBar、SessionMap、QuestionCard、Transcript、InputBar、PhoneSurface、HintBanner 和全局 Finish CTA 的 DOM anchor / a11y / responsive 行为；当前 UI 不包含独立辅助信息栏、固定辅助栏 CTA、会话内本地 persona switch、严格模式开关、语音转文字、跳过和语音分析。
- 只通过 generated client 消费 `getPracticeSession`、`appendSessionEvent`、`completePracticeSession`；`appendSessionEvent` 使用 `clientEventId` 且不带 `Idempotency-Key`，`completePracticeSession` 使用 `Idempotency-Key` 且 body 只包含 `clientCompletedAt`。
- `practiceMode` 不再作为用户可见 strict/assisted 开关；提示由用户主动请求并只记录 `hintUsed/hintCount`；`practiceGoal` 只表达 `baseline / retry_current_round / next_round` 数据来源，不能改变辅助度显隐。
- `completePracticeSession` 返回 `ReportWithJob` 后只 handoff 到 `generating`，路由参数携带稳定 `InterviewContext` ID 与 `PracticeDisplayContext`；当前简历绑定字段为 `resumeId`。
- 电话模式底层 turn 由 `practice-voice-mvp` owner 接管；本 plan 的 UI gate 证明用户可见 copy 统一为 `电话模式 / Phone`，并且 `createPracticeVoiceTurn` 不散落到 text event hook、completion handoff 或 report polling。

## 2 当前实现面

| Surface | 当前文件 | 合同 |
|---------|----------|------|
| Screen | `frontend/src/app/screens/practice/PracticeScreen.tsx` | `practice` route 渲染正式 screen；缺 session 进入 `PracticeSessionLostState`；`resumeId` 从 route / InterviewContext 传递 |
| Session load | `hooks/usePracticeSessionLoader.ts` | `getPracticeSession(sessionId)`，覆盖 loading / data / missing / error / refresh |
| Event loop | `hooks/usePracticeEvents.ts` | `answer_submitted / hint_requested / session_paused / session_resumed` 单 endpoint；retry 复用 `clientEventId`；正式 UI 不再发送 `turn_skipped` |
| Display policy | `hooks/usePracticeAssistance.ts` | `practiceGoal` 不参与显隐计算；strict/assisted 不再作为用户可见开关 |
| Completion | `hooks/useCompletePracticeSession.ts` | `completePracticeSession(sessionId,{clientCompletedAt},Idempotency-Key)`；replay、防抖、409/5xx error mapping |
| Handoff | `utils/practiceHandoffParams.ts` | 输出 `planId / targetJobId / jdId / resumeId / roundId / sessionId / reportId` + display context；禁止 raw text / prompt / model provenance |
| Voice boundary | `hooks/usePracticeVoiceTurn.ts` | 唯一允许调用 `createPracticeVoiceTurn` 的 practice runtime hook |

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + frontend + contract + BDD`。
- **TDD 策略**: 通过 `/implement` -> `/tdd` 执行。每个非文档 checklist item 必须先有 focused Vitest / typecheck / pixel parity / scenario wrapper 或 OpenAPI/backend contract test 断言，再修改实现；范围收敛项必须先补 current-boundary negative assertion，再调整 runtime surface。
- **BDD 策略**: 需要 BDD。真实面试会话涉及用户可见 UI、电话模式流程和跨层事件合同，继续使用 `bdd-plan.md` / `bdd-checklist.md`，并由主 checklist 的 `BDD-Gate:` 项引用 `E2E.P0.044`-`E2E.P0.047`。
- **真实环境验收 gate**: 完成标记必须包含本地真实前后端环境闭环测试和截图证据；不能只用 jsdom、fixture contract 或静态原型 parity 作为最终完成依据。
- **替代验证 gate**: 不适用；本计划是用户行为功能计划。

## 4 Coverage Matrix

| 行为 | 测试 / Gate | 证据 |
|------|-------------|------|
| PracticeScreen DOM / visual source parity | `PracticeScreen.test.tsx`、`PracticeScreenIntegration.test.tsx`、`frontend/tests/pixel-parity/practice.spec.ts` | ≥20 `practice-*` anchors、desktop/mobile theme smoke、no prototype runtime data |
| `getPracticeSession` refresh / missing-session | `hooks/usePracticeSessionLoader.test.tsx`、`practiceSessionLost.test.tsx` | 404 渲染 lost state；workspace CTA 保留 `resumeId` |
| `appendSessionEvent` body / retry / idempotency boundary | `hooks/usePracticeEvents.test.tsx`、`appendSessionEventBody.test.tsx`、`idempotencyContract.test.tsx` | 5 event kind body parity；append 不带 `Idempotency-Key` |
| AssistantAction rendering | `components/AssistantActionRenderer.test.tsx`、`PracticeScreenIntegration.test.tsx` | `ask_question / ask_follow_up / show_hint / session_wait / session_completed` 映射当前 UI |
| hint / goal policy | `hooks/usePracticeAssistance.test.ts`、`practiceGoalParity.test.tsx`、`practiceHints.test.tsx`、deleted strict-switch negative test | 提示由用户在会话中可选触发；`baseline / retry_current_round / next_round` 对显隐无副作用；不存在严格模式拦截 |
| Pause / session map | `practicePauseResume.test.tsx`、`SessionMap.test.tsx` | pause/resume disables controls；turn map 展示 done/active/pending/follow-up states；无 skip UI / event 正向路径 |
| Real-interview UI boundary | `PracticeScreen.test.tsx`、`practiceModeSwitch.test.tsx`、`outOfScopeNegative.test.ts`、`frontend/tests/pixel-parity/practice.spec.ts` | 无独立辅助信息栏 / 会话内本地 persona switch / strict switch / dictate / skip / voice metrics；phone mode 有字幕、切断、重新开始 |
| Completion handoff | `hooks/useCompletePracticeSession.test.tsx`、`completePracticeSessionBody.test.tsx`、`practiceCompletion.test.tsx`、`utils/practiceHandoffParams.test.ts` | body 只含 `clientCompletedAt`；handoff 参数使用 `resumeId` |
| Privacy / current boundary | `practicePrivacy.test.tsx`、`outOfScopeNegative.test.ts`、P0.044/P0.047 verify scripts | `getFeedbackReport` 不在 practice runtime；voice turn 只在 voice owner hook；raw answer/question/hint 不泄漏 |
| Scenario behavior | `test/scenarios/e2e/p0-044` 至 `p0-047` | assisted happy path、mode policy、failure recovery、complete + generating handoff |

## 5 Operation Matrix

| operationId | Frontend consumer | Fixture | Boundary |
|-------------|-------------------|---------|----------|
| `getPracticeSession` | `usePracticeSessionLoader` | `openapi/fixtures/PracticeSessions/getPracticeSession.json` | route `sessionId` 必填；404 进入 lost state |
| `appendSessionEvent` | `usePracticeEvents` | `openapi/fixtures/PracticeSessions/appendSessionEvent.json` | `clientEventId` in body；无 `Idempotency-Key` header；正向 UI 不再发送 `turn_skipped` |
| `completePracticeSession` | `useCompletePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json` | `Idempotency-Key` required；body 仅 `clientCompletedAt` |

## 6 BDD-Gate

- `E2E.P0.044`：文本面试 assisted happy path，覆盖 mount、answer、follow-up、Question advance、DOM anchor 和 runtime negative grep。
- `E2E.P0.045`：真实面试显示策略，覆盖 text/phone、hint optional、pause/resume、无独立辅助信息栏、无 skip、无 dictation、无 strict switch、无会话内本地 persona switch 和 forbidden input negative gate。
- `E2E.P0.046`：AI timeout、404 lost state、409 mismatch、hint retry/recovery 和 retry 复用；不得恢复严格模式冲突路径。
- `E2E.P0.047`：complete 202、idempotency replay、`generating` handoff、privacy redline。

## 7 实施步骤

### Phase 6: Real-interview session simplification

#### 6.1 UI truth source revision

Revise `docs/ui-design/module-practice-review.md` and `ui-design/src/screen-practice.jsx` so the current prototype uses the real-interview shell without independent side-panel controls, speech-to-text, skip, role switch, visible strict switch or voice analysis, and exposes phone mode with captions, hang-up and restart controls.

#### 6.2 Frontend runtime removal

Align the matching runtime components, hooks, i18n strings, tests and pixel parity expectations in `frontend/src/app/screens/practice`. The finish action remains available in the global top bar.

#### 6.3 Contract and backend removal

Remove `turn_skipped` as a positive OpenAPI/backend/frontend/scenario path. If generated schemas or backend tests still expose skip as a supported user action, revise B2/backend-practice artifacts and regenerate clients.

#### 6.4 Phone mode handoff

Keep the backend voice orchestration owner intact while making the user-visible mode `电话模式 / Phone`; phone mode must support hang-up/restart and captions without surfacing recording/submit-turn controls as the main user flow.

#### 6.5 Verification closeout

Run focused practice frontend tests, relevant backend/OpenAPI contract tests, pixel parity, BDD wrappers, context validation, doc/index checks, current-boundary negative searches, and a real local environment browser smoke with screenshot evidence.

## 8 收口证据索引

当前 owner 完成以最新 gate 为准，不引用旧 PASS 状态：

- `validate_context.py frontend-workspace-and-practice/002 frontend`
- `pnpm --filter @easyinterview/frontend test` focused practice suite
- `pnpm --filter @easyinterview/frontend exec tsc --noEmit`
- `make validate-fixtures`
- `test/scenarios/e2e/p0-044.../scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-045.../scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-046.../scripts/{setup,trigger,verify,cleanup}.sh`
- `test/scenarios/e2e/p0-047.../scripts/{setup,trigger,verify,cleanup}.sh`
- real local environment browser close-loop screenshot evidence for text and phone modes
- `sync-doc-index --check`
- `make docs-check`
- `git diff --check`

## 9 修订记录

| 日期 | 版本 | 说明 |
|------|------|------|
| 2026-07-10 | 1.12 | Align current-boundary wording with the real-interview shell: no independent side-panel controls, dictation, skip, role switch, visible strict switch or voice analysis; user-visible voice remains phone mode with hang-up/restart controls. |
| 2026-07-09 | 1.11 | Reopen plan for real-interview session simplification: current practice UI excludes side-panel controls, voice analysis, dictation, skip, role switch and visible strict switch; user-visible voice becomes phone mode with hang-up/restart controls. |
| 2026-07-07 | 1.10 | Compress owner docs to current text event loop, `resumeId` handoff, generated-client operations, voice owner boundary, BDD gates and executable evidence index. |
