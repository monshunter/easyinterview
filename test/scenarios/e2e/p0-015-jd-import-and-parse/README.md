# E2E.P0.015 — JD Import and Parse (Paste/Upload/URL)

> **Scenario ID**: E2E.P0.015
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the full JD import flow with three source variants:
- Home selects an existing ready resume from a dropdown before any import source can submit
- Paste JD → importTargetJob(manual_text) → parse loading → preview with route `resumeId`
- Upload file → createUploadPresign → importTargetJob(file) → parse loading → preview with route `resumeId`
- URL input → importTargetJob(url) → parse loading → preview with route `resumeId`
- Failed import (4xx) → inline error
- Failed parse (analysisStatus=failed) → failed UI

## Fixture Variants

- `openapi/fixtures/TargetJobs/importTargetJob.json`: manual_text/file/url success + 422 invalid source
- `openapi/fixtures/Uploads/createUploadPresign.json`: target_job_attachment success + 4xx
- `openapi/fixtures/TargetJobs/getTargetJob.json`: queued→processing→ready polling + failed

## Verification Points

- Home omits the non-current hero sub copy and `解析并确认面试` CTA
- Home integrates paste textarea with upload/URL source actions inside the JD input card, keeps the resume dropdown compact with create CTA on the same row, and places `立即面试` below resume selection
- Home requires explicit ready resume dropdown selection before importTargetJob or pending import
- importTargetJob discriminator (type + required fields)
- Successful Home import navigates to parse with the selected real `resumeId`
- Idempotency-Key header on all side-effect calls
- Real backend mode generated-client gate for upload presign, import, parse read, and update
- polling节奏 ≥600ms, progress step advances
- Ready response browser gate: Playwright opens `/parse?targetJobId=...` with
  a fixture-backed ready `getTargetJob` response, captures the loading DOM
  screenshot, and proves preview is absent for the required loading window
- Preview渲染: title/company/location/requirements/hidden signals/rounds
- JD raw text not in console/URL/localStorage/telemetry
- No AI provider/prompt registry/LLM endpoint calls

## Scripts

- `scripts/setup.sh` — select fixture variant (paste/upload/url)
- `scripts/trigger.sh` — execute import flow per variant
- `scripts/verify.sh` — assert Home resume pre-bind, request body schema, polling behavior, ready-response browser marker, privacy redline
- `scripts/cleanup.sh` — reset mock state

## Offline Limitations

- Requires mock transport fixture variant selection
- Upload path tests placeholder file metadata only (no real binary upload)

## Real Backend Overlay

- The trigger first runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  generated client routes `listTargetJobs`, `createUploadPresign`,
  `importTargetJob`, `getTargetJob`, and `updateTargetJob` to the real backend
  base URL with cookie credentials, Idempotency-Key side effects, and
  provenance roundtrip.
- Fixture-backed UI variants remain the deterministic source for paste/upload/URL
  DOM, 4xx, failed parse, polling, and privacy UI assertions. Real backend
  route/persistence/auth/IK/privacy/provenance semantics are paired with
  backend E2E.P0.010-P0.013.
