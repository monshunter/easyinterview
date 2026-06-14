# E2E.P0.074 Seed Input

## Users

- User A: authenticated candidate who owns ready and processing resume assets.
- User B: authenticated candidate without access to user A assets or versions.

## Flat Resume Rows

- Ready flat resumes owned by user A, each carrying `structured_profile`,
  `display_name`, source fields, parse state, and updated timestamps.
- Cross-user rows owned by user B, used to prove scoped 404/list isolation.
- Cursor pagination rows ordered by `updated_at DESC, id DESC`.

## Fixture Inputs

- `openapi/fixtures/Resumes/getResume.json`: `default`, `not-found`.
- `openapi/fixtures/Resumes/listResumes.json`: `default`, `empty`, `paginated`.
