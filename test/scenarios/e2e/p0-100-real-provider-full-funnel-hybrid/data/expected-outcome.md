# E2E.P0.100 Expected Outcome

The AI Agent executes the scenario preflight first. A human may then complete
the browser journey and attach redacted evidence under the same scenario output
directory.

## Agent-Verified Outcome

- Shared scenario environment is prepared through the top-level env lifecycle.
- The scenario has the standard `setup.sh`, `trigger.sh`, `verify.sh`, and
  `cleanup.sh` contract.
- `deploy/dev-stack/.env` is the single real local environment source; account
  material, JD/resume/answer inputs, and checklist are present.
- Mock/stub completion paths are rejected as evidence.
- If `deploy/dev-stack/.env` values or browser evidence are missing, `verify.sh` writes
  `result=MANUAL_REQUIRED` instead of passing the full journey.

## Human Or Browser-Agent Outcome

- Frontend runs with `VITE_EI_API_MODE=real`.
- Backend runs with `APP_ENV=dev` and a real OpenAI-compatible provider.
- Login uses `manual-uat-full-funnel@example.test` through Mailpit.
- Home -> Parse -> Workspace -> Practice -> Generating -> Report -> Next Round
  completes against real backend handlers, PostgreSQL, and AI provider calls.
- Evidence records provider/profile/model/latency/task-run count only, never
  prompt/response bodies, JD text, answer text, report prose, token values, or
  session cookies.
