# E2E.P0.079 Seed Input

- Route catalog where out-of-scope resume-version operationIds are absent.
- OpenAPI fixtures include:
  - `Resumes/updateResume.json`: `default`, `idempotency-replay`,
    `validation-error-422`
  - `Resumes/duplicateResume.json`: `default`, `idempotency-replay`,
    `validation-error-422`
  - `ResumeTailor/requestResumeTailor.json` and
    `ResumeTailor/getResumeTailorRun.json`
- Frontend fixtures include a ready flat resume rendered by the read-only
  detail view; out-of-scope Rewrites/Edit route params are negative inputs only.
