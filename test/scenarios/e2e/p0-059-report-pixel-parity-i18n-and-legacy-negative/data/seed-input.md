# Seed input

- `frontend/src/app/i18n/locales/{zh,en}.ts` carries the `report.*` + `generating.*` namespaces.
- `frontend/src/app/screens/{report,generating}/` is the implementation surface.
- `scripts/lint/frontend_report_dashboard_legacy.py` walks both screen trees and bans retired vocabulary.
