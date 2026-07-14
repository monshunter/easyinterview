# Expected Outcome

- `importTargetJob` accepts exact `{rawText,targetLanguage,resumeId}` and returns `202 + TargetJobWithJob` with `jobType=target_import` and `status=queued`.
- Replaying the same idempotency key returns the same TargetJob and does not add a second target row.
- The only persisted JD text fact is `target_jobs.raw_jd_text`; import/job/outbox payloads omit the raw text, `sourceType`, and `sourceUrl`.
- The parse executor writes requirements, summary provenance, fitSummary provenance, and `target.parsed` without a source refresh job.
- `listTargetJobs` and `getTargetJob` expose the ready TargetJob for the same user only.
- `updateTargetJob` accepts a legal status transition and preserves `analysisStatus=ready`.
