# Seed Input

- User A: `user-resume-a`
- User B: `user-resume-b`
- Register scenarios:
  - `upload`: `fileObjectId=01918fa0-0000-7000-8000-000000001100`
  - `paste`: `rawText` contains a private resume body and must not appear in scenario logs.
  - non-current `guided`: validation-only negative input; it must return 422 and
    must not create an asset/job.
- Fixture scenarios:
  - `openapi/fixtures/Resumes/registerResume.json`: `default`, `paste-text`
  - `openapi/fixtures/Resumes/getResume.json`: `default`, `not-found`
  - `openapi/fixtures/Resumes/listResumes.json`: `default`, `empty`, `paginated`
- Pagination seed: 25 resume assets with mixed `sourceType`, ordered by `updated_at DESC, id DESC`.
- Upload dependency: backend-upload `RegisterFileObject` confirms object existence and actual byte size before backend-resume creates an upload-backed asset.
