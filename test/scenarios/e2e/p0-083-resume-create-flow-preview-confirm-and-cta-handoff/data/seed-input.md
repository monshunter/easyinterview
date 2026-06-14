# P0.083 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Uploads/createUploadPresign.json` `default`
- `openapi/fixtures/Resumes/registerResume.json` `default`
- `openapi/fixtures/Resumes/getResume.json` `default`
- `openapi/fixtures/Resumes/updateResume.json` `default / idempotency-replay / validation-error-422`

## Synthetic Inputs

- Paste tab raw text: `Some pasted resume`
- Display name: derived from parsedSummary.identity.name (`Alice Example`)

## Polling Options

`window.__EI_RESUME_POLLING_OPTIONS__ = { initialDelayMs: 10, backoffFactor: 1, maxAttempts: 3, maxTotalMs: 500 }`

## Mock Overrides

- `updateResume`:
  - success: mockResolvedValue with a saved flat Resume
  - 422: mockRejectedValue with HTTP 422 wrapped error
  - generic failure: mockRejectedValue with HTTP 500 wrapped error
