# Owned by docs/spec/local-dev-stack/plans/001-bootstrap.
# Real implementation behind the root Makefile's dev-* targets.

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

DEV_STACK_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
COMPOSE := docker compose -f $(DEV_STACK_DIR)/docker-compose.yaml --project-directory $(DEV_STACK_DIR)

DEPENDENCY_SERVICES := postgres-dev redis-dev minio-dev
INIT_SERVICES := minio-init
DOCTOR_SCRIPT := $(DEV_STACK_DIR)/scripts/dev-doctor.sh
ENV_EXAMPLE := $(DEV_STACK_DIR)/.env.example
ENV_FILE := $(DEV_STACK_DIR)/.env

.PHONY: up down doctor reset logs pull help _ensure_env _wait_init _all_healthy

help: ## List dev-stack make targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

_ensure_env:
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "[dev-stack] .env not found, copying from $(notdir $(ENV_EXAMPLE))"; \
		cp "$(ENV_EXAMPLE)" "$(ENV_FILE)"; \
	fi

_all_healthy: # internal: prints "yes" iff all deps running+healthy AND init exited 0
	@all_ok=yes; \
	for svc in $(DEPENDENCY_SERVICES); do \
		state="$$($(COMPOSE) ps --format '{{.State}}/{{.Health}}' $$svc 2>/dev/null || echo)"; \
		case "$$state" in \
			running/healthy) ;; \
			*) all_ok=no; break;; \
		esac; \
	done; \
	if [ "$$all_ok" = "yes" ]; then \
		init_state="$$($(COMPOSE) ps -a --format '{{.State}}/{{.ExitCode}}' minio-init 2>/dev/null || echo)"; \
		case "$$init_state" in \
			exited/0) ;; \
			*) all_ok=no;; \
		esac; \
	fi; \
	echo "$$all_ok"

up: _ensure_env ## Start dev stack (idempotent; skip if already healthy)
	@status="$$($(MAKE) -s _all_healthy)"; \
	if [ "$$status" = "yes" ]; then \
		echo "[dev-stack] already healthy, skipping start"; \
		exit 0; \
	fi; \
	echo "[dev-stack] starting dependency services"; \
	$(COMPOSE) up -d --wait --wait-timeout 60 $(DEPENDENCY_SERVICES); \
	echo "[dev-stack] running init jobs"; \
	$(COMPOSE) up -d $(INIT_SERVICES); \
	$(MAKE) -s _wait_init; \
	if [ -x "$(DOCTOR_SCRIPT)" ]; then \
		echo "[dev-stack] running dev-doctor gate"; \
		"$(DOCTOR_SCRIPT)" || { \
			echo "[dev-stack] dev-doctor reported degraded/down services" >&2; \
			for svc in $(DEPENDENCY_SERVICES) $(INIT_SERVICES); do \
				echo "----- last 50 lines for $$svc -----" >&2; \
				$(COMPOSE) logs --tail 50 --no-color $$svc >&2 || true; \
			done; \
			exit 1; \
		}; \
	fi; \
	echo "[dev-stack] up complete"

_wait_init: # internal: wait until every init container exits 0 (timeout 60s)
	@for svc in $(INIT_SERVICES); do \
		attempts=0; \
		while :; do \
			state="$$($(COMPOSE) ps -a --format '{{.State}}/{{.ExitCode}}' $$svc 2>/dev/null || echo)"; \
			case "$$state" in \
				exited/0) break;; \
				exited/*) \
					echo "[dev-stack] init service $$svc exited non-zero: $$state" >&2; \
					$(COMPOSE) logs --tail 50 --no-color $$svc >&2 || true; \
					exit 1;; \
			esac; \
			attempts=$$((attempts + 1)); \
			if [ $$attempts -ge 60 ]; then \
				echo "[dev-stack] init service $$svc did not finish within 60s" >&2; \
				$(COMPOSE) logs --tail 50 --no-color $$svc >&2 || true; \
				exit 1; \
			fi; \
			sleep 1; \
		done; \
	done

down: ## Stop dev stack; named volumes preserved (C-4)
	@$(COMPOSE) down

doctor: ## Run structured health probe (D-6 contract)
	@if [ ! -x "$(DOCTOR_SCRIPT)" ]; then \
		echo "[dev-stack] doctor script missing or not executable: $(DOCTOR_SCRIPT)" >&2; \
		exit 1; \
	fi; \
	"$(DOCTOR_SCRIPT)"

logs: ## Tail container logs; pass SERVICE=<name> to scope to one (C-7)
	@if [ -n "$${SERVICE:-}" ]; then \
		$(COMPOSE) logs --tail 200 --no-color $$SERVICE; \
	else \
		$(COMPOSE) logs --tail 200 --no-color; \
	fi

reset: ## Stop stack AND delete named volumes; requires DEV_RESET_FORCE=1 or stdin yes (C-5)
	@volumes="$$(docker volume ls --filter name=easyinterview --format '{{.Name}}' | sort -u)"; \
	if [ -z "$$volumes" ]; then \
		echo "[dev-stack] no easyinterview-* volumes present; nothing to reset"; \
		exit 0; \
	fi; \
	echo "[dev-stack] WARN: about to delete the following named volumes:" >&2; \
	for v in $$volumes; do \
		size="$$(docker system df -v --format '{{json .Volumes}}' | python3 -c "import json,sys; vols=json.load(sys.stdin); print(next((x.get('Size','?') for x in vols if x.get('Name')=='$$v'), '?'))" 2>/dev/null || echo '?')"; \
		echo "  - $$v  (size=$$size)" >&2; \
	done; \
	if [ "$${DEV_RESET_FORCE:-0}" != "1" ]; then \
		printf "Type 'yes' to confirm volume deletion: " >&2; \
		read -r answer; \
		if [ "$$answer" != "yes" ]; then \
			echo "[dev-stack] reset aborted, volumes preserved" >&2; \
			exit 1; \
		fi; \
	else \
		echo "[dev-stack] DEV_RESET_FORCE=1, skipping confirmation" >&2; \
	fi; \
	$(COMPOSE) down --volumes

pull: ## Pre-pull pinned images for slow-network bootstrap
	@$(COMPOSE) pull
