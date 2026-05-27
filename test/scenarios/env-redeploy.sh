#!/usr/bin/env bash
# Framework-owned redeploy entrypoint for host-run local scenario environments.

set -euo pipefail

SCENARIO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCENARIO_DIR/../.." && pwd)"
TARGET="all"
DRY_RUN=0
LOCAL_DEV_RUNTIME="$SCENARIO_DIR/_shared/scripts/local-dev-runtime.sh"

# shellcheck disable=SC1090
. "$LOCAL_DEV_RUNTIME"

usage() {
  cat <<'USAGE'
Usage: test/scenarios/env-redeploy.sh [deps|backend|frontend|all] [--dry-run]

Rebuild/redeploy repo components for the shared local scenario environment.
In the current host-run topology, backend/frontend redeploy rebuilds artifacts,
restarts the matching host-run process, and prints debug endpoints/log paths.
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
    deps|backend|frontend|all)
      TARGET="$1"
      ;;
    --dry-run)
      DRY_RUN=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "env-redeploy: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

redeploy_deps() {
  run_root make dev-up
  run_root make dev-doctor
}

redeploy_backend() {
  run_root_shell "cd backend && go build ./cmd/..."
  if [ "$DRY_RUN" -eq 1 ]; then
    printf 'dry-run: restart backend host-run process\n'
    return 0
  fi
  restart_backend_runtime
}

redeploy_frontend() {
  run_root_shell '
set -euo pipefail
if [ ! -s deploy/dev-stack/.env ]; then
  echo "env-redeploy: missing deploy/dev-stack/.env; run test/scenarios/env-setup.sh or copy deploy/dev-stack/.env.example" >&2
  exit 1
fi
set -a
. deploy/dev-stack/.env
set +a
: "${VITE_EI_API_MODE:?env-redeploy: deploy/dev-stack/.env must set VITE_EI_API_MODE}"
: "${VITE_EI_API_BASE_URL:?env-redeploy: deploy/dev-stack/.env must set VITE_EI_API_BASE_URL}"
pnpm --filter @easyinterview/frontend build
'
  if [ "$DRY_RUN" -eq 1 ]; then
    printf 'dry-run: restart frontend host-run process\n'
    return 0
  fi
  restart_frontend_runtime
}

case "$TARGET" in
  deps)
    redeploy_deps
    ;;
  backend)
    redeploy_backend
    ;;
  frontend)
    redeploy_frontend
    ;;
  all)
    redeploy_deps
    redeploy_backend
    redeploy_frontend
    ;;
esac

if [ "$DRY_RUN" -eq 1 ]; then
  echo "dry-run: print local dev endpoints and debug commands"
else
  local_dev_summary
fi
