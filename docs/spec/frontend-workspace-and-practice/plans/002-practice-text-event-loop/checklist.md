# 002 — Practice Continuous Text Conversation Checklist

> **版本**: 2.6
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1: UI truth source
- [x] 1.1 RED-GREEN: update prototype tests/source to TopBar + full-width Conversation and delete question/hint/phone-positive source.
- [x] 1.2 RED-GREEN: update desktop/mobile source geometry expectations and stale-contract negative checks.

## Phase 2: Formal screen structure
- [x] 2.1 RED-GREEN: delete SessionMap/QuestionCard/PhoneSurface/hint/controller components and simplify PracticeScreen/TopBar.
- [x] 2.2 RED-GREEN: disabled phone icon has native disabled/a11y/unavailable copy and no route/API action.
- [x] 2.3 RED-GREEN: remove mode/modality/practiceMode/hint context/handoff/i18n/test contracts.
- [x] 2.4 BDD-Gate: P0.045 simplified UI and phone-disabled scenario passes.

## Phase 3: Message hooks and states
- [x] 3.1 RED-GREEN: loader renders ordered session.messages including refresh recovery.
- [x] 3.2 RED-GREEN: send hook handles success/replay/failure/same-ID retry without duplicate messages.
- [x] 3.3 RED-GREEN: loading/sending/error/local-paused/completing/session-lost states remain usable; pause has no backend event call and refresh resumes Running.
- [x] 3.4 BDD-Gate: P0.044 and P0.046 pass.

## Phase 4: Completion/generating
- [x] 4.1 RED-GREEN: finish handoff contains stable IDs only; generating copy is conversation-level.
- [x] 4.2 BDD-Gate: P0.047 passes.

> Ownership note (2026-07-12): the completed evidence above is historical. Current work stops at stable `reportId` handoff; GeneratingScreen is exclusively owned by `frontend-report-dashboard/001`.

## Phase 5: Parity and real scenario
- [x] 5.1 Run focused/full frontend, typecheck/build, UI contract and pixel parity desktop/mobile.
  <!-- verified: 2026-07-12 method=full-frontend-and-parity evidence="111 files/708 Vitest tests, typecheck, build, 45 UI contracts and 8 desktop/mobile practice Playwright cases pass" -->
- [x] 5.2 Run the then-current real backend/frontend path and capture redacted conversation/report screenshots.
- [x] 5.3 BDD-Gate: the then-current real fullstack screenshot evidence passes; current Practice screenshot ownership is Phase 9 P0.044/P0.046.

## Phase 6: Review remediation
- [x] 6.1 RED-GREEN: PracticeScreen retries loader, message and completion failures through the correct operation and preserves message/completion idempotency. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx`)
- [x] 6.2 RED-GREEN: Finish CTA is disabled during send, load, completion and non-mutable session states. (`pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx`; frontend typecheck)
- [x] 6.3 BDD-Gate: P0.046 and P0.047 screen-level failure/recovery and completion scenarios pass. (serial `setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`, both PASS)

## Phase 7: Zero-answer finish eligibility and backend authority

- [x] 7.1 RED-GREEN: PracticeScreen derives Finish eligibility only from server-loaded committed candidate `user` messages plus existing mutable/no-pending-reply/no-load/no-send/no-complete guards; opening assistant, composer draft and route params do not count. (`PracticeScreen.test.tsx` + completion hook tests)
  <!-- verified: 2026-07-12 method=focused-and-full-vitest evidence="PracticeScreen 8/8 PASS; related practice regression 24/24 PASS; full frontend 111 files/762 tests PASS" -->
- [x] 7.2 RED-GREEN: prototype and formal Finish are native disabled at zero answers and expose the same nearby zh/en reason through stable `aria-describedby`; first committed user message removes only the zero-answer reason. (ui-design source contract + i18n exact-set + DOM/a11y tests)
  <!-- verified: 2026-07-12 method=source-contract-dom-a11y-i18n evidence="ui-design contract 50/50 PASS; PracticeScreen and locale exact-set tests included in full Vitest PASS" -->
- [x] 7.3 RED-GREEN: direct zero-answer completion is still rejected by backend `VALIDATION_FAILED`, session remains mutable and no report/job/outbox/idempotency success is written; one-answer completion and replay remain green. (consume backend-practice/002 Phase 9 service/store/API/PostgreSQL markers; do not duplicate backend logic in frontend)
  <!-- verified: 2026-07-13 method=isolated-postgresql-service-store-api evidence="P0.047 service/store/API plus real PostgreSQL integration PASS; zero_answer_side_effect_count=0, pending_reply_side_effect_count=0, report-context snapshot/replay markers PASS; temporary migrated database removed." -->
- [x] 7.4 BDD-Gate: E2E.P0.047 composes `ZERO_ANSWER_FINISH_DISABLED_PASS` + `ZERO_ANSWER_COMPLETION_REJECTED_PASS`, then proves one-answer completion, stable reportId handoff and exact replay.
  <!-- verified: 2026-07-13 method=scenario-run evidence="Serial setup/trigger/verify/cleanup PASS; frontend marker, backend rejection marker, stable completion Idempotency-Key and report-context replay marker all verified." -->

## Phase 8: reportId-only completion handoff

- [x] 8.1 RED-GREEN: PracticeScreen completion navigation has exact query/state/context shape `{reportId}`; tests first fail on and then reject `targetJobId|planId|sessionId|resumeId|roundId|roundName|status|error` copies while preserving same-reportId completion replay.
  <!-- verified: 2026-07-12 method=screen-router-and-privacy-tests evidence="PracticeScreen, App, routeUrl, pendingAction and routing privacy cases included in 111-file/762-test PASS" -->
- [x] 8.2 REGRESSION-GATE: active PracticeScreen/context/router code contains no positive write of those copied fields to generating/report navigation; frontend-report consumes `getFeedbackReport(reportId)` as the sole downstream authority.
  <!-- verified: 2026-07-12 method=active-negative-and-route-tamper evidence="report/generating out-of-scope tests PASS; Playwright canonicalizes hostile report/generating URLs to reportId only" -->
- [x] 8.3 BDD-Gate: E2E.P0.047 one-answer completion asserts URL/history state contains only reportId, downstream request is keyed only by reportId, and idempotent replay returns the same locator.
  <!-- verified: 2026-07-13 method=screen-route-scenario evidence="PracticeScreen first navigates to /generating?reportId=... with null history state; downstream getFeedbackReport is keyed only by reportId; forbidden identity/status keys are absent; completion replay keeps the stable owner locator." -->

## Phase 9: Immediate user message, thinking state and row-local retry

- [x] 9.1 RED-GREEN: prototype tests/source append one user row and clear composer synchronously, render accessible interviewer-thinking only while pending/retrying, and render retry only beneath a failed user row.
- [x] 9.2 CONTRACT-DEPENDENCY-GATE: OpenAPI-generated user `PracticeMessage` exposes `clientMessageId + replyStatus=pending|retryable_failed|terminal_failed|complete` and typed `ApiClientError.apiError.retryable` with HTTP/envelope/transport metadata；backend durably projects reply status；the operation-matrix recovery fixtures are current and validated.
  <!-- verified: 2026-07-13 method=openapi+backend+fixtures evidence="Generated role union and typed error compile; backend persists/projects all four user states; get/send fixture matrix now validates pending/retryable/terminal/complete plus validation/auth/not-found/conflict/mismatch/timeout cases." -->
- [x] 9.3 RED-GREEN: formal Practice keeps transient `{text, clientMessageId, status}` only until first response/read convergence；reload/remount rehydrates pending/retryable/terminal/complete solely from `getPracticeSession`, with no URL/browser-storage retry persistence or `Error.message` parsing.
  <!-- verified: 2026-07-13 method=vitest+negative-search evidence="Optimistic state is in-memory only and is removed when server truth arrives; reload pending uses bounded read-only polling, retryable/terminal/complete rehydrate from getPracticeSession, and no URL/storage/error-string recovery path exists." -->
- [x] 9.4 RED-GREEN: typed retryable failure invokes the shared send path with server original text + same `clientMessageId`, preserves row/draft and restores one icon after repeated failure；AI failure → reload → same-ID retry converges to one user/reply pair；pending re-read never duplicate-sends；terminal failures have no retry.
  <!-- verified: 2026-07-13 method=vitest+real-postgres evidence="Same send path reuses original server text/client ID, preserves next draft and one row, repeated retryable failure restores one icon, pending never resends, terminal has no retry, and isolated PostgreSQL converges to one user/assistant pair." -->
- [x] 9.5 REGRESSION-GATE: pending/retryable-failed/retrying/terminal-recovery all keep Finish disabled；focused generated-client/Practice hooks/screen/i18n/a11y tests, UI source contracts, full frontend, typecheck/build and active negative searches pass.
  <!-- verified: 2026-07-13 method=frontend-regression evidence="Practice 35/35 and full frontend 114 files/810 tests PASS; typecheck/build and i18n/a11y gates PASS; unresolved states lock Finish and send failures expose only localized user copy." -->
- [x] 9.6 BDD-Gate: `E2E.P0.044` pending/reload/success and `E2E.P0.046` AI-failure/reload/same-ID retry/terminal recovery pass with prototype/formal DOM/style/bbox/viewport parity and exact 1440/390 screenshots.
  <!-- verified: 2026-07-13 method=scenario-run+playwright evidence="P0.044 ran 6 desktop/mobile tests and saved immediate/persisted pending screenshots; P0.046 ran 4 desktop/mobile tests and saved retryable/terminal screenshots. Verifiers confirmed CSS 1440x900 and 390x844 with exact PNG 1440x900 and 1170x2532." -->

## Phase 10: Lease-aligned timeout reconciliation and terminal plan recovery

- [x] 10.1 RED-GREEN: UI source tests first fail, then prototype preserves injected pending/retryable/terminal statuses and renders one localized generic terminal recovery CTA with no row retry.
  <!-- verified: 2026-07-14 method=ui-source-tdd evidence="The focused source contract first failed because transcript initialization overwrote injected user status with complete; GREEN preserves message.status, keeps retry only for retryable_failed, and renders one zh/en terminal alert whose sole recovery CTA returns to the current Parse interview plan. Focused contract PASS and full ui-design contract 59/59 PASS." -->
- [x] 10.2 RED: hook/screen fake-timer tests require AbortSignal forwarding, independently cancellable/bounded reads, exact 95,000 ms POST timeout, same-ID reconciliation and stale-response suppression before implementation.
  <!-- verified: 2026-07-14 method=focused-vitest-red evidence="Four focused files ran 38 tests: 22 historical behaviors remained green and 16 new assertions failed before implementation, precisely exposing missing signal forwarding, cleanup/read abort, bounded read timeout, pending fact preservation, stale-read fence, exact 95-second POST abort/reconcile, terminal CTA/i18n and fail-locked loader error behavior." -->
- [x] 10.3 GREEN: `usePracticeMessages` forwards request signal through generated request options；session loader/reconcile abort on cleanup/timeout without hand-editing generated output.
  <!-- verified: 2026-07-14 method=focused-vitest-green evidence="Signal forwarding, cleanup abort and independently bounded 10-second reads passed in the focused hook/loader suite; generated client output was not hand-edited for this behavior." -->
- [x] 10.4 GREEN: PracticeScreen aborts POST at exactly 95 seconds and adopts reconciled complete/pending/retryable/terminal server truth；missing-ID/read-failure preserves the original unresolved row/ID and keeps new-ID submit/Finish locked；late old responses cannot overwrite newer truth.
  <!-- verified: 2026-07-14 method=focused-vitest-green evidence="Fake-timer tests prove no abort at 94,999 ms, abort plus independent same-ID read at 95,000 ms, adoption of all four server statuses, uncertain-read fail-lock, and both late-POST and older-reconcile suppression." -->
  <!-- reopened: 2026-07-14 reason="Independent review found the untested inverse completion order where an older loader snapshot can invalidate and strand the newer same-ID reconciliation." -->
  <!-- reverified: 2026-07-14 method=adversarial-review+vitest evidence="Read-start tokens enforce latest-start-wins in both completion orders; stale reconciliation preserves only the same submission failure, newer missing-ID becomes retryable fail-lock, and authoritative truth clears the synchronous transient ref so old results cannot resurrect it." -->
- [x] 10.5 RED-GREEN: reloaded pending only re-reads until backend 90-second lease convergence；loader refresh failure preserves prior unresolved facts or fails locked and can never unlock a new ID.
  <!-- verified: 2026-07-14 method=focused-vitest-green evidence="Reloaded pending polls GET without resend; refresh/read failure retains same-session unresolved data, while missing data leaves composer and Finish fail-locked. Backend lease convergence remains owned by backend-practice Phase 11." -->
  <!-- reopened: 2026-07-14 reason="Independent review found missing same-mounted-component session isolation and stale local state reset coverage." -->
  <!-- reverified: 2026-07-14 method=adversarial-review+vitest evidence="A keyed session owner isolates optimistic/error/draft/pause/timer/poll and async send/complete continuations; completion refs are synchronously scoped by session, while pending refresh remains locked through the backend 90-second convergence contract." -->
  <!-- reopened: 2026-07-14 reason="A persisted retryable_failed row followed by an online/focus refresh failure retained row-local retry while also rendering global loader retry and disabling the draft composer; typed HTTP retryable=true lacked direct UI recovery proof." -->
  <!-- reverified: 2026-07-14 method=red-green+focused-vitest evidence="Persisted retryable_failed plus failed online refresh retains exactly one row-local retry and an editable next draft; typed HTTP retryable=true reuses the exact clientMessageId/text without duplicate or technical leakage. Practice/hooks focused suite passed 62/62 and frontend typecheck passed." -->
- [x] 10.6 RED-GREEN: terminal state shows safe zh/en copy and exactly `navigate({ name: "parse", params: { targetJobId } })` via “返回当前面试规划”；no retry icon, workspace fallback, `planId`, composer submit or technical error text.
  <!-- verified: 2026-07-14 method=focused-vitest-green evidence="Trusted terminal truth renders one localized alert and exact parse(targetJobId) CTA with secondary/sm source and DOM interaction locks; retry, workspace, planId, duplicate ErrorState and technical copy are negative assertions." -->
  <!-- reopened: 2026-07-14 reason="Independent review found that later authoritative terminal/complete truth can leave an earlier local ErrorState visible, producing duplicate failure UI." -->
  <!-- reverified: 2026-07-14 method=adversarial-review+vitest evidence="Authoritative terminal/complete clears local message error; terminal truth suppresses ErrorState and loader retry even after a later refresh failure, leaving exactly the current Parse plan CTA. Final independent narrow review found no remaining issue." -->
- [x] 10.7 PARITY-GATE: prototype/formal immediate pending, persisted pending, retryable failure and terminal CTA match DOM/a11y/computed-style/bbox/viewport/screenshot at desktop 1440 and mobile 390.
  <!-- verified: 2026-07-14 method=playwright-pixel-parity+ui-contract evidence="practice.spec.ts passed 16/16 across desktop 1440 and mobile 390 with DOM, a11y, computed-style, bbox, viewport and screenshot comparison; ui-design contract passed 60/60." -->
- [x] 10.8 BDD-Gate: P0.044/P0.046 use the shared tracked source manifest, trigger/current SHA-256 equality and per-PNG SHA-256/dimensions/viewport；all exact lease/fence/concurrency/timeout/terminal/fingerprint markers pass and historical artifacts fail closed.
  <!-- reverified: 2026-07-14 method=serial-scenario-run evidence="Current-source P0.044 run cd8c378d-6fcc-4045-a00f-c9129873e511 and P0.046 run dfc68a8f-41de-46e5-9b0c-1aec3fbb67fb passed setup/trigger/verify/cleanup against source SHA 3e644ae013ee2159937e4853c8f3f32a3f2bd1f1351fd6fe74d9b66aa2ea11d2; shared-verifier ownership, fingerprints, screenshots, exact markers and isolated PostgreSQL residual=0 all passed." -->
- [x] 10.9 Run focused/full Vitest, UI contract, typecheck/build, Playwright parity, scenario contract, serial P0.044/P0.046, context/docs/diff gates；only then record current evidence and restore completed lifecycle.
  <!-- reverified: 2026-07-14 method=current-full-aggregate evidence="Root make test passed UI contract 62/62, Python 590 tests/5181 subtests, all Go packages and frontend 121 files/977 tests; typecheck/build, P0.006 160 Playwright tests, current P0.044/P0.046, context/docs/diff gates and shared-verifier negative contract all pass." -->
