#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"

mkdir -p "$OUTPUT_DIR"
export DATABASE_URL="$PG_DSN"

cd "$REPO_ROOT/backend"
go test -v ./cmd/api -run '^TestE2EP0098' -count=1 | tee "$OUTPUT_DIR/trigger.log"
printf 'scenario=E2E.P0.098\nmethod=cmd-api-live-postgres\ntrigger_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/trigger.env"
echo "trigger: ok"
