#!/usr/bin/env bash
# Shared helpers for host-run backend/frontend local runtime management.

set -euo pipefail

LOCAL_DEV_OUTPUT_DIR="${LOCAL_DEV_OUTPUT_DIR:-$REPO_ROOT/.test-output/local-dev}"

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

api_port() {
  local listen_addr="${APP_LISTEN_ADDR:-:8080}"
  local port="${listen_addr##*:}"
  printf '%s\n' "${API_HOST_PORT:-$port}"
}

backend_listen_addr() {
  local listen_addr="${APP_LISTEN_ADDR:-:8080}"
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
  printf '%s\n' "${FRONTEND_HOST_PORT:-5173}"
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
  if [ ! -s "$pid_file" ]; then
    return 0
  fi
  local pid
  pid="$(cat "$pid_file")"
  if [ -z "$pid" ] || ! kill -0 "$pid" >/dev/null 2>&1; then
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

stop_port_listeners() {
  local port="$1"
  local pids
  pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  if [ -z "$pids" ]; then
    return 0
  fi
  echo "[local-dev] stopping port $port listeners: $pids"
  kill $pids >/dev/null 2>&1 || true
  sleep 1
  pids="$(lsof -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)"
  if [ -n "$pids" ]; then
    echo "[local-dev] force stopping port $port listeners: $pids"
    kill -KILL $pids >/dev/null 2>&1 || true
  fi
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
  mkdir -p "$LOCAL_DEV_OUTPUT_DIR"
  local port log_file pid_file
  port="$(api_port)"
  log_file="$LOCAL_DEV_OUTPUT_DIR/backend.log"
  pid_file="$LOCAL_DEV_OUTPUT_DIR/backend.pid"

  stop_pidfile_process_group "$pid_file"
  stop_port_listeners "$port"
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

  stop_pidfile_process_group "$pid_file"
  stop_port_listeners "$port"
  echo "[local-dev] starting frontend: pnpm --filter @easyinterview/frontend dev --host 127.0.0.1"
  start_detached "$REPO_ROOT/frontend" "$log_file" "$pid_file" pnpm --filter @easyinterview/frontend dev --host 127.0.0.1 >/dev/null
  wait_for_tcp_port "frontend" "$port" "$log_file"
}
