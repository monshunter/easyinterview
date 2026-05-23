# Seed Input

- Frontend surfaces: `DebriefFailureState`, `DebriefMissingContextState`, `DebriefTimeoutState`, Step 2 launcher, `useRequestAuth`, and replay handoff logic.
- Runner inputs:
  - `frontend-real-backend-gate.sh`
  - `DebriefScreen.test.tsx`
  - `privacyBoundary.test.ts`
  - `InterviewContext.test.tsx`
- Source-level handoff gate: `DebriefScreen.tsx` must reference `createPracticePlan`, `startPracticeSession`, and `sourceDebriefId`.
