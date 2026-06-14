# E2E.P0.075 Seed Input

## Users

- User A: authenticated candidate who owns the editable flat resume.
- User B: authenticated candidate without access to user A resumes.

## Resume Assets

- Ready flat resume row owned by user A with initial `structured_profile`
  containing `headline`, `summary`, `skills`, `sections`, and server provenance.

## Fixture Inputs

- `openapi/fixtures/Resumes/updateResume.json`: `default`,
  `idempotency-replay`, `validation-error-422`.
