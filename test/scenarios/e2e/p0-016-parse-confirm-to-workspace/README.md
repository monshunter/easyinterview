# E2E.P0.016 — Parse Readonly Receipt + Start Handoff

> **Scenario ID**: E2E.P0.016
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the parse success detail as a readonly saved-plan receipt:
- Basic fields render as readonly text; notes/edit inputs are absent.
- Requirement evidence badges are not toggle buttons.
- Hidden signals and round assumptions are display-only.
- The saved bound `resumeId` is inherited from TargetJob, with route `resumeId` only as a legacy first-import fallback.
- Missing bound resume disables Start without exposing resume picker, resume options, create-resume fallback, Save plan, Cancel, or success Re-parse controls.
- Start interview does not call `updateTargetJob`; it directly uses `getPracticePlan` / `createPracticePlan` / `startPracticeSession` and reaches `practice`.
- Unauthenticated users without a verified saved resume cannot trigger Start pendingAction.

## Fixture Variants

- `openapi/fixtures/TargetJobs/getTargetJob.json`: ready TargetJob with saved `resumeId` and `currentPracticePlanId`.
- `openapi/fixtures/Resumes/listResumes.json`: ready list + empty variant.
- `openapi/fixtures/PracticePlans/getPracticePlan.json` / `createPracticePlan.json`: existing or newly created plan.
- `openapi/fixtures/PracticeSessions/startPracticeSession.json`: started session.
- `openapi/fixtures/Auth/getMe.json`: authenticated + unauthenticated.

## Verification Points

- Parse detail exposes no editable inputs, requirement toggles, hidden remove buttons, resume picker, Save plan, Cancel, or success Re-parse.
- Parse detail does not call `updateTargetJob` during Start.
- Browser-level readonly detail keeps real ready `resumeId` visible and never emits `resume-unbound`.
- Browser-level Start reaches `/practice` directly with `targetJobId`, `resumeId`, `planId`, and `sessionId`.
- Real backend mode generated-client gate still proves TargetJobs read/import/update API routing; the UI subcase separately proves Parse success detail is not an `updateTargetJob` consumer.

## Scripts

- `scripts/setup.sh` — auth state selection (signed-in/out)
- `scripts/trigger.sh` — execute readonly detail and direct Start flow
- `scripts/verify.sh` — assert readonly controls, no target patch, practice route, no `resume-unbound`
- `scripts/cleanup.sh` — reset auth state

## Offline Limitations

- Requires getMe fixture variant for auth state selection.
- Requires listResumes fixture for ready/empty variants.
- Requires practice plan/session fixtures for direct Start handoff.

## Real Backend Overlay

- The trigger first runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  generated client routes `listTargetJobs`, `createUploadPresign`,
  `importTargetJob`, `getTargetJob`, and `updateTargetJob` to the real backend
  base URL with cookie credentials, Idempotency-Key side effects, and
  provenance roundtrip.
- The parse UI subcases remain fixture-backed for deterministic readonly DOM,
  missing-resume disabled gate, no-PATCH assertion, and direct practice handoff.

## Browser Route/Context Gate

- The trigger builds `frontend/dist` and runs Playwright
  `tests/pixel-parity/parse.spec.ts --grep "readonly plan detail exposes|start interview hands off directly"`.
- The browser gate opens `/parse?targetJobId=...`, mocks generated API
  responses, verifies the readonly receipt has only Start as the success action,
  asserts no `updateTargetJob` PATCH calls were made, verifies Start reaches
  `practice`, and rejects `resume-unbound` / `workspace-missing-resume` success markers.
