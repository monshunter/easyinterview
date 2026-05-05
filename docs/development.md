# Local Development Quality Gates

> **Owner**: [`ci-pipeline-baseline/001-local-quality-gates`](./spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md) (A5)
> **Updated**: 2026-04-30

This page is the canonical reference for the 5 local quality gates that every contributor runs before pushing changes. The project is currently in **single-developer P0 phase**: there is **no remote CI pipeline**. Required checks, branch protection, and runner secrets are deferred until A5 spec [D-5](./spec/ci-pipeline-baseline/spec.md#31-已锁定决策) trigger conditions land.

## 1 The 5 Local Gates

Run from repo root:

| Command | Purpose | Owner |
|---------|---------|-------|
| `make lint` | B1 conventions + A4 config + F1 observability (placeholder) + Go/TS lint | A5 aggregator → B1 / A4 / F1 |
| `make test` | Backend Go unit tests + frontend TypeScript unit tests; AI tests via stub/fixture only | A5 aggregator → backend / frontend owners |
| `make build` | Backend `go build ./cmd/...` + frontend bundle | A5 aggregator → backend / frontend owners |
| `make docs-check` | `sync-doc-index --check` (Header / INDEX drift) + relative-link sanity for `docs/` | A5 aggregator → `/sync-doc-index` skill + A5 `scripts/lint/check_md_links.py` |
| `make codegen-check` | B1 conventions generator + B2 OpenAPI generator + `git diff --exit-code` on generated outputs | A5 aggregator → B1 + B2 |

Each gate exits non-zero on failure. NOT-YET-LANDED owner sub-targets print a single line `not implemented yet: <owner>` and exit 0; landed sub-targets propagate their exit code (failures are not swallowed).

## 2 What Is **Not** Set Up Yet

This project does **not** currently have:

- `.github/workflows/*.yml` (no GitHub Actions, GitLab CI, or other remote pipeline)
- PR / push required checks or branch protection rules
- nightly / scheduled jobs, Dependabot, Codecov, GHCR push, Docker buildx remote cache
- artifact upload (coverage HTML, OpenAPI diff artifact, build outputs)
- runner secret management

**Documentation must not assert that this project has an enabled CI pipeline or a mandatory pull-request check** — see [A5 spec §4.3](./spec/ci-pipeline-baseline/spec.md#43-文档约束) for the exact forbidden phrasings.

## 3 When Remote CI Will Be Added (D-5 Triggers)

Per [A5 spec D-5](./spec/ci-pipeline-baseline/spec.md#31-已锁定决策), remote CI is reconsidered when **any** of these conditions land:

1. A second long-term contributor joins
2. A public release branch is created
3. Paying users come online
4. Automated release / release-gate is needed
5. Regression frequency grows beyond what local gates can catch

**Upgrade path**: revise A5 spec in place and add a new sequenced plan (default name `002-remote-ci`). Do not change the subject directory, do not create a sibling spec, and do not back-fill remote CI scope into the current `001-local-quality-gates` plan. See [A5 spec §7](./spec/ci-pipeline-baseline/spec.md#7-关联计划).

## 4 Secret Red Line

Local quality gates **do not read business secrets**:

- AI provider keys (`AI_PROVIDER_*`)
- Database / Redis / object-storage credentials (`DB_*` / `REDIS_*` / `MINIO_*`)
- Analytics keys (`POSTHOG_*`)
- Any secret listed in [`.env.example`](../config/) for production / staging contexts

AI unit tests are required to use stub / fixture providers; real AI provider calls happen only via `make dev-up` local-stack invocation, never via `make test`. See [A5 spec §4.2](./spec/ci-pipeline-baseline/spec.md#42-安全与权限约束).

When remote CI eventually lands, any new workflow that needs a runner secret **must first revise A5 spec / history to register the secret dictionary and permission boundary** before the workflow file is created. Workflows that introduce secrets without a prior spec revision must be rejected at review.

## 5 Common Workflows

| Scenario | Recommended Sequence |
|----------|---------------------|
| Pre-commit sanity | `make lint && make test` |
| Before pushing a docs-only change | `make docs-check` |
| After editing `shared/conventions.yaml` or `openapi/openapi.yaml` | `make codegen && make codegen-check` |
| After editing `migrations/` | `make migrate-check` |
| Full pre-push gate | `make lint && make test && make build && make docs-check && make codegen-check` |

## 6 Related Documents

- [A5 spec](./spec/ci-pipeline-baseline/spec.md) — local quality gate scope, D-5 deferral conditions
- [A5 plan 001-local-quality-gates](./spec/ci-pipeline-baseline/plans/001-local-quality-gates/plan.md) — implementation detail
- [A1 repo-scaffold spec](./spec/repo-scaffold/spec.md) — root Makefile structure (10 phony targets)
- [B1 shared-conventions-codified spec](./spec/shared-conventions-codified/spec.md) — `make codegen-conventions` / `make lint-conventions` owner
- [B2 openapi-v1-contract spec](./spec/openapi-v1-contract/spec.md) — `make codegen-openapi` / `make lint-openapi` owner
- [A4 secrets-and-config spec](./spec/secrets-and-config/spec.md) — `make lint-config` owner
- [F1 observability-stack spec](./spec/observability-stack/spec.md) — future `make lint-observability` owner
