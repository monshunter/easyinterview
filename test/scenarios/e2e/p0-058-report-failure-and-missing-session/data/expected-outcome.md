# Expected outcome

- All three backend packages execute and pass `TestE2EP0058ReportFailureBackendEvidence` with the six exact action-retry/fail-closed markers and no skipped/no-test/failure output.
- Context mismatch and oversized input perform no ready write; the in-memory 917,505-byte case causes zero provider calls; invalid→valid consumes two calls; a fourth invalid result ends the current action. One invocation waits exactly `10s/20s/40s`, destroys state on return, and a second independent invocation starts at attempt one; async job attempts do not affect the product attempt.
- Eight focused frontend files pass: missing only `reportId` performs no fetch; report status/context come only from API; timeout/network can continue-check; terminal/not-found/invalid direct/`REPORT_CONTEXT_TOO_LARGE` expose back only and never raw enums or fake reports. Trusted current/last report context returns Report/Generating to `/reports?targetJobId=...`; missing or malformed trusted context returns `/workspace`.
- `verify.sh` alone writes exact-key `report-backend-evidence.v3` with redacted `database` fail-closed facts, separate `runtime` action/reset/separation facts and `result=PASS`.
- Cleanup preserves only the approved evidence artifact.
