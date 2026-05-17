# P0.083 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Uploads/createUploadPresign.json` `default`
- `openapi/fixtures/Resumes/registerResume.json` `default`
- `openapi/fixtures/Resumes/getResume.json` `default`
- `openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` `default / idempotency-replay / already-exists-409 / validation-422`
- `openapi/fixtures/Resumes/listResumeVersions.json` `default`

## Synthetic Inputs

- Paste tab raw text: `Some pasted resume`
- Display name: derived from parsedSummary.identity.name (`Alice Example`)

## Polling Options

`window.__EI_RESUME_POLLING_OPTIONS__ = { initialDelayMs: 10, backoffFactor: 1, maxAttempts: 3, maxTotalMs: 500 }`

## Mock Overrides

- `confirmResumeStructuredMaster`:
  - success: mockResolvedValue with a synthetic ResumeVersion
  - replay: same key/body second call → same version
  - 409: mockRejectedValue with HTTP 409 wrapped error
  - 422: mockRejectedValue with HTTP 422 wrapped error
- `listResumeVersions`: mockResolvedValue with single structured_master item
