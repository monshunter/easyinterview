# Expected Outcome

- `createPracticePlan` returns `201 + PracticePlan` with `status=ready`, `goal=baseline`, and the expected `targetJobId`.
- Repeating the same request with the same `Idempotency-Key` replays the first response and leaves plan / audit side effects at one row each.
- `getPracticePlan` returns the created plan for User A.
- User B receives `404 + PRACTICE_PLAN_NOT_FOUND` for the same `planId`.
- Audit metadata includes only `plan_id`, `goal`, `mode`, `language`, and `target_job_id`; it excludes question / answer / hint / prompt / response text.
- PracticeMode generated/runtime code has zero `legacy debrief replay value` or `debrief_replay` literal.
