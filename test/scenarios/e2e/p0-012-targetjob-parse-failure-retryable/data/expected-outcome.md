# Expected Outcome

- Provider timeout is retryable.
- Invalid output, disabled / unsupported registry resolution, and missing provider secret are non-retryable.
- `target.analysis.failed` is emitted for each failure class.
- Failed target jobs are deleted with requirement rows cascade-deleted; no source row, file reference, or refresh job exists.
- `GET /targets/{id}` returns 404 and `GET /targets` does not list the failed job.
- Failure evidence does not leak prompt, response, provider secret, raw JD text, or authorization headers.
