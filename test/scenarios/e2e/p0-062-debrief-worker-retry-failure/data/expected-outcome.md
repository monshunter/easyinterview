# Expected Outcome

- failure outcomes include B1 AI error codes
- retryable outcome requeues with backoff before max attempts
- max-attempt outcome is non-retryable and finalizes as failed
- no completed debrief update occurs on failure
