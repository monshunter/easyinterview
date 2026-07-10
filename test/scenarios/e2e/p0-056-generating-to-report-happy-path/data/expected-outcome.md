# Expected outcome

- Five focused Vitest files pass: owner preflight, poll hook, GeneratingScreen, ReportScreen and DetailSurface.
- Generating tests cover the current DOM and ready/failed route handoff; report tests separately cover dashboard/error/missing states, context labels and all five detail tabs.
- Read-only generated-client calls do not send `Idempotency-Key`; ContextStrip does not read raw resume/JD body fields.
- `listTargetJobReports` is never invoked from report or generating implementation code.
- Out-of-scope vocabulary (`reportLayout`, `mistakesQueue`, voice surface imports, etc.) is absent.
- The real-mode bootstrap gate and fixture-backed focused tests both pass, without representing a single live-backend/browser journey.
