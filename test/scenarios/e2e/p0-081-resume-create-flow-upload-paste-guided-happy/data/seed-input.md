# P0.081 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Auth/getRuntimeConfig.json` `default`
- `openapi/fixtures/Auth/getMe.json` `authenticated`
- `openapi/fixtures/Uploads/createUploadPresign.json` `default`
- `openapi/fixtures/Resumes/registerResume.json` `default` (upload) / `paste-text`
- `openapi/fixtures/Resumes/getResume.json` `default`

## Synthetic User Inputs

- Upload tab：`alice.pdf` size 2048 bytes, mime `application/pdf`
- Paste tab：`Hello, I am Alice and I lead frontend work.`
- Retired guided tab：negative only; the tab and panel must not render.

## Mock Harness Notes

- `getResume` mock returns parseStatus="processing" for non-terminal polling assertions
- `fetch` global is spied to intercept the signed URL PUT and return 200
