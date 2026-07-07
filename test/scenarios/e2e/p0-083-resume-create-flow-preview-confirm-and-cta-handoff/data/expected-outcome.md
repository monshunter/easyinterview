# P0.083 Expected Outcome

## Direct Create Outcome

- `resume-create-flow` renders for `resume_versions?flow=create`
- Upload/Paste register success navigates directly to `resume_versions?resumeId=<id>`
- `resume-preview-confirm` does not render
- `updateResume` is not called by create flow

## Idempotency

- `registerResume` request carries `Idempotency-Key` matching `v1.<unix>.<uuidv7>`
- Upload presign and PUT still use the existing upload path

## Home / Workspace CTA

- Home CTA testid: `home-resume-create`
- Workspace MissingResumeState CTA: navigates to `resume_versions?flow=create`
- Unauthenticated route lands on auth login pending-action state; pendingAction
  params contain `{ flow: "create" }` and NOT raw text / structuredProfile /
  rawText

## Privacy

- rawText and structuredProfile JSON content do NOT appear in nav params
- localStorage / sessionStorage receive no setItem calls during create handoff

## Trigger Log Assertions

- `Test Files +\d+ passed` matches
- Linked test files present in log
- Test names mentioning `Home`, `Workspace`, and direct detail navigation are exercised
