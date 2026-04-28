# openapi/fixtures

Single source of truth for HTTP-level mock responses (and the matching request
shapes) of every operation in `openapi/openapi.yaml`. Owner subspec:
[openapi-v1-contract](../../docs/spec/openapi-v1-contract/spec.md) (B2 in the
engineering roadmap; spec §2.1 / §4.7 lock the fixtures contract).

`openapi/fixtures/<tag>/<operationId>.json` is the **only** place where a
canonical example response for an operation may live. The OpenAPI document
itself does not carry hand-written `examples`; Phase 3 of plan `002-fixtures-and-mock-source`
projects fixtures into a derived `openapi/.generated/openapi-with-fixtures.yaml`
for Prism / docs-site consumption.

## Layout

```
openapi/fixtures/
├── README.md
├── PROTOTYPE_MAPPING.md          # data.jsx ↔ operationId mapping table
└── <Tag>/
    └── <operationId>.json        # one fixture per operation (36 in v1.0.0)
```

The 14 tag directories follow
[spec §2.1](../../docs/spec/openapi-v1-contract/spec.md#2-范围) declaration
order; the 36 operationIds are listed in
[spec §3.1.1](../../docs/spec/openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表).

## File shape

Each fixture is JSON with the following structure:

```json
{
  "operationId": "getMe",
  "scenarios": {
    "default": {
      "request": { "headers": {}, "body": {} },
      "response": { "status": 200, "headers": {}, "body": {} }
    },
    "prototype-baseline": {
      "response": { "status": 200, "headers": {}, "body": {} }
    }
  }
}
```

Rules:

- `operationId` **must** equal the filename stem.
- `scenarios.default` is required; it must be the **first** key in `scenarios`.
- `request` is present only on operations declaring a `requestBody`; the
  `body` field is required when `request` is present.
- `response.status` must be an integer status code declared by the operation,
  or a 4xx/5xx covered by the operation's `default` response. The single P0
  exception is `POST /api/v1/privacy/exports`, which is locked to `501` with
  `error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`.
- `response.body` must be schema-valid against the operation's request /
  response schema in `openapi.yaml`. The validator runs both the `default`
  scenario and any extra scenarios.

## Scenario naming

| Scenario name | Owner | Source | Required? |
|---------------|-------|--------|-----------|
| `default` | this directory | hand-authored | ✅ every fixture |
| `prototype-baseline` | this directory | `make sync-fixtures-from-prototype` from `easyinterview-ui/src/data.jsx` | optional; required for the 8 P0 closed-loop endpoints listed in [PROTOTYPE_MAPPING.md §3](./PROTOTYPE_MAPPING.md#3-p0-闭环关键-endpoint-覆盖plan-24-自检) |
| `<purpose>-<variant>` | the consumer adding it | hand-authored | optional; e.g. `error-conflict`, `boundary-empty-list`, `slow-network` |

The first key in the JSON map (`default`) is canonical. Any additional
scenarios may follow it in arbitrary order, but `prototype-baseline` (when
present) is rendered second so consumers can find the prototype-derived
flavour without scanning the rest.

`make sync-fixtures-from-prototype` is the **only** way the
`prototype-baseline` scenario is produced. Hand edits to that scenario will be
overwritten the next time the sync tool runs; if the prototype data is wrong,
fix it in `easyinterview-ui/src/data.jsx` and re-run the sync.

## Consumer scenario-selection contract

Mock-server consumers must implement the following selection rules:

1. If the request carries `Prefer: example=<scenario>` (or the equivalent
   query / header understood by the mock), use that scenario when present.
2. Otherwise, fall back to `scenarios.default`.
3. If a request asks for a scenario the fixture does not declare, mock servers
   must **fail loudly** rather than silently fall back to `default`. Tests
   that intentionally exercise scenario fallback should assert against
   `default` directly.

Frontend `msw` and the local backend mock server (E1
[`mock-contract-suite`](../../docs/spec/engineering-roadmap/spec.md#55-layer-e--integration4-份))
must consume `openapi/fixtures/` as the same source. **Hardcoding mock responses
inside the frontend is forbidden** per spec §2.1 / §4.7 / §5 — if a consumer
needs a new variant, it adds a new scenario here and re-runs the consumer's
mock to pick it up.

## Local commands

| Target | Purpose |
|--------|---------|
| `make validate-fixtures` | Schema-validates every scenario against `openapi.yaml`, enforces AI-schema provenance, runs the privacy allowlist + UUIDv7 + `tmp_` scans, and verifies all 36 operationIds are present. |
| `make sync-fixtures-from-prototype` | Re-renders the `prototype-baseline` scenario of every supported fixture from `data.jsx`. Idempotent — `git diff --exit-code -- openapi/fixtures` stays clean across re-runs. |

The two are independent; the sync tool calls `validate-fixtures` internally as
its final gate but the validate gate also runs standalone for hand-edited
fixtures.

## Privacy posture

Per spec §4.7, fixtures **must not** contain real personal information:

- Email addresses use `example.com` / `example.org` / `example.net` or any
  `*.example` reserved domain (per RFC 6761).
- Phone numbers use the reserved range `+1-555-0100`..`+1-555-0199`.
- Employer names are kept generic (`Acme`, `Lumen Labs`, …); the validator's
  blacklist rejects real-employer brand names that have appeared in the
  prototype data. AI-vendor names (`anthropic`, `openai`, …) only appear in
  `provenance.modelId` infrastructure metadata and are exempt from the
  employer blacklist.
- All `format: uuid` values are UUIDv7 literals; any `tmp_*` prefix is
  forbidden.

If a new placeholder pattern is added, update both this README and
`scripts/lint/validate_fixtures.py`'s allowlist tables.

## Prism smoke (B2 002 Phase 3)

Once Phase 3.1 of plan `002-fixtures-and-mock-source` lands the OpenAPI
examples projection, fixtures can be served locally through Prism:

```sh
npx @stoplight/prism-cli mock openapi/.generated/openapi-with-fixtures.yaml -p 4010
```

Phase 3.2 of the plan documents the curl smoke matrix used to confirm
byte-equal parity between the fixture's `default` response body and the Prism
response. Phase 3 has not landed yet — this section will be updated when it
does.
