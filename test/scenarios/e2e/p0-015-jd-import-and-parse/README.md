# E2E.P0.015 — Paste JD Import and Parse

> **Scenario ID**: E2E.P0.015
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated
> **Isolation**: repo-tracked Vitest / Playwright with stub fetch
> **parallel-safe**: No

## Scope

Verifies the single paste-only JD import flow:
- Home exposes one JD textarea, selects an existing ready resume, and rejects empty / whitespace-only input before dispatch
- Paste JD → exact `importTargetJob({rawText,targetLanguage,resumeId})` → parse loading → preview with route `targetJobId + resumeId`
- Signed-out submission stores only `opaquePendingImportId`; normal login atomically consumes the process-memory intent once, while missing / expired / duplicate consume dispatches no import
- Failed import (422 / 4xx) → inline error
- Failed parse (analysisStatus=failed) → failed UI

## Fixture Variants

- `openapi/fixtures/TargetJobs/importTargetJob.json`: `default`, `paste-primary`, and canonical `validation-blank-raw-text`
- `openapi/fixtures/TargetJobs/getTargetJob.json`: default parsed body used by focused Parse tests to construct queued / processing / ready / failed states
- `openapi/fixtures/TargetJobs/listTargetJobs.json`: Home recent default / empty / one-job / twelve-plus states

## Verification Points

- Home omits the out-of-scope hero sub copy and `解析并确认面试` CTA
- Home contains only the paste textarea, keeps the resume dropdown compact with create CTA on the same row, places `立即面试` below resume selection, and exposes no source controls, upload / URL trigger, or assist modal
- Home requires non-blank JD text and explicit ready resume selection before importTargetJob or pending import
- `importTargetJob` request is exactly `{rawText,targetLanguage,resumeId}` with no source discriminator or title/company hint
- Successful Home import navigates to parse with the selected real `resumeId`
- Idempotency-Key header on all side-effect calls
- Real backend mode generated-client gate for paste import, parse read, and update; any `createUploadPresign(purpose=resume)` assertion remains an independent Resume capability guard, not a JD intake path
- polling节奏 ≥600ms, progress step advances
- Browser gates run at desktop 1440×900 and mobile 390×844. Home formal/prototype parity captures the paste-only surface; Parse opens `/parse?targetJobId=...` with
  a fixture-backed ready `getTargetJob` response, captures the loading DOM
  screenshot, and proves preview is absent for the required loading window
- Preview渲染: title/company/location/requirements/hidden signals/rounds
- JD raw text not in console/URL/localStorage/telemetry
- No AI provider/prompt registry/LLM endpoint calls

## Scripts

- `scripts/setup.sh` — initialize isolated scenario output
- `scripts/trigger.sh` — execute generated-client, Home paste/auth, Parse state, build, and desktop/mobile browser gates
- `scripts/verify.sh` — assert exact paste-only request, one-shot auth intent, polling, desktop/mobile browser markers, old-intake negatives, and privacy redline
- `scripts/cleanup.sh` — reset mock state

## Offline Limitations

- Uses deterministic fixture-backed transport and does not require a live backend or provider.

## Real Backend Overlay

- The trigger first runs `src/api/targetJob.realApiMode.test.ts` with
  `VITE_EI_API_MODE=real` and
  `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, proving the production
  generated client routes `listTargetJobs`, paste-only `importTargetJob`,
  `getTargetJob`, and `updateTargetJob` to the real backend base URL with cookie
  credentials, Idempotency-Key side effects, exact flattened request body, and
  provenance roundtrip. The independent Resume-purpose presign assertion is not
  counted as JD intake evidence.
- Fixture-backed UI variants remain the deterministic source for paste-only
  DOM, 4xx, failed parse, polling, and privacy UI assertions. Real backend
  route/persistence/auth/IK/privacy/provenance semantics are paired with
  backend E2E.P0.010 / E2E.P0.012.
