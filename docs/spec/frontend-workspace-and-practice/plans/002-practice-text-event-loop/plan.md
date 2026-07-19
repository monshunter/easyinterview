# 002 — Practice Continuous Text Conversation

> **版本**: 3.1
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 Practice UI 维持为 Top Bar + 全宽连续聊天。Phase 11 为 user/assistant persisted text 增加安全 Markdown/GFM view projection：`react-markdown + remark-gfm`，启用 `skipHtml` 且不使用 `rehypeRaw`，禁止 remote image/unsafe URI；发送和 same-ID retry 始终使用原始 `message.text/clientMessageId`。

## 2 Operation Matrix

| operationId | fixture current / planned | frontend consumer | backend handler / service / store | persistence | AI | scenario |
|-------------|---------------------------|-------------------|-----------------------------------|-------------|----|----------|
| `getPracticeSession` | current fixtures | loader/rehydration + safe Markdown/GFM | practice read owner | messages/reply facts | none | 当前无 text-loop E2E owner；root `make test` |
| `sendPracticeMessage` | current fixtures | send/retry exact raw text | practice send owner | messages/reply facts | `practice.session.chat` | 当前无真实 E2E owner；root `make test` |
| `completePracticeSession` | current fixtures | Finish CTA | practice completion owner | session/report/job/outbox/idempotency | report job after completion | `E2E.P0.098` 仅 completion API 与 progress refresh；不覆盖 text loop |
| `getTargetJob` | current progress fixtures | target display | targetjob read owner | requirements/progress | none | `E2E.P0.098` 仅 progress/detail read |
| `createPracticeVoiceTurn` | disabled fixture | none | fail-closed handler | none | none | 当前无真实 E2E owner；root `make test` |

## 3 质量门禁分类

- **Plan 类型**: user-visible UI + API consumer + refactor。
- **TDD 策略**: focused source/client/DOM/hook tests provide development feedback；阶段完成由根 `make test` 承接，禁止用 `Error.message` 或浏览器 storage 绕过合同。
- **BDD 策略**: text chat、session start 与 recovery 保留行为合同但当前无真实 E2E owner；`E2E.P0.098` 不覆盖这些行为。
- **替代验证 gate**: formal implementation contract, computed style/bounding box, desktop/mobile screenshot, typecheck/build, stale-contract grep.

## 4 Coverage Matrix

| Source | Category | Phase | Verification | UI anchor | Negative |
|--------|----------|-------|--------------|-----------|----------|
| full-width chat | source structure | 1-2 | prototype/UI contract tests | screen-practice::PracticeScreen | SessionMap/QuestionCard |
| geometry | UX/visual | 1/5 | Playwright bbox/screenshots | updated conversation layout | 260px sidebar gap |
| finish race guard | lifecycle/UX | 6 | PracticeScreen Vitest | Finish CTA | finish enabled while send/loading/completing |

## 5 实施步骤

### Phase 1: UI design document
- Rewrite `frontend/src` and data to TopBar + full-width Conversation.
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
- Run Vitest/typecheck/build plus formal component/responsive/accessibility assertions.
- Run the Practice owner scenarios and capture redacted desktop/mobile conversation screenshots; report-page real screenshot ownership remains in `frontend-report-dashboard`.

### Phase 6: Review remediation
- Track the error source and bind retry to `loader.refresh`, same-ID message send, or completion retry as appropriate.
- Disable Finish CTA while message send/session load/completion is active or the session is no longer mutable.

### Phase 7: Zero-answer finish eligibility and backend authority

- Derive frontend eligibility only from server-loaded `messages`: at least one committed candidate `user` message, no pending assistant reply, and the existing mutable/not-loading/not-sending/not-completing guards must all hold. Opening assistant content, composer drafts and route state never count.
- In `frontend/src` first, then formal PracticeScreen, render Finish as native disabled for zero answers and expose a nearby zh/en reason with stable `aria-describedby`; the reason disappears when the first committed user message makes the action eligible.
- Keep backend authoritative: direct zero-answer `completePracticeSession` returns typed `VALIDATION_FAILED`, leaves the session mutable and writes no report/job/outbox/idempotency success. Frontend tests prove UX; backend-practice/002 Phase 9 supplies service/store/API/PostgreSQL evidence.

### Phase 8: reportId-only completion handoff

- Treat the completion response `reportId` as the only navigation locator. PracticeScreen must navigate to Generating with no copied `targetJobId`, `planId`, `sessionId`, `resumeId`, `roundId`, `roundName`, status or error fields in query, route state or screen context.
- RED/GREEN route tests first prove the current multi-field handoff, then require exact route/state shape `{reportId}` and reject restoring any copied business identifier. Generating/Report fetch all state and frozen context from `getFeedbackReport(reportId)` under the frontend-report owner.

### Phase 9: Immediate user message, thinking state and row-local retry

- Dependency: execute and close the pre-existing Phase 7.3/7.4 and 8.3 checklist gates first; Phase 9 must not bypass the active owner plan's original order.
- Update `frontend/src` first: submit immediately appends one user row and clears composer；pending/retrying disables composer and renders an assistant-style accessible thinking animation；failure removes thinking and renders one retry icon only beneath that failed user row.
- **Contract dependency gate**: OpenAPI owner must generate user-message `clientMessageId + replyStatus=pending|retryable_failed|terminal_failed|complete` and typed `ApiClientError.apiError.retryable` while preserving HTTP status, `code/requestId/retryable/details` and transport cause；backend owner must durably project reply status. The planned fixture variants in the matrix are missing current work, not historical PASS.
- Formal Practice keeps `{text, clientMessageId, status}` only as in-memory submit-to-first-response/read feedback. Reload/remount reconstructs every unresolved state from `getPracticeSession`; retry identity must never enter URL, `localStorage`, `sessionStorage` or IndexedDB.
- A reloaded `pending` row keeps composer/Finish disabled, shows thinking, performs bounded single-flight re-read and never sends again. Reloaded `retryable_failed` restores exactly one icon under the server row；`terminal_failed` has no icon and enters fact recovery；`complete` renders the unique user/reply pair.
- Retry calls the same send path with the server/original text and same `clientMessageId`. After failure, textarea may retain a new draft, but submit remains disabled with localized guidance until row-local retry succeeds；retry never consumes or replaces that draft.
- Row-local retry requires transport failure without an HTTP response or typed `ApiClientError.apiError.retryable=true`. Intentional abort/unmount, validation, auth, not-found, conflict and mismatch are terminal and recover through loader/auth/session-lost server truth；Practice must not parse `Error.message`.
- RED/GREEN covers deferred success, typed retryable/terminal classification, reload during `pending`, AI failure → reload → same-ID retry → exactly one user/reply pair, Finish disabled across unresolved states, retry-in-flight, repeated failure, draft preservation, no retry before failure and accessible disabled/thinking/retry semantics.

### Phase 10: Lease-aligned timeout reconciliation and terminal plan recovery

- **10.1 UI source first**: add failing `ui-design` contract cases, then update `screen-practice.jsx` so persisted terminal data can render a generic safe recovery state with one “返回当前面试规划 / Back to this interview plan” CTA. The prototype must preserve injected `replyStatus` instead of forcing every user row to complete, and it must expose pending/retryable/terminal states for parity tests.
- **10.2 Request plumbing**: add RED hook tests that require `usePracticeMessages.sendMessage(submission, { signal })` to forward `AbortSignal` through the generated client request options；add cancellable/bounded session reads and cleanup aborts. Do not hand-edit generated client output.
- **10.3 95-second timeout and reconcile**: with fake timers, prove the POST is still pending before 95,000 ms, aborts exactly at 95,000 ms, and immediately starts an independently bounded `getPracticeSession` reconciliation for the same `clientMessageId`. Adopt authoritative complete/pending/retryable/terminal rows；if the ID is absent or read fails, keep the original row/ID unresolved and lock new-ID submit/Finish. A request sequence/generation guard must ignore any late response from the aborted POST or an older reconcile.
- **10.4 Reload and loader safety**: reloaded pending remains thinking/locked and only re-reads；backend 90-second lease/GET lazy convergence ends immortal pending. A refresh/read failure must preserve the last same-session unresolved facts when available, or fail-locked when unavailable；it may never clear the pending identity and enable a new ID.
- **10.5 Terminal recovery (historical Phase 10 contract)**: authoritative `terminal_failed` has no row-local retry. Phase 10 originally routed the single localized CTA to `parse(targetJobId)`；Phase 11 supersedes that navigation target while retaining the established no-retry, no-composer-submit and no-technical-copy behavior. Auth/session-lost keep their existing global recovery owners.

### Phase 11: Safe Markdown/GFM conversation projection

- Add a shared message-body renderer using `react-markdown` with `remark-gfm`. Use `skipHtml`; do not add `rehypeRaw`. Both persisted user and assistant messages use it; optimistic raw user text must converge to the same projection after server adoption.
- Disable remote images entirely: image Markdown must not create a network-fetching `<img>`. Raw HTML is not rendered. Links allow only safe protocols, open with hardened `rel="noopener noreferrer"` when external/new-tab behavior applies, and reject `javascript:`/unsafe URI. No script/event-handler execution is possible.
- Preserve exact raw `message.text` and `clientMessageId` as business payload. Renderer output/DOM text/normalized Markdown must never feed send or retry. Same-ID retry sends the exact server/original raw text byte-for-byte while preserving the next draft.
- Render GFM headings, emphasis, links, lists, blockquotes, inline/fenced code and tables within existing prototype-derived message typography. On 390px mobile, pre/code/table content stays inside the message viewport using local horizontal scrolling/wrapping rules; document overflow is zero.
- Supersede the Phase 10 terminal CTA target: authoritative `terminal_failed` must call exactly `navigate({ name: "workspace", params: { targetJobId: loader.data.targetJobId } })` and resolve to `/workspace?targetJobId=...` read-only detail. Query-free workspace, `parse(targetJobId)`, `planId`, row retry, composer submit and technical error text are negative gates; active spec/plan/tests/source must have zero positive current-scope `parse(targetJobId)` recovery references.

### Phase 12: Required runtime message and session guards

- Practice consumes required `AppRuntimeProvider.contentLimits.practiceMessageBytes/practiceSessionTextBytes` and the shared UTF-8 byte helper。Focused tests inject small message/session values；overflow preserves the draft and makes zero send calls without constructing default-sized strings；existing DOM/styles, raw retry and pending state remain unchanged。
- Required 子字段无 per-field fallback；只有整体 runtime source 不可用时可沿用既有 bootstrap fallback。

### Phase 13: Reference-aligned active interview surface

以提供的 active Practice 参考图为 desktop 视觉合同，保留真实 session/messages/completion 与所有 pending/retry/terminal 语义，重构 `PracticeScreen`、Session Header、Transcript、Composer 和 Finish/Send controls。先由现有 `PracticeScreen.test.tsx` / `Transcript.test.tsx` 与新增视觉合同证明 `calc(100dvh - 76px)`、共享内容边界、正式 SVG icon、消息 surface、composer 高度和 390px containment，再移除对应内联视觉样式。Chrome 使用正式 frontend repository fixture 验收视觉状态；未运行的真实 active-session 业务动作不得报告 PASS。

### Phase 14: Fixed Composer and helper anchoring

`Transcript` 是会话卡内唯一滚动区，Composer 整体以 `flex: 0 0 auto` 固定在会话卡底部；聊天记录从短到长、滚动到任意位置都不得改变输入框坐标。说明胶囊从 `Transcript` 滚动内容移交给 `InputBar` / Composer 固定区。RED source/component gate 必须拒绝 `Transcript` 的 `helperText` prop 和 helper DOM，并要求 `InputBar` 在输入 shell 正上方拥有 helper；Chrome 分别构造短聊天与可滚动长聊天，证明 Composer 坐标和 helper/input 垂直间距不随消息数量或 Transcript scrollTop 改变。该说明只解释作答方式，不恢复已删除的业务 hint 能力。

### Phase 15: Composer inner-surface send anchoring

把 `InputBar` 中的 textarea 与 send 收进同一个内层 input surface，send 位于该表面内部的底部 action area 并右对齐；action area 不与 textarea 叠加，避免长文本、placeholder 或光标被覆盖，也不以大块右内边距压缩窄屏正文。RED source/component/CSS gate 先拒绝当前悬浮在内外边框之间的 actions row，并拒绝 overlay + 固定右内边距方案；GREEN 只调整 DOM 与 CSS，不改变发送点击、Ctrl/Meta+Enter、disabled、pending/retry、helper、Composer 固定位置或 API 合同。该阶段是纯 frontend 视觉结构修订，Operation Matrix 全部 `operationId`、fixture、handler、persistence 与 AI 状态保持不变；Chrome 在 desktop/mobile 验证内层边界、按钮 containment、文本完整宽度、固定 Composer 与零横向溢出。

## 6 验收标准

- No left rail, question count or QuestionCard at any viewport.
- Only ordered chat and composer occupy the body.
- Phone icon is visible, grey and not actionable.
- Refresh/retry/complete work against real backend.
- Source and formal screenshots/geometry match.
- Zero-answer Finish is natively disabled with a localized accessible reason; one committed candidate message enables it only when all existing lifecycle guards also pass, while backend independently rejects direct zero-answer completion.
- Completion handoff exposes only reportId; no mutable business identity or report status is copied through navigation state.
- User messages appear immediately；reload restores server `clientMessageId + replyStatus` without browser persistence；pending/retryable-failed/retrying/terminal-recovery all block Finish；thinking only appears while awaiting a response；retry exists only beneath an explicitly retryable failed user row, reuses server original text and identity, and AI failure → reload → retry converges to one server user/reply pair。
- A send never waits beyond 95 seconds without reconciliation；timeout uses the same ID, preserves fail-locked state on uncertain reads and ignores stale responses. Terminal failure offers one generic CTA to the exact current `/workspace?targetJobId` read-only detail and never a row retry, query-free workspace fallback or `parse(targetJobId)` recovery.
- Persisted user and assistant text renders through safe Markdown/GFM; raw HTML, remote images and unsafe URIs cannot execute/request; safe links are hardened; retry preserves exact raw text/clientMessageId; mobile code/table content does not overflow the document.
- Runtime guard 使用 required fields；默认/override/invalid 归 A4，frontend 以小型 injected values 验证 overflow 保留 draft 且 zero send，backend aggregate remains authoritative。
- Send 与 textarea 共享一个内层 input surface；按钮在不覆盖 textarea 的底部 action area 右对齐，窄屏文本保持完整宽度，且按钮不得悬浮在内外边框之间。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-20 | 3.1 | Add Phase 15 so textarea and send share one input surface with a non-overlapping bottom action area and full-width narrow-screen text. |
| 2026-07-19 | 3.0 | Add Phase 14 so the helper capsule belongs to the fixed Composer instead of the scrolling Transcript. |
| 2026-07-14 | 2.7 | Add Phase 11 safe react-markdown/remark-gfm projection, security negatives, exact raw retry, mobile code-overflow gates, and supersede terminal recovery to Workspace detail. |
| 2026-07-19 | 2.9 | Reopen Phase 13 for the supplied active-interview reference: available viewport height, shared content grid, structured session controls, message surfaces and large composer. |
| 2026-07-12 | 2.4 | Reopen Phase 8 to enforce reportId-only completion navigation and remove six copied business identifiers from PracticeScreen handoff. |
| 2026-07-12 | 2.2 | Clarify that this owner stops at stable reportId handoff; GeneratingScreen is exclusively owned by frontend-report-dashboard/001. |
| 2026-07-12 | 2.1 | Reopen for source-aware retry wiring and send/complete UI race guards. |
| 2026-07-12 | 2.0 | Reopen for full-width continuous chat and disabled phone entry. |
