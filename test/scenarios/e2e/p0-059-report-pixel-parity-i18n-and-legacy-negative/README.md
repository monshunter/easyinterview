# E2E.P0.059 — Pixel parity + i18n + legacy negative gates

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-13 / C-14 / D-12
> **Execution**: Vitest + Python lint + frontend build + Playwright pixel parity

Runs the composable gates:

- `reportDashboardI18nCoverage.test.ts` — namespace sync + B1 AI_* enum coverage.
- `legacyNegative.test.ts` — implementation code negative grep.
- `scripts/lint/frontend_report_dashboard_legacy.py` — repository-wide scoped grep.
- `pnpm --filter @easyinterview/frontend build` — compile the frontend bundle before visual parity.
- `pnpm --filter @easyinterview/frontend test:pixel-parity -- tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts` — execute the generating/report pixel-parity specs for desktop and mobile viewports.

The verify step reads `trigger.log` and fails unless both Playwright spec paths and the Playwright pass marker are present.
