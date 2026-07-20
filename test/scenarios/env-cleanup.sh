#!/usr/bin/env bash
# Framework-owned cleanup entrypoint for the shared local scenario environment.

set -euo pipefail

SCENARIO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCENARIO_DIR/../.." && pwd)"
DRY_RUN=0
WITH_VOLUMES=0
LOCAL_DEV_RUNTIME="$SCENARIO_DIR/_shared/scripts/local-dev-runtime.sh"

# shellcheck disable=SC1090
. "$LOCAL_DEV_RUNTIME"

usage() {
  cat <<'USAGE'
Usage: test/scenarios/env-cleanup.sh [--with-volumes|--reset] [--dry-run]

Clean the shared local scenario environment.
Default cleanup stops containers and preserves named volumes.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --with-volumes|--reset)
      WITH_VOLUMES=1
      ;;
    --dry-run)
      DRY_RUN=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "env-cleanup: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

stop_host_runtimes

if [ "$WITH_VOLUMES" -eq 1 ]; then
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "dry-run: DEV_RESET_FORCE=1 make dev-reset"
    exit 0
  fi
  (cd "$REPO_ROOT" && DEV_RESET_FORCE=1 make dev-reset)
else
  if [ "$DRY_RUN" -eq 1 ]; then
    echo "dry-run: make dev-down"
    exit 0
  fi
  (cd "$REPO_ROOT" && make dev-down)
fi
