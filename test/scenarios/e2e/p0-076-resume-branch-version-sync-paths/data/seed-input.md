# E2E.P0.076 Seed Input

## Users

- User A: authenticated candidate who owns the parent resume version and target job.
- User B: authenticated candidate without access to user A parent version or target job.

## Resume Assets And Versions

- Ready resume asset owned by user A.
- Active structured master version owned by user A with `structured_profile` containing headline, summary, skills, sections, and server provenance.

## Target Jobs

- Ready target job owned by user A.
- Foreign target job owned by user B for cross-user 404.

## Fixture Inputs

- `openapi/fixtures/Resumes/branchResumeVersion.json`: `default`, `copy-master-sync`, `blank-sync`, `idempotent-replay`, `validation-error-422`.
