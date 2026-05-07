#!/bin/sh
# Owned by docs/spec/local-dev-stack/plans/001-bootstrap (Phase 3).
# Structured health probe for the easyinterview local dev stack.
#
# Output contract (spec D-6):
#   {
#     "services": [{"name","type":"dependency|app","status":"OK|DEGRADED|DOWN","reason"}],
#     "summary":  {"ok","degraded","down","total"}
#   }
# Exit 0 iff summary.down == 0 && summary.degraded == 0.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
DEV_STACK_DIR=$(CDPATH= cd -- "${SCRIPT_DIR}/.." && pwd)
COMPOSE_FILE="${DEV_STACK_DIR}/docker-compose.yaml"
PROJECT="${COMPOSE_PROJECT_NAME:-easyinterview-dev}"
ROLE_LABEL="easyinterview.dev-stack.role"
HEALTHZ_LABEL="easyinterview.dev-stack.healthz"
METRICS_LABEL="easyinterview.dev-stack.metrics"
AICLIENT_LABEL="easyinterview.dev-stack.aiclient"
HOSTPORT_LABEL="easyinterview.dev-stack.host-port"

if [ -f "${DEV_STACK_DIR}/.env" ]; then set -a; . "${DEV_STACK_DIR}/.env"; set +a; fi

POSTGRES_USER="${POSTGRES_USER:-easyinterview}"
POSTGRES_DB="${POSTGRES_DB:-easyinterview}"
POSTGRES_HOST_PORT="${POSTGRES_HOST_PORT:-5432}"
REDIS_HOST_PORT="${REDIS_HOST_PORT:-6379}"
MINIO_API_HOST_PORT="${MINIO_API_HOST_PORT:-9000}"
OBJECT_STORAGE_BUCKET="${OBJECT_STORAGE_BUCKET:-easyinterview-dev}"
OBJECT_STORAGE_ACCESS_KEY="${OBJECT_STORAGE_ACCESS_KEY:-dev-access-key}"
OBJECT_STORAGE_SECRET_KEY="${OBJECT_STORAGE_SECRET_KEY:-dev-secret-key}"

command -v jq >/dev/null 2>&1 || { echo "dev-doctor: jq is required" >&2; exit 2; }

compose() { docker compose -f "${COMPOSE_FILE}" --project-directory "${DEV_STACK_DIR}" "$@"; }
container_state() { compose ps -a --format '{{.State}}/{{.Health}}' "$1" 2>/dev/null; }

emit() {
  jq -nc --arg n "$1" --arg t "$2" --arg s "$3" --arg r "$4" \
    '{name:$n,type:$t,status:$s,reason:$r}'
}

port_owner() {
  command -v lsof >/dev/null 2>&1 || { echo ""; return; }
  lsof -nP -iTCP:"$1" -sTCP:LISTEN +c 0 2>/dev/null \
    | awk 'NR>1 {printf "pid=%s cmd=%s|", $2, $1}' | sed 's/|$//' | head -c 200
}

down_or_conflict() {
  state=${4:-missing}
  case "$state" in
    running/starting) emit "$1" "$2" DEGRADED "container still starting"; return;;
  esac
  owner=$(port_owner "$3")
  case "$owner" in
    ""|*com.docker*|*Docker*|*docker*|*vpnkit*|*dockerd*) emit "$1" "$2" DOWN "container state=$state";;
    *) emit "$1" "$2" DOWN "port conflict: host port $3 held by $owner";;
  esac
}

probe_postgres() {
  state=$(container_state postgres-dev)
  case "$state" in
    running/healthy) ;;
    *) down_or_conflict postgres-dev dependency "$POSTGRES_HOST_PORT" "$state"; return ;;
  esac
  if ! compose exec -T postgres-dev pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
    emit postgres-dev dependency DEGRADED "pg_isready failed"; return
  fi
  if ! compose exec -T postgres-dev psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -tAc "select 1" 2>/dev/null | tr -d '\r\n ' | grep -qx '1'; then
    emit postgres-dev dependency DEGRADED "select 1 failed"; return
  fi
  emit postgres-dev dependency OK ""
}

probe_redis() {
  state=$(container_state redis-dev)
  case "$state" in
    running/healthy) ;;
    *) down_or_conflict redis-dev dependency "$REDIS_HOST_PORT" "$state"; return ;;
  esac
  out=$(compose exec -T redis-dev sh -c 'redis-cli set __doctor__ ok EX 5 >/dev/null && redis-cli get __doctor__ && redis-cli del __doctor__ >/dev/null' 2>/dev/null | tr -d '\r\n ' || true)
  if [ "$out" != "ok" ]; then
    emit redis-dev dependency DEGRADED "set/get/del probe failed"; return
  fi
  emit redis-dev dependency OK ""
}

probe_minio() {
  state=$(container_state minio-dev)
  case "$state" in
    running/healthy) ;;
    *) down_or_conflict minio-dev dependency "$MINIO_API_HOST_PORT" "$state"; return ;;
  esac
  if ! out=$(compose run --rm --no-deps \
      -e MC_HOST_local="http://${OBJECT_STORAGE_ACCESS_KEY}:${OBJECT_STORAGE_SECRET_KEY}@minio-dev:9000" \
      --entrypoint mc minio-init ls "local/${OBJECT_STORAGE_BUCKET}" 2>&1); then
    reason=$(printf '%s' "$out" | tr '\n' ' ' | sed 's/  */ /g' | head -c 200)
    emit minio-dev dependency DEGRADED "mc ls failed: ${reason}"; return
  fi
  emit minio-dev dependency OK ""
}

list_app_services() {
  docker ps -a \
    --filter "label=com.docker.compose.project=${PROJECT}" \
    --filter "label=${ROLE_LABEL}=app" \
    --format '{{.Names}}' 2>/dev/null
}

label_value() {
  docker inspect --format "{{ index .Config.Labels \"$2\" }}" "$1" 2>/dev/null || true
}

env_value() {
  docker inspect --format '{{range .Config.Env}}{{println .}}{{end}}' "$1" 2>/dev/null \
    | awk -F= -v k="$2" '$1==k {sub(/^[^=]*=/,""); print; exit}'
}

probe_app() {
  container=$1
  service=$(label_value "$container" "com.docker.compose.service")
  : "${service:=$container}"
  state=$(docker inspect --format '{{.State.Status}}/{{if .State.Health}}{{.State.Health.Status}}{{else}}n/a{{end}}' "$container" 2>/dev/null || echo missing)
  case "$state" in
    running/healthy|running/n/a) ;;
    *) emit "$service" app DOWN "container state=$state"; return ;;
  esac
  if [ "$(label_value "$container" "$AICLIENT_LABEL")" = "true" ]; then
    base=$(env_value "$container" AI_PROVIDER_BASE_URL)
    key=$(env_value "$container" AI_PROVIDER_API_KEY)
    if [ -z "$base" ] || [ -z "$key" ]; then
      emit "$service" app DOWN "missing real AI provider config: AI_PROVIDER_BASE_URL=${base:+set}${base:-empty} AI_PROVIDER_API_KEY=${key:+set}${key:-empty}"; return
    fi
  fi
  port=$(label_value "$container" "$HOSTPORT_LABEL")
  healthz=$(label_value "$container" "$HEALTHZ_LABEL")
  metrics=$(label_value "$container" "$METRICS_LABEL")
  if [ -n "$port" ] && [ -n "$healthz" ]; then
    if ! curl -fsS --max-time 3 "http://localhost:${port}${healthz}" >/dev/null 2>&1; then
      emit "$service" app DEGRADED "GET ${healthz} on host port ${port} failed"; return
    fi
  fi
  if [ -n "$port" ] && [ -n "$metrics" ]; then
    body=$(curl -fsS --max-time 3 "http://localhost:${port}${metrics}" 2>/dev/null || true)
    if [ -z "$body" ]; then
      emit "$service" app DEGRADED "GET ${metrics} returned empty body"; return
    fi
  fi
  emit "$service" app OK ""
}

{
  probe_postgres
  probe_redis
  probe_minio
  for c in $(list_app_services); do probe_app "$c"; done
} | jq -s '{
  services: .,
  summary: {
    ok:       map(select(.status=="OK")) | length,
    degraded: map(select(.status=="DEGRADED")) | length,
    down:     map(select(.status=="DOWN")) | length,
    total:    length
  }
}' | tee /tmp/dev-doctor.$$ >/dev/null
result=$(cat /tmp/dev-doctor.$$)
rm -f /tmp/dev-doctor.$$
printf '%s\n' "$result"
down=$(printf '%s' "$result" | jq -r '.summary.down')
degraded=$(printf '%s' "$result" | jq -r '.summary.degraded')
[ "$down" -eq 0 ] && [ "$degraded" -eq 0 ] || exit 1
