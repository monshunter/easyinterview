# E2E.P0.034 resume register and list

## 1. Purpose

Validate the backend-resume baseline path for resume registration, get/list retrieval, idempotency replay, upload handoff, cursor pagination, cross-user hiding, and privacy redlines.

## 2. Requirements

- backend-resume C-1, C-2, C-5, C-6, C-7, C-8
- `docs/spec/backend-resume/plans/001-asset-register-parse-and-listing/bdd-plan.md`

## 3. Given / When / Then

Given two authenticated users and the two supported resume input modes
(`upload`, `paste`), plus B2 fixtures for `registerResume`, `getResume`, and
`listResumes`. Unsupported sourceType values are covered as invalid-input
regressions.

When user A registers resumes, replays the same idempotency key, fetches one resume, lists the collection with cursor pagination, and user B attempts to fetch user A's resume.

Then detail and list APIs return their distinct checked-in fixtures: `getResume` owns the full asset, while every `listResumes.items[]` contains exactly the nine scalar summary fields `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`. The list SQL performs one closed scalar projection with no detail JSON/blob scan or per-item detail fetch. Registration remains atomic, configured active/paste boundaries are checked with in-memory exact and limit+1 values before store, invalid input is rejected, cross-user resumes are hidden as 404, and raw resume body values stay out of logs and evidence.

## 4. Scripts

- `scripts/setup.sh`: prepares output directories and copies expected evidence notes.
- `scripts/trigger.sh`: runs the focused `cmd/api` HTTP scenario, exact ResumeSummary projection/mapper/fixture gates, upload register validation, store state-machine tests, and the live DB integration gate.
- `scripts/verify.sh`: rejects skips/no-op focused gates, checks required test evidence, reruns fixture parity, and performs privacy / current-scope negative searches.
- `scripts/cleanup.sh`: records cleanup completion while preserving logs under `.test-output/`.

## 5. Evidence

Scenario evidence is written to `.test-output/e2e/p0-034-resume-register-and-list/`:

- `setup.log`
- `trigger.log`
- `verify.log`
- `cleanup.log`
- `expected-outcome.md`

## 6. Baseline

This scenario is backend-owned. It proves that backend-resume exposes the baseline resume registration and listing surface required before frontend workspace can switch Resume Picker from disabled-list to active-list mode.

## 7. Offline Limits

Focused unit and fixture parity gates are necessary but not sufficient. The scenario also runs the integration-tag resume store gate with a concrete `DATABASE_URL`; missing DB availability is a scenario failure, not a PASS.
