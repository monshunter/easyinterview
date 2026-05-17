# P0.082 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Resumes/registerResume.json` `default`
- `openapi/fixtures/Resumes/getResume.json` `default` (only ready scenario in fixture)
- `openapi/fixtures/Auth/getRuntimeConfig.json` + `getMe.json authenticated`

## Mock Harness Stepping

- Test 1 (failed → retry succeed): mock `getResume` resolves to parseStatus="failed" then "ready"
- Test 2 (timeout): mock `getResume` resolves to parseStatus="processing" for ≥ maxAttempts
- Test 3 (cancel preservation): mock `getResume` resolves to parseStatus="processing"; user cancels

## Polling Options

`window.__EI_RESUME_POLLING_OPTIONS__ = { initialDelayMs: 10, backoffFactor: 1, maxAttempts: 3, maxTotalMs: 500 }`

## Privacy Sentinel Values

- `PRIVATE_PARSED_TEXT_SNAPSHOT_DO_NOT_RENDER`
- `PRIVATE_ORIGINAL_TEXT_DO_NOT_RENDER`

These values are injected into parsedTextSnapshot / originalText fields and the
ParseFlow DOM is asserted not to contain them.
