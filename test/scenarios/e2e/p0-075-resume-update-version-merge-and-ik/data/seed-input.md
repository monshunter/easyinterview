# E2E.P0.075 Seed Input

## Users

- User A: authenticated candidate who owns the editable resume version.
- User B: authenticated candidate without access to user A versions.

## Resume Assets

- Ready resume asset owned by user A.

## Version Rows

- One active `structured_master` row for user A with initial `structured_profile` containing `headline`, `summary`, `skills`, `sections`, and server provenance.
- One soft-deleted version row for user A used to prove deleted rows return 404.

## Fixture Inputs

- `openapi/fixtures/Resumes/updateResumeVersion.json`: `default`, `idempotency-replay`, `validation-error-422`.
