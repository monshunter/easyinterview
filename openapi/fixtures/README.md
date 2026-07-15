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
    └── <operationId>.json        # one fixture per operation (37 in the current contract)
```

The 10 tag directories follow
[spec §2.1](../../docs/spec/openapi-v1-contract/spec.md#2-范围) declaration
order; the 37 operationIds are listed in
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
- `request` is present on operations declaring a `requestBody`; the `body`
  field is required in that case. Header-only idempotent operations such as
  `deleteMe` may declare `request.headers` without a body.
- `response.status` must be an integer status code declared by the operation,
  or a 4xx/5xx covered by the operation's `default` response. The P0 export
  exceptions are `POST /api/v1/privacy/exports` and
  `POST /api/v1/resumes/{resumeId}/exports`, locked to `501`
  with `PRIVACY_EXPORT_NOT_AVAILABLE` / `RESUME_EXPORT_NOT_AVAILABLE`.
- `response.body` must be schema-valid against the operation's request /
  response schema in `openapi.yaml`. The validator runs both the `default`
  scenario and any extra scenarios.

## Scenario naming

| Scenario name | Owner | Source | Required? |
|---------------|-------|--------|-----------|
| `default` | this directory | hand-authored | ✅ every fixture |
| `prototype-baseline` | this directory | `make sync-fixtures-from-prototype` from `ui-design/src/data.jsx` | optional; required for the 6 P0 closed-loop endpoints listed in [PROTOTYPE_MAPPING.md §3](./PROTOTYPE_MAPPING.md#3-p0-闭环关键-endpoint-覆盖plan-24-自检) |
| `<purpose>-<variant>` | the consumer adding it | hand-authored | optional; e.g. `error-conflict`, `boundary-empty-list`, `slow-network` |

The first key in the JSON map (`default`) is canonical. Any additional
scenarios may follow it in arbitrary order, but `prototype-baseline` (when
present) is rendered second so consumers can find the prototype-derived
flavour without scanning the rest.

`make sync-fixtures-from-prototype` is the **only** way the
`prototype-baseline` scenario is produced. Hand edits to that scenario will be
overwritten the next time the sync tool runs; if the prototype data is wrong,
fix it in `ui-design/src/data.jsx` and re-run the sync.

## Consumer scenario-selection contract

Mock-server consumers must implement the following selection rules:

1. If the request carries `Prefer: example=<scenario>` (or the equivalent
   query / header understood by the mock), use that scenario when present.
2. Otherwise, fall back to `scenarios.default`.
3. If a request asks for a scenario the fixture does not declare, mock servers
   must **fail loudly** rather than silently fall back to `default`. Tests
   that intentionally exercise scenario fallback should assert against
   `default` directly.

Frontend `msw` and the local backend mock server workstream
([engineering-roadmap S1](../../docs/spec/engineering-roadmap/spec.md#62-s1--contract-backed-mock-runway))
must consume `openapi/fixtures/` as the same source. **Hardcoding mock responses
inside the frontend is forbidden** per spec §2.1 / §4.7 / §5 — if a consumer
needs a new variant, it adds a new scenario here and re-runs the consumer's
mock to pick it up.

## Local commands

| Target | Purpose |
|--------|---------|
| `make validate-fixtures` | Schema-validates every scenario against `openapi.yaml`, enforces AI-schema provenance, runs the privacy allowlist + UUIDv7 + `tmp_` scans, and verifies all 37 operationIds are present. |
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
  prototype data.
- `provenance.modelId` uses provider-neutral fixture ids. Generic contract
  fixtures may use `model-profile:contract.default`; fixtures that model an
  actual A3 resolved model id use `fixture-model:<name>`. Fixtures must not
  hardcode vendor or concrete model names.
- All `format: uuid` values are UUIDv7 literals; any `tmp_*` prefix is
  forbidden.

If a new reserved synthetic-value pattern is added, update both this README and
`scripts/lint/validate_fixtures.py`'s allowlist tables.

## Prism smoke (B2 002 Phase 3)

Phase 3.1 of plan `002-fixtures-and-mock-source` projects every fixture's
`scenarios.default.response.body` into named OpenAPI examples at
`openapi/.generated/openapi-with-fixtures.yaml`. That derived file is the
**only** input to the Prism smoke; do **not** point Prism at `openapi.yaml`
directly (it carries no examples by design — fixtures own them).

Refresh the projection and start Prism on port 4010:

```sh
make render-openapi-fixture-examples
npx @stoplight/prism-cli mock openapi/.generated/openapi-with-fixtures.yaml -p 4010
```

The local verification used `@stoplight/prism-cli@5.14.2` against Node
v23.10.0. Prism does not enforce the `sessionCookie` security scheme as a
hard gate, but it returns `401 Unauthorized` with the documented error
envelope when a cookie is missing — the smoke calls below therefore include
`Cookie: ei_session=fake` to exercise the success branch.

### Live smoke matrix (14 fixed operations)

The repeatable smoke matches all selected defaults byte-for-byte. It retains
the original read/handoff checks and also covers the Resume summary-list/full-detail
split, report overview, and every TargetJob default affected by removal of `latestReportId`. Status codes
other than `200` use `Prefer: code=<status>, example=default` so Prism selects
the declared response.

| operationId | Method/path | Expected |
|-------------|-------------|----------|
| `getMe` | `GET /me` | 200 |
| `listResumes` | `GET /resumes` | 200 |
| `getResume` | `GET /resumes/{resumeId}` | 200 |
| `listTargetJobs` | `GET /targets` | 200 |
| `getTargetJob` | `GET /targets/{targetJobId}` | 200 |
| `importTargetJob` | `POST /targets/import` | 202 |
| `updateTargetJob` | `PATCH /targets/{targetJobId}` | 200 |
| `archiveTargetJob` | `POST /targets/{targetJobId}/archive` | 202 |
| `getPracticeSession` | `GET /practice/sessions/{sessionId}` | 200 |
| `getFeedbackReport` | `GET /reports/{reportId}` | 200 |
| `getReportConversation` | `GET /reports/{reportId}/conversation` | 200 |
| `listTargetJobReports` | `GET /targets/{targetJobId}/reports` | 200 |
| `createPracticePlan` | `POST /practice/plans` | 201 |
| `requestPrivacyExport` | `POST /privacy/exports` | 501 |

Run the matrix with:

```sh
python3 scripts/codegen/prism_fixture_smoke.py
```

For each call, the response body must match the fixture's
`scenarios.default.response.body` byte-for-byte (jq diff against
`openapi/fixtures/<tag>/<operationId>.json#/scenarios/default/response/body`).

A repeatable verifier lives at
[`scripts/codegen/prism_fixture_smoke.py`](../../scripts/codegen/prism_fixture_smoke.py)
— start Prism, then run `python3 scripts/codegen/prism_fixture_smoke.py` from
the repo root. The verifier exits 0 when every byte-equal check passes.

### Out of scope here

The plan **does not** stand up a long-running mock server — that belongs to the
mock contract workstream tracked by
[engineering-roadmap S1](../../docs/spec/engineering-roadmap/spec.md#62-s1--contract-backed-mock-runway).
Phase 3 of B2 002 only proves the fixtures→OpenAPI examples→Prism response
loop is byte-stable; the future mock workstream will pick up the same
`openapi/fixtures/` and `openapi/.generated/openapi-with-fixtures.yaml`
artefacts as inputs.
