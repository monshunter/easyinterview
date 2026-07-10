# Expected Outcome

- `resume_parse` moves the resume asset to ready.
- `target_import` moves the target job to `analysisStatus=ready` and writes
  requirements.
- `createPracticePlan(baseline)` creates a ready plan bound to both target job
  and resume asset.
- `startPracticeSession` creates a running session and first asked turn.
- `appendSessionEvent` records an answer event and advances the event loop.
- `completePracticeSession` queues `report_generate`.
- `report_generate` creates a ready feedback report with `next_round` in
  `nextActions`.
- `createPracticePlan(next_round, sourceReportId)` creates a different ready
  derived plan.
- Replayed idempotency keys do not create duplicate sessions, reports, outbox
  events, or derived plans.
- Empty `focusCompetencyCodes` is accepted and stored as cardinality `0`.
- Trigger logs contain concrete Go test pass markers and no skip/no-test marker.
- Trigger logs and observable payload checks do not expose private JD text,
  answer text, report prose, prompt body, response body, or provider secrets.
- API-level core loop artifacts do not contain out-of-scope `Debriefs` / `Profile`
  operations, `candidate_profiles` / `experience_cards` tables,
  `source_debrief_id`, `debrief_generate`, or debrief event tokens.
