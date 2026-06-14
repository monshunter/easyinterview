# Seed Input

- Auth state: authenticated user session.
- Generated client surfaces: `listTargetJobs`, `listPracticeSessions`, `listResumes`, `getTargetJob`, `getResume`, and `getPracticeSession`.
- Route normalization: historical `debrief_full` alias maps to canonical `debrief`.
- Runner inputs:
  - `frontend-real-backend-gate.sh` for generated-client real backend mode evidence.
  - `DebriefScreen.test.tsx`
  - `DebriefHeader.test.tsx`
  - `DebriefContextStrip.test.tsx`
  - `DebriefStepper.test.tsx`
  - `normalizeRoute.test.ts`
