# Seed Input

- User A: `scenario-p0-033-user-a@example.test`
- User B: `scenario-p0-033-user-b@example.test`
- Purpose: `resume`
- Fixture: `openapi/fixtures/Uploads/createUploadPresign.json` scenario `default`
- Files:
  - `resume-small.pdf`: 1 MiB binary
  - `resume-boundary.bin`: 5 MiB binary
  - `resume-oversize.bin`: 11 MiB binary, expected validation failure

No generated file content is committed. `scripts/setup.sh` creates deterministic binary inputs under `.test-output/`.
