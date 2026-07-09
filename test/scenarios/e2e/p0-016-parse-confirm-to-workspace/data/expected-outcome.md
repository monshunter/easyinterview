# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJob read/update/import-path operations route to the real backend base URL with cookie credentials and side-effect `Idempotency-Key`.
- `ParseEdit.test.tsx`, `ParseAuthGate.test.tsx`, and `ParseResumeBinding.test.tsx` run and report passed tests.
- Parse success detail renders Basic fields as readonly text; notes/edit inputs are absent.
- Requirement evidence, hidden signals, round assumptions, and bound resume display have no edit/toggle/remove/picker controls.
- Parse preview inherits the saved TargetJob `resumeId`; route `resumeId` is accepted only as a legacy fallback when TargetJob lacks one.
- Parse disables Start when no saved bound resume is available, without offering in-place binding.
- Browser-level readonly detail exposes only Start as the success footer action and makes no `updateTargetJob` request.
- Browser-level Start reaches `/practice` with real ready `resumeId`, `targetJobId`, `planId`, and `sessionId`.
- Unauthenticated users without a verified saved resume cannot trigger Start pendingAction.
- Trigger log contains no success markers for `resume-unbound`, `workspace-missing-resume`, or workspace auto-start.
