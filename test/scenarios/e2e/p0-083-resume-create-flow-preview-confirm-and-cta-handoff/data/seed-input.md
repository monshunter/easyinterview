# P0.083 Seed Input

## Fixture Snapshots Used

- `openapi/fixtures/Uploads/createUploadPresign.json` `default`
- `openapi/fixtures/Resumes/registerResume.json` `default`

## Synthetic Inputs

- Paste tab raw text: `Some pasted resume`
- Temporary title: derived from the first meaningful raw-text line

## Direct Handoff

- Home and Workspace CTAs open `resume_versions?flow=create`.
- Register success opens `resume_versions?resumeId=<id>` directly.
- Auth pendingAction carries only route params.
