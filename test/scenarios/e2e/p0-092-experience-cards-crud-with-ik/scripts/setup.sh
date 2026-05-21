#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-082-experience-cards-crud-with-ik"

mkdir -p "$OUTPUT_DIR"
printf 'scenario=E2E.P0.082\nsetup_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/setup.env"

# Verify dev stack Postgres is reachable.
PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"
if ! psql "$PG_DSN" -c "select 1" >/dev/null 2>&1; then
  echo "setup: dev-stack postgres unreachable at $PG_DSN" >&2
  exit 1
fi
