# report screen

Source: `docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/`.
UI truth: `ui-design/src/screen-report.jsx` + `docs/ui-design/report-dashboard.md`.

## Composition

- `ReportScreen.tsx` — three-way dispatch:
  - `reportStatus=failed` → `components/ReportFailureState.tsx`
  - missing `sessionId` → `components/ReportMissingSessionState.tsx`
  - otherwise → `components/ReportDashboard.tsx`
- `components/ReportDashboard.tsx` — top-level dashboard container that drives:
  - `hooks/useFeedbackReport.ts` — single-shot `getFeedbackReport(reportId)` read; 4-state machine (`idle / loading / data / error / notFound`) with HTTP 404 mapped to `notFound + REPORT_NOT_FOUND`.
  - `hooks/useReportContextData.ts` — fetch from `getTargetJob` + `getResume`; labels fall back to raw IDs, while the structured TargetJob rounds also gate next-round availability.
  - Header / ContextStrip / three summary metrics / conversation-level dimensions, evidence, issues and next-actions sections.
- The report has no per-question tabs, turn selectors, retry turn IDs, hint state or modality branch.

## Replay handoff

`useReplayCtaHandlers` builds the path A / path B payloads via `handoff.ts`:

- Path A `goReplay()` carries `practiceGoal=retry_current_round` and the report-derived competency focus codes.
- Path B `goNextRound()` resolves only the immediate ordered successor from `TargetJob.summary.interviewRounds[]` and carries `practiceGoal=next_round` with that round context.
- Missing/unknown/final/duplicate round state fails closed. While either CTA is starting, both CTAs are disabled and repeated clicks create at most one plan/session.
- Authenticated users land on `practice` directly; unauthenticated users route through `useRequestAuth({type:'replay_practice'})` so the post-login resume re-issues the same payload.

`pendingAction` round-trip lives in `src/app/auth/pendingAction.ts`; the gate is exercised by `src/app/auth/__tests__/pendingActionReplayPractice.test.ts`.

## Cross-owner boundary

- Backend handler delivery: `docs/spec/backend-review/plans/001-report-generation-baseline`.
- Frontend workspace + practice owner: `docs/spec/frontend-workspace-and-practice` (consumer of the `nav("generating", …)` exit + `replay_practice` resume).
- The implementation must never call practice message/session mutation operations, voice turn operations, or workspace insight-only APIs. The `outOfScopeNegative.test.ts` and `scripts/lint/frontend_report_dashboard_out_of_scope.py` enforce the boundary.
