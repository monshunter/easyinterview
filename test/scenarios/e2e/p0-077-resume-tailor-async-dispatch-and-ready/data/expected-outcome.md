# E2E.P0.077 Expected Outcome

## Request And Read Outcomes

- `POST /api/v1/resume/tailor` requires `Idempotency-Key`, accepts only `gap_review` / `bullet_suggestions`, and returns 202 with `ResumeTailorRunWithJob`.
- RequestTailor creates a queued `async_jobs(job_type='resume_tailor')` row with `resource_type='resume_tailor_run'` and a `payload.resumeId` owner scope.
- `GET /api/v1/resume/tailor-runs/{tailorRunId}` returns queued, generating, ready, and failed status variants from async job status/result, scoped to the authenticated user through the flat `resumes` table.
- Store transitions support queued or retryable failed async jobs to generating, generating to ready, and generating to failed; concurrent claim from queued allows exactly one successful transition.
- Fixture parity covers `requestResumeTailor` `default` / `idempotency-replay` and `getResumeTailorRun` `default` / `queued` / `generating` / `failed`.

## Ready Drainer Outcomes

- `cmd/api` registers an in-process `resume_tailor` drainer handler.
- `RunOnce(resume_tailor)` calls the A3 AIClient through the F3 `resume.tailor.gap_review` feature key.
- A successful AI response writes match summary, provenance, and three
  ephemeral suggestions into `async_jobs.result` for the tailor run.
- A typed `ai_task_runs` row is written for `task_type='resume_tailor'`, `resource_type='resume_tailor_run'`, and `output_schema_version='resume.tailor.v1'`.
- Exactly one `resume.tailor.completed` outbox event is written on ready; its
  payload only contains `tailorRunId`, `resumeId`, `targetJobId`, `mode`, and
  `status`.

## Evidence Outcomes

- `method=cmd-api-http` appears in verify output.
- Fixture parity covers `requestResumeTailor` and `getResumeTailorRun`.
- Store/job tests cover queued async job dispatch, state transitions, concurrent
  claim, rollback, ready suggestions in task output, and ready-only outbox.
- Trigger log contains no skipped or no-op focused gates.
- Out-of-scope vocabulary and privacy negative greps return zero hits for backend resume implementation and scenario evidence.
