# frontend/src/api

Fixture-backed API entrypoint for frontend workstreams.

## Mock Runtime Handoff

- Use `EasyInterviewClient` from `./generated/client` with `createFixtureBackedFetch(...)` from `./mockTransport`.
- Production bootstrap imports `createAppClient(...)` from `./clientFactory`; Vite dev defaults to fixture-backed mock mode so pages render without a real backend.
- To hit the real backend during dev, run `VITE_EI_API_MODE=real pnpm --filter @easyinterview/frontend dev`. The default dev real API base is `http://localhost:8080/api/v1`; override it with `VITE_EI_API_BASE_URL=<full-api-base>`.
- Build the registry with `createFixtureRegistry([...])`, using JSON files from `openapi/fixtures/<tag>/<operationId>.json`.
- Scenario selection is explicit: pass `Prefer: example=<scenario>` through request headers. Missing scenarios fail loudly instead of falling back to `default`.
- Seed profiles currently available for frontend shell and follow-on D2-D6 workstreams:
  - `getMe`: `default`, `authenticated`, `unauthenticated`, `prototype-baseline`
  - `getPracticeSession`: `default`, `missing-session`, `prototype-baseline`
  - `createPracticePlan`: `default`, `missing-resume`
  - `getFeedbackReport`: `default`, `report-generating`, `prototype-baseline`
  - `requestPrivacyDelete`: `default`, `privacy-delete-requested`

## Blockers

- Do not import `ui-design/src/data.jsx` from `frontend/src`; prototype data is only an upstream source for B2 fixture projection.
- If a flow needs a new mock state, add a named scenario under `openapi/fixtures/` and run `make lint-mock-contract`.
- If an operation is missing, fix the B2 OpenAPI/fixture truth source first; do not hardcode response data in frontend code.
- Do not rely on relative `/api/v1` for local real-backend dev. In the browser it resolves to the Vite origin (`localhost:5173`), not the Go backend (`localhost:8080`).
