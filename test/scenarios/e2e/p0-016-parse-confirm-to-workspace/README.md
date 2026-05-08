# E2E.P0.016 — Parse Confirm + Edit → Workspace (with Auth Gate)

> **Scenario ID**: E2E.P0.016
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the parse preview editing and confirm flow:
- Edit title/company/location/notes in Basic fields
- Level/language read-only (no input elements)
- Hit toggle cycling (false → true → partial → false)
- Confirm calls updateTargetJob with only supplied fields + Idempotency-Key
- Successful confirm navigates to workspace with interviewContext params
- 4xx inline error preserves editing state
- Authenticated: direct confirm → workspace
- Unauthenticated: confirm → requestAuth → auth_login with pendingAction=confirm_interview

## Fixture Variants

- `openapi/fixtures/TargetJobs/updateTargetJob.json`: success + 4xx validation
- `openapi/fixtures/Auth/getMe.json`: authenticated + unauthenticated

## Verification Points

- updateTargetJob request body: titleHint/companyNameHint/locationText/notes only
- Hit toggle state NOT in request body
- Level/language NOT in request body
- Idempotency-Key header present
- Nav to workspace with 7 interviewContext fields
- Auth pending action triggers correctly
- 4xx error inline, edit state preserved

## Scripts

- `scripts/setup.sh` — auth state selection (signed-in/out)
- `scripts/trigger.sh` — execute confirm flow (A: auth, B: unauth, C: 4xx)
- `scripts/verify.sh` — assert request body schema, auth gate, nav params
- `scripts/cleanup.sh` — reset auth state

## Offline Limitations

- Requires getMe fixture variant for auth state selection
- Requires updateTargetJob fixture for success/error paths
