#!/usr/bin/env bash
# Framework-owned setup entrypoint for the shared local scenario environment.

set -euo pipefail

SCENARIO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCENARIO_DIR/../.." && pwd)"
DRY_RUN=0
WITH_MIGRATIONS=0
LOCAL_DEV_RUNTIME="$SCENARIO_DIR/_shared/scripts/local-dev-runtime.sh"

# shellcheck disable=SC1090
. "$LOCAL_DEV_RUNTIME"

usage() {
  cat <<'USAGE'
Usage: test/scenarios/env-setup.sh [--with-migrations] [--dry-run]

Prepare the shared local scenario environment independently from any scenario.
Default setup starts Docker Compose external dependencies and verifies them.
USAGE
}

run_root() {
  if [ "$DRY_RUN" -eq 1 ]; then
    printf 'dry-run: %s\n' "$*"
    return 0
  fi
  (cd "$REPO_ROOT" && "$@")
}

run_root_shell() {
  if [ "$DRY_RUN" -eq 1 ]; then
    printf 'dry-run: %s\n' "$1"
    return 0
  fi
  (cd "$REPO_ROOT" && bash -lc "$1")
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --with-migrations)
      WITH_MIGRATIONS=1
      ;;
    --dry-run)
      DRY_RUN=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "env-setup: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

run_root make dev-up
run_root make dev-doctor

if [ "$WITH_MIGRATIONS" -eq 1 ]; then
  run_root_shell '
set -euo pipefail
if [ ! -s deploy/dev-stack/.env ]; then
  echo "env-setup: missing deploy/dev-stack/.env; run make dev-up or copy deploy/dev-stack/.env.example" >&2
  exit 1
fi
set -a
. deploy/dev-stack/.env
set +a
DEFAULT_DATABASE_URL="postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable"
if [ -z "${DATABASE_URL:-}" ] || [ "$DATABASE_URL" = "$DEFAULT_DATABASE_URL" ]; then
  export POSTGRES_USER="${POSTGRES_USER:-easyinterview}"
  export POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-dev}"
  export POSTGRES_DB="${POSTGRES_DB:-easyinterview}"
  export POSTGRES_HOST_PORT="${POSTGRES_HOST_PORT:-5432}"
  DATABASE_URL="$(python3 - <<PY
import os
import urllib.parse

def enc(value):
    return urllib.parse.quote(value, safe="")

user = os.environ["POSTGRES_USER"]
password = os.environ["POSTGRES_PASSWORD"]
database = os.environ["POSTGRES_DB"]
port = os.environ["POSTGRES_HOST_PORT"]
print(f"postgres://{enc(user)}:{enc(password)}@localhost:{port}/{enc(database)}?sslmode=disable")
PY
)"
  export DATABASE_URL
fi
make migrate-up
'
fi

if [ "$DRY_RUN" -eq 1 ]; then
  echo "dry-run: print local dev endpoints and debug commands"
else
  local_dev_summary
fi
