# E2E.P0.074 Seed Input

## Users

- User A: authenticated candidate who owns ready and processing resume assets.
- User B: authenticated candidate without access to user A assets or versions.

## Resume Assets

- First asset: `parse_status='ready'`, used for `confirmResumeStructuredMaster`, `getResumeVersion`, and populated `listResumeVersions`.
- Second asset: `parse_status='ready'`, used for empty `listResumeVersions`.
- Third asset: `parse_status='processing'`, used for `PARSE_NOT_READY` validation.

## Version Rows

- One active `structured_master` row for the first asset after confirm.
- Additional targeted rows seeded by the store integration gate to prove cursor pagination and `updated_at DESC, id DESC` ordering before Phase 5 branch endpoints exist.

## Fixture Inputs

- `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json`: `default`, `idempotency-replay`, `already-exists-409`, `validation-422`.
- `openapi/fixtures/Resumes/getResumeVersion.json`: `default`, `not-found-404`.
- `openapi/fixtures/Resumes/listResumeVersions.json`: `default`, `empty`, `paginated`.
