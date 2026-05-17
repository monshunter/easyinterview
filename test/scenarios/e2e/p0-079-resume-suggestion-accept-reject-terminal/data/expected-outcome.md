# E2E.P0.079 Expected Outcome

- `acceptResumeTailorSuggestion` returns `200` with the target suggestion status set to `accepted` and `decidedAt` populated.
- `rejectResumeTailorSuggestion` returns `200` with the target suggestion status set to `rejected` and `decidedAt` populated.
- Reusing the same `Idempotency-Key` returns replay output without calling the service again.
- Re-deciding an accepted or rejected suggestion returns `409` with `VALIDATION_FAILED` and `details.reason=SUGGESTION_ALREADY_DECIDED`.
- Cross-user suggestion or version access returns 404.
- `resume_versions.structured_profile` remains JSON-equal before and after accept/reject.
