# Expected outcome

- Four focused Vitest files pass: owner preflight, poll hook, GeneratingScreen and ConversationReport.
- Generating tests cover the current DOM and ready/failed route handoff; report tests cover conversation-level readiness, dimensions, evidence and actions.
- Read-only generated-client calls do not send `Idempotency-Key`; ContextStrip does not read raw resume/JD body fields.
- `listTargetJobReports` is never invoked from report or generating implementation code.
- Out-of-scope vocabulary (`reportLayout`, `mistakesQueue`, voice surface imports, etc.) is absent.
- The real-mode bootstrap gate and fixture-backed focused tests both pass, without representing a single live-backend/browser journey.
