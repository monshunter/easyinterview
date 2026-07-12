# Expected outcome

Provider failure returns the typed AI error. Missing resume evidence returns `VALIDATION_FAILED` with zero AI/opening message. A valid retry succeeds with one running conversation and no duplicate opening message.
