# Honest Grounded Report Screen BDD Checklist

> **版本**: 3.4
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.056 Honest generating and ready report

- [x] Setup prepares isolated queued/ready direct-shape fixtures and redacted IDs.
- [x] Trigger composes required backend `REPORT_CONTEXT_SNAPSHOT_PASS` / `REPORT_DIRECT_READY_PASS` markers with focused generating/report Vitest markers.
- [x] Verify proves honest state/actions, frozen context, direct summary/dimensions/evidence/actions and route-tamper negatives; cleanup removes artifacts/data.
- [x] Verify exact maxAttempts49 timeline：one report action's local10s/20s/40s waits remain queued/generating，no attempt/progress is visible，ready transitions normally，and ~6m04s client exhaustion remains continue-check rather than failed.
  <!-- verified: 2026-07-13 evidence="focused frontend timing suite PASS; P0.058 v3 keeps product waits and async attempts separate" -->
- [x] Verify hidden/blur during timer and in-flight request preserves attempt/delay；visible/focus resumes n+1 with no reset1/repeat/concurrent call and one run<=49.
  <!-- verified: 2026-07-13 evidence="current P0.056 artifact plus focused polling17/17, generating23/23 and full frontend789 prove honest state and monotonic max49 resume" -->

## E2E.P0.057 Replay and next UI paths

- [x] Prepare retry-first with empty generic focus, retry-first with valid non-empty issue-backed focus, next-first/review-first and unavailable-round data.
- [x] Refresh setup / trigger / verify / cleanup UI assertions.
- [x] Execute and record action priority, no client focus payload, generic/focused server projection and fresh-session evidence.
  <!-- verified: 2026-07-13 scenario="E2E.P0.057" result="PASS; six frontend files/49 tests" -->

## E2E.P0.070 Server-owned focus

- [x] Prepare ready source-report focus data.
- [x] Refresh setup / trigger / verify / cleanup request-negative assertions.
- [x] Execute and record server-projected focus evidence.
  <!-- verified: 2026-07-13 scenario="E2E.P0.070" result="PASS; generic/non-empty/next focus and idempotency markers" -->

## E2E.P0.072 Replay isolation and failure

- [x] Prepare missing/cross-user/mismatch/non-ready/invalid-focus data.
- [x] Refresh setup / trigger / verify / cleanup failure assertions.
- [x] Execute and record fail-closed/privacy evidence.
  <!-- verified: 2026-07-13 scenario="E2E.P0.072" result="PASS; 12 isolation failures, zero-insert/privacy/legacy-negative markers" -->

## E2E.P0.058 Failure states

- [x] Setup prepares failed/missing/timeout/network/invalid contract, invalid non-empty focus cross-references and conflicting route params.
- [x] Trigger composes backend v3 action-retry/fail-closed markers with failure/poll/report Vitest markers.
  <!-- verified: 2026-07-13 evidence="P0.058 trigger PASS with six backend v3 markers and seven frontend files/51 tests" -->
- [x] Verify proves timeout/network continue-check, terminal back-only, API-only status/context and no raw enum/fake report; cleanup removes artifacts/data.
  <!-- verified: 2026-07-13 evidence="P0.058 verify+cleanup PASS with typed failure UI and redaction checks" -->
- [x] Verify action attempt4/nonretryable API failed is terminal，while maxAttempts49/network exhaustion never mutates server state；action-local retry, async job attempts and outbox infra schedules are not conflated in UI copy or timing.
  <!-- verified: 2026-07-13 evidence="v3 runtime records attempt4/reset/separation; frontend 51 tests reject fabricated terminal/progress copy" -->
- [x] Revision 2026-07-14 covers pending/normal/failed/recoverable generating with trusted target -> `/reports?targetJobId=...`, plus missing reportId/404/first-load network/invalid payload without trusted target -> `workspace` fallback.
- [x] Revision 2026-07-14 rejects route/title/reportId target inference and proves Report/Generating never consume `listTargetJobReports`；only ReportsScreen does.
  <!-- verified: 2026-07-14 scenario=E2E.P0.058 evidence="Final four-stage wrapper PASS; frontend 8 files/82 tests and three backend exact packages execute, including the named three-field UUID-free Context Strip marker and trusted/missing-context navigation matrix." -->

## E2E.P0.059 Strong parity

- [x] Setup locks identical fixtures, locale/timezone/Date/DPR, fonts and disabled motion for prototype/formal at 1440/390.
- [x] Trigger records DOM/style/bbox and prototype/formal/diff PNGs; changed-pixel ratio must be ≤0.5% using pixelmatch threshold 0.1.
- [x] Verify rejects non-empty-buffer-only evidence and requires visual + active stale-contract negatives; cleanup keeps failure artifacts only.
  <!-- verified: 2026-07-13 scenario="E2E.P0.059" evidence="source/DOM/style/bbox/pixel threshold gate, stale-contract negatives, production build and 12 Playwright cases PASS; cleanup PASS" -->
- [x] Revision 2026-07-13 removes session/report UUID from prototype/formal visible and accessible Context Strip, then executes clean 1440/390 DOM/style/bbox/viewport/pixel parity.
- [x] Revision 2026-07-13 updates P0.059 README/INDEX to C-12, verifies distinct UUID sentinels are UI/a11y-negative but contract/CTA-positive, preserves normal PASS cleanup, then captures the same formal real ready report at exact 1440x1200 / 390x844 with `fullPage: true` into `.test-output/acceptance/report-context-strip/<run-id>/`.
- [x] Acceptance directory contains only `report-context-strip-desktop-1440x1200.png`, `report-context-strip-mobile-390x844.png`, `manifest.json`; recomputed SHA-256 matches each manifest row, `state=ready`/viewport/fullPage are exact, target/round/resume are visible, and report/session sentinel absence is linked to passing DOM/a11y negative evidence.
- [x] Revision 2026-07-14 independent ReportsScreen covers populated/empty/loading/error × 1440/390，only current TargetJob canonical rounds/current/latest，A/B isolation、mismatch/stale fail-close、different-ready status and no full history.
- [x] Revision 2026-07-14 ready report Back uses the same response's trusted `targetJobId` and reaches `/reports?targetJobId=...`; current report URL remains reportId-only and existing report parity is unchanged.
  <!-- verified: 2026-07-14 scenario=E2E.P0.059 evidence="Full wrapper PASS with 137/137 Vitest and 16/16 Playwright; max changedRatio 0.000060. Formal real acceptance manifests prove current-plan-only rows and UUID-free Context Strip at both viewports." -->
- [x] Revision 2026-07-14 Phase 11 proves Reports Back reaches `/workspace?targetJobId=...` read-only detail directly at 1440/390, emits no Workspace query key except `targetJobId`, exposes no Parse detour in browser history, and triggers no Parse animation/import/poll while preserving the existing list isolation/parity assertions.
  <!-- verified: 2026-07-14 evidence="Fresh P0.059 desktop/mobile PASS proves direct Workspace detail, targetJobId-only URL, no Parse history/import/poll/animation and preserved report-list isolation/parity; P0.058 recovery matrix remains PASS." -->

## E2E.P0.099 Real full-stack screenshots

- [x] Setup prepares zh needs-practice and en well-prepared real reports with long summary/evidence plus a generating state; record redacted IDs.
- [x] Setup creates current-run en/zh ready rows and binds each row's DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest；P0.100 output digest is not required.
- [x] Trigger captures exactly six `fullPage: true` images at 1440x1200 / 390x844；both mobile real report images fully cover action regions and show actual `<=24-whitespace-word` English / `<=64-Unicode-code-point` zh-CN labels.
- [x] Deterministic ui-design/OpenAPI fixtures with exactly 24 English whitespace words / 64 zh-CN Unicode code points pass prototype/formal pixel parity with complete wrapping and no clipping/ellipsis/hiding/overflow.
- [x] Verify desktop+390 complete wrapping/no overflow and redaction；over24/64 is typed invalid/no raw；200-code-point fuse and18/52 repair-margin evidence are rejected as UX PASS.
  <!-- verified: 2026-07-13 run=e2e-p0-099-20260713T095144Z-12381 evidence="trigger+verify PASS; six fullPage images/three states; DB/API canonical digests, manual content audit and raw-debug-absent PASS" -->
