# Seed input

- Route: `report?reportId=…&sessionId=…&targetJobId=…&resumeId=…&roundId=round-tech-1`.
- Fixture: `getFeedbackReport=default` (status=ready, `retryFocusTurnIds=['turn-1','turn-3']`, issues populated).
- Generated responses: `createPracticePlan` returns a ready derived plan; `startPracticeSession` returns a fresh running session.
- Auth state: signed-in for both Header CTA paths; signed-out for the report-route auth gate.
- Pending action: `type=replay_practice`, `route=report`, with report/session/target/resume/round display params only.
