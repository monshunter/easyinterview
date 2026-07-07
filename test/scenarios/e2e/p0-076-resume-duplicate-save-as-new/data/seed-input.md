# E2E.P0.076 Seed Input

## Users

- User A: authenticated candidate who owns the source flat resume.
- User B: authenticated candidate without access to user A source resume.

## Resume Rows

- Ready source resume owned by user A with source snapshot fields and
  `structured_profile` containing headline, summary, skills, sections, and
  server provenance.

## Fixture Inputs

- `openapi/fixtures/Resumes/duplicateResume.json`: `default`,
  `idempotency-replay`, `validation-error-422`.
