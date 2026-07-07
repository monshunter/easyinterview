# Expected outcome

- `generating-screen` testid mounts, the 5-phase progress + evidence stream render, and once the poll observes `status=ready` the route flips to `report`.
- `report-dashboard` testid surfaces with header, context strip, summary cards, detail surface (default `questions` tab) and dashboard body sections.
- The generated client never sends `Idempotency-Key` for `getFeedbackReport` / `getTargetJob` / `getResume`.
- `listTargetJobReports` is never invoked from report or generating implementation code.
- Non-current vocabulary (`reportLayout`, `mistakesQueue`, voice surface imports, etc.) is absent.
