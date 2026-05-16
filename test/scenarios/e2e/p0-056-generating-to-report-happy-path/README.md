# E2E.P0.056 — Generating → Report happy path

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-1 / C-2 / C-5 / C-8 / C-11
> **Execution**: Vitest (`pnpm --filter @easyinterview/frontend test`)

This scenario verifies the end-to-end mount → poll → report dispatch loop:

1. `generating?reportId=…&sessionId=…` mounts `GeneratingScreen` with the
   5-phase progress + live evidence stream.
2. `useReportGenerationPoll` polls `getFeedbackReport(reportId)` through the
   generated client, observes `status='ready'`, and navigates to `report` with
   the same identifiers + InterviewContext pass-through.
3. `ReportScreen` lands on `ReportDashboard` with header, ContextStrip, the
   four summary cards, the detail surface (default `questions` tab), and the
   dashboard body sections.

The verify step asserts the testid set is present, the read path never carries
`Idempotency-Key`, `listTargetJobReports` is never invoked, and the negative
vocabulary (legacy `reportLayout`, `mistakesQueue`, voice surface imports, etc.)
stays out of the implementation.

## Pipeline

| Script | Responsibility |
|--------|----------------|
| `scripts/setup.sh` | Record run metadata under `.test-output/e2e/p0-056…/` |
| `scripts/trigger.sh` | Run focused Vitest gates: preflight, GeneratingScreen, useReportGenerationPoll, ReportScreen, DetailSurface |
| `scripts/verify.sh` | Assert trigger log + testid coverage + lint script |
| `scripts/cleanup.sh` | Remove the `.test-output/e2e/p0-056…/` workspace |

## Privacy

All payloads exchanged within the harness use the InterviewContext owner IDs +
display knobs only; the scenario asserts that no raw answer / question / hint
text is persisted in route params or transcripts.
