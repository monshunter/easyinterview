# E2E.P0.056 — Generating → Report happy path

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-1 / C-2 / C-5 / C-8 / C-11
> **Execution**: Vitest (`pnpm --filter @easyinterview/frontend test`)

This scenario composes four focused owner test files:

1. `preflight.test.ts` checks shared report/OpenAPI contracts and scenario ownership.
2. `useReportGenerationPoll.test.tsx` covers poll states, backoff, ready/failed callbacks, cancellation and read-only headers.
3. `GeneratingScreen.test.tsx` covers generating DOM, missing-report, i18n and ready/failed route handoff behavior.
4. `ConversationReport.test.tsx` covers conversation-level readiness, dimensions, evidence and action rendering.

It is not a single browser or live-backend journey. The shared real-mode gate proves production client configuration separately; the four owner files use deterministic test clients. Verify requires all four pass markers, counts current implementation testids, runs the scoped vocabulary lint and rejects `listTargetJobReports` in report/generating runtime code.

## Pipeline

| Script | Responsibility |
|--------|----------------|
| `scripts/setup.sh` | Record run metadata under `.test-output/e2e/p0-056…/` |
| `scripts/trigger.sh` | Run focused Vitest gates: preflight, GeneratingScreen, useReportGenerationPoll, ConversationReport |
| `scripts/verify.sh` | Assert all four test-file markers + testid coverage + lint/list negatives |
| `scripts/cleanup.sh` | Remove the `.test-output/e2e/p0-056…/` workspace |

## Privacy

The focused report gate verifies that the rendered report remains conversation-level. Broader route/storage payload privacy belongs to the dedicated CTA/failure scenarios rather than this composed runner.
