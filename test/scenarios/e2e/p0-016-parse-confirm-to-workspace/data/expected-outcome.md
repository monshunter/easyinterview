# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJob read/update/import-path operations route to the real backend base URL with cookie credentials and side-effect `Idempotency-Key`.
- `ParseEdit.test.tsx` and `ParseAuthGate.test.tsx` run and report passed tests.
- Confirm sends only editable fields such as title, company, location, and notes; hit toggle state, level, and language are not submitted.
- Authenticated confirm navigates to workspace with interview context fields.
- Browser-level confirm reaches `/workspace` with all 7 context query params:
  `targetJobId`, `jobId`, `jdId`, `planId`, `resumeId`, `roundId`, and
  `roundName`, then renders the `workspace-missing-resume` next-step state for
  the default `resume-unbound` handoff.
- Unauthenticated confirm requests auth with a `confirm_interview` pending action.
- Validation error remains inline and preserves edit state.
- Trigger log contains no read-only field identifiers such as `parse-basics-level` or `parse-basics-language`.
