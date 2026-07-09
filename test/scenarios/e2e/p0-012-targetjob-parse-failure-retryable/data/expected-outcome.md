# Expected Outcome

- Timeout and fallback exhaustion are retryable.
- Invalid output, unsupported capability, missing secret, invalid config, and invalid source are non-retryable.
- `target.analysis.failed` is emitted for each failure class.
- Failed target jobs are deleted with source / requirement rows cascade-deleted.
- `GET /targets/{id}` returns 404 and `GET /targets` does not list the failed job.
- Failure evidence does not leak prompt, response, provider secret, raw JD text, or authorization headers.
