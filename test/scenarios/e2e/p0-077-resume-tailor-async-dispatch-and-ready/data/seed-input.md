# E2E.P0.077 Seed Input

## Users

- User A: authenticated candidate who owns the flat resume and target job.

## Resume Rows

- Ready flat resume owned by user A.

## Target Jobs

- Ready target job owned by user A.

## Fixture Inputs

- `openapi/fixtures/ResumeTailor/requestResumeTailor.json`: `default`,
  `idempotency-replay`.
- `openapi/fixtures/ResumeTailor/getResumeTailorRun.json`: `default`,
  `queued`, `generating`, `failed`.
