# 002 — Practice Continuous Text Conversation

> **版本**: 2.8
> **状态**: active
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 Test Plan**: [test-plan](./test-plan.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 1 目标

把 Practice UI 维持为 Top Bar + 全宽连续聊天。Phase 11 为 user/assistant persisted text 增加安全 Markdown/GFM view projection：`react-markdown + remark-gfm`，启用 `skipHtml` 且不使用 `rehypeRaw`，禁止 remote image/unsafe URI；发送和 same-ID retry 始终使用原始 `message.text/clientMessageId`。

## 2 Operation Matrix

| operationId | fixture current / planned | frontend consumer | backend handler / service / store | persistence | AI | scenario |
|-------------|---------------------------|-------------------|-----------------------------------|-------------|----|----------|
| `getPracticeSession` | current fixtures unchanged | loader/rehydration + safe Markdown/GFM projection of raw user/assistant text | existing handler/service/repository | raw text/reply facts unchanged | none | `E2E.P0.044`, `E2E.P0.046` |
| `sendPracticeMessage` | current fixtures unchanged | send/retry exact raw text + Markdown/GFM view projection | existing handler/service/store | raw text/clientMessageId unchanged | `practice.session.chat` | `E2E.P0.044`, `E2E.P0.046` |
| `completePracticeSession` | `openapi/fixtures/PracticeSessions/completePracticeSession.json`: current `default`, `replay`, `mismatch`, `cross-user-not-found`, `session-already-completed`; **planned** `zero-answer-rejected`, `one-answer-ready` | `useCompletePracticeSession` + Finish CTA | `Handler.CompletePracticeSession` → `Service.CompletePracticeSession` → completion store transaction | zero-answer none；success session/report/job/outbox/idempotency | report async after valid completion | `E2E.P0.047` |
| `getTargetJob` | `openapi/fixtures/TargetJobs/getTargetJob.json`: current `default`, `not-started-progress`, `all-completed-progress`, `prototype-baseline` | `usePracticeTargetDisplay` | `targetjob.Handler.GetTargetJob` → `Service.GetTargetJob` → `SQLStore.GetTargetJobByUser` | target requirements/progress projection | none | `E2E.P0.045`, `E2E.P0.098` |
| `createPracticeVoiceTurn` | `openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json`: current `default` disabled response | none | `practice.Handler.CreatePracticeVoiceTurn` fail-closed | none | none | `E2E.P0.007`, `E2E.P0.045` |

## 3 质量门禁分类

- **Plan 类型**: user-visible UI + API consumer + refactor。
- **TDD 策略**: source prototype contract tests first, then generated typed-error/client contract, formal DOM/hooks and server-rehydration states；Phase 9 先以当前缺口建立 RED，再由 OpenAPI/backend owner 落地依赖，禁止用 `Error.message` 或浏览器 storage 绕过合同。Phase 10 严格按 UI source RED/GREEN → hook signal/timeout RED/GREEN → screen reconcile/route RED/GREEN → parity → fresh BDD evidence 执行。
- **BDD 策略**: P0.044 happy/pending conversation, P0.045 simplified/disabled UI, P0.046 message recovery, P0.047 completion；本次 Practice desktop/mobile 截图由 P0.044/P0.046 与 pixel-parity gate 闭环，report full-stack screenshot scenario 继续由 report owner 独占。
- **替代验证 gate**: source parity, computed style/bounding box, desktop/mobile screenshot, typecheck/build, stale-contract grep.

## 4 Coverage Matrix

| Source | Category | Phase | Verification | UI anchor | Negative |
|--------|----------|-------|--------------|-----------|----------|
| full-width chat | source structure | 1-2 | prototype/UI contract tests | screen-practice::PracticeScreen | SessionMap/QuestionCard |
| geometry | UX/visual | 1/5 | Playwright bbox/screenshots | updated conversation layout | 260px sidebar gap |
| ordered messages | primary | 3/9 | Vitest + P0.044 | Transcript/Composer/optimistic or server-rehydrated row/thinking | delayed/duplicated user row; question/followUp labels |
| send failure | recovery | 3/6/9 | typed generated error + Vitest + P0.046 | failed user-row retry icon | `Error.message` parsing; global retry; duplicate user message; changed retry payload |
| reload after AI failure | recovery/idempotency | 9 | server projection + P0.046 | original text/clientMessageId/replyStatus | browser retry persistence; new ID after reload; duplicate reply |
| terminal send failure | failure/authority | 9 | generated error classification + P0.046 | loader/auth/session-lost recovery | retry loop on validation/auth/not-found/conflict/mismatch |
| disabled phone | alternate/negative | 2 | DOM/a11y + P0.045/P0.007 | disabled topbar icon | PhoneSurface/click handler |
| completion | primary | 4 | P0.047 | Finish CTA | hint/mode handoff |
| pending/failure screenshot | integration/visual | 9 | P0.044/P0.046 + pixel parity + browser artifacts | desktop 1440/mobile 390 | retry before failure; missing thinking; fixture-only evidence |
| retry routing | failure/recovery | 6 | PracticeScreen Vitest + P0.046/P0.047 | ErrorState retry | completion retry calls send; loader retry absent |
| finish race guard | lifecycle/UX | 6 | PracticeScreen Vitest | Finish CTA | finish enabled while send/loading/completing |
| zero-answer finish | boundary/a11y | 7 | PracticeScreen+i18n tests + P0.047 backend marker | Finish CTA + described reason | opening/draft/route counted as answer; UI treated as authority |
| bounded POST reconciliation | recovery/timeout | 10 | fake-timer hook/screen tests + P0.046 | optimistic row/thinking/reconcile | wait forever; blind resend; new ID; stale response overwrites truth |
| lease-bounded reload | recovery/persistence | 10 | server fixture states + P0.044/P0.046 | pending → retryable/complete | client clock mutates server status; loader error unlocks submit |
| terminal plan recovery | failure/navigation/a11y | 10/11 | prototype/formal route+i18n tests + P0.046 | generic terminal state + `/workspace?targetJobId` detail CTA | row retry; query-free workspace fallback; `parse(targetJobId)`; planId; technical copy |
| evidence freshness | BDD/visual | 10 | shared source fingerprint + screenshot hashes/parity | P0.044/P0.046 artifacts | historical screenshots accepted after source change |
| Markdown/GFM projection | content/security/visual | 11 | renderer/unit/component + P0.044/P0.046 | user/assistant message body | raw HTML/remote image/unsafe URI; retry from DOM text; mobile code overflow |

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
- Run the Practice owner scenarios and capture redacted desktop/mobile conversation screenshots; report-page real screenshot ownership remains in `frontend-report-dashboard`.

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

### Phase 9: Immediate user message, thinking state and row-local retry

- Dependency: execute and close the pre-existing Phase 7.3/7.4 and 8.3 checklist gates first; Phase 9 must not bypass the active owner plan's original order.
- Update `ui-design/src/screen-practice.jsx` first: submit immediately appends one user row and clears composer；pending/retrying disables composer and renders an assistant-style accessible thinking animation；failure removes thinking and renders one retry icon only beneath that failed user row.
- **Contract dependency gate**: OpenAPI owner must generate user-message `clientMessageId + replyStatus=pending|retryable_failed|terminal_failed|complete` and typed `ApiClientError.apiError.retryable` while preserving HTTP status, `code/requestId/retryable/details` and transport cause；backend owner must durably project reply status. The planned fixture variants in the matrix are missing current work, not historical PASS.
- Formal Practice keeps `{text, clientMessageId, status}` only as in-memory submit-to-first-response/read feedback. Reload/remount reconstructs every unresolved state from `getPracticeSession`; retry identity must never enter URL, `localStorage`, `sessionStorage` or IndexedDB.
- A reloaded `pending` row keeps composer/Finish disabled, shows thinking, performs bounded single-flight re-read and never sends again. Reloaded `retryable_failed` restores exactly one icon under the server row；`terminal_failed` has no icon and enters fact recovery；`complete` renders the unique user/reply pair.
- Retry calls the same send path with the server/original text and same `clientMessageId`. After failure, textarea may retain a new draft, but submit remains disabled with localized guidance until row-local retry succeeds；retry never consumes or replaces that draft.
- Row-local retry requires transport failure without an HTTP response or typed `ApiClientError.apiError.retryable=true`. Intentional abort/unmount, validation, auth, not-found, conflict and mismatch are terminal and recover through loader/auth/session-lost server truth；Practice must not parse `Error.message`.
- RED/GREEN covers deferred success, typed retryable/terminal classification, reload during `pending`, AI failure → reload → same-ID retry → exactly one user/reply pair, Finish disabled across unresolved states, retry-in-flight, repeated failure, draft preservation, no retry before failure and accessible disabled/thinking/retry semantics.
- P0.044/P0.046 plus formal/prototype pixel parity capture pending and failed states at exact 1440 desktop and 390 mobile viewports, including DOM/computed-style/bbox/viewport and screenshot evidence.

### Phase 10: Lease-aligned timeout reconciliation and terminal plan recovery

- **10.1 UI source first**: add failing `ui-design` contract cases, then update `screen-practice.jsx` so persisted terminal data can render a generic safe recovery state with one “返回当前面试规划 / Back to this interview plan” CTA. The prototype must preserve injected `replyStatus` instead of forcing every user row to complete, and it must expose pending/retryable/terminal states for parity tests.
- **10.2 Request plumbing**: add RED hook tests that require `usePracticeMessages.sendMessage(submission, { signal })` to forward `AbortSignal` through the generated client request options；add cancellable/bounded session reads and cleanup aborts. Do not hand-edit generated client output.
- **10.3 95-second timeout and reconcile**: with fake timers, prove the POST is still pending before 95,000 ms, aborts exactly at 95,000 ms, and immediately starts an independently bounded `getPracticeSession` reconciliation for the same `clientMessageId`. Adopt authoritative complete/pending/retryable/terminal rows；if the ID is absent or read fails, keep the original row/ID unresolved and lock new-ID submit/Finish. A request sequence/generation guard must ignore any late response from the aborted POST or an older reconcile.
- **10.4 Reload and loader safety**: reloaded pending remains thinking/locked and only re-reads；backend 90-second lease/GET lazy convergence ends immortal pending. A refresh/read failure must preserve the last same-session unresolved facts when available, or fail-locked when unavailable；it may never clear the pending identity and enable a new ID.
- **10.5 Terminal recovery (historical Phase 10 contract)**: authoritative `terminal_failed` has no row-local retry. Phase 10 originally routed the single localized CTA to `parse(targetJobId)`；Phase 11 supersedes that navigation target while retaining the established no-retry, no-composer-submit and no-technical-copy behavior. Auth/session-lost keep their existing global recovery owners.
- **10.6 Parity and fresh evidence**: formal/prototype DOM, computed style, key bounding boxes, viewport overflow and screenshot diff must cover immediate pending, persisted pending, retryable failure and terminal CTA at desktop 1440 and mobile 390. P0.044/P0.046 share a tracked `test/scenarios/e2e/practice-source-fingerprint-paths.json` manifest spanning UI docs/source, formal Practice/i18n/route/client, OpenAPI + fixtures, backend practice/store/migration and both scenario directories. Trigger records SHA-256；verify recomputes it and records each PNG SHA-256/dimensions/viewport. Source drift invalidates every historical artifact.
- Required markers: P0.044 emits `PRACTICE_IMMEDIATE_PENDING_PASS`, `PRACTICE_PERSISTED_PENDING_PASS`, `PRACTICE_EVIDENCE_FINGERPRINT_PASS`；P0.046 emits `PRACTICE_PENDING_LEASE_RECOVERY_PASS`, `PRACTICE_STALE_GENERATION_FENCED_PASS`, `PRACTICE_CONCURRENT_RESERVATION_PASS`, `PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS`, `PRACTICE_TERMINAL_PLAN_RECOVERY_PASS`, `PRACTICE_EVIDENCE_FINGERPRINT_PASS`.

### Phase 11: Safe Markdown/GFM conversation projection

- Add a shared message-body renderer using `react-markdown` with `remark-gfm`. Use `skipHtml`; do not add `rehypeRaw`. Both persisted user and assistant messages use it; optimistic raw user text must converge to the same projection after server adoption.
- Disable remote images entirely: image Markdown must not create a network-fetching `<img>`. Raw HTML is not rendered. Links allow only safe protocols, open with hardened `rel="noopener noreferrer"` when external/new-tab behavior applies, and reject `javascript:`/unsafe URI. No script/event-handler execution is possible.
- Preserve exact raw `message.text` and `clientMessageId` as business payload. Renderer output/DOM text/normalized Markdown must never feed send or retry. Same-ID retry sends the exact server/original raw text byte-for-byte while preserving the next draft.
- Render GFM headings, emphasis, links, lists, blockquotes, inline/fenced code and tables within existing prototype-derived message typography. On 390px mobile, pre/code/table content stays inside the message viewport using local horizontal scrolling/wrapping rules; document overflow is zero.
- Extend P0.044 for user+assistant GFM and desktop/mobile parity; extend P0.046 for malicious raw HTML/remote image/unsafe link negatives and exact raw same-ID retry. Existing scenario IDs remain; no sibling.
- Supersede the Phase 10 terminal CTA target: authoritative `terminal_failed` must call exactly `navigate({ name: "workspace", params: { targetJobId: loader.data.targetJobId } })` and resolve to `/workspace?targetJobId=...` read-only detail. Query-free workspace, `parse(targetJobId)`, `planId`, row retry, composer submit and technical error text are negative gates; active spec/plan/tests/source must have zero positive current-scope `parse(targetJobId)` recovery references.

### Phase 12: Runtime-configured message and session byte limits

- RED tests use ASCII and multibyte UTF-8 inputs at 32KiB/32KiB+1 and server-loaded session totals at 256KiB/256KiB+1；the old 8,000-rune path must fail.
- GREEN consumes `AppRuntimeProvider.contentLimits.practiceMessageBytes/practiceSessionTextBytes`, uses the shared byte helper and A4-matching missing-field defaults. Overflow preserves the draft and makes zero send calls; existing DOM/styles, raw retry and pending state remain unchanged.
- `BDD-Gate: E2E.P0.046` proves limit/+1, reload, server-authoritative aggregate and typed backend rejection. Optimistic rows are never persisted or counted as final facts.

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
- P0.044/P0.046 evidence is current only when the tracked source fingerprint and every screenshot SHA-256/geometry match at verify time.
- Persisted user and assistant text renders through safe Markdown/GFM; raw HTML, remote images and unsafe URIs cannot execute/request; safe links are hardened; retry preserves exact raw text/clientMessageId; mobile code/table content does not overflow the document.
- Runtime/default message/session limits are 32KiB/256KiB UTF-8 bytes；limit sends, limit+1 preserves draft with zero send，backend aggregate remains authoritative。

## 7 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-14 | 2.8 | Reopen Phase 12 for RuntimeConfig 32KiB message / 256KiB session UTF-8 limits and removal of 8,000-rune local truth. |
| 2026-07-14 | 2.7 | Add Phase 11 safe react-markdown/remark-gfm projection, security negatives, exact raw retry, mobile code-overflow gates, and supersede terminal recovery to Workspace detail. |
| 2026-07-14 | 2.6 | Add Phase 10 for a 95-second abort-and-reconcile timeout aligned to the backend lease, stale-response guards, generic terminal recovery to `parse(targetJobId)`, parity and fingerprint-bound P0.044/P0.046 evidence. |
| 2026-07-13 | 2.5 | Add immediate optimistic user rows, interviewer-thinking, row-local same-ID retry, draft-safe failure handling and P0.044/P0.046 desktop/mobile screenshot closure. |
| 2026-07-12 | 2.4 | Reopen Phase 8 to enforce reportId-only completion navigation and remove six copied business identifiers from PracticeScreen handoff. |
| 2026-07-12 | 2.3 | Reopen Phase 7 for zero-answer Finish eligibility, localized accessible reason and composed backend-authoritative P0.047 evidence. |
| 2026-07-12 | 2.2 | Clarify that this owner stops at stable reportId handoff; GeneratingScreen is exclusively owned by frontend-report-dashboard/001. |
| 2026-07-12 | 2.1 | Reopen for source-aware retry wiring and send/complete UI race guards. |
| 2026-07-12 | 2.0 | Reopen for full-width continuous chat and disabled phone entry. |
