# P0.083 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Uploads/createUploadPresign.json` `default`
- `openapi/fixtures/Resumes/registerResume.json` `default`

## Synthetic Inputs

- Paste tab raw text: `Some pasted resume`
- Register request title remains neutral; readable `displayName` is owned by backend parse output, not by the pasted first line.

## Direct Handoff

- Home CTA opens `resume_versions?flow=create`.
- Register success opens `resume_versions?resumeId=<id>` directly.
- Auth pendingAction carries only route params.
