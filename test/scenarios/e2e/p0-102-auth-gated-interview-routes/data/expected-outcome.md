# Expected Outcome

- Signed-out Home keeps the JD import surface visible but hides Recent mock interviews.
- Signed-out Home does not call `listTargetJobs` and does not show backend raw unauthorized or mock fixture errors.
- Signed-out Home does not expose the out-of-scope debrief CTA and never produces `pendingRoute=debrief`.
- Home business CTAs and direct protected URLs route to `auth_login` with pendingAction context.
- Protected route loading state renders `auth-route-gate` and avoids business API calls.
- Backend focused tests pass and prove protected business APIs return the auth middleware envelope when no session is present.
- Generated-client tests pass the same-key coalescing, separation, signal/mutation bypass, read/auth epoch fencing and settle-then-retry matrix.
- Runtime, Home and Workspace StrictMode loaders each issue one underlying same-key initial GET.
- Parse queued/ready initial read issues one underlying GET and each scheduler tick issues one later GET.
- The authenticated Resume Workshop route issues one list GET, zero detail GETs before Open and exactly one detail GET after Open.
- A rejected Resume Workshop list GET is evicted and user retry emits one new successful transport.
- `trigger.log` contains runner pass markers and does not contain `FAIL`, `no tests to run`, or the known raw Home error string `missing fixture for operationId: listTargetJobs`.
