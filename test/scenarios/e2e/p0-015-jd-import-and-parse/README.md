# E2E.P0.015 — JD Import and Parse (Paste/Upload/URL)

> **Scenario ID**: E2E.P0.015
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the full JD import flow with three source variants:
- Paste JD → importTargetJob(manual_text) → parse loading → preview
- Upload file → createUploadPresign → importTargetJob(file) → parse loading → preview
- URL input → importTargetJob(url) → parse loading → preview
- Failed import (4xx) → inline error
- Failed parse (analysisStatus=failed) → failed UI

## Fixture Variants

- `openapi/fixtures/TargetJobs/importTargetJob.json`: manual_text/file/url success + 422 invalid source
- `openapi/fixtures/Uploads/createUploadPresign.json`: target_job_attachment success + 4xx
- `openapi/fixtures/TargetJobs/getTargetJob.json`: queued→processing→ready polling + failed

## Verification Points

- importTargetJob discriminator (type + required fields)
- Idempotency-Key header on all side-effect calls
- polling节奏 ≥600ms, progress step advances
- Preview渲染: title/company/location/requirements/hidden signals/rounds
- JD raw text not in console/URL/localStorage/telemetry
- No AI provider/prompt registry/LLM endpoint calls

## Scripts

- `scripts/setup.sh` — select fixture variant (paste/upload/url)
- `scripts/trigger.sh` — execute import flow per variant
- `scripts/verify.sh` — assert request body schema, polling behavior, privacy redline
- `scripts/cleanup.sh` — reset mock state

## Offline Limitations

- Requires mock transport fixture variant selection
- Upload path tests placeholder file metadata only (no real binary upload)
