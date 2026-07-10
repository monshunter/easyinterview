# E2E.P0.056 — Generating → Report happy path

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-1 / C-2 / C-5 / C-8 / C-11
> **Execution**: Vitest (`pnpm --filter @easyinterview/frontend test`)

This scenario composes five focused owner test files:

1. `preflight.test.ts` checks shared report/OpenAPI contracts and scenario ownership.
2. `useReportGenerationPoll.test.tsx` covers poll states, backoff, ready/failed callbacks, cancellation and read-only headers.
3. `GeneratingScreen.test.tsx` covers generating DOM, missing-report, i18n and ready/failed route handoff behavior.
4. `ReportScreen.test.tsx` covers report states, context labels, read-only generated-client calls and raw resume/JD field avoidance.
5. `DetailSurface.test.tsx` covers the five detail tabs and current local replay-marker behavior.

It is not a single browser or live-backend journey. The shared real-mode gate proves production client configuration separately; the five owner files use deterministic test clients. Verify requires all five pass markers, counts current implementation testids, runs the scoped vocabulary lint and rejects `listTargetJobReports` in report/generating runtime code.

## Pipeline

| Script | Responsibility |
|--------|----------------|
| `scripts/setup.sh` | Record run metadata under `.test-output/e2e/p0-056…/` |
| `scripts/trigger.sh` | Run focused Vitest gates: preflight, GeneratingScreen, useReportGenerationPoll, ReportScreen, DetailSurface |
| `scripts/verify.sh` | Assert all five test-file markers + testid coverage + lint/list negatives |
| `scripts/cleanup.sh` | Remove the `.test-output/e2e/p0-056…/` workspace |

## Privacy

The focused ReportScreen gate verifies that ContextStrip does not read raw resume or JD body fields. Broader route/storage payload privacy belongs to the dedicated CTA/failure scenarios rather than this composed runner.
