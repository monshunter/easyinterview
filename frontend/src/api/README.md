# frontend/src/api

Fixture-backed API entrypoint for frontend workstreams.

## Mock Runtime Handoff

- Use `EasyInterviewClient` from `./generated/client` with `createFixtureBackedFetch(...)` from `./mockTransport`.
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
