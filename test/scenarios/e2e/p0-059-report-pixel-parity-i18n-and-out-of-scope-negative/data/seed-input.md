# Seed input

- Identical prototype/formal report fixtures: zh needs-practice and en well-prepared direct semantic reports with frozen context, long model prose, dimensions, evidence and actions.
- Identical prototype/formal generating fixture with API-owned `generating` status and no fake progress/observation/notification/records surface.
- Deterministic browser controls: 1440×900 and 390×844, deviceScaleFactor 1, UTC timezone, fixed clock, fixed locale, `document.fonts.ready`, animation disabled and transition disabled.
- `frontend/tests/pixel-parity/report-parity-helpers.ts` supplies normalized DOM text, computed-style/absolute-bbox snapshots and pixelmatch threshold 0.1.
- Active i18n and out-of-scope sources for the same report/generating implementation trees.
