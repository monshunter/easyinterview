# Fixture-backed Mock Runtime Checklist

> **版本**: 1.11
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## 1 Fixture registry and coverage

- [x] 1.1 Registry reads `openapi/fixtures/<tag>/<operationId>.json` and exposes operationId lookup for all current 37 operations.<!-- verified: 2026-07-10 method=python-unittest tests=scripts.mock_contract.fixture_registry_test,frontend devMockClient.test.ts -->
- [x] 1.2 Registry tests fail on missing fixture, extra fixture, unexpected tag directory, mismatched operationId and out-of-scope mock/API token.
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
- [x] 4.4 Current mock/API token scan rejects out-of-scope routes, tags, schema keys and config paths while allowing practice-voice owner operation names and fixtures.

## 5 Current owner compression gate

- [x] 5.1 `spec.md`, `plan.md`, `checklist.md`, `context.yaml` and plans INDEX align to the current 37-operation fixture-backed mock runtime contract.<!-- verified: 2026-07-10 method=targeted-doc-update+context-validation -->
  <!-- verified: 2026-07-10 method=current-inventory-reconcile evidence="Updated mock-contract-suite spec.md to v1.14, plan.md/checklist.md to v1.8, context specVersion to v1.14, added archiveTargetJob to apiNames, and synced docs/spec plus mock-contract-suite plans INDEX. PASS: validate_context.py mock-contract-suite/001 tooling; make lint-mock-contract; python3 -m pytest scripts/lint/openapi_diff_test.py scripts/lint/openapi_inventory_test.py scripts/mock_contract/fixture_registry_test.py -q; codegen-openapi generated diff hash unchanged before/after rerun." -->

## 6 Backend mockruntime test helper cleanup

- [x] 6.1 删除 `mockruntime_test.go` 中无调用的 `assertJSONField` / `lookupJSONPath` helper，保留完整 response parity 与 unknown-scenario 断言；验证: scoped staticcheck、`go test ./internal/api/mockruntime -count=1`、`make lint-mock-contract` 与 owner docs gates 通过。
  <!-- verified: 2026-07-10 method=mockruntime-dead-test-helper-removal evidence="RED: backend staticcheck reported U1000 for assertJSONField and its only callee lookupJSONPath. GREEN: removed both helpers; staticcheck ./internal/api/mockruntime/... and go test ./internal/api/mockruntime -count=1 PASS; make lint-mock-contract validates 37 fixtures, 10 tags, 37 operations and mock runtime boundary; owner contexts and pruning gate PASS." -->

## 7 Frontend dev registry observation cleanup

- [x] 7.1 RED/GREEN: dev mock source gate detects and then rejects the production `getDevMockFixtureOperationIds` test observer.<!-- verified: 2026-07-10 method=vitest-red-green evidence="RED failed only on the observer while five behavior tests passed; GREEN passed all 6 devMockClient tests after deletion." -->
- [x] 7.2 Operation parity compares real registry keys with generated `ALL_OPERATION_IDS`; focused frontend mock tests, typecheck and `make lint-mock-contract` pass.<!-- verified: 2026-07-10 method=vitest+typecheck+lint-mock-contract evidence="Frontend mock passed 3 files/14 tests; typecheck passed; lint validated 37 fixtures, 10 tags, 37 operations and boundary tests." -->
