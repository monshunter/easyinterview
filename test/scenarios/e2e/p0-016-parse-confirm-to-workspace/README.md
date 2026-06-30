# E2E.P0.016 — Parse Resume Binding + Save/Start Handoff

> **Scenario ID**: E2E.P0.016
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the parse preview editing and resume-bound launch flow:
- Edit title/company/location/notes in Basic fields
- Level/language read-only (no input elements)
- Hit toggle cycling (false → true → partial → false)
- Parse preview reads ready resumes and requires a bound resume before exit
- Save plan / Start interview call updateTargetJob with only supplied fields + Idempotency-Key
- Successful save navigates to workspace with real resumeId interviewContext params
- Successful start uses workspace autoStartPractice=1 and reaches practice
- 4xx inline error preserves editing state
- Unauthenticated without a verified ready resume: Save/Start disabled, no pendingAction

## Fixture Variants

- `openapi/fixtures/TargetJobs/updateTargetJob.json`: success + 4xx validation
- `openapi/fixtures/Resumes/listResumes.json`: ready list + empty variant
- `openapi/fixtures/Auth/getMe.json`: authenticated + unauthenticated

## Verification Points

- updateTargetJob request body: titleHint/companyNameHint/locationText/notes only
- Hit toggle state NOT in request body
- Level/language NOT in request body
- Idempotency-Key header present
- Real backend mode generated-client gate for TargetJobs read/update and import path operations
- Nav to workspace with real `resumeId` interviewContext fields
- Browser-level Save plan reaches `/workspace` with real ready `resumeId` and does not render `workspace-missing-resume`
- Browser-level Start interview reaches `practice` through workspace `autoStartPractice=1`
- `resume-unbound` never appears in success markers
- 4xx error inline, edit state preserved

## Scripts

- `scripts/setup.sh` — auth state selection (signed-in/out)
- `scripts/trigger.sh` — execute save/start flow (A: bound save, B: bound start, C: empty resume gate, D: 4xx)
- `scripts/verify.sh` — assert request body schema, disabled unauth/empty gate, nav params, no `resume-unbound`
- `scripts/cleanup.sh` — reset auth state

## Offline Limitations

- Requires getMe fixture variant for auth state selection
- Requires updateTargetJob fixture for success/error paths
- Requires listResumes fixture for ready/empty variants

## Real Backend Overlay

- The trigger first runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  generated client routes `listTargetJobs`, `createUploadPresign`,
  `importTargetJob`, `getTargetJob`, and `updateTargetJob` to the real backend
  base URL with cookie credentials, Idempotency-Key side effects, and
  provenance roundtrip.
- The parse edit/auth UI subcases remain fixture-backed for deterministic
  request-body, empty/unauth disabled gate, 4xx, and workspace/practice handoff assertions.
  Backend E2E.P0.010-P0.013 pair this frontend routing proof with live
  TargetJob route/persistence/auth/IK/privacy/provenance semantics.

## Browser Route/Context Gate

- The trigger builds `frontend/dist` and runs Playwright
  `tests/pixel-parity/parse.spec.ts --grep "save plan navigates|start interview hands off"`.
- The browser gate opens `/parse?targetJobId=...`, mocks generated API
  responses, clicks Save plan and Start interview, asserts `updateTargetJob` body/Idempotency-Key,
  verifies `/workspace` carries `targetJobId / jobId / jdId / planId /
  resumeId / roundId / roundName` with a real ready resume, verifies Start reaches
  `practice`, and rejects `workspace-missing-resume` / `resume-unbound` success markers.
