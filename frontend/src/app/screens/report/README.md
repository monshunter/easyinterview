# report screen

Source: `docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/`.
UI truth: `ui-design/src/screen-report.jsx` + `docs/ui-design/report-dashboard.md`.

## Composition

- `ReportScreen.tsx` (D-?) — three-way dispatch:
  - `reportStatus=failed` → `components/ReportFailureState.tsx`
  - missing `sessionId` → `components/ReportMissingSessionState.tsx`
  - otherwise → `components/ReportDashboard.tsx`
- `components/ReportDashboard.tsx` — top-level dashboard container that drives:
  - `hooks/useFeedbackReport.ts` — single-shot `getFeedbackReport(reportId)` read; 4-state machine (`idle / loading / data / error / notFound`) with HTTP 404 mapped to `notFound + REPORT_NOT_FOUND`.
  - `hooks/useReportContextData.ts` — label-only fetch from `getTargetJob` + `getResume`; failures fall back to the raw ID (decorative strip, never blocks the dashboard).
  - Header / ContextStrip / SummaryCards / DetailSurface / Dashboard body sections.
- `components/DetailSurface.tsx` — ARIA tablist for the 5 detail tabs; default tab is `questions`.
- `components/tabs/{Readiness,Dimensions,Questions,Evidence,Next}Tab.tsx` — source-level mirrors of `ui-design/src/screen-report.jsx` lines 311-516.

## Replay handoff

`useReplayCtaHandlers` builds the path A / path B payloads via `handoff.ts`:

- Path A `goReplay()` carries `practiceGoal=retry_current_round` and the report-derived `retryFocusTurnIds`.
- Path B `goNextRound()` rotates `roundId` via `inferNextRoundId` and carries `practiceGoal=next_round`.
- Authenticated users land on `practice` directly; unauthenticated users route through `useRequestAuth({type:'replay_practice'})` so the post-login resume re-issues the same payload.

`pendingAction` round-trip lives in `src/app/auth/pendingAction.ts`; the gate is exercised by `src/app/auth/__tests__/pendingActionReplayPractice.test.ts`.

## Cross-owner boundary

- Backend handler delivery: `docs/spec/backend-review/plans/001-report-generation-baseline`.
- Frontend workspace + practice owner: `docs/spec/frontend-workspace-and-practice` (consumer of the `nav("generating", …)` exit + `replay_practice` resume).
- The implementation must never call `listTargetJobReports`, `appendSessionEvent`, `completePracticeSession`, `createPracticeVoiceTurn`, or `getCompanyIntel`. The `legacyNegative.test.ts` and `scripts/lint/frontend_report_dashboard_legacy.py` enforce the boundary.
