# P0.085 Seed Input

- `Resumes/branchResumeVersion.json ai-select-202-with-job` (job.resourceId becomes tailorRunId).
- `ResumeTailor/requestResumeTailor.json` `default` / `idempotency-replay` (both carry IK header).
- `ResumeTailor/getResumeTailorRun.json` four status variants: `default(ready) / queued / generating / failed`.
- `Resumes/getResumeVersion.json targeted-with-suggestions` (1+ suggestions for refetch verification).
- Authenticated user; deterministic fake-timer harness (no real wall clock).
