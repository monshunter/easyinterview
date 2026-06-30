# Seed Input

- TargetJob update fixture: `openapi/fixtures/TargetJobs/updateTargetJob.json` with success and validation error variants.
- Resume list fixture: `openapi/fixtures/Resumes/listResumes.json` with ready and empty variants.
- Auth fixture: `openapi/fixtures/Auth/getMe.json` with authenticated and unauthenticated variants.
- Parse edit fields: title, company, location, notes, hit toggles, read-only level, and read-only language.
- Navigation target: workspace with interview context parameters for target job, JD, real ready resume, round, and related context; Start interview adds `autoStartPractice=1`.
- Real backend overlay: `targetJob.realApiMode.test.ts` under `VITE_EI_API_MODE=real`.
