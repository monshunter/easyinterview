# Seed Input

- User email: `full-funnel-journey@example.com`
- Backend seed:
  - authenticated session cookie emitted in `state.json`
  - ready resume asset created through `registerResume`
  - resume readiness produced by the real `resume_parse` runner
- UI input:
  - paste JD text on Home
  - answer text in Practice
- Runtime:
  - `VITE_EI_API_MODE=real`
  - `VITE_EI_API_BASE_URL=http://127.0.0.1:18099/api/v1`
  - Playwright browser base URL `http://127.0.0.1:4174`

No frontend mock transport is used for the scenario journey.
