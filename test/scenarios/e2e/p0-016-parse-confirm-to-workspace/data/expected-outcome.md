# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJob read/update/import-path operations route to the real backend base URL with cookie credentials and side-effect `Idempotency-Key`.
- `ParseEdit.test.tsx`, `ParseAuthGate.test.tsx`, and `ParseResumeBinding.test.tsx` run and report passed tests.
- Save plan / Start interview send only editable fields such as title, company, location, and notes; hit toggle state, level, and language are not submitted.
- Parse preview reads ready resumes, leaves Save/Start disabled until the user explicitly selects a resume, and disables Save/Start when no ready resume exists.
- Authenticated Save plan navigates to workspace with interview context fields and the explicitly selected real ready `resumeId`.
- Browser-level Save plan reaches `/workspace` with all context query params:
  `targetJobId`, `jobId`, `jdId`, `planId`, `resumeId`, `roundId`, and
  `roundName`; `resumeId` is a real ready resume id and `workspace-missing-resume`
  is not rendered.
- Browser-level Start interview uses the explicitly selected `resumeId`, then uses workspace `autoStartPractice=1` and reaches `practice`.
- Unauthenticated users without a verified ready resume cannot trigger Save/Start pendingAction.
- Validation error remains inline and preserves edit state.
- Trigger log contains no read-only field identifiers such as `parse-basics-level` or `parse-basics-language`.
- Trigger log contains no `resume-unbound` success marker.
