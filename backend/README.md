# backend

Go 后端服务根目录：HTTP API、领域模块（auth / upload / target-job / practice / review / resume / async-runtime）与共享基础设施落点。

Current truth sources: [openapi-v1-contract](../docs/spec/openapi-v1-contract/spec.md)、[event-and-outbox-contract](../docs/spec/event-and-outbox-contract/spec.md)、[db-migrations-baseline](../docs/spec/db-migrations-baseline/spec.md)、[secrets-and-config](../docs/spec/secrets-and-config/spec.md) 与 [shared-conventions-codified](../docs/spec/shared-conventions-codified/spec.md)。后端 domain workstream 进入实现时按 [engineering-roadmap S2](../docs/spec/engineering-roadmap/spec.md#63-s2--backend-domain-implementation) 创建对应 child spec / plan。

## 1 Frontend / Backend Contract Boundary

Cross-layer development follows [docs/development.md §2](../docs/development.md#2-frontend--backend-contract-workflow). Backend implementation must consume the same `operationId`, generated DTO/envelope, fixtures, shared error codes, and scenario IDs that frontend and contract plans use.

Backend owners must keep an operation matrix for each new or materially changed API:

| Field | Backend expectation |
|-------|---------------------|
| `operationId` | Matches `openapi/openapi.yaml` and generated artifacts |
| `fixture` | Existing fixture scenarios used for frontend/scenario development |
| `frontend consumer` | Known screen/hook/client consumer, or `none yet` |
| `backend handler` | Real handler/store/job path, or explicit `not-yet-implemented` |
| `persistence` | Migration/table/object/queue path, or `none` |
| `AI dependency` | AI profile/capability, `stub/fixture only`, or `none` |
| `scenario coverage` | Scenario ID or substitute gate |

`backend/cmd/api/main.go` currently wires real auth, `/me`, privacy delete handoff, and `/runtime-config` routes. Other OpenAPI operations may exist as contract/fixture/generated-client surface before their real handlers land; do not treat fixture availability as backend implementation evidence.

## 2 Local Implementation Gates

- API or DTO changes start in `openapi/openapi.yaml`, then `openapi/fixtures/`, then `make codegen && make codegen-check`.
- Persistent behavior requires migrations plus `make migrate-check` when schema changes.
- Handler work must cover success, failure, auth/session, idempotency where applicable, privacy red lines, and generated error envelope shape.
- AI-enabled backend code must route through `internal/ai/aiclient`; unit tests use stub/fixture only, while dev/Kind/staging/prod must fail fast without real provider config.
- Before handoff to frontend integration, run focused Go tests and the relevant repo-level gates from [docs/development.md §1](../docs/development.md#1-the-5-local-gates).
