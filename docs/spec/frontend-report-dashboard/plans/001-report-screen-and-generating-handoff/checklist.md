# 001 — Honest Grounded Report Screen Checklist

> **版本**: 3.2
> **状态**: completed
> **更新日期**: 2026-07-14

**关联计划**: [plan](./plan.md)

## Phase 1-5: Conversation-level baseline（历史已完成）

- [x] 1-5 Prototype/formal conversation report, data states, replay/next and baseline parity completed.

## Phase 6: UI truth source and honest generating

- [x] 6.1 RED-GREEN: docs/ui-design and ui-design source converge on three metrics + four always-visible sections, readiness summary, no tabs; source contract tests reject four-tab/four-card drift. (ui-design contract tests + report preflight)
  <!-- verified: 2026-07-12 method=source-contract-and-preflight evidence="ui-design contract 50/50 PASS; frontend preflight included in full 111-file/762-test Vitest PASS" -->
- [x] 6.2 RED-GREEN: after ownership transfer, prototype/formal Generating removes timer-driven percentage/completion, fixed observations, fake notify and records promise; queued/generating auto-poll, timeout/network continue-check, and terminal failed/not-found/invalid/REPORT_CONTEXT_TOO_LARGE back-only actions are truthful; oversize copy points to a shorter-input new session. (focused GeneratingScreen Vitest + source tests)
  <!-- verified: 2026-07-12 method=vitest-and-playwright evidence="full frontend 111 files/762 tests PASS; generating desktop/mobile parity and queued/context-too-large action matrix PASS" -->
- [x] 6.3 BDD-Gate: E2E.P0.056 honest queued/generating markers pass with no fake-live/records/retry-terminal markers.
  <!-- verified: 2026-07-13 evidence="current P0.056 schema-valid backend artifact retained; focused generating/report owner tests and active fake-live/stale-contract negatives PASS" -->
- [x] 6.4 RED-GREEN + TIMING-GATE: poll uses exact `maxAttempts=49`、initial1.5s、multiplier1.5、cap8s（约6m04s）；covers one action's4×60s+10s/20s/40s=5m10s with ~54s margin。Queued/generating remains honest during action-local waits；no attempt/retry/progress surface；window exhaustion is continue-check, and only API terminal failed is back-only。Runner attempts/outbox schedules cannot be treated as product timing.
  <!-- verified: 2026-07-13 method=vitest+scenario evidence="seven frontend files/51 tests PASS; P0.058 v3 proves timing/reset/async separation without UI attempt leakage" -->
- [x] 6.5 RED-GREEN + RESUME-FENCE: hide/blur during scheduled wait or in-flight request preserves current attempt and next delay；visible/focus resumes at `n+1`, never attempt1/repeat-n。Repeated pause/resume is idempotent，no concurrent request，single run total calls<=49；only explicit continue-check or reportId/client change resets. <!-- verified: 2026-07-13 method=fake-clock-vitest evidence="scheduled wait, unresolved in-flight abort, same-tick timer fence, repeated pause/resume, max49 started calls; focused17/17 and full789 PASS" -->

## Phase 7: Direct semantic dashboard and server-owned handoff

- [x] 7.1 DEPENDENCY + RED-GREEN: after backend-review 6.1, frontend consumes generated summary/frozen context/code+label/status/confidence/dimensionCode/retryFocusDimensionCodes; empty focus fixture remains valid for generic same-round Replay, while every non-empty code must reference a needs-work dimension and same-code issue or fail closed. Missing/unknown fields also fail closed. (generated-client, ConversationReport/useFeedbackReport/fixture table tests)
  <!-- verified: 2026-07-12 method=generated-contract-fixture-and-component-tests evidence="validate-fixtures 37/37 PASS; full frontend 111 files/762 tests PASS; direct ready/empty-focus/invalid-contract fixture assertions PASS" -->
- [x] 7.2 RED-GREEN: UI chrome/status/confidence/readiness/CTA are localized by UI locale while model summary/dimension/evidence/action labels remain byte-semantically in report language; first action switches only existing CTA variants and accessible disabled reason is exposed. (mixed-locale i18n exact-set + ReportHeader/ConversationReport tests)
  <!-- verified: 2026-07-12 method=exact-i18n-and-identical-fixture-parity evidence="zh needs-practice and en well-prepared desktop direct-report parity PASS; model prose injected unchanged; full i18n/component Vitest PASS" -->
- [x] 7.3 RED-GREEN: ReportScreen derives status/error and ContextStrip/CTA identity only from API frozen context; reportId-only deep link works and conflicting route status/target/resume/round is ignored. Long context/summary/labels/evidence/actions wrap; desktop grid and 390px single-column pass. (route-tamper component + report Playwright tests)
  <!-- verified: 2026-07-12 method=route-tamper-and-mobile-readability evidence="report desktop/mobile Playwright 6/6 PASS; shared TopBar absolute geometry is equal at zh 91.25px and en 128.25px, both sides scrollWidth=390; long EN capability labels use readable 309x20px rows instead of per-character vertical wrapping; focused ConversationReport 5/5 PASS" -->
- [x] 7.3a RED-GREEN: frontend mirrors English24 ECMAScript `/\s/u` whitespace words / zh-CN64 Unicode code points and fail-closes over-limit ready payloads to typed invalid without raw label/truncation/ellipsis/rewrite；paired backend/frontend tests lock U+FEFF as delimiter and U+0085 as non-delimiter。Schema200 remains outer fuse only；18/52 is not a frontend boundary。Previous 14/40 evidence is historical and does not close this item.
  <!-- verified: 2026-07-13 method=ui-contract+focused-playwright evidence="ui-design contract 54/54 PASS; focused Playwright 34/34 PASS; exact en24/zh64 fixtures remain complete at desktop/mobile and over-limit payloads fail closed without raw output" -->
- [x] 7.4 DEPENDENCY + RED-GREEN: after backend-practice/004 Phase 3, replay/next route/request removes settings/identity/focus/evidence gaps and sends only goal+sourceReportId; empty projected focus creates a generic same-round plan, non-empty focus remains server-derived, context.hasNextRound disables Next, and the fresh session starts once. (handoff/buildCreatePlanRequest/ReplayCta tests + backend integration)
  <!-- verified: 2026-07-13 evidence="P0.057 six files/49 tests PASS; P0.070/P0.072 PostgreSQL projection/isolation/idempotency/privacy markers PASS" -->
- [x] 7.5 BDD-Gate: P0.056/P0.058 v3 direct/typed UI markers prove honest polling plus action-local timing/async-attempt separation; P0.057 UI paths and registry-owned P0.070/P0.072 server focus markers remain regression inputs.
  <!-- verified: 2026-07-13 method=scenario-composition evidence="P0.058 four-stage v3 PASS and seven frontend files/51 tests PASS; P0.056/057/070/072 markers remain composed regression inputs" -->

## Phase 8: Strong parity and real acceptance

- [x] 8.1 RED-GREEN: report/generating Playwright uses identical fixtures, fixed locale/timezone/Date/DPR, font wait and disabled animation; compares DOM/computed style/key bbox at 1440/390 plus pixelmatch changed ratio ≤0.5%, retaining prototype/formal/diff failures. Non-empty buffer alone cannot pass. (report.spec.ts/generating.spec.ts)
  <!-- verified: 2026-07-12 method=source-geometry-full-page-pixel-parity evidence="UI contract source RED then 52/52 GREEN; report/generating desktop+mobile Playwright 10/10 PASS with root and fullPage pixelmatch; six deterministic full-page pairs have identical dimensions, max changedRatio 0.000726 (0.0726%), generating 0, and mobile formal/prototype scrollWidth=390" -->
- [x] 8.2 RED-GREEN: active fixture/scenario/docs/runtime negative gate removes stale question fields, fake-live copy, raw enum UI and client focus authority while preserving historical evidence records. (frontend report out-of-scope lint + repo scoped search)
- [x] 8.3 BDD-Gate: E2E.P0.059 source/geometry/screenshot parity passes.
  <!-- verified: 2026-07-13 scenario="E2E.P0.059" evidence="20 focused Vitest tests, active out-of-scope lint, 3 Python tests, production build and 12 desktop/mobile Playwright parity tests PASS" -->
- [x] 8.4 BDD-Gate: E2E.P0.099 captures exactly six redacted `fullPage: true` images；each ready row binds current-run DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest。Both 390x844 real report images cover the action region and show actual `<=24-whitespace-word` / `<=64-Unicode-code-point` labels fully with no clipping/ellipsis/hiding/overflow；P0.100 output digest is not a prerequisite。
  <!-- verified: 2026-07-13 run=e2e-p0-099-20260713T095144Z-12381 evidence="trigger+verify PASS; exact six fullPage images across three states; current-run DB/API canonical digests and manual content audit PASS; raw-debug absent; privacy redacted" -->
- [x] 8.5 UX-BOUNDARY-GATE: exact24/64 fixtures pass1440x1200+390x844 parity with complete wrapping；200-code-point malformed fixture only proves typed invalid/no raw，18/52 only proves repair margin，neither can replace UX evidence。
  <!-- verified: 2026-07-13 method=ui-contract+focused-playwright evidence="54/54 + 34/34 PASS; exact en24/zh64 labels wrap completely at 1440x1200 and 390x844; malformed over-limit fixture is typed invalid with no raw output" -->
- [x] 8.6 Run full frontend, UI contract, typecheck/build, privacy/negative, docs/index and diff gates; record current evidence only.
  <!-- verified: 2026-07-13 evidence="frontend 112 files/795 tests, production build and UI source contracts PASS; P0.099 current exact-six desktop/390px screenshots PASS; privacy/negative, docs/index/context and diff gates PASS" -->

## Phase 9: User-visible internal locator removal

- [x] 9.1 RED-GREEN: prototype source/contract removes session/report UUID from Context Strip and preserves target/round/resume layout at desktop/mobile.
- [x] 9.2 RED-GREEN: formal `ReportContextStrip` removes distinct report/session sentinel UUIDs from text/title/tooltip/aria/accessibility DOM while OpenAPI/report contract UUID validation and CTA `sourceReportId` authority stay green.
- [x] 9.3 REGRESSION-GATE: delete orphan `report.context.session` zh/en keys；update P0.059 README/INDEX C-12 metadata；focused Report/i18n tests, UI source contracts, active negative, full frontend and typecheck/build pass.
  <!-- verified: 2026-07-14 method=focused+full evidence="Report/i18n and Context Strip negatives pass; UI 62/62, frontend 121 files/977 tests, typecheck/build and refreshed P0.059 metadata are green." -->
- [x] 9.4 BDD-Gate: `E2E.P0.059` passes refreshed 1440/390 DOM/style/bbox/viewport/pixel parity under normal PASS cleanup；then `/agent-browser` captures the same real ready report from formal frontend as exact 1440x1200 / 390x844 `fullPage: true` files under `.test-output/acceptance/report-context-strip/<run-id>/`. The directory contains only the two fixed PNG names plus `manifest.json`, whose per-image path/SHA-256/state/viewport/fullPage/report+session sentinel absence and linked DOM/a11y audit all validate.
  <!-- verified: 2026-07-14 scenario=E2E.P0.059 evidence="P0.059 passes 16/16 Playwright with max changedRatio 0.000060; acceptance directory 20260714-final contains exactly two PNGs plus manifest and recomputed SHA-256/state/viewport/fullPage/sentinel audits pass." -->

## Phase 10: Independent current-plan reports list and Back recovery

- [x] 10.1 RED-GREEN: prototype/formal 新增保留 App chrome 但不进入 TopBar 的 ReportsScreen；`/reports?targetJobId=...` 组合 `getTargetJob + listTargetJobReports`，覆盖 loading/empty/error/ready 与 1440/390 source parity。
  <!-- verified: 2026-07-14 method=ui-contract+focused+pixel-parity evidence="UI 62/62, ReportsScreen 15/15 and P0.059 16/16 pass for ready/loading/empty/error and exact current-target rendering without a TopBar entry." -->
- [x] 10.2 RED-GREEN: validator 精确闭合 response target、2~5 canonical round IDs/sequences/order/display、report locator 单轮唯一归属、same-ID-only-ready 与 current/latest cross-field nullability；A/B 规划隔离、跨用户/mismatch/extra/missing/unknown round、target switch 首帧清旧 rows 与 stale response fence 均 fail closed。
  <!-- verified: 2026-07-14 method=focused-vitest evidence="TargetJob reports validator 40/40 + ReportsScreen 15/15 pass; negative cases cover cross-round locator reuse, same-ID non-ready, latest-ready without current, current without latest, canonical identity, A/B isolation, switch clearing and stale fences." -->
- [x] 10.3 RED-GREEN: 每轮只展示 current report 与 latest attempt；ready/current links、queued/generating link、failed typed no-Retry、same-ID de-dup、different-ID latest-ready status 均正确，且没有完整历史列表或其他规划 sentinel。
  <!-- verified: 2026-07-14 method=focused-vitest evidence="ReportsScreen 15-test matrix proves current/latest-only rendering, queued/generating links, typed failed no-Retry, same-ID de-dup, different ready status without a history action, and no foreign-plan sentinel." -->
- [x] 10.4 RED-GREEN + BDD-Gate: Reports Back 精确回当前 `parse?targetJobId=...`；Report/Generating ready/pending/failed/recoverable/normal-generating 有 trusted target 时回 `/reports?targetJobId=...`，无可信 identity 回 workspace。P0.058/P0.059 覆盖该矩阵；report/generating route 保持 reportId-only。
  <!-- verified: 2026-07-14 method=focused+P0.058+P0.059 evidence="Trusted target returns to Reports, missing/invalid/404/network-first-load falls back by replace to workspace without loops; report/generating URLs remain reportId-only." -->
- [x] 10.5 POST-PASS: focused Reports/Report/Generating/route/i18n tests、UI source contract、typecheck/build、P0.058/P0.059、owner contexts、docs/diff，以及“Parse/Report/Generating 零 list consumer、ReportsScreen 唯一 consumer、TopBar 无入口、无 `section=reports`”负向 gate 全部通过后完成 Phase 10。
  <!-- reverified: 2026-07-14 method=current-aggregate evidence="Root make test passed UI 62/62, Python 590 tests/5181 subtests, all Go packages and frontend 121 files/977 tests. Current P0.059 passed 10 focused Vitest files/137 tests, production build and 16 desktop/mobile Playwright parity tests; verify/cleanup, context, docs/diff and all scoped consumer/route/TopBar negatives pass." -->
