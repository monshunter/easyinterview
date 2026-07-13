# E2E.P0.058 — Report failure, recovery and missing report ID

> **Owner**: backend-review/001-report-generation-baseline + frontend-report-dashboard/001-report-screen-and-generating-handoff
> **Coverage tags**: C-1 / C-9
> **Execution**: exact three-package Go evidence test + seven focused frontend Vitest files

## Given / When / Then

- **Given** a schema-valid P0.047 completion artifact plus isolated missing/mismatched/oversized/output-retry/four-invalid/action-reset/retry-layer-separation backend cases and API-owned frontend report states.
- **When** the exact `TestE2EP0058ReportFailureBackendEvidence` command and seven focused frontend files run.
- **Then** context mismatch, 48,001-byte input, fourth invalid output and invalid direct report shape fail closed; each `GenerateReport` invocation owns initial plus at most three retries and waits `10s/20s/40s`; returning destroys that retry state and the next independent invocation starts at attempt one. `async_jobs.attempts/max_attempts` are infrastructure-only and do not affect product attempts. Only timeout/network transport checks can continue; terminal report failures, `REPORT_CONTEXT_TOO_LARGE`, not-found and invalid direct contracts are back-only.

The report route is keyed by `reportId` only. Missing-state means only a missing `reportId`; `sessionId`, target/resume identity, status and error route params are neither prerequisites nor business truth. Current OpenAPI has no failed-report regenerate operation, so reset is proved by two independent backend `GenerateReport` invocations rather than a claimed UI/API retry. The scenario composes focused unit/integration evidence and does not claim a live browser journey.

## Evidence contract

- `setup.sh` validates and consumes P0.047 `practice-completion-evidence.v1`; completion ownership is never recreated.
- `trigger.sh` runs the exact backend test across `internal/review`, `internal/store/review` and `internal/api/reports`, then seven frontend files including `GeneratingScreen.test.tsx` for the explicit context-too-large terminal UI.
- `verify.sh` is the sole writer of `backend-evidence.json` (`report-backend-evidence.v3`) after all backend/frontend markers, redacted database/runtime assertions and privacy negatives pass.
- Required backend markers are `REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS`, `REPORT_CONTEXT_TOO_LARGE_PASS`, `REPORT_OUTPUT_RETRY_PASS`, `REPORT_FOUR_INVALID_FAIL_CLOSED_PASS`, `REPORT_ACTION_RETRY_RESET_PASS` and `REPORT_RETRY_LAYER_SEPARATION_PASS`.

## Pipeline

| Script | Responsibility |
|--------|----------------|
| `scripts/setup.sh` | Validate owner evidence, clear stale evidence and record a redacted correlation ID |
| `scripts/trigger.sh` | Run the exact backend evidence command and seven focused frontend files |
| `scripts/verify.sh` | Require exact failure/recovery evidence and write the redacted artifact |
| `scripts/cleanup.sh` | Remove transient logs/setup metadata while preserving the approved artifact |

## Privacy

The persisted evidence contains only status enums, booleans, counts and the fixed retry schedule. Cookie, raw JD/resume/transcript, prompt body, provider output, anchors and content-bearing IDs are forbidden.
