# Honest Grounded Report Screen Test Plan

> **版本**: 2.9
> **状态**: completed
> **更新日期**: 2026-07-13

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
