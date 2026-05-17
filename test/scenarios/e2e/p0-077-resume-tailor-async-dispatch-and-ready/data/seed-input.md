# E2E.P0.077 Seed Input

## Users

- User A: authenticated candidate who owns the parent resume version and target job.

## Resume Assets And Versions

- Ready resume asset owned by user A.
- Active structured master version owned by user A.

## Target Jobs

- Ready target job owned by user A.

## Fixture Inputs

- `openapi/fixtures/Resumes/branchResumeVersion.json`: `ai-select-202-with-job`.
- Later phases extend this scenario with `requestResumeTailor.json` and `getResumeTailorRun.json`.
