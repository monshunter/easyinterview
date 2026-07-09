# Seed Input

- TargetJob detail fixture: `openapi/fixtures/TargetJobs/getTargetJob.json` with ready status, saved `resumeId`, optional `currentPracticePlanId`, and `summary.interviewRounds[]`.
- TargetJob list fixture: `openapi/fixtures/TargetJobs/listTargetJobs.json` with ready cards and `summary.interviewRounds[]` for recent-card round rails.
- Resume list fixture: `openapi/fixtures/Resumes/listResumes.json` with ready and empty variants.
- Practice fixtures: `openapi/fixtures/PracticePlans/getPracticePlan.json`, `openapi/fixtures/PracticePlans/createPracticePlan.json`, and `openapi/fixtures/PracticeSessions/startPracticeSession.json`.
- Auth fixture: `openapi/fixtures/Auth/getMe.json` with authenticated and unauthenticated variants.
- Parse readonly fields: title, company, location, level, language, requirements, hidden signals, round assumptions, and bound resume display.
- Navigation target: direct `practice` route with target job, JD, real ready resume, round, plan, and session context.
- Real backend overlay: `targetJob.realApiMode.test.ts` under `VITE_EI_API_MODE=real`.
