# backend/internal/api/mockruntime

Fixture-backed HTTP mock runtime for local backend tests and dev harnesses.

## Entry Points

- `LoadRegistry(<repo>/openapi/fixtures)` loads every `openapi/fixtures/<tag>/<operationId>.json` file.
- `NewHandler(registry)` returns an `http.Handler` that matches generated `generated.AllRoutes`.
- Requests may use `/api/v1/...` paths or the generated route path without the prefix.
- Scenario selection uses `Prefer: example=<scenario>`. Unknown scenarios return `400` and do not fall back to `default`.

## Seed Profiles

- `getMe`: `default`, `authenticated`, `unauthenticated`, `prototype-baseline`
- `getPracticeSession`: `default`, `missing-session`, `prototype-baseline`
- `createPracticePlan`: `default`, `missing-resume`
- `getFeedbackReport`: `default`, `report-generating`, `prototype-baseline`
- `requestPrivacyDelete`: `default`, `privacy-delete-requested`

## Blockers

- The only mock data source is `openapi/fixtures`; do not duplicate fixture JSON in backend tests or handlers.
- Add new states as named fixture scenarios, then run `make validate-fixtures` and `make lint-mock-contract`.
- If a route or schema is missing, revise the B2 OpenAPI contract first instead of adding a mock-only endpoint.
