# P0.082 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Resumes/registerResume.json` `default`
- `openapi/fixtures/Auth/getRuntimeConfig.json` + `getMe.json authenticated`

## Mock Harness Scope

- Register success returns a resume id.
- Create flow immediately routes to detail with that resume id.
- Parser and preview-confirm DOM anchors are asserted absent.

## Privacy Sentinel Values

- `PRIVATE_ORIGINAL_TEXT_DO_NOT_RENDER`

The value is injected into paste raw text and asserted absent from URL,
pendingAction, and storage surfaces.
