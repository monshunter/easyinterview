# E2E.P0.059 — Deterministic report/generating parity and stale-contract negatives

> **Owner**: frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-5 / C-6 / C-8 / C-10
> **Execution**: Vitest + Python lint + frontend build + Playwright semantic/pixel parity

## Given / When / Then

- **Given** prototype and formal surfaces receive identical deterministic API fixtures, fixed locale/timezone/Date, DPR 1, loaded fonts and disabled animation/transition.
- **When** the runner executes zh needs-practice, en well-prepared and honest generating surfaces at 1440×900 and 390×844.
- **Then** normalized DOM text, selected computed style and absolute bounding boxes match; `pixelmatch` uses threshold 0.1 and the changed-pixel ratio is at most 0.5%. A mismatch retains prototype/formal/diff images for diagnosis.

The source/geometry checks are first-class gates. A screenshot capture by itself is not acceptance evidence, and viewport-relative root coordinates must not be normalized away.

## Composed gates

- `preflight.test.ts` binds active owner docs, both Playwright specs and the shared comparison helper to DOM/style/bbox, absolute viewport geometry and pixel-difference evidence.
- `reportDashboardI18nCoverage.test.ts` proves exact zh/en UI-key coverage while model-owned report-language prose remains unchanged.
- Report/generating out-of-scope tests plus `frontend_report_dashboard_out_of_scope.py` reject fake-live copy, client focus authority and stale report identifiers.
- Frontend build succeeds before browser execution.
- `generating.spec.ts` and `report.spec.ts` use the identical fixture bridge for prototype/formal pages, compare deterministic semantic snapshots and enforce pixelmatch ≤0.5% at both viewport widths.

## Failure artifacts

Playwright attaches prototype/formal/diff PNGs only when dimensions or changed-pixel ratio fail. Cleanup removes successful-run output and preserves failure diagnostics for investigation.
