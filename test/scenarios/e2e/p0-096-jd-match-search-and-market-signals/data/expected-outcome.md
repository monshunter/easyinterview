# E2E.P0.096 expected outcome

- `searchJobs` success: sync 200 + searchRunId + items JOIN existing
  recommendations + jd_match_search_runs new row.
- `searchJobs` 30s timeout: 502 + `AI_PROVIDER_TIMEOUT`.
- `searchJobs` output_invalid: 502 + `AI_OUTPUT_INVALID` or generic
  AI_PROVIDER_TIMEOUT mapping per cmd/api adapter.
- IK replay: identical searchRunId + body.
- `getMarketSignals?window=7d/14d/30d` returns 200 + 4 signals + asOf.
- `getMarketSignals?window=invalid` returns 422 VALIDATION_FAILED.
- query / filters never appear in log / audit / outbox.
