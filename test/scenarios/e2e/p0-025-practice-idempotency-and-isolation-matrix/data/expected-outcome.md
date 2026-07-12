# Expected Outcome

- Replaying a completed message pair returns the original messages without another AI call.
- Retrying a pending user message reuses the original row and its preceding history.
- Reusing a client message ID with different text returns `409 PRACTICE_SESSION_CONFLICT`.
- Cross-user session access returns `404 PRACTICE_SESSION_NOT_FOUND`.
- Every focused gate must execute named tests; `no tests to run` is rejected.
