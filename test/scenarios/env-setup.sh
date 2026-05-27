#!/usr/bin/env bash
# Framework-owned setup entrypoint for the shared local scenario environment.

set -euo pipefail

SCENARIO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCENARIO_DIR/../.." && pwd)"
DRY_RUN=0
WITH_MIGRATIONS=0
DEFAULT_DATABASE_URL="postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable"

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
  run_root_shell "DATABASE_URL=\"\${DATABASE_URL:-$DEFAULT_DATABASE_URL}\" make migrate-up"
fi
