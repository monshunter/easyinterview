# E2E.P0.058 — Report failure + missing session + cross-user 404

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-4 / C-12 / D-15
> **Execution**: Vitest

Covers the four failure variants:

- `reportStatus=failed` + AI_* enum → ReportFailureState with retry / back CTAs.
- Missing sessionId → ReportMissingSessionState (no `getFeedbackReport` call).
- HTTP 404 cross-user → useFeedbackReport.state=notFound with REPORT_NOT_FOUND copy.
- Generating timeout → GeneratingErrorState with retry CTA.
