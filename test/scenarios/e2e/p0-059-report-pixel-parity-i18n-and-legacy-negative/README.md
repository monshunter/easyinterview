# E2E.P0.059 — Pixel parity + i18n + legacy negative gates

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-13 / C-14 / D-12
> **Execution**: Vitest + Python lint (Playwright spec files staged)

Runs the three composable gates:

- `reportDashboardI18nCoverage.test.ts` — namespace sync + B1 AI_* enum coverage.
- `legacyNegative.test.ts` — implementation code negative grep.
- `scripts/lint/frontend_report_dashboard_legacy.py` — repository-wide scoped grep.

The Playwright pixel-parity specs live under `frontend/tests/pixel-parity/{generating,report}.spec.ts` and are wired for desktop / mobile viewports. They run via `pnpm --filter @easyinterview/frontend test:pixel-parity`, which requires the local Playwright Chromium binary; the scenario verify therefore only confirms the spec files are staged.
