SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GIT_HOOKS_DIR := $(ROOT_DIR)/scripts/git-hooks

.PHONY: help fmt lint lint-conventions lint-config lint-getenv-boundary lint-env-dict lint-secrets-pattern lint-observability lint-events lint-openapi openapi-diff validate-fixtures sync-fixtures-from-prototype render-openapi-fixture-examples test build dev-up dev-down dev-doctor dev-reset dev-logs dev-pull codegen codegen-conventions codegen-events codegen-openapi codegen-events-check codegen-check docs-check docs-openapi migrate migrate-up migrate-down migrate-status migrate-create migrate-check privacy-delete-dry-run install-hooks

help: ## List all top-level make targets with their descriptions
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

fmt: ## Format Go and frontend sources (delegates to backend/ and frontend/)
	@$(call recurse_target,fmt,backend/Makefile,backend)
	@$(call recurse_target,fmt,frontend/Makefile,frontend)

lint: lint-conventions lint-config lint-observability ## Lint Go and frontend sources (lint-conventions (B1) / lint-config (A4) / lint-observability (F1), then backend golangci-lint + frontend pnpm lint)
	@cd "$(ROOT_DIR)/backend" && golangci-lint run ./...
	@pnpm --filter @easyinterview/frontend lint

lint-conventions: ## lint-conventions (B1): validate shared/conventions.yaml structure and error-code casing/boundary
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_yaml.py" "$(ROOT_DIR)/shared/conventions.yaml"
	@python3 "$(ROOT_DIR)/scripts/lint/error_codes.py"

lint-config: lint-getenv-boundary lint-env-dict lint-secrets-pattern ## lint-config (A4): env boundary + .env.example drift + secret-pattern scan

lint-getenv-boundary: ## Reject os.Getenv usage outside platform/config / platform/secrets / cmd/{api,worker} (spec C-7)
	@cd "$(ROOT_DIR)" && go run scripts/lint/getenv_boundary.go -root backend

lint-env-dict: ## Verify .env.example, code-side os.Getenv calls, and spec §3.1.1 dictionary stay aligned (spec C-9 / C-11)
	@python3 "$(ROOT_DIR)/scripts/lint/env_dict.py" --repo-root "$(ROOT_DIR)"

lint-secrets-pattern: ## Scan staged + tracked files for AKIA / sk- / xox secret prefixes (defense-in-depth; pre-commit hook is the primary gate)
	@bash "$(ROOT_DIR)/scripts/lint/gitleaks.sh" "$(ROOT_DIR)"

lint-observability: ## lint-observability (F1): observability-stack metrics/log lint hook (NOT-YET-LANDED placeholder, exit 0 until F1 lands)
	@echo "not implemented yet: F1 observability-stack"

lint-events: ## Validate event/job baselines and local B3 contract drift
	@python3 "$(ROOT_DIR)/scripts/lint/lint_events.py" --repo-root "$(ROOT_DIR)"

test: ## A5 test aggregator: backend Go + frontend pnpm; AI tests use stub/fixture only, no provider secrets
	@cd "$(ROOT_DIR)/backend" && go test ./...
	@pnpm --filter @easyinterview/frontend test

build: ## A5 build aggregator: backend cmd binaries + frontend bundle
	@cd "$(ROOT_DIR)/backend" && go build ./cmd/...
	@pnpm --filter @easyinterview/frontend build

dev-up: ## Start local dev dependencies (postgres+pgvector / redis / minio + project components)
	@$(MAKE) -C "$(ROOT_DIR)/deploy/dev-stack" up

dev-down: ## Stop local dev dependencies; named volumes preserved
	@$(MAKE) -C "$(ROOT_DIR)/deploy/dev-stack" down

dev-doctor: ## Structured health probe (D-6 JSON contract)
	@$(MAKE) -C "$(ROOT_DIR)/deploy/dev-stack" doctor

dev-reset: ## Stop dev stack AND delete named volumes (DEV_RESET_FORCE=1 to skip prompt)
	@$(MAKE) -C "$(ROOT_DIR)/deploy/dev-stack" reset

dev-logs: ## Tail dev-stack container logs (SERVICE=<name> to scope)
	@$(MAKE) -C "$(ROOT_DIR)/deploy/dev-stack" logs

dev-pull: ## Pre-pull dev-stack pinned images for slow-network bootstrap
	@$(MAKE) -C "$(ROOT_DIR)/deploy/dev-stack" pull

codegen: codegen-conventions codegen-events codegen-openapi ## Run all code generators in dependency order (B1 conventions → B3 events → B2 openapi)

codegen-conventions: ## codegen-conventions (B1): render shared/conventions.yaml into Go and TS shared lib files
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/codegen/conventions \
		-yaml "$(ROOT_DIR)/shared/conventions.yaml" \
		-repo-root "$(ROOT_DIR)"

codegen-events: codegen-conventions ## Render shared/events.yaml and shared/jobs.yaml into Go, TS, JSON Schema, and baseline artefacts
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/codegen/events \
		-events "$(ROOT_DIR)/shared/events.yaml" \
		-jobs "$(ROOT_DIR)/shared/jobs.yaml" \
		-repo-root "$(ROOT_DIR)"

# B3 local drift gate only. Remote required CI wiring is deferred until A5 ci-pipeline-baseline triggers.
codegen-events-check: ## Local B3 event/job contract drift gate
	@$(MAKE) --no-print-directory codegen-events
	@$(MAKE) --no-print-directory lint-events
	@git -C "$(ROOT_DIR)" diff --exit-code -- \
		"$(ROOT_DIR)/shared/events.yaml" \
		"$(ROOT_DIR)/shared/jobs.yaml" \
		"$(ROOT_DIR)/backend/internal/shared/events" \
		"$(ROOT_DIR)/backend/internal/shared/jobs" \
		"$(ROOT_DIR)/frontend/src/lib/events" \
		"$(ROOT_DIR)/frontend/src/lib/jobs" \
		"$(ROOT_DIR)/shared/events/schemas" \
		"$(ROOT_DIR)/shared/events/refs" \
		"$(ROOT_DIR)/shared/events/baseline" \
		"$(ROOT_DIR)/shared/jobs/baseline"

codegen-openapi: codegen-conventions ## codegen-openapi (B2): render openapi/openapi.yaml into Go and TS API artefacts (idempotent)
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/codegen/openapi \
		-openapi "$(ROOT_DIR)/openapi/openapi.yaml" \
		-conventions "$(ROOT_DIR)/shared/conventions.yaml" \
		-templates "$(ROOT_DIR)/openapi/templates" \
		-repo-root "$(ROOT_DIR)"

lint-openapi: ## Validate openapi.yaml structurally + enforce spec §3.1.1 inventory invariants
	@npx --yes -p @apidevtools/swagger-cli@4.0.4 swagger-cli validate "$(ROOT_DIR)/openapi/openapi.yaml"
	@python3 "$(ROOT_DIR)/scripts/lint/openapi_inventory.py" "$(ROOT_DIR)/openapi/openapi.yaml"

openapi-diff: ## Compare openapi/openapi.yaml against the latest baseline under openapi/baseline/ (003 — spec §4.4 breaking-change gate; BASELINE_VERSION=v1.0.0 to pin)
	@python3 "$(ROOT_DIR)/scripts/lint/openapi_diff.py" \
		--repo-root "$(ROOT_DIR)" \
		$(if $(BASELINE_VERSION),--baseline-version $(BASELINE_VERSION),) \
		$(if $(HISTORY_REF),--history-ref $(HISTORY_REF),) \
		--fail-on-incompatible

validate-fixtures: ## Validate openapi/fixtures/*.json against openapi.yaml (B2 002 — schema, provenance, privacy, UUIDv7, 36-coverage)
	@python3 "$(ROOT_DIR)/scripts/lint/validate_fixtures.py" --repo-root "$(ROOT_DIR)"

sync-fixtures-from-prototype: ## Refresh `scenarios.prototype-baseline` of every P0 fixture from easyinterview-ui/src/data.jsx (B2 002, idempotent — fail-fast on mapping gaps)
	@python3 "$(ROOT_DIR)/scripts/codegen/sync_fixtures_from_prototype.py" --repo-root "$(ROOT_DIR)"

render-openapi-fixture-examples: ## Project openapi/fixtures/<tag>/<operationId>.json default scenario into openapi/.generated/openapi-with-fixtures.yaml (B2 002 — Prism / docs-site source; idempotent)
	@python3 "$(ROOT_DIR)/scripts/codegen/render_openapi_fixture_examples.py" --repo-root "$(ROOT_DIR)"

docs-check: ## A5 docs gate aggregator: sync-doc-index Header/INDEX drift (skill-owned) + relative-link sanity for docs/ (A5-owned check_md_links.py; TEMPLATES.md placeholders excluded)
	@python3 "$(ROOT_DIR)/.agent-skills/sync-doc-index/scripts/sync-doc-index.py" --check
	@python3 "$(ROOT_DIR)/scripts/lint/check_md_links.py" "$(ROOT_DIR)/docs" --ignore '**/TEMPLATES.md'

codegen-check: ## A5 codegen drift gate aggregator: B1 conventions + B2 OpenAPI generators, re-run lint + `git diff --exit-code` on B1/B2 generated outputs. Remote CI required-check deferred until A5 D-5 trigger (spec §4.5 / §5).
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_yaml.py" "$(ROOT_DIR)/shared/conventions.yaml"
	@python3 "$(ROOT_DIR)/scripts/lint/error_codes.py"
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_drift.py" --repo-root "$(ROOT_DIR)"
	@$(MAKE) --no-print-directory codegen-openapi
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_drift.py" --repo-root "$(ROOT_DIR)"
	@$(MAKE) --no-print-directory lint-openapi
	@git -C "$(ROOT_DIR)" diff --exit-code -- \
		"$(ROOT_DIR)/shared/conventions.yaml" \
		"$(ROOT_DIR)/openapi/openapi.yaml" \
		"$(ROOT_DIR)/backend/internal/api/generated" \
		"$(ROOT_DIR)/frontend/src/api/generated" \
		"$(ROOT_DIR)/backend/internal/shared/types/enums.go" \
		"$(ROOT_DIR)/backend/internal/shared/types/http_dto.go" \
		"$(ROOT_DIR)/backend/internal/shared/errors/codes.go" \
		"$(ROOT_DIR)/backend/internal/shared/idx/generated.go" \
		"$(ROOT_DIR)/backend/internal/shared/ai" \
		"$(ROOT_DIR)/frontend/src/lib/conventions/enums.ts" \
		"$(ROOT_DIR)/frontend/src/lib/conventions/errors.ts" \
		"$(ROOT_DIR)/frontend/src/lib/conventions/ai.ts" \
		"$(ROOT_DIR)/frontend/src/lib/conventions/pagination.ts" \
		"$(ROOT_DIR)/frontend/src/lib/ids/generated.ts"

docs-openapi: ## Render openapi/openapi.yaml as a single-file HTML site at openapi/dist/index.html (Redocly; local artefact only — A5 deferred for CI upload)
	@mkdir -p "$(ROOT_DIR)/openapi/dist"
	@npx --yes -p @redocly/cli@2.30.1 redocly build-docs \
		"$(ROOT_DIR)/openapi/openapi.yaml" \
		-o "$(ROOT_DIR)/openapi/dist/index.html" \
		--title "easyinterview API"
	@echo "rendered: $(ROOT_DIR)/openapi/dist/index.html"

migrate: ## Show DB migration wrapper help
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate --help

migrate-up: ## Run all pending DB schema migrations
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate \
		--migrations-dir "$(ROOT_DIR)/migrations" \
		--backfill-manifest "$(ROOT_DIR)/migrations/backfill/manifest.yaml" \
		up

migrate-down: ## Roll back DB schema migrations; refused in APP_ENV=prod unless explicitly forced
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate \
		--migrations-dir "$(ROOT_DIR)/migrations" \
		--backfill-manifest "$(ROOT_DIR)/migrations/backfill/manifest.yaml" \
		down

migrate-status: ## Print current DB migration version
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate \
		--migrations-dir "$(ROOT_DIR)/migrations" \
		--backfill-manifest "$(ROOT_DIR)/migrations/backfill/manifest.yaml" \
		status

migrate-create: ## Create paired migration files; usage: make migrate-create NAME=add_example
	@if [ -z "$(NAME)" ]; then echo "ERROR: NAME is required, e.g. make migrate-create NAME=add_example" >&2; exit 2; fi
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate \
		--migrations-dir "$(ROOT_DIR)/migrations" \
		create "$(NAME)"

migrate-check: ## Run B4 local migration gate: up -> down -> up plus lint/probe checks
	@python3 "$(ROOT_DIR)/scripts/lint/migrations_lint.py" --repo-root "$(ROOT_DIR)"
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate \
		--migrations-dir "$(ROOT_DIR)/migrations" \
		--backfill-manifest "$(ROOT_DIR)/migrations/backfill/manifest.yaml" \
		check

privacy-delete-dry-run: ## Print the B4 privacy deletion table matrix for a dry run
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/migrate privacy-matrix --dry-run

install-hooks: ## Symlink scripts/git-hooks/{pre-commit,commit-msg} into .git/hooks
	@hooks_dir="$$(git -C "$(ROOT_DIR)" rev-parse --git-path hooks 2>/dev/null || true)"; \
	if [ -z "$$hooks_dir" ]; then \
		echo "ERROR: $(ROOT_DIR) is not a git worktree"; exit 1; \
	fi; \
	case "$$hooks_dir" in \
		/*) ;; \
		*) hooks_dir="$(ROOT_DIR)/$$hooks_dir";; \
	esac; \
	mkdir -p "$$hooks_dir"; \
	for hook in pre-commit commit-msg; do \
		src="$(GIT_HOOKS_DIR)/$$hook"; \
		dst="$$hooks_dir/$$hook"; \
		if [ ! -f "$$src" ]; then \
			echo "ERROR: missing $$src"; exit 1; \
		fi; \
		ln -sf "$$src" "$$dst"; \
		echo "linked: $$dst -> $$src"; \
	done

# recurse_target $1=target $2=child Makefile path $3=child label
define recurse_target
	if [ -f "$(ROOT_DIR)/$2" ] && grep -qE "^$1:" "$(ROOT_DIR)/$2"; then \
		$(MAKE) -C "$(ROOT_DIR)/$3" $1; \
	else \
		echo "TODO: $1 implemented by $3 child subspec"; \
	fi
endef
