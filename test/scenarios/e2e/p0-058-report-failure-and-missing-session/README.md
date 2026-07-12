# E2E.P0.058 — Report failure + missing session + cross-user 404

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-3 / C-4 / C-6 / C-7 / D-6
> **Execution**: Vitest

This scenario composes six focused owner test files:

- `preflight.test.ts` keeps owner/scenario claims bound to executable evidence.
- `ReportFailureState.test.tsx` covers AI_* / REPORT_NOT_FOUND copy and retry/back handlers.
- `ReportMissingSessionState.test.tsx` covers missing-session and missing-report variants.
- `useFeedbackReport.test.tsx` covers ready, 404, missing-report and 5xx/refresh hook states.
- `ConversationReport.test.tsx` covers the current conversation-level report rendering contract.
- `useReportGenerationPoll.test.tsx` covers failed, 404 and timeout hook states.

The runner does not mount `GeneratingScreen`; timeout evidence stops at the poll-hook state. The shared real-mode gate proves production client configuration separately from these deterministic tests.
