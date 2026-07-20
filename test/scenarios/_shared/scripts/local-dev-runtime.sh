#!/usr/bin/env bash
# Shared helpers for host-run backend/frontend local runtime management.

set -euo pipefail

LOCAL_DEV_OUTPUT_DIR="${LOCAL_DEV_OUTPUT_DIR:-$REPO_ROOT/.test-output/local-dev}"
LOCAL_DEV_COMPOSE_FILE="$REPO_ROOT/deploy/dev-stack/docker-compose.yaml"
LOCAL_DEV_COMPOSE_DIR="$REPO_ROOT/deploy/dev-stack"

secure_dev_stack_env() {
  local env_file="$REPO_ROOT/deploy/dev-stack/.env"
  if [ ! -s "$env_file" ]; then
    echo "local-dev-runtime: missing $env_file; run test/scenarios/env-setup.sh first" >&2
    return 1
  fi
  python3 - "$env_file" <<'PY'
import os
import stat
import sys

path = sys.argv[1]
os.chmod(path, 0o600)
mode = stat.S_IMODE(os.stat(path).st_mode)
if mode != 0o600:
    raise SystemExit("local-dev-runtime: dev-stack env permission hardening failed")
PY
}

load_dev_stack_env() {
  local env_file="$REPO_ROOT/deploy/dev-stack/.env"
  secure_dev_stack_env
  set -a
  # shellcheck disable=SC1090
  . "$env_file"
  set +a
}

assert_host_mailpit_smtp_route() {
  local provider="${EMAIL_PROVIDER:-}"
  local host="${EMAIL_SMTP_HOST:-}"
  local backend_port="${EMAIL_SMTP_PORT:-1025}"
  local mailpit_host_port="${MAILPIT_SMTP_HOST_PORT:-1025}"

  case "$provider:$host" in
    mailpit:127.0.0.1|mailpit:localhost|mailpit:::1) ;;
    *) return 0 ;;
  esac
  if [ "$backend_port" != "$mailpit_host_port" ]; then
    echo "local-dev-runtime: host Mailpit SMTP route mismatch: EMAIL_SMTP_PORT=$backend_port but MAILPIT_SMTP_HOST_PORT=$mailpit_host_port" >&2
    return 1
  fi
}

secure_raw_capture_path() {
  case "${AI_DEBUG_CAPTURE_RAW_IO:-}" in
    true|TRUE|1) ;;
    *) return 0 ;;
  esac
  if [ -z "${AI_DEBUG_RAW_IO_PATH:-}" ]; then
    echo "local-dev-runtime: AI raw capture is enabled but its dedicated path is empty" >&2
    return 1
  fi
  python3 - "$REPO_ROOT" "$AI_DEBUG_RAW_IO_PATH" <<'PY'
import os
import stat
import sys
from pathlib import Path

root = Path(sys.argv[1])
configured = Path(sys.argv[2])
config_dir = (root / "config").resolve(strict=True)
anchor = config_dir.parent
candidate = configured if configured.is_absolute() else anchor / configured
candidate = Path(os.path.abspath(candidate))

current = Path(candidate.anchor)
for component in candidate.parts[1:]:
    current /= component
    try:
        mode = os.lstat(current).st_mode
    except FileNotFoundError:
        continue
    if stat.S_ISLNK(mode):
        raise SystemExit("local-dev-runtime: raw capture path contains a symlink component")

parent = candidate.parent
if os.path.ismount(parent):
    raise SystemExit("local-dev-runtime: raw capture parent must not be a filesystem or volume root")
parent.mkdir(parents=True, exist_ok=True)
if stat.S_ISLNK(os.lstat(parent).st_mode) or not parent.is_dir():
    raise SystemExit("local-dev-runtime: raw capture parent must be a regular directory")
os.chmod(parent, 0o700)
try:
    target_mode = os.lstat(candidate).st_mode
except FileNotFoundError:
    target_mode = None
if target_mode is not None:
    if stat.S_ISLNK(target_mode) or not stat.S_ISREG(target_mode):
        raise SystemExit("local-dev-runtime: raw capture target must be a regular file")
    os.chmod(candidate, 0o600)
PY
}

stop_full_container_app_role() {
  local role="$1"
  case "$role" in
    backend-dev)
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        echo "dry-run: docker compose rm -sf backend-dev (dependencies and volumes preserved)"
      else
        docker compose -f "$LOCAL_DEV_COMPOSE_FILE" --project-directory "$LOCAL_DEV_COMPOSE_DIR" --profile full-container rm -sf backend-dev
      fi
      ;;
    frontend-dev)
      if [ "${DRY_RUN:-0}" -eq 1 ]; then
        echo "dry-run: docker compose rm -sf frontend-dev (dependencies and volumes preserved)"
      else
        docker compose -f "$LOCAL_DEV_COMPOSE_FILE" --project-directory "$LOCAL_DEV_COMPOSE_DIR" --profile full-container rm -sf frontend-dev
      fi
      ;;
    *)
      echo "local-dev-runtime: unsupported app role: $role" >&2
      return 2
      ;;
  esac
}

compose_app_role_running() {
  local role="$1" running
  if ! running="$(docker compose -f "$LOCAL_DEV_COMPOSE_FILE" --project-directory "$LOCAL_DEV_COMPOSE_DIR" --profile full-container ps --status running --services "$role" 2>/dev/null)"; then
    echo "local-dev-runtime: unable to inspect full-container role $role" >&2
    return 2
  fi
  [ "$running" = "$role" ]
}

pid_command_matches_role() {
  local pid="$1" role="$2" command
  command="$(ps -p "$pid" -o command= 2>/dev/null || true)"
  case "$role:$command" in
    backend-dev:*"go run ./backend/cmd/api"*|backend-dev:*"easyinterview-api"*) return 0 ;;
    frontend-dev:*"pnpm --filter @easyinterview/frontend dev"*|frontend-dev:*"vite"*) return 0 ;;
    *) return 1 ;;
  esac
}

process_cwd_is_repo_owned() {
  local pid="$1" cwd
  cwd="$(readlink "/proc/$pid/cwd" 2>/dev/null || true)"
  if [ -z "$cwd" ] && command -v lsof >/dev/null 2>&1; then
    cwd="$(lsof -a -p "$pid" -d cwd -Fn 2>/dev/null | sed -n 's/^n//p' | head -n 1)"
  fi
  case "$cwd" in
    "$REPO_ROOT"|"$REPO_ROOT"/*) return 0 ;;
    *) return 1 ;;
  esac
}

host_app_role_running() {
  local role="$1" pid_file port pid pids candidates
  case "$role" in
    backend-dev)
      pid_file="$LOCAL_DEV_OUTPUT_DIR/backend.pid"
      port="$(api_port)"
      ;;
    frontend-dev)
      pid_file="$LOCAL_DEV_OUTPUT_DIR/frontend.pid"
      port="$(frontend_port)"
      ;;
    *) return 2 ;;
  esac

  if [ -s "$pid_file" ]; then
    pid="$(cat "$pid_file")"
    if [[ "$pid" =~ ^[0-9]+$ ]] && kill -0 "$pid" >/dev/null 2>&1 && pid_command_matches_role "$pid" "$role"; then
      return 0
    fi
  fi

  if command -v lsof >/dev/null 2>&1; then
    pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null | head -n 32 || true)"
    while IFS= read -r pid; do
      [ -n "$pid" ] || continue
      if pid_command_matches_role "$pid" "$role" || process_cwd_is_repo_owned "$pid"; then
        return 0
      fi
    done <<< "$pids"
  fi

  case "$role" in
    backend-dev)
      candidates="$(ps -axo pid=,command= 2>/dev/null | awk '/go run \.\/backend\/cmd\/api|easyinterview-api/ {print $1}' | head -n 32 || true)"
      ;;
    frontend-dev)
      candidates="$(ps -axo pid=,command= 2>/dev/null | awk '/pnpm .*@easyinterview\/frontend .*dev|vite .*--host/ {print $1}' | head -n 32 || true)"
      ;;
  esac
  while IFS= read -r pid; do
    [ -n "$pid" ] || continue
    if process_cwd_is_repo_owned "$pid"; then
      return 0
    fi
  done <<< "$candidates"
  return 1
}

assert_single_app_runners() {
  local role compose_state host_state conflicts=()
  load_dev_stack_env
  for role in backend-dev frontend-dev; do
    compose_state=0
    host_state=0
    compose_app_role_running "$role" || compose_state="$?"
    if [ "$compose_state" -eq 2 ]; then
      return 1
    fi
    host_app_role_running "$role" || host_state="$?"
    if [ "$host_state" -gt 1 ]; then
      return 1
    fi
    if [ "$compose_state" -eq 0 ] && [ "$host_state" -eq 0 ]; then
      conflicts+=("$role")
    fi
  done
  if [ "${#conflicts[@]}" -ne 0 ]; then
    echo "local-dev-runtime: runner conflict: host and full-container app coexist for role(s): ${conflicts[*]}" >&2
    return 1
  fi
}

api_port() {
  local listen_addr="${APP_LISTEN_ADDR:-:10901}"
  local port="${listen_addr##*:}"
  printf '%s\n' "${API_HOST_PORT:-$port}"
}

backend_listen_addr() {
  local listen_addr="${APP_LISTEN_ADDR:-:10901}"
  local port
  port="$(api_port)"

  case "$listen_addr" in
    :*|0.0.0.0:*)
      printf '127.0.0.1:%s\n' "$port"
      ;;
    *)
      printf '%s\n' "$listen_addr"
      ;;
  esac
}

frontend_port() {
  printf '%s\n' "${FRONTEND_HOST_PORT:-10900}"
}

mailpit_web_port() {
  printf '%s\n' "${MAILPIT_WEB_HOST_PORT:-8025}"
}

minio_console_port() {
  printf '%s\n' "${MINIO_CONSOLE_HOST_PORT:-9001}"
}

local_dev_summary() {
  load_dev_stack_env
  local api frontend mailpit minio_console
  api="$(api_port)"
  frontend="$(frontend_port)"
  mailpit="$(mailpit_web_port)"
  minio_console="$(minio_console_port)"

  cat <<SUMMARY
[local-dev] endpoints
  frontend dev: http://127.0.0.1:${frontend}/
  backend API:  http://127.0.0.1:${api}/api/v1
  Mailpit:      http://127.0.0.1:${mailpit}
  MinIO UI:     http://127.0.0.1:${minio_console}
[local-dev] logs
  backend:      tail -f .test-output/local-dev/backend.log
  frontend:     tail -f .test-output/local-dev/frontend.log
  containers:   make dev-logs SERVICE=<postgres-dev|redis-dev|minio-dev|mailpit-dev>
[local-dev] process files
  backend pid:  .test-output/local-dev/backend.pid
  frontend pid: .test-output/local-dev/frontend.pid
[local-dev] redeploy
  all:          test/scenarios/env-redeploy.sh all
  backend:      test/scenarios/env-redeploy.sh backend
  frontend:     test/scenarios/env-redeploy.sh frontend
SUMMARY
}

stop_pidfile_process_group() {
  local pid_file="$1"
  local role="$2"
  if [ ! -s "$pid_file" ]; then
    return 0
  fi
  local pid
  pid="$(cat "$pid_file")"
  if [ -z "$pid" ] || ! [[ "$pid" =~ ^[0-9]+$ ]] || ! kill -0 "$pid" >/dev/null 2>&1; then
    rm -f "$pid_file"
    return 0
  fi
  if ! pid_command_matches_role "$pid" "$role" || ! process_cwd_is_repo_owned "$pid"; then
    echo "[local-dev] removing stale pidfile without stopping unowned process: ${pid_file#$REPO_ROOT/}" >&2
    rm -f "$pid_file"
    return 0
  fi
  kill -TERM "-$pid" >/dev/null 2>&1 || kill -TERM "$pid" >/dev/null 2>&1 || true
  for _ in $(seq 1 20); do
    if ! kill -0 "$pid" >/dev/null 2>&1; then
      rm -f "$pid_file"
      return 0
    fi
    sleep 0.25
  done
  kill -KILL "-$pid" >/dev/null 2>&1 || kill -KILL "$pid" >/dev/null 2>&1 || true
  rm -f "$pid_file"
}

stop_host_runtimes() {
  if [ "${DRY_RUN:-0}" -eq 1 ]; then
    echo "dry-run: stop repo-managed backend/frontend host-run processes"
    return 0
  fi
  stop_pidfile_process_group "$LOCAL_DEV_OUTPUT_DIR/backend.pid" backend-dev
  stop_pidfile_process_group "$LOCAL_DEV_OUTPUT_DIR/frontend.pid" frontend-dev
}

assert_port_available() {
  local port="$1"
  local pids
  pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  if [ -z "$pids" ]; then
    return 0
  fi
  echo "[local-dev] port $port is held by an unowned or manual process; refusing to stop it" >&2
  return 1
}

start_detached() {
  local cwd="$1"
  local log_file="$2"
  local pid_file="$3"
  shift 3

  umask 077
  mkdir -p "$(dirname "$log_file")"
  : > "$log_file"
  chmod 600 "$log_file"
  python3 - "$cwd" "$log_file" "$pid_file" "$@" <<'PY'
import os
import subprocess
import sys

cwd, log_file, pid_file, *cmd = sys.argv[1:]
os.chmod(log_file, 0o600)
with open(log_file, "ab", buffering=0) as log:
    proc = subprocess.Popen(
        cmd,
        cwd=cwd,
        env=os.environ.copy(),
        stdin=subprocess.DEVNULL,
        stdout=log,
        stderr=subprocess.STDOUT,
        start_new_session=True,
    )
with open(pid_file, "w", encoding="utf-8") as handle:
    handle.write(f"{proc.pid}\n")
os.chmod(pid_file, 0o600)
print(proc.pid)
PY
}

wait_for_tcp_port() {
  local name="$1"
  local port="$2"
  local log_file="$3"
  for _ in $(seq 1 60); do
    if (echo >/dev/tcp/127.0.0.1/"$port") >/dev/null 2>&1; then
      echo "[local-dev] $name is listening on 127.0.0.1:$port"
      return 0
    fi
    sleep 0.5
  done
  echo "[local-dev] $name did not listen on 127.0.0.1:$port within 30s" >&2
  echo "----- $name log tail -----" >&2
  tail -80 "$log_file" >&2 || true
  return 1
}

restart_backend_runtime() {
  load_dev_stack_env
  assert_host_mailpit_smtp_route
  secure_raw_capture_path
  mkdir -p "$LOCAL_DEV_OUTPUT_DIR"
  local port log_file pid_file
  port="$(api_port)"
  log_file="$LOCAL_DEV_OUTPUT_DIR/backend.log"
  pid_file="$LOCAL_DEV_OUTPUT_DIR/backend.pid"

  stop_pidfile_process_group "$pid_file" backend-dev
  assert_port_available "$port"
  APP_LISTEN_ADDR="$(backend_listen_addr)"
  export APP_LISTEN_ADDR
  echo "[local-dev] starting backend: APP_LISTEN_ADDR=$APP_LISTEN_ADDR go run ./backend/cmd/api -config-dir config"
  start_detached "$REPO_ROOT" "$log_file" "$pid_file" go run ./backend/cmd/api -config-dir config >/dev/null
  wait_for_tcp_port "backend" "$port" "$log_file"
}

restart_frontend_runtime() {
  load_dev_stack_env
  mkdir -p "$LOCAL_DEV_OUTPUT_DIR"
  local port log_file pid_file
  port="$(frontend_port)"
  log_file="$LOCAL_DEV_OUTPUT_DIR/frontend.log"
  pid_file="$LOCAL_DEV_OUTPUT_DIR/frontend.pid"

  stop_pidfile_process_group "$pid_file" frontend-dev
  assert_port_available "$port"
  echo "[local-dev] starting frontend: pnpm --filter @easyinterview/frontend dev --host 127.0.0.1"
  start_detached "$REPO_ROOT/frontend" "$log_file" "$pid_file" pnpm --filter @easyinterview/frontend dev --host 127.0.0.1 >/dev/null
  wait_for_tcp_port "frontend" "$port" "$log_file"
}
