# 002 — Practice Continuous Text Conversation

> **版本**: 2.0
> **状态**: completed
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
| `completePracticeSession` | default | finish hook | backend-practice | report/job | report async | P0.047 |
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

### Phase 4: Completion/generating handoff
- Finish sends only clientCompletedAt and navigates with stable IDs.
- Generating uses conversation-level progress copy.

### Phase 5: Parity and real scenario
- Run Vitest/typecheck/build/UI contract/pixel parity.
- Reset/redeploy real local environment, run P0.099 path and capture redacted 1440x900 screenshots of conversation and report handoff.

## 6 验收标准

- No left rail, question count or QuestionCard at any viewport.
- Only ordered chat and composer occupy the body.
- Phone icon is visible, grey and not actionable.
- Refresh/retry/complete work against real backend.
- Source and formal screenshots/geometry match.

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-12 | 2.0 | Reopen for full-width continuous chat and disabled phone entry. |
