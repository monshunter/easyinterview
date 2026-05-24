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
- Real backend mode generated-client gate for TargetJobs read/update and import path operations
- Nav to workspace with 7 interviewContext fields
- Browser-level Confirm click reaches `/workspace` with all 7 interviewContext query params and renders the `workspace-missing-resume` next-step state for the default `resume-unbound` handoff
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

## Real Backend Overlay

- The trigger first runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  generated client routes `listTargetJobs`, `createUploadPresign`,
  `importTargetJob`, `getTargetJob`, and `updateTargetJob` to the real backend
  base URL with cookie credentials, Idempotency-Key side effects, and
  provenance roundtrip.
- The parse edit/auth UI subcases remain fixture-backed for deterministic
  request-body, auth pending action, 4xx, and workspace navigation assertions.
  Backend E2E.P0.010-P0.013 pair this frontend routing proof with live
  TargetJob route/persistence/auth/IK/privacy/provenance semantics.

## Browser Route/Context Gate

- The trigger builds `frontend/dist` and runs Playwright
  `tests/pixel-parity/parse.spec.ts --grep "confirm navigates to workspace missing-resume with complete interview context"`.
- The browser gate opens `/parse?targetJobId=...`, mocks generated API
  responses, clicks Confirm, asserts `updateTargetJob` body/Idempotency-Key,
  verifies `/workspace` carries `targetJobId / jobId / jdId / planId /
  resumeVersionId / roundId / roundName`, and captures a non-empty screenshot
  of `workspace-missing-resume`.
