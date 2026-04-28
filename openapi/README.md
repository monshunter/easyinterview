# openapi

`openapi/` holds the v1.0.0 HTTP contract truth source for easyinterview and
the codegen pipeline that materialises it into Go and TypeScript artefacts.

Owner subspec: [openapi-v1-contract](../docs/spec/openapi-v1-contract/spec.md)
(B2 in the engineering roadmap; v1.0.0 freeze is scoped by that spec §3.1).

## Layout

```
openapi/
├── openapi.yaml              # single-root OpenAPI 3.1 contract (hand-authored)
├── README.md                 # this file
└── templates/
    ├── go/{types,server,spec}.tmpl   # Go renderer templates
    └── ts/{types,client,spec}.tmpl   # TS renderer templates
```

`openapi.yaml` is the only authoritative HTTP contract document. Generated
DTOs / clients live under `backend/internal/api/generated/` and
`frontend/src/api/generated/` and **must never be edited by hand** — they are
regenerated whenever `openapi.yaml` changes.

## Codegen pipeline

```
shared/conventions.yaml ─┐
                         ▼
openapi/openapi.yaml ──► backend/cmd/codegen/openapi
                         │
                         ├─► backend/internal/api/generated/{types,server,spec}.gen.go
                         └─► frontend/src/api/generated/{types,client,spec}.ts
```

Two truth sources feed the contract:

- `shared/conventions.yaml` (B1, owned by
  [shared-conventions-codified](../docs/spec/shared-conventions-codified/spec.md))
  — enums, error codes, `ApiError` / `PageInfo` / `JobStatus`. The codegen
  re-renders the `# === B1-AUTO-START ... # === B1-AUTO-END ===` block of
  `openapi.yaml` from this file, so any B1 enum value change automatically
  surfaces as drift in `openapi.yaml` (and from there into the Go/TS
  artefacts).
- `openapi/openapi.yaml` itself — paths, operations, request / response
  schemas, the v1.0.0 freeze contract. This file is hand-authored outside
  the B1-AUTO block.

### `make` entry points

| Target            | Purpose |
|-------------------|---------|
| `make codegen-openapi` | Re-render the B1-AUTO block, then regenerate Go and TS artefacts. Idempotent. |
| `make codegen` | Convenience composite: runs `codegen-conventions` first (so B1's libs are fresh), then `codegen-openapi`. |
| `make lint-openapi` | `swagger-cli validate` + `scripts/lint/openapi_inventory.py` against `openapi/openapi.yaml`. |
| `make codegen-check` | Local **drift gate**: `codegen-openapi` + `lint-openapi` + `git diff --exit-code` over `openapi.yaml`, `backend/internal/api/generated/`, and `frontend/src/api/generated/`. |

`make codegen-check` is the **only required** drift gate today. A remote CI
required-check is deferred until A5
[`ci-pipeline-baseline`](../docs/spec/ci-pipeline-baseline/spec.md) signals
that the project is ready to wire CI; this plan does not modify any A5
workflow (per spec §4.5 / §5).

### Running the generator directly

```sh
go run ./backend/cmd/codegen/openapi \
    -openapi openapi/openapi.yaml \
    -conventions shared/conventions.yaml \
    -templates openapi/templates \
    -repo-root .
```

Output is byte-stable across runs — running twice in a row produces no
diffs. The generator copies `openapi.yaml` into
`backend/internal/api/generated/` so `spec.gen.go`'s `//go:embed openapi.yaml`
resolves a sibling file.

## Tag inventory

The 14 OpenAPI tags follow
[spec §2.1](../docs/spec/openapi-v1-contract/spec.md#2-范围) in declaration
order: Auth, Uploads, Profile, Resumes, TargetJobs, PracticePlans,
PracticeSessions, Reports, Mistakes, ResumeTailor, Debriefs, Growth, Jobs,
Privacy. The 36 operations are catalogued in
[spec §3.1.1](../docs/spec/openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表);
`scripts/lint/openapi_inventory.py` enforces tag order, operation enumeration,
default `ApiError` refs, the `Idempotency-Key` mutex (per ADR-Q1 +
`clientEventId`), the unique `POST /privacy/exports` 501 example, and the
`GenerationProvenance` reachability invariant for AI-generated schemas
(spec §4.6).

## `$ref` topology with B1

```
shared/conventions.yaml ─► (codegen) ─► openapi.yaml's B1-AUTO block
                                       (ApiError, ApiErrorCode, PageInfo,
                                        JobStatus, 14 enums)
                                              │
                                              ▼
            openapi.yaml hand-authored schemas use $ref into the same block
                                              │
                                              ▼
            Go/TS codegen aliases B1-owned types to
                backend/internal/shared/{types,errors}/…
                frontend/src/lib/conventions/{enums,errors,pagination,…}
```

Domain schemas (`UserContext`, `TargetJob`, …) become full Go structs / TS
interfaces in the generated artefacts. AI-generated schemas listed in
[spec §4.6](../docs/spec/openapi-v1-contract/spec.md#46-ai-生成结果-provenance-约束)
must remain reachable from `GenerationProvenance` via `$ref`; the inventory
linter rejects any topology that breaks that path.

## Authentication

Per [ADR-Q1](../docs/spec/engineering-roadmap/decisions/ADR-Q1-auth.md), the
contract locks a single P0 security scheme — `sessionCookie` (`type: apiKey`,
`in: cookie`). The `Authorization: Bearer` form is **not** a default scheme;
reintroducing it requires revising both ADR-Q1 and the B2 spec.

## Tooling versions

- Go ≥ 1.24 (see `backend/go.mod`); the generator only depends on
  `gopkg.in/yaml.v3` at runtime.
- Node ≥ 20 (used solely for `npx @apidevtools/swagger-cli` at
  `make lint-openapi`).
- Python ≥ 3.11 for `scripts/lint/openapi_inventory.py`.

If the generator templates change, run `make codegen-check` to confirm the
output stays idempotent.
