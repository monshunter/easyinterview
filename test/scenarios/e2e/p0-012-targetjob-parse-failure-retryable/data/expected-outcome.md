# Expected Outcome

- Timeout and fallback exhaustion are retryable.
- Invalid output, unsupported capability, missing secret, invalid config, and invalid source are non-retryable.
- `target.analysis.failed` is emitted for each failure class.
- Failed jobs preserve source records for user-initiated re-import.
- Failure evidence does not leak prompt, response, provider secret, raw JD text, or authorization headers.
