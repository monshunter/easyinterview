# generating screen

Source: `docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/`.
UI truth: `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` (lines 269-399).

## Composition

- `GeneratingScreen.tsx` — drives the visual phase indicator + live evidence stream; routes to `report` when the poll observes `status=ready` / `status=failed`.
- `hooks/useReportGenerationPoll.ts` — 7-state poller for `getFeedbackReport(reportId)`:
  - States: `idle / polling / ready / failed / timeout / error / paused`.
  - Exponential backoff: initial 1.5s, factor 1.5, cap 8s, max attempts 30.
  - Visibility / focus pause-resume.
  - Read-only contract: no `Idempotency-Key`. HTTP 404 surfaces `failed + REPORT_NOT_FOUND`.
  - `onReady(report)` / `onFailed(errorCode)` are debounced via `handoffNavigatedRef` so the same observation cannot nav twice.
- `components/HeaderHero / ProgressBar / PhaseList / LiveEvidenceStream / SlaHint / GeneratingErrorState` — DOM mirrors of the prototype composition.

## Route boundary

`generating` lives in `frontend/src/app/routes.ts::NO_CHROME_ROUTES` — the screen owns the full viewport while polling. Without `reportId`, the screen renders `GeneratingErrorState` and never invokes the read.

## Handoff contract

On success the screen navigates to `report` with the 13-key handoff plus the resolved `reportId` and `sessionId`. On failure it navigates to `report?reportStatus=failed&errorCode=…`. On `timeout` it stays put and surfaces the retry CTA; the user can either retry (resetting the attempt counter) or back out to workspace.

## Negative red lines

- No imports from `ui-design/src/data*` or `window.EI_DATA`.
- No imports from `ui-design/src/screen-practice`.
- No practice operation calls (`getPracticeSession`, `sendPracticeMessage`, etc.).
- No `listTargetJobReports` invocation (dashboard-only D-7).

Enforced by `src/app/screens/generating/__tests__/outOfScopeNegative.test.ts` and `scripts/lint/frontend_report_dashboard_out_of_scope.py`.
