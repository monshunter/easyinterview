# Fixture-backed Mock Runtime Checklist

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## 1 Fixture registry and coverage

- [x] 1.1 Registry reads `openapi/fixtures/<tag>/<operationId>.json` and exposes operationId lookup for all current 35 operations.
- [x] 1.2 Registry tests fail on missing fixture, extra fixture, unexpected tag directory, mismatched operationId and non-current mock/API token.
- [x] 1.3 `make lint-mock-contract` includes `make validate-fixtures`, `make lint-openapi`, fixture registry tests and mock runtime boundary lint.

## 2 Frontend mock transport and dev client

- [x] 2.1 `createFixtureBackedFetch` returns typed generated-client responses from fixtures.
- [x] 2.2 Frontend mock transport supports named scenario selection, unknown scenario failure, mock delay and abort behavior.
- [x] 2.3 `createDevMockClient` covers every generated operationId, returns fixture-backed dev preview data, models auth session state and handles declared export fallback responses.
- [x] 2.4 `createAppClient` defaults Vite dev to fixture-backed mode, keeps production on `/api/v1`, and requires `VITE_EI_API_BASE_URL` when dev opts into real API mode.

## 3 Backend mockruntime

- [x] 3.1 Backend mockruntime maps HTTP method/path to operationId and returns the same status/body as the selected fixture scenario.
- [x] 3.2 Named seed scenarios cover unauthenticated, authenticated, missing session, missing resume, report generating and privacy delete request states.
- [x] 3.3 Unknown scenario requests fail loudly and do not silently fall back to `default`.

## 4 Boundary lint

- [x] 4.1 Frontend runtime code cannot import `ui-design/src/data.jsx`.
- [x] 4.2 Fixture response bodies cannot contain prototype-only display fields.
- [x] 4.3 `openapi/fixtures/` tag directories must exactly match current OpenAPI tags.
- [x] 4.4 Current mock/API token scan rejects non-current routes, tags, schema keys and config paths while allowing practice-voice owner operation names and fixtures.

## 5 Current owner compression gate

- [x] 5.1 `spec.md`, `plan.md`, `checklist.md`, `context.yaml` and plans INDEX align to the current 35-operation fixture-backed mock runtime contract.
  <!-- verified: 2026-07-07 method=current-owner-compression evidence="Updated mock-contract-suite spec.md to v1.13, plan.md/checklist.md to v1.7, context specVersion to v1.13, and synced docs/spec plus mock-contract-suite plans INDEX. PASS: targeted stale-wording grep returned no matches except current completeMyProfile operation when broad Profile token is included; validate_context.py mock-contract-suite/001 tooling PASS; make lint-mock-contract PASS; python3 -m pytest scripts/lint/mock_runtime_boundary_test.py -q PASS (7 tests); go test ./backend/internal/api/mockruntime -count=1 PASS; pnpm --filter @easyinterview/frontend test src/api/mockTransport.test.ts src/api/devMockClient.test.ts src/api/clientFactory.test.ts PASS (13 tests); make codegen-check PASS." -->
