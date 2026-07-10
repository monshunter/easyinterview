# E2E.P0.059 — Pixel parity + i18n + out-of-scope negative gates

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-13 / C-14 / D-12
> **Execution**: Vitest + Python lint + frontend build + Playwright pixel parity

Runs the composable gates:

- `report/__tests__/preflight.test.ts` — keeps the active spec, six plan artifacts, both Playwright sources and this scenario runner bound to executable browser evidence.
- `reportDashboardI18nCoverage.test.ts` — namespace sync + B1 AI_* enum coverage.
- `outOfScopeNegative.test.ts` — implementation code negative grep.
- `scripts/lint/frontend_report_dashboard_out_of_scope.py` — repository-wide scoped grep.
- `pnpm --filter @easyinterview/frontend build` — compile the frontend bundle before visual parity.
- `pnpm --filter @easyinterview/frontend test:pixel-parity -- tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts` — execute seven generating/report states across desktop and mobile; every state captures an in-memory screenshot and asserts that its buffer is non-empty.

The verify step reads `trigger.log` and fails unless the owner/browser preflight marker, both Playwright spec paths and the Playwright pass marker are present. The scenario does not create or maintain image-comparison files.
