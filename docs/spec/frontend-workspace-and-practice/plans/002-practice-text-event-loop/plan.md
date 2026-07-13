# 002 — Practice Continuous Text Conversation

> **版本**: 2.4
> **状态**: active
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 Practice UI 从“左侧题目地图 + 当前题卡 + 对话”改为“Top Bar + 全宽连续聊天”，删除专用 hint/mode/phone surface；电话图标保留为 disabled affordance。正式实现必须先修改 `ui-design/src/screen-practice.jsx`，再源级迁移。

## 2 Operation Matrix

| operationId | fixture | consumer | backend | persistence | AI | scenario |
|-------------|---------|----------|---------|-------------|----|----------|
| `getPracticeSession` | conversation variants | loader | backend-practice | session/messages | none | P0.044/P0.046 |
| `sendPracticeMessage` | happy/failure/replay | message hook | backend-practice | practice_messages | practice.session.chat | P0.044/P0.046 |
| `completePracticeSession` | zero-answer-rejected/one-answer-ready/replay | finish hook with localized disabled reason | backend-practice authoritative validation | zero-answer none；success report/job | report async after valid completion | P0.047 |
| `getTargetJob` | default/errors | target display | backend-targetjob | target | none | P0.045 |
| `createPracticeVoiceTurn` | disabled negative | none | fail-closed | none | none | P0.007/P0.045 |

## 3 质量门禁分类

- **Plan 类型**: user-visible UI + API consumer + refactor。
- **TDD 策略**: source prototype contract tests first, then formal DOM/hooks; each phase has focused Vitest assertions.
- **BDD 策略**: P0.044 happy conversation, P0.045 simplified/disabled UI, P0.046 recovery, P0.047 completion; P0.099 real fullstack screenshot closure.
- **替代验证 gate**: source parity, computed style/bounding box, desktop/mobile screenshot, typecheck/build, stale-contract grep.

## 4 Coverage Matrix

| Source | Category | Phase | Verification | UI anchor | Negative |
|--------|----------|-------|--------------|-----------|----------|
| full-width chat | source structure | 1-2 | prototype/UI contract tests | screen-practice::PracticeScreen | SessionMap/QuestionCard |
| geometry | UX/visual | 1/5 | Playwright bbox/screenshots | updated conversation layout | 260px sidebar gap |
| ordered messages | primary | 3 | Vitest + P0.044 | Transcript/Composer | question/followUp labels |
| send failure | recovery | 3 | Vitest + P0.046 | Error/retry | duplicate user message |
| disabled phone | alternate/negative | 2 | DOM/a11y + P0.045/P0.007 | disabled topbar icon | PhoneSurface/click handler |
| completion | primary | 4 | P0.047 | Finish CTA | hint/mode handoff |
| real screenshot | integration | 5 | P0.099 + browser artifacts | desktop 1440x900/mobile | fixture-only evidence |
| retry routing | failure/recovery | 6 | PracticeScreen Vitest + P0.046/P0.047 | ErrorState retry | completion retry calls send; loader retry absent |
| finish race guard | lifecycle/UX | 6 | PracticeScreen Vitest | Finish CTA | finish enabled while send/loading/completing |
| zero-answer finish | boundary/a11y | 7 | PracticeScreen+i18n tests + P0.047 backend marker | Finish CTA + described reason | opening/draft/route counted as answer; UI treated as authority |

## 5 实施步骤

### Phase 1: UI truth source
- Rewrite `ui-design/src/screen-practice.jsx` and data to TopBar + full-width Conversation.
- Remove all question/hint/phone-positive prototype state/components/copy.
- Update docs/ui-design and prototype contract/parity expectations.

### Phase 2: Formal screen structure
- Delete SessionMap, QuestionCard, PhoneSurface and hint components/controller/hooks.
- Simplify TopBar props; disabled phone button has native disabled/a11y/copy.
- Remove mode/modality/practiceMode/hint route/context/handoff fields.

### Phase 3: Message hooks and states
- Loader consumes session.messages.
- New message hook sends/retries `clientMessageId`, adopts server message pair and prevents duplicates.
- Cover loading/running/sending/error/retry/local-paused/completing/session-lost states；local pause 不调用 backend event API，刷新后回到 Running。

### Phase 4: Completion/report handoff
- Finish sends only clientCompletedAt and navigates with stable `reportId`.
- 本阶段只拥有 completion handoff；GeneratingScreen 的状态、文案与动作自 2026-07-12 起由 `frontend-report-dashboard/001` 唯一承接。

### Phase 5: Parity and real scenario
- Run Vitest/typecheck/build/UI contract/pixel parity.
- Reset/redeploy real local environment, run P0.099 path and capture redacted 1440x900 screenshots of conversation and report handoff.

### Phase 6: Review remediation
- Track the error source and bind retry to `loader.refresh`, same-ID message send, or completion retry as appropriate.
- Disable Finish CTA while message send/session load/completion is active or the session is no longer mutable.
- Extend P0.046/P0.047 evidence so the screen-level recovery actions execute, not only the underlying hooks.

### Phase 7: Zero-answer finish eligibility and backend authority

- Derive frontend eligibility only from server-loaded `messages`: at least one committed candidate `user` message, no pending assistant reply, and the existing mutable/not-loading/not-sending/not-completing guards must all hold. Opening assistant content, composer drafts and route state never count.
- In `ui-design/src/screen-practice.jsx` first, then formal PracticeScreen, render Finish as native disabled for zero answers and expose a nearby zh/en reason with stable `aria-describedby`; the reason disappears when the first committed user message makes the action eligible.
- Keep backend authoritative: direct zero-answer `completePracticeSession` returns typed `VALIDATION_FAILED`, leaves the session mutable and writes no report/job/outbox/idempotency success. Frontend tests prove UX; backend-practice/002 Phase 9 supplies service/store/API/PostgreSQL evidence.
- Refresh P0.047 owner assertions to compose frontend `ZERO_ANSWER_FINISH_DISABLED_PASS` with backend `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, then prove one-answer completion and exact replay still succeed.

### Phase 8: reportId-only completion handoff

- Treat the completion response `reportId` as the only navigation locator. PracticeScreen must navigate to Generating with no copied `targetJobId`, `planId`, `sessionId`, `resumeId`, `roundId`, `roundName`, status or error fields in query, route state or screen context.
- RED/GREEN route tests first prove the current multi-field handoff, then require exact route/state shape `{reportId}` and reject restoring any copied business identifier. Generating/Report fetch all state and frozen context from `getFeedbackReport(reportId)` under the frontend-report owner.
- Refresh E2E.P0.047 after one-answer completion to assert the browser URL/history state and downstream API request use only reportId; idempotent replay returns the same locator without duplicating report state.

## 6 验收标准

- No left rail, question count or QuestionCard at any viewport.
- Only ordered chat and composer occupy the body.
- Phone icon is visible, grey and not actionable.
- Refresh/retry/complete work against real backend.
- Source and formal screenshots/geometry match.
- Zero-answer Finish is natively disabled with a localized accessible reason; one committed candidate message enables it only when all existing lifecycle guards also pass, while backend independently rejects direct zero-answer completion.
- Completion handoff exposes only reportId; no mutable business identity or report status is copied through navigation state.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.4 | Reopen Phase 8 to enforce reportId-only completion navigation and remove six copied business identifiers from PracticeScreen handoff. |
| 2026-07-12 | 2.3 | Reopen Phase 7 for zero-answer Finish eligibility, localized accessible reason and composed backend-authoritative P0.047 evidence. |
| 2026-07-12 | 2.2 | Clarify that this owner stops at stable reportId handoff; GeneratingScreen is exclusively owned by frontend-report-dashboard/001. |
| 2026-07-12 | 2.1 | Reopen for source-aware retry wiring and send/complete UI race guards. |
| 2026-07-12 | 2.0 | Reopen for full-width continuous chat and disabled phone entry. |
