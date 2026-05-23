# Seed Input

- TargetJob update fixture: `openapi/fixtures/TargetJobs/updateTargetJob.json` with success and validation error variants.
- Auth fixture: `openapi/fixtures/Auth/getMe.json` with authenticated and unauthenticated variants.
- Parse edit fields: title, company, location, notes, hit toggles, read-only level, and read-only language.
- Navigation target: workspace with interview context parameters for target job, JD, resume/version, round, and related context.
- Real backend overlay: `targetJob.realApiMode.test.ts` under `VITE_EI_API_MODE=real`.
