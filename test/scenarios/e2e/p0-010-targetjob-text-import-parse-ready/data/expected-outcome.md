# Expected Outcome

- `importTargetJob` returns `202 + TargetJobWithJob` with `jobType=target_import` and `status=queued`.
- `target.import.requested.sourceType` is `text`, not `manual_text`, and its payload omits raw JD text.
- The parse executor writes requirements, summary provenance, fitSummary provenance, `latest_parse_job_id`, `target.parsed`, and `source_refresh`.
- `listTargetJobs` and `getTargetJob` expose the ready TargetJob for the same user only.
- `updateTargetJob` accepts a legal status transition and preserves `analysisStatus=ready`.
- Store-level idempotency gates prevent duplicate rows / outbox for repeated keys.
