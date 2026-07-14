# Expected Outcome

- `POST /api/v1/resumes` requires an authenticated session and `Idempotency-Key`.
- `sourceType=upload` requires `fileObjectId`; `paste` requires `rawText`;
  unsupported sourceType values and other invalid combinations return 422.
- Upload-backed registration calls backend-upload `RegisterFileObject` with expected purpose `resume`; missing object or byte-size mismatch does not create a resume.
- A successful register transaction creates one `resumes` row and one queued `async_jobs(job_type=resume_parse, resource_type=resume_asset)` row.
- Idempotency replay returns the original accepted response without creating duplicate resume/job/outbox rows.
- `GET /api/v1/resumes/{resumeId}` hides cross-user resumes as 404.
- `GET /api/v1/resumes` returns only the authenticated user's resumes with stable cursor pagination; every item has exactly `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`.
- List rows contain no `originalText`, `parsedTextSnapshot`, `structuredProfile`, `parsedSummary`, source-object metadata, timestamps unrelated to display, or archive/detail state.
- Store/service/handler gates prove one scalar list projection, no full-detail mapper reuse, no detail JSON/blob scan and no N+1 `getResume` loop.
- `registerResume`, `getResume`, and `listResumes` responses remain aligned with their B2 fixture scenarios.
- Scenario logs do not contain raw resume text, object keys, or unsupported module terms.
