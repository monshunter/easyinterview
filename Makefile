SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GIT_HOOKS_DIR := $(ROOT_DIR)/scripts/git-hooks

.PHONY: help fmt lint lint-conventions lint-openapi openapi-diff validate-fixtures sync-fixtures-from-prototype render-openapi-fixture-examples test build dev-up dev-down dev-doctor dev-reset dev-logs dev-pull codegen codegen-conventions codegen-openapi codegen-check docs-openapi migrate install-hooks

help: ## List all top-level make targets with their descriptions
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

fmt: ## Format Go and frontend sources (delegates to backend/ and frontend/)
	@$(call recurse_target,fmt,backend/Makefile,backend)
	@$(call recurse_target,fmt,frontend/Makefile,frontend)

lint: lint-conventions ## Lint Go and frontend sources (delegates to backend/ and frontend/ after the cross-language conventions gate)
	@$(call recurse_target,lint,backend/Makefile,backend)
	@$(call recurse_target,lint,frontend/Makefile,frontend)

lint-conventions: ## Validate shared/conventions.yaml structure and error-code casing/boundary (B1 local gate)
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_yaml.py" "$(ROOT_DIR)/shared/conventions.yaml"
	@python3 "$(ROOT_DIR)/scripts/lint/error_codes.py"

test: ## Run unit tests (delegates to backend/ and frontend/)
	@$(call recurse_target,test,backend/Makefile,backend)
	@$(call recurse_target,test,frontend/Makefile,frontend)

build: ## Build Go binaries and frontend bundle (delegates to backend/ and frontend/)
	@$(call recurse_target,build,backend/Makefile,backend)
	@$(call recurse_target,build,frontend/Makefile,frontend)

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

codegen: codegen-conventions codegen-openapi ## Run all code generators in dependency order (B1 conventions → B2 openapi)

codegen-conventions: ## Render shared/conventions.yaml into Go and TS shared lib files
	@cd "$(ROOT_DIR)/backend" && go run ./cmd/codegen/conventions \
		-yaml "$(ROOT_DIR)/shared/conventions.yaml" \
		-repo-root "$(ROOT_DIR)"

codegen-openapi: codegen-conventions ## Render openapi/openapi.yaml into Go and TS API artefacts (idempotent)
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
		--fail-on-incompatible

validate-fixtures: ## Validate openapi/fixtures/*.json against openapi.yaml (B2 002 — schema, provenance, privacy, UUIDv7, 36-coverage)
	@python3 "$(ROOT_DIR)/scripts/lint/validate_fixtures.py" --repo-root "$(ROOT_DIR)"

sync-fixtures-from-prototype: ## Refresh `scenarios.prototype-baseline` of every P0 fixture from easyinterview-ui/src/data.jsx (B2 002, idempotent — fail-fast on mapping gaps)
	@python3 "$(ROOT_DIR)/scripts/codegen/sync_fixtures_from_prototype.py" --repo-root "$(ROOT_DIR)"

render-openapi-fixture-examples: ## Project openapi/fixtures/<tag>/<operationId>.json default scenario into openapi/.generated/openapi-with-fixtures.yaml (B2 002 — Prism / docs-site source; idempotent)
	@python3 "$(ROOT_DIR)/scripts/codegen/render_openapi_fixture_examples.py" --repo-root "$(ROOT_DIR)"

codegen-check: codegen-openapi lint-openapi ## Local drift gate: re-run codegen + lint, then `git diff --exit-code` on generated outputs and openapi.yaml. Currently the only required gate; remote CI required-check is deferred until A5 ci-pipeline-baseline triggers (spec §4.5 / §5).
	@git -C "$(ROOT_DIR)" diff --exit-code -- \
		"$(ROOT_DIR)/openapi/openapi.yaml" \
		"$(ROOT_DIR)/backend/internal/api/generated" \
		"$(ROOT_DIR)/frontend/src/api/generated"

docs-openapi: ## Render openapi/openapi.yaml as a single-file HTML site at openapi/dist/index.html (Redocly; local artefact only — A5 deferred for CI upload)
	@mkdir -p "$(ROOT_DIR)/openapi/dist"
	@npx --yes -p @redocly/cli@2.30.1 redocly build-docs \
		"$(ROOT_DIR)/openapi/openapi.yaml" \
		-o "$(ROOT_DIR)/openapi/dist/index.html" \
		--title "easyinterview API"
	@echo "rendered: $(ROOT_DIR)/openapi/dist/index.html"

migrate: ## Run DB schema migrations; implemented by B4 db-migrations-baseline
	@echo "TODO: implemented by B4 db-migrations-baseline"

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
