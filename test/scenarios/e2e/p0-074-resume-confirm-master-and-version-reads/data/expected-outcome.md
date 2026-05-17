# E2E.P0.074 Expected Outcome

## API Outcomes

- `POST /api/v1/resumes/{resumeAssetId}/structured-master` returns 201 with `versionType='structured_master'`.
- Same idempotency key replays the original 201 response and does not call the service again.
- A new idempotency key for an already confirmed active master returns `409 RESUME_STRUCTURED_MASTER_ALREADY_EXISTS`.
- Blank display name and processing assets return `422 VALIDATION_FAILED`; processing includes `details.reason='PARSE_NOT_READY'`.
- `GET /api/v1/resume-versions/{resumeVersionId}` returns the saved version for the owner and 404 for cross-user access.
- `GET /api/v1/resumes/{resumeAssetId}/versions` returns default, empty, paginated, and invalid-cursor outcomes with stable ordering.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Fixture parity covers all P0.074 confirm/get/list scenarios.
- Live DB integration covers partial UNIQUE index, soft-delete replacement, cross-user read hiding, invalid cursor, and pagination.
- Trigger log contains no skipped or no-op focused gates.
- Retired vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
