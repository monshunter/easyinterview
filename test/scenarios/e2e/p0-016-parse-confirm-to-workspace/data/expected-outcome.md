# Expected Outcome

- `targetJob.realApiMode.test.ts` proves TargetJob read/update/import-path operations route to the real backend base URL with cookie credentials and side-effect `Idempotency-Key`.
- `ParseScreen.test.tsx`, `ParseEdit.test.tsx`, `ParseAuthGate.test.tsx`, `ParseResumeBinding.test.tsx`, `MockInterviewCard.test.tsx`, `HomeRecentMocks.test.tsx`, and `navigation/interviewContext.test.ts` run and report passed tests.
- Parse success detail renders Basic fields as readonly text; notes/edit inputs are absent.
- Requirement evidence, hidden signals, round assumptions, and bound resume display have no edit/toggle/remove/picker controls.
- Parse round cards and Home recent round rails display saved 2~5 item `TargetJob.summary.interviewRounds[]` count, type/name, duration, and focus instead of static round/default rail labels.
- Playwright readonly detail acceptance attaches a screenshot and logs `E2E.P0.016 ... screenshotBytes=`.
- Shared navigation context derives `roundId` / `roundName` from the TargetJob round-assumption mapper and does not emit the old `Technical Round 1` string.
- Parse preview inherits the saved TargetJob `resumeId`; route `resumeId` is accepted only as a legacy fallback when TargetJob lacks one.
- Parse disables Start when no saved bound resume is available, without offering in-place binding.
- Browser-level readonly detail exposes only Start as the success footer action and makes no `updateTargetJob` request.
- Browser-level Start reaches `/practice` with real ready `resumeId`, `targetJobId`, `planId`, and `sessionId`.
- Unauthenticated users without a verified saved resume cannot trigger Start pendingAction.
- Trigger log contains no success markers for `resume-unbound`, `workspace-missing-resume`, or workspace auto-start.
