# Expected Outcome

- Signed-out Home keeps the JD import surface visible but hides Recent mock interviews.
- Signed-out Home does not call `listTargetJobs` and does not show backend raw unauthorized or mock fixture errors.
- Signed-out Home does not expose the retired debrief CTA and never produces `pendingRoute=debrief`.
- Home business CTAs and direct protected URLs route to `auth_login` with pendingAction context.
- Protected route loading state renders `auth-route-gate` and avoids business API calls.
- Backend focused tests pass and prove protected business APIs return the auth middleware envelope when no session is present.
- `trigger.log` contains runner pass markers and does not contain `FAIL`, `no tests to run`, or the known raw Home error string `missing fixture for operationId: listTargetJobs`.
