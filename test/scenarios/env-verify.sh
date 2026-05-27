#!/usr/bin/env bash
# Framework-owned verification entrypoint for the shared local scenario environment.

set -euo pipefail

SCENARIO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCENARIO_DIR/../.." && pwd)"
DRY_RUN=0

usage() {
  cat <<'USAGE'
Usage: test/scenarios/env-verify.sh [--dry-run]

Verify that the shared local scenario environment is ready.
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "env-verify: unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

if [ "$DRY_RUN" -eq 1 ]; then
  echo "dry-run: make dev-doctor"
  exit 0
fi

(cd "$REPO_ROOT" && make dev-doctor)
