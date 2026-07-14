# generating screen

Source: `docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/`.
UI truth: `ui-design/src/screens-p0-complete.jsx::ReportGeneratingScreen` (lines 269-399).

## Composition

- `GeneratingScreen.tsx` — displays only the observed `queued` / `generating` state and routes to `report` when the poll observes `status=ready`.
- `hooks/useReportGenerationPoll.ts` — 8-state poller for `getFeedbackReport(reportId)`:
  - States: `idle / polling / ready / failed / invalid / timeout / error / paused`.
  - Exponential backoff: initial 1.5s, factor 1.5, cap 8s, max attempts 49 (about 6 minutes), so status checks remain active across four 60-second provider calls plus the report-specific 10s / 20s / 40s retry delays.
  - Visibility / focus pause-resume preserves the current run's monotonic attempt and next delay, including when pause aborts an in-flight read. Resume waits before starting `n+1`; repeated pause/resume never restarts attempt 1 or creates concurrent reads.
  - Every 200 response passes the shared all-status report validator before becoming visible. Malformed payloads terminate as `invalid` and never overwrite the last trusted response.
  - Retained responses are owned by the exact `{client, reportId}` pair. A client or report identity change fails closed during render, before passive-effect cleanup; retrying with the same pair keeps the last trusted response available for recovery navigation.
  - Read-only contract: no `Idempotency-Key`. HTTP 404 surfaces `failed + REPORT_NOT_FOUND`.
  - `onReady(report)` / `onFailed(errorCode)` are debounced via `handoffNavigatedRef` so the same observation cannot nav twice.
- `components/HeaderHero / GeneratingErrorState` — source-level mirrors of the truthful prototype composition. There is no synthetic progress, phase, evidence stream, SLA, notification or records promise.

## Route boundary

`generating` lives in `frontend/src/app/routes.ts::NO_CHROME_ROUTES` — the screen owns the full viewport while polling. Without `reportId`, the screen renders `GeneratingErrorState` and never invokes the read.

## Handoff contract

On success the screen navigates to `report` with only the resolved `reportId`; the report API supplies all state and display context. Terminal report failures stay on the generating screen with a back action. On `timeout` or a recoverable read error it stays put and offers checking the same report again.

## Negative red lines

- No imports from `ui-design/src/data*` or `window.EI_DATA`.
- No imports from `ui-design/src/screen-practice`.
- No practice operation calls (`getPracticeSession`, `sendPracticeMessage`, etc.).
- No `listTargetJobReports` invocation; that operation belongs only to the target-scoped `ReportsScreen` current/latest overview.

Enforced by `src/app/screens/generating/__tests__/outOfScopeNegative.test.ts` and `scripts/lint/frontend_report_dashboard_out_of_scope.py`.
