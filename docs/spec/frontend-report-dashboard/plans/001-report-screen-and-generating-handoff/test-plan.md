# Honest Grounded Report Screen Test Plan

> **版本**: 3.4
> **状态**: completed
> **更新日期**: 2026-07-14

## Phase 1-5: Historical baseline

- Existing conversation-level report/replay/state tests remain regression evidence; historical screenshot buffer PASS is not visual proof.

## Phase 6: UI truth and generating

- Prototype/docs/source tests assert three metrics, four always-visible sections, readiness summary, no tab model and exclusive report-owner GeneratingScreen.
- Generating fake-clock tests turn RED, then prove no percentage/autocomplete/fixed observations/fake notify/records promise; queued/generating auto-poll, timeout/network continue-check and terminal back-only behavior remain. `REPORT_CONTEXT_TOO_LARGE` has localized actionable copy and never exposes a retry/continue-check action for the same report.
- Fake-clock tests assert exact `maxAttempts=49`, delays1.5s×1.5 capped8s and total≈6m04s。The timeline covers one action's4×60s calls plus action-local10s/20s/40s waits=5m10s，leaving≈54s；49th exhaustion remains client timeout/continue-check and never fabricates server failed。Tests reject maxAttempts30、internal attempt/progress copy、runner attempts/max_attempts and outbox/infra30s/2m/10m as report product timing assumptions.
- Pause/resume fake-clock tests hide/blur during scheduled timer and unresolved in-flight request；after queued/error resolution while paused，resume continues n+1 with preserved delay。Repeated hidden/visible/blur/focus causes no duplicate/in-flight overlap and cannot exceed49 calls。Explicit retry and identity change are the only reset cases.
- Copy tests cover truthful zh/en state/action wording without promising notification or a nonexistent records path.

## Phase 7: Direct dashboard and handoff

- Generated-client/fixture/component table tests cover summary, immutable context, code+label dimensions, dimensionCode evidence and retryFocusDimensionCodes: `[]` stays valid and keeps generic Replay enabled; each non-empty code must resolve to a needs-work dimension plus same-code issue; invalid references, missing/unknown fields and malformed ready states fail closed.
- i18n tests cover UI chrome/DimensionStatus/Confidence/Readiness while preserving report-language model summary/dimension/evidence/action labels under mismatched UI locale; raw enum/code cannot appear in user text.
- CTA tests cover retry-first with empty/non-empty focus, next-first, review-first, unavailable round, pending start and accessible disabled reasons.
- Deep-link/route-tamper tests prove reportId-only load and API-only status/context/CTA identity; conflicting route params cannot override.
- Context/detail tests cover long target/round/resume/summary/dimension/evidence/action, full-value access, wrap and mobile single column.
- The 200-code-point schema boundary is malformed-output coverage only。Consumer tests count English ECMAScript `/\s/u` whitespace words at24/25 and zh-CN Unicode code points at64/65，including paired U+FEFF-delimiter / U+0085-non-delimiter parity with backend；over-limit fails closed to typed invalid with no raw label，and proves zero client truncation/rewrite。Exact24/64 fixtures assert unchanged text and complete desktop+390 wrapping；18/52 is upstream repair-margin coverage, not a frontend acceptance boundary.
- Route/request negative tests reject settings/identity/focus/evidence gaps on derived create, assert hasNextRound disabled reason, and prove goal+sourceReportId reaches a server-derived focused fresh session.

## Phase 8: Parity and real UAT

- Source-structure tests compare report/generating DOM, controls, labels, aria state and primary interactions under identical fixtures.
- Visual tests fix locale/timezone/Date/DPR, await fonts, disable animation/transition, compare computed style/key bbox/1440/390 layout and enforce pixelmatch changed ratio ≤0.5%; failure writes prototype/formal/diff artifacts.
- Real UAT captures the exact six-image full-page matrix from current-run en/zh ready rows；each row binds DB/API `canonical_report_content_digest`、`action_length_audit`、`content_audit`、`screenshot_sha256` and report/session/context digest。The 390x844 report images fully include the action region and actual `<=24-whitespace-word` / `<=64-Unicode-code-point` labels；deterministic parity separately proves exact boundaries.
- Every real screenshot maps to report/session ID, redacted DB/API assertions and a fact→judgment→action audit table.

## Phase 9: Internal locator cleanup

- Prototype/formal fixtures use distinct report/session UUID sentinels. DOM tests reject each exact value from visible text, title/tooltip, every `aria-*` and accessible name while retaining target/round/resume；contract tests keep both UUIDs required and CTA replay/next continues using `sourceReportId`.
- Remove the orphan `report.context.session` locale keys and require active zero-reference. Update P0.059 README/INDEX mapping to C-12, then rerun deterministic 1440/390 DOM/computed-style/bbox/viewport/pixel parity under its normal PASS cleanup.
- Final successful screenshots are captured separately through `/agent-browser` from the same formal real-backend ready report into `.test-output/acceptance/report-context-strip/<run-id>/`, so scenario cleanup does not erase user-facing evidence. Require exact `report-context-strip-desktop-1440x1200.png` and `report-context-strip-mobile-390x844.png`, exact viewports 1440x1200 / 390x844, `fullPage: true`, and no prototype/fixture-only/cropped/extra-state substitutes.
- Validate the directory has only those two PNGs plus `manifest.json`. For each image recompute SHA-256 and compare the manifest relative path/hash/`state=ready`/viewport/`fullPage=true`; require the same redacted report locator/digest, visible target/round/resume, and `reportSentinelAbsent=true` / `sessionSentinelAbsent=true` backed by the linked text/title/tooltip/`aria-*`/accessible-name audit.

## Phase 10: Independent current-plan reports list and Back recovery

- ReportsScreen table tests combine `getTargetJob + listTargetJobReports` for populated/empty/loading/network/invalid-contract states; target/round count/order/ID/sequence mismatches、跨轮 locator 复用、same-ID non-ready、latest-ready without current 与 current without latest 全部 fail closed and render no stale links.
- A/B target tests prove exact request binding, cross-target sentinel absence, target-switch first-commit clearing and stale-response fencing. The screen never calls `listTargetJobs`, so it cannot become a global report center.
- Per-round tests cover current ready, queued/generating latest, failed typed/no-Retry, same-ID ready de-dup and different-ID latest-ready status while exposing no full history list.
- Phase 10 的 Reports Back destination 断言仅保留为已被 Phase 11 取代的历史测试事实；missing/invalid Reports target uses `replaceRoute(workspace)` with no push/back-loop。Report/Generating table tests cover ready、pending、queued/generating、failed、timeout/network with current/last trusted `targetJobId`; Back URL is exactly `/reports?targetJobId=<id>`. Missing reportId、404、first-load network and invalid payload with no trusted target fall back to workspace.
- Route/security negatives reject route-provided target/status/round authority and title/reportId inference. Source spies prove only ReportsScreen calls `listTargetJobReports`; report/generating safe params remain reportId-only, reports safe params only targetJobId, Parse has no section compatibility, and TopBar has no report entry.
- P0.058/P0.059 and desktop/mobile source parity rerun；P0.059 adds populated/empty/loading/error ReportsScreen at 1440/390, while existing polling、typed failure、Context Strip sentinel and screenshot manifest assertions remain required.

## Phase 11: Reports Back direct read-only workspace detail

- ReportsScreen route tests first fail on the legacy Back destination, then require the exact trusted URL `/workspace?targetJobId=<id>` and `/workspace` replace fallback when identity is unavailable. Assert that no `resumeId`、`planId`、`reportId`、`section` or other query key is emitted.
- Component/source spies prove the Back read path renders the Workspace read-only detail and does not mount Parse, render parsing animation, call JD import, or start parse-status polling. Existing Report/Generating matrices continue to require trusted `/reports?targetJobId=...` and untrusted `/workspace` list fallback.
- Route/history tests prove one click reaches Workspace detail directly and browser Back exposes no Parse detour. Current-scope source/docs negative scanning excludes explicit revision/history records and rejects any positive Reports-to-Parse contract.
- P0.059 reruns current-target isolation and desktop/mobile parity, then clicks Reports Back and verifies the targetJobId-only Workspace detail plus no Parse animation/import/poll evidence. P0.058 reruns unchanged Report/Generating trusted/untrusted recovery.
