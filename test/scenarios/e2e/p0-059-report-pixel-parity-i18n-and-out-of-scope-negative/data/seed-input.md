# Seed input

- `frontend/src/app/i18n/locales/{zh,en}.ts` carries the `report.*` + `generating.*` namespaces.
- `frontend/src/app/screens/{report,generating}/` is the implementation surface.
- `frontend/tests/pixel-parity/{generating,report}.spec.ts` is the browser-evidence surface.
- `docs/spec/frontend-report-dashboard/spec.md` and `plans/001-report-screen-and-generating-handoff/` form the seven owner artifacts checked by preflight.
- `scripts/lint/frontend_report_dashboard_out_of_scope.py` walks both screen trees and bans out-of-scope vocabulary.
