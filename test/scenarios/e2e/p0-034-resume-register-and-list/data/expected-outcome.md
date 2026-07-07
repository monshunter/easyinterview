# Expected Outcome

- `POST /api/v1/resumes` requires an authenticated session and `Idempotency-Key`.
- `sourceType=upload` requires `fileObjectId`; `paste` requires `rawText`;
  non-current `guided` and other invalid combinations return 422.
- Upload-backed registration calls backend-upload `RegisterFileObject` with expected purpose `resume`; missing object or byte-size mismatch does not create a resume asset.
- A successful register transaction creates one `resume_assets` row and one queued `async_jobs(job_type=resume_parse, resource_type=resume_asset)` row.
- Idempotency replay returns the original accepted response without creating duplicate asset/job/outbox rows.
- `GET /api/v1/resumes/{resumeId}` hides cross-user resumes as 404.
- `GET /api/v1/resumes` returns only the authenticated user's assets with stable cursor pagination.
- `registerResume`, `getResume`, and `listResumes` responses remain aligned with their B2 fixture scenarios.
- Scenario logs do not contain raw resume text, object keys, or non-current module terms.
