# 002 — Practice Continuous Text Conversation Checklist

> **版本**: 3.1
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: UI design document
- [x] 1.1 RED-GREEN: update prototype tests/source to TopBar + full-width Conversation and delete question/hint/phone-positive source.
- [x] 1.2 RED-GREEN: update desktop/mobile source geometry expectations and stale-contract negative checks.

## Phase 2: Formal screen structure
- [x] 2.1 RED-GREEN: delete SessionMap/QuestionCard/PhoneSurface/hint/controller components and simplify PracticeScreen/TopBar.
- [x] 2.2 RED-GREEN: disabled phone icon has native disabled/a11y/unavailable copy and no route/API action.
- [x] 2.3 RED-GREEN: remove mode/modality/practiceMode/hint context/handoff/i18n/test contracts.

## Phase 3: Message hooks and states
- [x] 3.1 RED-GREEN: loader renders ordered session.messages including refresh recovery.
- [x] 3.2 RED-GREEN: send hook handles success/replay/failure/same-ID retry without duplicate messages.
- [x] 3.3 RED-GREEN: loading/sending/error/local-paused/completing/session-lost states remain usable; pause has no backend event call and refresh resumes Running.

## Phase 4: Completion/generating
- [x] 4.1 RED-GREEN: finish handoff contains stable IDs only; generating copy is conversation-level.

> Ownership note (2026-07-12): the completed evidence above is historical. Current work stops at stable `reportId` handoff; GeneratingScreen is exclusively owned by `frontend-report-dashboard/001`.

## Phase 5: Parity and real scenario
- [x] 5.1 仓库根 `make test` 完成前后端全量单测回归；typecheck/build、UI contract 与 desktop/mobile component/responsive assertions 作为独立 gates。
- [x] 5.2 Run the then-current real backend/frontend path and capture redacted conversation/report screenshots.

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: PracticeScreen retries loader, message and completion failures through the correct operation and preserves message/completion idempotency. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx`)
- [x] 6.2 RED-GREEN: Finish CTA is disabled during send, load, completion and non-mutable session states. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx`; frontend typecheck)

## Phase 7: Zero-answer finish eligibility and backend authority

- [x] 7.1 RED-GREEN: PracticeScreen derives Finish eligibility only from server-loaded committed candidate `user` messages plus existing mutable/no-pending-reply/no-load/no-send/no-complete guards; opening assistant, composer draft and route params do not count. (`PracticeScreen.test.tsx` + completion hook tests)
- [x] 7.2 RED-GREEN: prototype and formal Finish are native disabled at zero answers and expose the same nearby zh/en reason through stable `aria-describedby`; first committed user message removes only the zero-answer reason. (ui-design source contract + i18n exact-set + DOM/a11y tests)
- [x] 7.3 RED-GREEN: direct zero-answer completion is still rejected by backend `VALIDATION_FAILED`, session remains mutable and no report/job/outbox/idempotency success is written; one-answer completion and replay remain green. (consume backend-practice/002 Phase 9 service/store/API/PostgreSQL markers; do not duplicate backend logic in frontend)

## Phase 8: reportId-only completion handoff

- [x] 8.1 RED-GREEN: PracticeScreen completion navigation has exact query/state/context shape `{reportId}`; tests first fail on and then reject `targetJobId|planId|sessionId|resumeId|roundId|roundName|status|error` copies while preserving same-reportId completion replay.
  <!-- verified: 2026-07-12 method=screen-router-and-privacy-tests evidence="PracticeScreen, App, routeUrl, pendingAction and routing privacy cases included in 111-file/762-test PASS" -->
- [x] 8.2 REGRESSION-GATE: active PracticeScreen/context/router code contains no positive write of those copied fields to generating/report navigation; frontend-report consumes `getFeedbackReport(reportId)` as the sole downstream authority.
  <!-- verified: 2026-07-12 method=active-negative-and-route-tamper evidence="report/generating out-of-scope tests PASS; Playwright canonicalizes hostile report/generating URLs to reportId only" -->
  <!-- verified: 2026-07-13 method=screen-route-scenario evidence="PracticeScreen first navigates to /generating?reportId=... with null history state; downstream getFeedbackReport is keyed only by reportId; forbidden identity/status keys are absent; completion replay keeps the stable owner locator." -->

## Phase 9: Immediate user message, thinking state and row-local retry

- [x] 9.1 RED-GREEN: prototype tests/source append one user row and clear composer synchronously, render accessible interviewer-thinking only while pending/retrying, and render retry only beneath a failed user row.
- [x] 9.2 CONTRACT-DEPENDENCY-GATE: OpenAPI-generated user `PracticeMessage` exposes `clientMessageId + replyStatus=pending|retryable_failed|terminal_failed|complete` and typed `ApiClientError.apiError.retryable` with HTTP/envelope/transport metadata；backend durably projects reply status；the operation-matrix recovery fixtures are current and validated.
  <!-- verified: 2026-07-13 method=openapi+backend+fixtures evidence="Generated role union and typed error compile; backend persists/projects all four user states; get/send fixture matrix now validates pending/retryable/terminal/complete plus validation/auth/not-found/conflict/mismatch/timeout cases." -->
- [x] 9.3 RED-GREEN: formal Practice keeps transient `{text, clientMessageId, status}` only until first response/read convergence；reload/remount rehydrates pending/retryable/terminal/complete solely from `getPracticeSession`, with no URL/browser-storage retry persistence or `Error.message` parsing.
- [x] 9.4 RED-GREEN: typed retryable failure invokes the shared send path with server original text + same `clientMessageId`, preserves row/draft and restores one icon after repeated failure；AI failure → reload → same-ID retry converges to one user/reply pair；pending re-read never duplicate-sends；terminal failures have no retry.
- [x] 9.5 REGRESSION-GATE: pending/retryable-failed/retrying/terminal-recovery all keep Finish disabled；focused generated-client/Practice hooks/screen/i18n/a11y tests, UI source contracts, full frontend, typecheck/build and active negative searches pass.

## Phase 10: Lease-aligned timeout reconciliation and terminal plan recovery

- [x] 10.1 RED-GREEN: UI source tests first fail, then prototype preserves injected pending/retryable/terminal statuses and renders one localized generic terminal recovery CTA with no row retry.
- [x] 10.2 RED: hook/screen fake-timer tests require AbortSignal forwarding, independently cancellable/bounded reads, exact 95,000 ms POST timeout, same-ID reconciliation and stale-response suppression before implementation.
- [x] 10.3 GREEN: `usePracticeMessages` forwards request signal through generated request options；session loader/reconcile abort on cleanup/timeout without hand-editing generated output.
- [x] 10.4 GREEN: PracticeScreen aborts POST at exactly 95 seconds and adopts reconciled complete/pending/retryable/terminal server truth；missing-ID/read-failure preserves the original unresolved row/ID and keeps new-ID submit/Finish locked；late old responses cannot overwrite newer truth.
  <!-- reopened: 2026-07-14 reason="Independent review found the untested inverse completion order where an older loader snapshot can invalidate and strand the newer same-ID reconciliation." -->
- [x] 10.5 RED-GREEN: reloaded pending only re-reads until backend 90-second lease convergence；loader refresh failure preserves prior unresolved facts or fails locked and can never unlock a new ID.
  <!-- reopened: 2026-07-14 reason="Independent review found missing same-mounted-component session isolation and stale local state reset coverage." -->
  <!-- reopened: 2026-07-14 reason="A persisted retryable_failed row followed by an online/focus refresh failure retained row-local retry while also rendering global loader retry and disabling the draft composer; typed HTTP retryable=true lacked direct UI recovery proof." -->
- [x] 10.6 HISTORICAL RED-GREEN: Phase 10 terminal state showed safe zh/en copy and routed “返回当前面试规划” to `parse(targetJobId)`；Phase 11 supersedes only this route target while retaining the verified no-retry/no-composer-submit/no-technical-copy behavior.
  <!-- reopened: 2026-07-14 reason="Independent review found that later authoritative terminal/complete truth can leave an earlier local ErrorState visible, producing duplicate failure UI." -->
- [x] 10.7 PARITY-GATE: prototype/formal immediate pending, persisted pending, retryable failure and terminal CTA match DOM/a11y/computed-style/bbox/viewport/screenshot at desktop 1440 and mobile 390.
  <!-- verified: 2026-07-14 method=playwright-responsive-browser+ui-contract evidence="practice.spec.ts passed 16/16 across desktop 1440 and mobile 390 with DOM, a11y, computed-style, bbox, viewport and screenshot comparison; ui-design contract passed 60/60." -->

## Phase 11: Safe Markdown/GFM projection and current-plan recovery

- [x] 11.1 RED-GREEN: add one shared `react-markdown + remark-gfm` message-body renderer and apply it to persisted user and assistant rows with `skipHtml` and no `rehypeRaw`.
- [x] 11.2 SECURITY-GATE: raw HTML is inert, remote images create no requestable `<img>`, `javascript:`/unsafe URIs are rejected, safe external links use hardened `rel`, and script/event handlers cannot execute.
- [x] 11.3 PAYLOAD-GATE: send and row-local same-ID retry use the exact original raw `message.text/clientMessageId`; rendered DOM text or normalized Markdown never becomes the business payload, and the next draft remains untouched.
- [x] 11.4 PARITY-GATE: headings/lists/blockquote/inline and fenced code/tables preserve prototype-derived typography; at desktop 1440 and mobile 390, pre/code/table stays inside the message viewport and document horizontal overflow is zero.

## Phase 12: Required runtime message and session guards

- [x] 12.1 RED/GREEN: Practice 消费 required RuntimeConfig fields/shared UTF-8 helper；small injected values 覆盖多字节 overflow、draft preservation 与 zero request，DOM/styles 不变。
- [x] 12.2 FALLBACK-GATE: required 子字段无 per-field fallback；仅整体 runtime source 不可用时保留既有 bootstrap fallback。
  <!-- verified: 2026-07-14 method=ui-contract-and-playwright-red-green evidence="Unknown prototype markdown-gfm state first failed both projects; source-owned semantic demo/CSS then passed UI contract 64/64 and Playwright 2/2 at 1440x900 and 390x844 with matching DOM/style/bbox/pixels, local pre overflow, mobile table overflow, bounded surfaces and zero document horizontal overflow." -->
- [x] 11.5 RED-GREEN: terminal recovery navigates exactly to `{ name: "workspace", params: { targetJobId } }` / `/workspace?targetJobId=...` read-only detail; query-free workspace, `planId` and current-scope `parse(targetJobId)` recovery are negative assertions.

## BDD Gate

- [x] BDD-Gate: `BDD.PRACTICE.TEXT.001` 由 [BDD checklist](./bdd-checklist.md) 关联 send/retry/completion/recovery owner behavior tests；不创建或声明真实 E2E PASS。

## Phase 13: Reference-aligned active interview surface

- [x] 13.1 RED: Practice visual tests 固化全局 TopBar 下可用视口高度、约 1708px desktop 内容面、Session Header 分组、消息 surface、大 Composer 与 390px containment。<!-- verified: 2026-07-19 method=vitest-red evidence="PracticeVisual expected failures captured the old unconstrained layout before implementation" -->
- [x] 13.2 GREEN: 重构 PracticeScreen / TopBar / Transcript / InputBar / FinishCta 视觉 DOM/CSS 与 SVG icon，保持 session/message/completion/disabled 语义不变。<!-- verified: 2026-07-19 method=focused-vitest evidence="practice focused suite 54 tests PASS after square role markers and icon prompt refinement" -->
- [x] 13.3 BDD-Gate: `BDD.PRACTICE.TEXT.001` 在新视觉层级下继续覆盖发送、pending、retry、completion 与终态 fail-closed；正式 frontend 的 repository running fixture 完成 Chrome 1916×821 / 390×844 containment 验收，未声明真实 active-session E2E PASS。<!-- verified: 2026-07-19 method=chrome-formal-frontend-fixture evidence="desktop frame x=105 width=1706, avatar 50x50 radius=7, prompt 320x42 radius=7 with icon; mobile x=16 width=358; documentOverflow=0" -->
- [x] 13.4 REGRESSION-GATE: focused frontend、根 `make test`、typecheck/build、context/docs/index/diff 通过后恢复 completed。<!-- verified: 2026-07-19 method=focused+root-regression evidence="practice Phase 13 focused 54 PASS; final root Python 615/4615 subtests, Go all packages, frontend 131 files/1054 tests PASS; typecheck/build PASS" -->

## Phase 14: Fixed Composer and helper anchoring

- [x] 14.1 RED: source/component tests 证明 helper 仍由 `Transcript` 渲染并会随滚动内容移动，同时要求它成为 `InputBar` / Composer 子元素。<!-- verified: 2026-07-19 method=vitest-red evidence="PracticeVisual 2 expected failures plus InputBar helper DOM 1 expected failure" -->
- [x] 14.2 GREEN: 明确 `Transcript` 是唯一滚动区、Composer 以 `flex: 0 0 auto` 固定在会话卡底部，并把 helper capsule 与 sparkle icon 移入 `InputBar`；保持文案、样式、a11y、消息与发送语义不变。<!-- verified: 2026-07-19 method=focused-vitest evidence="PracticeVisual, InputBar, Transcript and PracticeScreen: 4 files / 56 tests PASS" -->
- [x] 14.3 CHROME: desktop/mobile 的短聊天与可滚动长聊天中，Composer/input 坐标和 helper/input gap 在 Transcript scroll 前后保持不变且 document 无横向溢出；fixture 证据不声明真实 active-session E2E PASS。<!-- verified: 2026-07-19 method=chrome-formal-frontend-fixture evidence="standard desktop helper 320x42 and mobile 292x42, gap=8 and overflow=0; compact desktop scrollTop 4→151 and mobile 0→175.5 while helperY/shellY/gap remained identical" -->
- [x] 14.4 REGRESSION-GATE: focused Practice、根 `make test`、typecheck/build、context/docs/index/diff 通过后恢复 completed。<!-- verified: 2026-07-19 method=focused+root+docs evidence="focused 4 files/56 tests; root Python 615/4615 subtests, Go all packages, frontend 131 files/1054 tests; typecheck/build/context/docs/index/diff PASS; all lint owners except pre-existing change-intake generic-api-key false positive PASS" -->
