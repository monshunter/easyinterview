# E2E.P0.056 — Honest Generating → frozen direct report

> **Owner**: backend-review/001-report-generation-baseline + frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-1 / C-2 / C-9 / C-10
> **Execution**: exact three-package Go evidence test + four focused frontend Vitest files

## Given / When / Then

- **Given** P0.047 has already published a schema-valid, redacted `practice-completion-evidence.v1` artifact and the report API exposes queued/generating/ready projections from the frozen `report-context.v1` snapshot.
- **When** the exact backend evidence test and four focused owner Vitest files run.
- **Then** the backend proves direct-ready persistence/read, immutable context and legacy-identifier absence; in-memory 62,397-byte regression and 917,504-byte boundary inputs each reach the provider exactly once without committed `input-*.json`; the UI shows an honest generating state and renders the direct semantic report from `reportId` alone without mutable context reads or route overrides.

This is a composed backend/frontend gate, not a single browser journey. The four focused owner test files are `preflight.test.ts`, `useReportGenerationPoll.test.tsx`, `GeneratingScreen.test.tsx` and `ConversationReport.test.tsx`. Frontend-only PASS cannot replace the backend evidence contract.

## Evidence contract

- `setup.sh` validates and consumes the existing P0.047 artifact; it does not recreate completion or copy raw content.
- `trigger.sh` runs `TestE2EP0056ReportBackendEvidence` in `internal/review`, `internal/store/review` and `internal/api/reports`, plus the four frontend files.
- `verify.sh` is the sole writer of `backend-evidence.json` (`report-backend-evidence.v1`) after exact RUN/PASS, marker, redacted database and frontend evidence all pass.
- Required backend markers are `REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS`, `REPORT_DIRECT_READY_PASS`, `REPORT_FROZEN_CONTEXT_READ_PASS`, `REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS`, `REPORT_62397_PROVIDER_ADMISSION_PASS` and `REPORT_917504_PROVIDER_ADMISSION_PASS`.

## Pipeline

| Script | Responsibility |
|--------|----------------|
| `scripts/setup.sh` | Validate P0.047 owner evidence, clear stale output and record only a redacted correlation ID |
| `scripts/trigger.sh` | Run the exact backend owner command and four focused frontend files |
| `scripts/verify.sh` | Reject partial/no-test/raw-content evidence and write the schema-valid redacted artifact |
| `scripts/cleanup.sh` | Remove transient logs/setup metadata while preserving the approved artifact |

## Privacy

No cookie, raw JD, resume, transcript, prompt body, provider output, message anchor or content-bearing ID is retained. The artifact contains only command/test status, owner-marker booleans and redacted report-state assertions.
