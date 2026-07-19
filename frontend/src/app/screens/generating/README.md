# generating screen

Source: `docs/spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/`.
UI truth: `formal frontend implementation` (lines 269-399).

## Composition

- `GeneratingScreen.tsx` — projects only the observed `queued` / `generating` state through the shared `AsyncTransitionScene` report composition and routes to `report` when the poll observes `status=ready`.
- `hooks/useReportGenerationPoll.ts` — 8-state poller for `getFeedbackReport(reportId)`:
  - States: `idle / polling / ready / failed / invalid / timeout / error / paused`.
  - Exponential backoff: initial 1.5s, factor 1.5, cap 8s, max attempts 49 (about 6 minutes), so status checks remain active across four 60-second provider calls plus the report-specific 10s / 20s / 40s retry delays.
  - Visibility / focus pause-resume preserves the current run's monotonic attempt and next delay, including when pause aborts an in-flight read. Resume waits before starting `n+1`; repeated pause/resume never restarts attempt 1 or creates concurrent reads.
  - Every 200 response passes the shared all-status report validator before becoming visible. Malformed payloads terminate as `invalid` and never overwrite the last trusted response.
  - Retained responses are owned by the exact `{client, reportId}` pair. A client or report identity change fails closed during render, before passive-effect cleanup; retrying with the same pair keeps the last trusted response available for recovery navigation.
  - Read-only contract: no `Idempotency-Key`. HTTP 404 surfaces `failed + REPORT_NOT_FOUND`.
  - `onReady(report)` / `onFailed(errorCode)` are debounced via `handoffNavigatedRef` so the same observation cannot nav twice.
- `components/GeneratingErrorState` — typed terminal/recoverable error surface. The shared transition has an indeterminate visual rule but no synthetic percentage, phase, evidence stream, SLA, notification or records promise.

## Route boundary

`generating` keeps the shared App TopBar and maps to the Interview primary navigation context. The transition owns the remaining viewport while polling. Without `reportId`, the screen renders `GeneratingErrorState` and never invokes the read.

## Handoff contract

On success the screen navigates to `report` with only the resolved `reportId`; the report API supplies all state and display context. Terminal report failures stay on the generating screen with a back action. On `timeout` or a recoverable read error it stays put and offers checking the same report again.

## Negative red lines

- No imports from `formal frontend implementation` or `window.EI_DATA`.
- No imports from `formal frontend implementation`.
- No practice operation calls (`getPracticeSession`, `sendPracticeMessage`, etc.).
- No `listTargetJobReports` invocation; that operation belongs only to the target-scoped `ReportsScreen` current/latest overview.

Enforced by `src/app/screens/generating/__tests__/outOfScopeNegative.test.ts` and `scripts/lint/frontend_report_dashboard_out_of_scope.py`.
