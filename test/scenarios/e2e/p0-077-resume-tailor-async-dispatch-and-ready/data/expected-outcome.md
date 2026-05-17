# E2E.P0.077 Expected Outcome

## Phase 5 Dispatch Outcomes

- `POST /api/v1/resume-versions` with `seedStrategy=ai_select` returns 202 and `BranchResumeVersionAccepted`.
- The provisional targeted `resume_versions` row is committed.
- `resume_tailor_runs.status='queued'` and `mode='gap_review'`.
- `async_jobs.status='queued'`, `job_type='resume_tailor'`, and `resource_type='resume_tailor_run'`.
- If async job insertion fails, the provisional version and tailor run are rolled back.
- Async job payload contains only IDs / mode / seed strategy and not raw profile or suggestion text.

## Deferred Outcomes

- Phase 6 adds requestTailor and getTailorRun endpoint assertions.
- Phase 7 adds drainer, ready suggestions, `ai_task_runs`, and outbox assertions.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Fixture parity covers `branchResumeVersion` `ai-select-202-with-job`.
- Live DB integration covers queued run/job dispatch and rollback.
- Trigger log contains no skipped or no-op focused gates.
- Retired vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
