#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"

mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR/trigger.log" "$OUTPUT_DIR/trigger.env"
printf 'scenario=E2E.P0.098\nsetup_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/setup.env"

if ! psql "$PG_DSN" -c "select 1" >/dev/null 2>&1; then
  echo "setup: dev-stack postgres unreachable at $PG_DSN" >&2
  exit 1
fi

(cd "$REPO_ROOT" && DATABASE_URL="$PG_DSN" make migrate-status >/dev/null)

echo "setup: ok"
