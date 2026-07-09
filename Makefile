SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GIT_HOOKS_DIR := $(ROOT_DIR)/scripts/git-hooks
EVAL_OUTPUT_DIR ?= $(ROOT_DIR)/.test-output/evals
EVAL_PROMPTFOO_CONFIG := $(EVAL_OUTPUT_DIR)/promptfooconfig.yaml
SCENARIO_ENV_SETUP := $(ROOT_DIR)/test/scenarios/env-setup.sh
SCENARIO_ENV_STATUS := $(ROOT_DIR)/test/scenarios/env-status.sh
SCENARIO_ENV_VERIFY := $(ROOT_DIR)/test/scenarios/env-verify.sh
SCENARIO_ENV_CLEANUP := $(ROOT_DIR)/test/scenarios/env-cleanup.sh
SCENARIO_ENV_REDEPLOY := $(ROOT_DIR)/test/scenarios/env-redeploy.sh
TARGET ?= all

.PHONY: help fmt lint lint-conventions lint-config lint-getenv-boundary lint-env-dict lint-ai-provider-terminology lint-ai-profile-coverage lint-backend-practice-non-current lint-runner-non-current lint-prompts lint-rubrics lint-prompts-hardcode lint-mock-contract lint-core-loop-pruning-surface lint-secrets-pattern lint-observability lint-events lint-runtime-topology lint-openapi openapi-diff validate-fixtures sync-fixtures-from-prototype render-openapi-fixture-examples test build eval-offline eval-offline-resolve dev-up dev-down dev-doctor dev-reset dev-logs dev-pull scenario-env-setup scenario-env-status scenario-env-verify scenario-env-cleanup scenario-env-redeploy scenario-env-reset-redeploy codegen codegen-conventions codegen-events codegen-openapi codegen-events-check codegen-check docs-check docs-openapi migrate migrate-up migrate-down migrate-status migrate-create migrate-check privacy-delete-dry-run install-hooks

help: ## List all top-level make targets with their descriptions
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

fmt: ## Format Go and frontend sources (delegates to backend/ and frontend/)
	@$(call recurse_target,fmt,backend/Makefile,backend)
	@$(call recurse_target,fmt,frontend/Makefile,frontend)

lint: lint-conventions lint-config lint-ai-profile-coverage lint-backend-practice-non-current lint-runner-non-current lint-prompts lint-rubrics lint-prompts-hardcode lint-mock-contract lint-core-loop-pruning-surface lint-runtime-topology lint-observability ## Lint Go and frontend sources (B1/A4/A3-F3/E1/runtime-topology/F1 local gates, then backend golangci-lint + frontend pnpm lint)
	@cd "$(ROOT_DIR)/backend" && golangci-lint run ./...
	@pnpm --filter @easyinterview/frontend lint

lint-conventions: ## lint-conventions (B1): validate shared/conventions.yaml structure and error-code casing/boundary
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_yaml.py" "$(ROOT_DIR)/shared/conventions.yaml"
	@python3 "$(ROOT_DIR)/scripts/lint/error_codes.py"

lint-config: lint-getenv-boundary lint-env-dict lint-ai-provider-terminology lint-secrets-pattern ## lint-config (A4): env boundary + .env.example drift + provider terminology + secret-pattern scan

lint-getenv-boundary: ## Reject os.Getenv usage outside platform/config / platform/secrets / cmd/{api,migrate} (spec C-7)
	@cd "$(ROOT_DIR)" && go run scripts/lint/getenv_boundary.go -root backend

lint-env-dict: ## Verify .env.example, code-side os.Getenv calls, and spec §3.1.1 dictionary stay aligned (spec C-9 / C-11)
	@python3 "$(ROOT_DIR)/scripts/lint/env_dict.py" --repo-root "$(ROOT_DIR)"

lint-ai-provider-terminology: ## Reject non-current AI gateway terminology in active AI provider surfaces
	@python3 "$(ROOT_DIR)/scripts/lint/ai_provider_terminology.py" --repo-root "$(ROOT_DIR)"

lint-ai-profile-coverage: ## Validate A3/F3/Product-UI AI profile coverage
	@python3 "$(ROOT_DIR)/scripts/lint/ai_profile_coverage.py" --repo-root "$(ROOT_DIR)"

lint-backend-practice-non-current: ## Reject backend-practice plan 001 non-current mode/module terms
	@python3 "$(ROOT_DIR)/scripts/lint/backend_practice_non_current.py" --repo-root "$(ROOT_DIR)" --phase all

lint-runner-non-current: ## Reject backend-async-runner non-current runner/drainer/dispatcher entry points
	@python3 "$(ROOT_DIR)/scripts/lint/runner_non_current.py" --repo-root "$(ROOT_DIR)" --phase all

lint-prompts: ## lint-prompts (F3): validate config/prompts/ schema, hash, and seed-migration drift
	@python3 "$(ROOT_DIR)/scripts/lint/prompt_lint.py" --prompts-dir "$(ROOT_DIR)/config/prompts" --migrations-dir "$(ROOT_DIR)/migrations"

lint-rubrics: ## lint-rubrics (F3): validate config/rubrics/ schema, weight tolerance, and dimension allowlist
	@python3 "$(ROOT_DIR)/scripts/lint/rubric_lint.py" --rubrics-dir "$(ROOT_DIR)/config/rubrics"

lint-prompts-hardcode: ## lint-prompts-hardcode (F3): reject raw / long-string prompt assignments outside config/prompts/
	@cd "$(ROOT_DIR)" && python3 scripts/lint/prompt_hardcode_lint.py

lint-mock-contract: validate-fixtures lint-openapi ## Validate fixture coverage, registry metadata, mock runtime boundaries, and non-current mock/API tokens
	@python3 -m unittest scripts.mock_contract.fixture_registry_test
	@python3 "$(ROOT_DIR)/scripts/lint/mock_runtime_boundary.py" --repo-root "$(ROOT_DIR)"

lint-core-loop-pruning-surface: ## Bucket D-22 non-current Debrief/Profile/JD Match runtime/generated references and fail real residuals
	@python3 "$(ROOT_DIR)/scripts/lint/core_loop_pruning_surface.py" --repo-root "$(ROOT_DIR)"

lint-secrets-pattern: ## Scan staged + tracked files for AKIA / sk- / xox secret prefixes (defense-in-depth; pre-commit hook is the primary gate)
	@bash "$(ROOT_DIR)/scripts/lint/gitleaks.sh" "$(ROOT_DIR)"

lint-observability: ## lint-observability (F1): observability-stack metrics/log lint hook (NOT-YET-LANDED placeholder, exit 0 until F1 lands)
	@echo "not implemented yet: F1 observability-stack"

lint-events: ## Validate event/job baselines and local B3 contract drift
	@python3 "$(ROOT_DIR)/scripts/lint/lint_events.py" --repo-root "$(ROOT_DIR)"

lint-runtime-topology: ## Reject non-current standalone backend worker process terminology in active code/docs
	@python3 "$(ROOT_DIR)/scripts/lint/runtime_topology.py" --repo-root "$(ROOT_DIR)"

test: ## A5 test aggregator: backend Go + frontend pnpm; AI tests use stub/fixture only, no provider secrets
	@cd "$(ROOT_DIR)/backend" && go test ./...
	@pnpm --filter @easyinterview/frontend test

build: ## A5 build aggregator: backend cmd binaries + frontend bundle
	@cd "$(ROOT_DIR)/backend" && go build ./cmd/...
	@pnpm --filter @easyinterview/frontend build

eval-offline-resolve: ## F3: regenerate the committed registry-resolved single-source export for the offline eval suite
	@cd "$(ROOT_DIR)" && go run ./backend/cmd/evalkit resolve

eval-offline: ## F3 offline eval (NOT in make test): single-source drift gate + >=36 count + Promptfoo runner over recorded fixtures; EVAL_LIVE=1 opts into real provider/judge calls
	@cd "$(ROOT_DIR)" && go build -o backend/bin/evalkit ./backend/cmd/evalkit
	@cd "$(ROOT_DIR)" && ./backend/bin/evalkit drift-check
	@cd "$(ROOT_DIR)" && ./backend/bin/evalkit run $(if $(EVAL_LIVE),--live,)
	@mkdir -p "$(EVAL_OUTPUT_DIR)/promptfoo/logs"
	@cd "$(ROOT_DIR)" && ./backend/bin/evalkit prompts-tests --out "$(EVAL_OUTPUT_DIR)/promptfoo_tests.yaml"
	@sed \
		-e "s#__PROMPTFOO_PROVIDER_FILE__#$(ROOT_DIR)/config/evals/promptfoo_provider.js#g" \
		-e "s#__PROMPTFOO_ASSERT_FILE__#$(ROOT_DIR)/config/evals/promptfoo_assert.js#g" \
		-e "s#__PROMPTFOO_TESTS_FILE__#$(EVAL_OUTPUT_DIR)/promptfoo_tests.yaml#g" \
		"$(ROOT_DIR)/config/evals/promptfooconfig.yaml" > "$(EVAL_PROMPTFOO_CONFIG)"
	@cd "$(ROOT_DIR)" && PROMPTFOO_CONFIG_DIR="$(EVAL_OUTPUT_DIR)/promptfoo" PROMPTFOO_LOG_DIR="$(EVAL_OUTPUT_DIR)/promptfoo/logs" PROMPTFOO_DISABLE_WAL_MODE=true PROMPTFOO_DISABLE_TELEMETRY=1 PROMPTFOO_DISABLE_UPDATE=1 EVALKIT_BIN="$(ROOT_DIR)/backend/bin/evalkit" pnpm exec promptfoo eval -c "$(EVAL_PROMPTFOO_CONFIG)" --no-cache

dev-up: ## Start local dev external dependencies (postgres / redis / minio / mailpit; optional app services only when explicitly added)
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

scenario-env-setup: ## Setup shared scenario/local-integration environment (ARGS="--with-migrations")
	@$(SCENARIO_ENV_SETUP) $(ARGS)

scenario-env-status: ## Print shared scenario/local-integration environment status
	@$(SCENARIO_ENV_STATUS) $(ARGS)

scenario-env-verify: ## Verify shared scenario/local-integration environment readiness
	@$(SCENARIO_ENV_VERIFY) $(ARGS)

scenario-env-cleanup: ## Cleanup shared scenario/local-integration environment (ARGS="--with-volumes")
	@$(SCENARIO_ENV_CLEANUP) $(ARGS)

scenario-env-redeploy: ## Rebuild/redeploy local env components (TARGET=deps|backend|frontend|all)
	@$(SCENARIO_ENV_REDEPLOY) $(TARGET) $(ARGS)

scenario-env-reset-redeploy: ## Reset data, run migrations, rebuild/redeploy backend+frontend, then verify (ARGS=--dry-run to preview)
	@$(SCENARIO_ENV_CLEANUP) --with-volumes $(ARGS)
	@$(SCENARIO_ENV_SETUP) --with-migrations $(ARGS)
	@$(SCENARIO_ENV_REDEPLOY) all $(ARGS)
	@$(SCENARIO_ENV_VERIFY) $(ARGS)

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
		"$(ROOT_DIR)/backend/internal/shared/events/envelope.go" \
		"$(ROOT_DIR)/backend/internal/shared/events/events.go" \
		"$(ROOT_DIR)/backend/internal/shared/jobs/jobs.go" \
		"$(ROOT_DIR)/frontend/src/lib/events/envelope.ts" \
		"$(ROOT_DIR)/frontend/src/lib/events/events.ts" \
		"$(ROOT_DIR)/frontend/src/lib/jobs/jobs.ts" \
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

validate-fixtures: ## Validate openapi/fixtures/*.json against openapi.yaml (B2 002 — schema, provenance, privacy, UUIDv7, operation coverage)
	@python3 "$(ROOT_DIR)/scripts/lint/validate_fixtures.py" --repo-root "$(ROOT_DIR)"

sync-fixtures-from-prototype: ## Refresh `scenarios.prototype-baseline` of every P0 fixture from ui-design/src/data.jsx (B2 002, idempotent — fail-fast on mapping gaps)
	@python3 "$(ROOT_DIR)/scripts/codegen/sync_fixtures_from_prototype.py" --repo-root "$(ROOT_DIR)"

render-openapi-fixture-examples: ## Project openapi/fixtures/<tag>/<operationId>.json default scenario into openapi/.generated/openapi-with-fixtures.yaml (B2 002 — Prism / docs-site source; idempotent)
	@python3 "$(ROOT_DIR)/scripts/codegen/render_openapi_fixture_examples.py" --repo-root "$(ROOT_DIR)"

docs-check: ## A5 docs gate aggregator: Header/INDEX drift + docs relative links + docs/spec heading fragments
	@python3 "$(ROOT_DIR)/.agent-skills/sync-doc-index/scripts/sync-doc-index.py" --check
	@python3 "$(ROOT_DIR)/scripts/lint/check_md_links.py" "$(ROOT_DIR)/docs" --ignore '**/TEMPLATES.md'
	@python3 "$(ROOT_DIR)/scripts/lint/check_md_links.py" "$(ROOT_DIR)/docs/spec" --ignore '**/TEMPLATES.md' --check-fragments

codegen-check: ## A5 codegen drift gate aggregator: B1 conventions + B3 events/jobs + B2 OpenAPI generators, re-run lint + drift checks. Remote CI required-check deferred until A5 D-5 trigger (spec §4.5 / §5).
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_yaml.py" "$(ROOT_DIR)/shared/conventions.yaml"
	@python3 "$(ROOT_DIR)/scripts/lint/error_codes.py"
	@python3 "$(ROOT_DIR)/scripts/lint/conventions_drift.py" --repo-root "$(ROOT_DIR)"
	@$(MAKE) --no-print-directory codegen-events-check
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
