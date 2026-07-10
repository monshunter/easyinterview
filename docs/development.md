# Local Development Workflow

> **Owner**: [`ci-pipeline-baseline/001-local-quality-gates`](./spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md) (A5)
> **Updated**: 2026-07-10

This page is the canonical reference for local development workflow and the 5 local quality gates. The project is currently in **single-developer P0 phase**: there is **no remote CI pipeline**. Required checks, branch protection, and runner secrets are deferred until A5 spec [D-5](./spec/ci-pipeline-baseline/spec.md#31-已锁定决策) trigger conditions land.

## 1 The 5 Local Gates

Run from repo root:

Install Python gate dependencies once in an isolated environment with `python3 -m pip install -r requirements-dev.txt`.

| Command | Purpose | Owner |
|---------|---------|-------|
| `make lint` | B1 conventions + A4 config + A3/F3/E1/runtime-topology local gates + Go lint + frontend typecheck-backed lint | A5 aggregator → B1 / A4 / A3 / F3 / E1 / backend / frontend |
| `make test` | UI prototype Node contract + Python tooling/skill contracts + backend Go unit tests + frontend TypeScript unit tests; AI tests via stub/fixture only | A5 aggregator → product/UI / scripts / skills / backend / frontend owners |
| `make build` | Backend `go build ./cmd/...` + frontend bundle | A5 aggregator → backend / frontend owners |
| `make docs-check` | `sync-doc-index --check` (Header / INDEX drift) + relative-link sanity for `docs/` | A5 aggregator → `/sync-doc-index` skill + A5 `scripts/lint/check_md_links.py` |
| `make codegen-check` | B1 conventions generator + B2 OpenAPI generator + `git diff --exit-code` on generated outputs | A5 aggregator → B1 + B2 |

Each gate exits non-zero on failure. A5 only calls landed local sub-targets; future owner gates are added when the owning implementation exposes a real command, and called sub-target failures are not swallowed.

## 2 Frontend / Backend Contract Workflow

Frontend and backend workstreams are split by contract, not by informal agreement. A feature that crosses UI, API, persistence, or AI behavior must move through these sources in order.

| Step | Source / Gate | Owner Boundary |
|------|---------------|----------------|
| 1. UI truth source | `ui-design/` + `docs/ui-design/`; formal frontend must source-level mirror DOM shape, controls, layout, tokens, and interactions | Frontend design / frontend implementation |
| 2. API contract | `openapi/openapi.yaml`; every frontend/backend API change starts here before generated client/server artifacts are consumed | B2 OpenAPI owner + feature owner |
| 3. Fixture data | `openapi/fixtures/<tag>/<operationId>.json`; mock data is fixture-backed and selected with `Prefer: example=<scenario>` | Contract owner + feature owner |
| 4. Codegen | `make codegen && make codegen-check`; generated Go/TS artifacts are not hand-edited | B1/B2/B3 owners |
| 5. Frontend implementation | `frontend/src/api/generated/client.ts` + `frontend/src/api/mockTransport.ts`; frontend may complete user-visible UI against fixtures before backend handler completion | Frontend owner |
| 6. Backend implementation | Real handler/store/migration/job code implements the same `operationId` and response envelope; tests prove SQL, auth, privacy, idempotency, and error paths | Backend owner |
| 7. Local integration | `make dev-up` provides Docker Compose external dependencies; backend/frontend default to host-managed dev processes unless a component owner explicitly adds an optional compose app service | Backend + frontend owner |
| 8. Scenario verification | `test/scenarios/` scripts provide BDD/E2E gates through shell/Python orchestration around repo-tracked local runners such as existing package tests, Vitest, Playwright, and browser smoke; scenario-owned dependencies must not be implemented as new `backend/cmd` / Go helper processes. Kind / K8s / Helm are not the default P0 scenario target | Scenario owner + feature owner |

### 2.1 Operation Matrix Requirement

Any plan that adds or materially changes a user-facing API, business flow, or cross-layer data dependency must keep an operation matrix in the plan/checklist or a referenced spec section:

| Field | Required Meaning |
|-------|------------------|
| `operationId` | OpenAPI operation name or `N/A` for non-HTTP UI-only work |
| `fixture` | Fixture file and scenario names used by frontend/scenario tests |
| `frontend consumer` | Screen/hook/client method consuming the operation |
| `backend handler` | Real handler/store/job path, or explicit `mock-only` / `not-yet-implemented` status |
| `persistence` | Tables, object storage, queues, or `none` |
| `AI dependency` | Profile/capability used, `stub/fixture only`, or `none` |
| `scenario coverage` | Scenario ID or substitute gate |

This matrix is the handoff proof that frontend mock progress and backend real implementation status are not being conflated.

### 2.2 Frontend-First Path

Use this path when the UI and interaction shape can be developed before real backend implementation:

1. Update `ui-design/` and `docs/ui-design/` first.
2. Update OpenAPI/fixtures if the UI consumes new or changed data.
3. Run `make codegen && make codegen-check` after contract changes.
4. Implement frontend against generated client + fixture-backed fetch.
5. Run focused frontend tests, visual parity gates when UI changes, then `make test` / `make build` as appropriate.
6. Keep the operation matrix marking backend status as `mock-only` or `not-yet-implemented` until a backend owner lands the real handler.

### 2.3 Backend-First Path

Use this path when correctness depends on auth, persistence, jobs, privacy, or AI provider behavior:

1. Update OpenAPI/fixtures and migrations before handler code.
2. Implement handler/store/job code against generated DTO and shared conventions.
3. Add unit/contract tests for success, failure, auth, idempotency, and privacy red lines.
4. Run focused Go tests, migration gates when needed, then `make test` / `make build`.
5. Wire frontend to the real generated client method only after the operation matrix shows the backend handler and persistence path.

### 2.4 AI Provider Boundary

Automated local quality gates do not call a real LLM. `APP_ENV=test` may use stub/fixture AI providers; Docker Compose dependency-backed local app runs, staging, and production must fail fast when selected active profiles need a real provider but `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` are missing. Product UI and backend orchestration may be developed with fixtures, but model quality, latency, cost, streaming, STT, and fallback behavior require an explicit real-provider smoke/eval gate before release-level acceptance.

### 2.5 Agent Enforcement

AI agents must read this section and the relevant module README files before implementing or validating related code. The enforcement points are:

- `AGENTS.md` §2.1.3 and §4.4 require the pre-read for implementation, validation, and L2 review.
- `/implement` blocks `/tdd` handoff when a cross-layer plan lacks the operation matrix.
- `/tdd` repeats the preflight before editing files or marking checklist items.
- `/plan-review` treats missing or ambiguous operation matrix rows as L1 document findings.
- `/plan-code-review` treats missing operation matrix evidence as a blocking L2 finding and must include the contract pre-read in Deep Evidence.

## 3 What Is **Not** Set Up Yet

This project does **not** currently have:

- `.github/workflows/*.yml` (no GitHub Actions, GitLab CI, or other remote pipeline)
- PR / push required checks or branch protection rules
- nightly / scheduled jobs, Dependabot, Codecov, GHCR push, Docker buildx remote cache
- artifact upload (coverage HTML, OpenAPI diff artifact, build outputs)
- runner secret management

**Documentation must not assert that this project has an enabled CI pipeline or a mandatory pull-request check** — see [A5 spec §4.3](./spec/ci-pipeline-baseline/spec.md#43-文档约束) for the exact forbidden phrasings.

## 4 When Remote CI Will Be Added (D-5 Triggers)

Per [A5 spec D-5](./spec/ci-pipeline-baseline/spec.md#31-已锁定决策), remote CI is reconsidered when **any** of these conditions land:

1. A second long-term contributor joins
2. A public release branch is created
3. Paying users come online
4. Automated release / release-gate is needed
5. Regression frequency grows beyond what local gates can catch

**Upgrade path**: revise A5 spec in place and add a new sequenced plan (default name `002-remote-ci`). Do not change the subject directory, do not create a sibling spec, and do not back-fill remote CI scope into the current `001-local-quality-gates` plan. See [A5 spec §7](./spec/ci-pipeline-baseline/spec.md#7-关联计划).

## 5 Secret Red Line

Local quality gates **do not read business secrets**:

- AI provider keys (`AI_PROVIDER_*`)
- Database / Redis / object-storage credentials (`DB_*` / `REDIS_*` / `MINIO_*`)
- Analytics keys (`POSTHOG_*`)
- Any secret listed in [`.env.example`](../config/) for production / staging contexts

AI unit tests are required to use stub / fixture providers; real AI provider calls happen only in explicit non-test app runtime or smoke/eval runs after required local dependencies are available, never via `make test`. See [A5 spec §4.2](./spec/ci-pipeline-baseline/spec.md#42-安全与权限约束).

When remote CI eventually lands, any new workflow that needs a runner secret **must first revise A5 spec / history to register the secret dictionary and permission boundary** before the workflow file is created. Workflows that introduce secrets without a prior spec revision must be rejected at review.

## 6 Common Workflows

| Scenario | Recommended Sequence |
|----------|---------------------|
| Pre-commit sanity | `make lint && make test` |
| Before pushing a docs-only change | `make docs-check` |
| After editing `shared/conventions.yaml` or `openapi/openapi.yaml` | `make codegen && make codegen-check` |
| After editing `migrations/` | `make migrate-check` |
| Frontend feature with mock data | `make codegen && pnpm --filter @easyinterview/frontend test && pnpm --filter @easyinterview/frontend build` |
| Backend API implementation | `cd backend && go test ./...` then `make test && make build` |
| Full pre-push gate | `make lint && make test && make build && make docs-check && make codegen-check` |

## 7 Related Documents

- [A5 spec](./spec/ci-pipeline-baseline/spec.md) — local quality gate scope, D-5 deferral conditions
- [A5 plan 001-local-quality-gates](./spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md) — implementation detail
- [Frontend README](../frontend/README.md) — frontend implementation boundaries, UI truth source, generated client, mock transport
- [Backend README](../backend/README.md) — backend implementation boundaries and operation status expectations
- [Local dev stack](../deploy/dev-stack/README.md) — Docker Compose dependency stack, host-run app boundary, and scenario-runner boundary
- [Scenario framework](../test/scenarios/README.md) — BDD/E2E scenario environment and script contract
- [A1 repo-scaffold spec](./spec/repo-scaffold/spec.md) — root Makefile structure (10 phony targets)
- [B1 shared-conventions-codified spec](./spec/shared-conventions-codified/spec.md) — `make codegen-conventions` / `make lint-conventions` owner
- [B2 openapi-v1-contract spec](./spec/openapi-v1-contract/spec.md) — `make codegen-openapi` / `make lint-openapi` owner
- [A4 secrets-and-config spec](./spec/secrets-and-config/spec.md) — `make lint-config` owner
- [F1 observability-stack spec](./spec/observability-stack/spec.md) — metric / log contract owner; A5 will add a local command only after F1 exposes a real helper
