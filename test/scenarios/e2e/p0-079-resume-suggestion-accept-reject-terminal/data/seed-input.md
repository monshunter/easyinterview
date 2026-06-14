# E2E.P0.079 Seed Input

- D-20 route catalog with retired resume-version operationIds removed.
- OpenAPI fixtures include:
  - `Resumes/updateResume.json`: `default`, `idempotency-replay`,
    `validation-error-422`
  - `Resumes/duplicateResume.json`: `default`, `idempotency-replay`,
    `validation-error-422`
  - `ResumeTailor/requestResumeTailor.json` and
    `ResumeTailor/getResumeTailorRun.json`
- Frontend fixtures include ephemeral Rewrites suggestions rendered by
  `ResumeRewritesTab`.
