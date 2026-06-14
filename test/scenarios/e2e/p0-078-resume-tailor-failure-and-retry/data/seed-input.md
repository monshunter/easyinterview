# E2E.P0.078 Seed Input

- User A owns one ready flat resume and one ready target job.
- Three resume tailor async jobs are queued for the same resume / target:
  - M1: AI provider timeout.
  - M2: invalid AI JSON output.
  - M3: AI provider timeout followed by retry success.
- The F3 `resume.tailor.gap_review` prompt and rubric entries are active.
- The A3 AIClient stub emits deterministic timeout, invalid, and success responses.
