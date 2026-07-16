# Owned by docs/spec/local-dev-stack/plans/001-bootstrap.
# Real implementation behind the root Makefile's dev-* targets.

SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c

DEV_STACK_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
REPO_ROOT := $(abspath $(DEV_STACK_DIR)/../..)
COMPOSE := docker compose -f $(DEV_STACK_DIR)/docker-compose.yaml --project-directory $(DEV_STACK_DIR)

DEPENDENCY_SERVICES := postgres-dev redis-dev minio-dev mailpit-dev
INIT_SERVICES := minio-init
FULL_CONTAINER_PROFILE := full-container
FULL_CONTAINER_SERVICES := backend-dev frontend-dev
DOCTOR_SCRIPT := $(DEV_STACK_DIR)/scripts/dev-doctor.sh
POSTGRES_VOLUME_GUARD := $(DEV_STACK_DIR)/scripts/check-postgres-volume-layout.sh
ENV_EXAMPLE := $(DEV_STACK_DIR)/.env.example
ENV_FILE := $(DEV_STACK_DIR)/.env
HOST_BACKEND_PID := $(REPO_ROOT)/.test-output/local-dev/backend.pid
HOST_FRONTEND_PID := $(REPO_ROOT)/.test-output/local-dev/frontend.pid

.PHONY: up down doctor reset logs pull container-up container-down container-doctor container-logs help _ensure_env _check_container_env _check_postgres_volume _stop_host_runtimes _wait_init _all_healthy

help: ## List dev-stack make targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

_ensure_env:
	@if [ ! -f "$(ENV_FILE)" ]; then \
		echo "[dev-stack] .env not found, copying from $(notdir $(ENV_EXAMPLE))"; \
		cp "$(ENV_EXAMPLE)" "$(ENV_FILE)"; \
	fi

_check_container_env: _ensure_env
	@set -a; source "$(ENV_FILE)"; set +a; \
	missing=(); \
	for key in SESSION_COOKIE_SECRET AUTH_CHALLENGE_TOKEN_PEPPER AI_PROVIDER_BASE_URL AI_PROVIDER_API_KEY; do \
		if [ -z "$${!key:-}" ]; then missing+=("$$key"); fi; \
	done; \
	if [ "$${#missing[@]}" -ne 0 ]; then \
		echo "[dev-stack] full-container runtime requires non-empty local values for: $${missing[*]}" >&2; \
		echo "[dev-stack] set them in $(ENV_FILE); values will not be printed" >&2; \
		exit 1; \
	fi

_check_postgres_volume:
	@if [ -x "$(POSTGRES_VOLUME_GUARD)" ]; then \
		"$(POSTGRES_VOLUME_GUARD)"; \
	fi

_stop_host_runtimes:
	@for pid_file in "$(HOST_BACKEND_PID)" "$(HOST_FRONTEND_PID)"; do \
		if [ ! -s "$$pid_file" ]; then continue; fi; \
		pid="$$(cat "$$pid_file")"; \
		if [ -z "$$pid" ] || ! kill -0 "$$pid" >/dev/null 2>&1; then \
			rm -f "$$pid_file"; \
			continue; \
		fi; \
		echo "[dev-stack] stopping repo-managed host runtime from $${pid_file#$(REPO_ROOT)/}"; \
		kill -TERM "-$$pid" >/dev/null 2>&1 || kill -TERM "$$pid" >/dev/null 2>&1 || true; \
		for _ in $$(seq 1 20); do \
			if ! kill -0 "$$pid" >/dev/null 2>&1; then break; fi; \
			sleep 0.25; \
		done; \
		if kill -0 "$$pid" >/dev/null 2>&1; then \
			kill -KILL "-$$pid" >/dev/null 2>&1 || kill -KILL "$$pid" >/dev/null 2>&1 || true; \
		fi; \
		rm -f "$$pid_file"; \
	done

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

up: _ensure_env _check_postgres_volume ## Start dev stack (idempotent; skip if already healthy)
	@status="$$(make -s -C "$(DEV_STACK_DIR)" _all_healthy)"; \
	if [ "$$status" = "yes" ]; then \
		echo "[dev-stack] already healthy, skipping start"; \
		exit 0; \
	fi; \
	echo "[dev-stack] starting dependency services"; \
	if ! $(COMPOSE) up -d --wait --wait-timeout 60 $(DEPENDENCY_SERVICES); then \
		echo "[dev-stack] compose up --wait failed; scanning host ports for conflicts" >&2; \
		for port in $${POSTGRES_HOST_PORT:-5432} $${REDIS_HOST_PORT:-6379} $${MINIO_API_HOST_PORT:-9000} $${MINIO_CONSOLE_HOST_PORT:-9001} $${MAILPIT_WEB_HOST_PORT:-8025} $${MAILPIT_SMTP_HOST_PORT:-1025}; do \
			if command -v lsof >/dev/null 2>&1; then \
				owners="$$(lsof -nP -iTCP:$$port -sTCP:LISTEN +c 0 2>/dev/null | awk 'NR>1 && $$1 !~ /docker|com.docker|vpnkit|dockerd/ {print}')"; \
				if [ -n "$$owners" ]; then \
					echo "[dev-stack] host port $$port held by non-docker process:" >&2; \
					printf '%s\n' "$$owners" >&2; \
				fi; \
			fi; \
		done; \
		exit 1; \
	fi; \
	echo "[dev-stack] running init jobs"; \
	$(COMPOSE) up -d $(INIT_SERVICES); \
	make -s -C "$(DEV_STACK_DIR)" _wait_init; \
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

container-up: _check_container_env _check_postgres_volume ## Build and start dependencies plus containerized backend/frontend
	@$(MAKE) -s -C "$(DEV_STACK_DIR)" up
	@$(MAKE) -s -C "$(DEV_STACK_DIR)" _stop_host_runtimes
	@echo "[dev-stack] building and starting full-container application services"
	@provider="$${EMAIL_PROVIDER:-$$(awk -F= '$$1 == "EMAIL_PROVIDER" { print $$2; exit }' "$(ENV_FILE)")}"; \
	provider="$$(printf '%s' "$$provider" | tr '[:upper:]' '[:lower:]')"; \
	if [ "$$provider" = "mailpit" ]; then \
		EMAIL_SMTP_HOST=mailpit-dev EMAIL_SMTP_PORT=1025 \
			$(COMPOSE) --profile $(FULL_CONTAINER_PROFILE) up -d --build --wait --wait-timeout 180 $(FULL_CONTAINER_SERVICES); \
	else \
		$(COMPOSE) --profile $(FULL_CONTAINER_PROFILE) up -d --build --wait --wait-timeout 180 $(FULL_CONTAINER_SERVICES); \
	fi
	@$(MAKE) -s -C "$(DEV_STACK_DIR)" container-doctor
	@echo "[dev-stack] frontend: http://127.0.0.1:$${FULL_CONTAINER_FRONTEND_HOST_PORT:-10800}"
	@echo "[dev-stack] backend:  http://127.0.0.1:$${FULL_CONTAINER_API_HOST_PORT:-10801}/api/v1"

container-down: ## Stop the full-container stack; named volumes preserved
	@$(COMPOSE) --profile $(FULL_CONTAINER_PROFILE) down

container-doctor: ## Run the structured health probe for dependencies and application containers
	@$(MAKE) -s -C "$(DEV_STACK_DIR)" doctor

container-logs: ## Tail full-container logs; pass SERVICE=<name> to scope
	@$(COMPOSE) --profile $(FULL_CONTAINER_PROFILE) logs --tail 200 --no-color $(if $(SERVICE),$(SERVICE),$(FULL_CONTAINER_SERVICES))

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
