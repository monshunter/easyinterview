# E2E.P0.059 — Current-plan reports and report-return parity

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-12 / C-13 / C-14
> **Execution**: Vitest + Python source lint + frontend build + Playwright

## Given / When / Then

- **Given** `/reports?targetJobId=<uuid>` resolves one trusted TargetJob and its canonical report overview, while Report/Generating retain their reportId-only contracts.
- **When** `ReportsScreen` renders ready, loading, empty, error, latest-ready, or mismatched-target data at 1440x900 and 390x844.
- **Then** it shows only the current plan's canonical rounds, each round's `currentReport` and `latestAttempt`, with current/latest-only actions and no full history or global Report Center. Back returns to the same Parse plan; Report/Generating trusted Back remains covered as a regression.

## ReportsScreen contract

- Reads `getTargetJob(targetJobId)` and `listTargetJobReports(targetJobId)` together, validates target and canonical round identity, and fails closed on mismatch.
- Renders current report, queued/generating link, typed failed status, and a latest-ready status without adding a second ready/history action.
- Clears rows for loading, empty, network/contract error and cross-target mismatch; other-plan identifiers and stale report IDs never enter visible or accessible DOM.
- `ReportsScreen.tsx` is the only production screen consumer of `listTargetJobReports`; Parse, Report and Generating remain zero consumers.
- Back returns to `/parse?targetJobId=<trusted uuid>`; the route contains only `targetJobId` and the TopBar contains no Reports item.

## Parity evidence

- `frontend/tests/pixel-parity/reports.spec.ts` compares formal/prototype normalized DOM, computed style, absolute viewport bounding boxes, responsive width and pixelmatch ≤0.5% for ready/loading/empty/error/latest-ready/mismatch at both viewports.
- Ready formal/prototype screenshots are attached per desktop/mobile project; state mismatch keeps formal/prototype/diff diagnostics.
- Existing `report.spec.ts` and `generating.spec.ts` remain in the same run, preserving report/generating content and trusted-return parity.
- `frontend_report_dashboard_out_of_scope.py` rejects history/global-center vocabulary and asserts the one exact production list consumer.

## Scripts

- `scripts/setup.sh` — resets scenario-owned output and writes `setup.env`.
- `scripts/trigger.sh` — runs script/source self-tests, focused Vitest, lint self-tests, frontend build, and three desktop/mobile Playwright specs.
- `scripts/verify.sh` — binds actual filenames, test titles, unique-consumer output, current-plan/current-latest/Back markers and Playwright PASS.
- `scripts/cleanup.sh` — removes successful scenario-owned output only; failures keep diagnostics.

## Environment

The scenario uses fixture-backed runners and the repo static parity server. No shared Docker or host-run application environment is required.
