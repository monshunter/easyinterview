# P0.086 Seed Input

- `Resumes/getResumeVersion.json targeted-with-suggestions` (>=1 pending suggestion id `0195f2d0-...0302`).
- `Resumes/acceptResumeTailorSuggestion.json` 3 scenarios (default 200, idempotency-replay, already-decided-409 with details.reason=SUGGESTION_ALREADY_DECIDED).
- `Resumes/rejectResumeTailorSuggestion.json` 3 scenarios (matching shape).
- `Resumes/updateResumeVersion.json` 3 scenarios (default, idempotency-replay, validation-error-422).
- User A authenticated; User B authenticated (cross-user).
