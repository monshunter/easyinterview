SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := help

ROOT_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
GIT_HOOKS_DIR := $(ROOT_DIR)/scripts/git-hooks
GIT_HOOKS_INSTALL_DIR := $(ROOT_DIR)/.git/hooks

.PHONY: help fmt lint test build dev-up dev-down codegen migrate install-hooks

help: ## List all top-level make targets with their descriptions
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

fmt: ## Format Go and frontend sources (delegates to backend/ and frontend/)
	@$(call recurse_target,fmt,backend/Makefile,backend)
	@$(call recurse_target,fmt,frontend/Makefile,frontend)

lint: ## Lint Go and frontend sources (delegates to backend/ and frontend/)
	@$(call recurse_target,lint,backend/Makefile,backend)
	@$(call recurse_target,lint,frontend/Makefile,frontend)

test: ## Run unit tests (delegates to backend/ and frontend/)
	@$(call recurse_target,test,backend/Makefile,backend)
	@$(call recurse_target,test,frontend/Makefile,frontend)

build: ## Build Go binaries and frontend bundle (delegates to backend/ and frontend/)
	@$(call recurse_target,build,backend/Makefile,backend)
	@$(call recurse_target,build,frontend/Makefile,frontend)

dev-up: ## Start local dev dependencies; implemented by A2 local-dev-stack
	@echo "TODO: implemented by A2 local-dev-stack"

dev-down: ## Stop local dev dependencies; implemented by A2 local-dev-stack
	@echo "TODO: implemented by A2 local-dev-stack"

codegen: ## Run code generators; implemented by B2 openapi-v1-contract
	@echo "TODO: implemented by B2 openapi-v1-contract"

migrate: ## Run DB schema migrations; implemented by B4 db-migrations-baseline
	@echo "TODO: implemented by B4 db-migrations-baseline"

install-hooks: ## Symlink scripts/git-hooks/{pre-commit,commit-msg} into .git/hooks
	@if [ ! -d "$(GIT_HOOKS_INSTALL_DIR)" ]; then \
		echo "ERROR: $(GIT_HOOKS_INSTALL_DIR) does not exist; not a git worktree"; exit 1; \
	fi
	@for hook in pre-commit commit-msg; do \
		src="$(GIT_HOOKS_DIR)/$$hook"; \
		dst="$(GIT_HOOKS_INSTALL_DIR)/$$hook"; \
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
