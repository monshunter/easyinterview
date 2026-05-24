#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"

mkdir -p "$OUTPUT_DIR"
export DATABASE_URL="$PG_DSN"
export EI_PLAYWRIGHT_OUTPUT_DIR="$OUTPUT_DIR/playwright"

cd "$REPO_ROOT"
pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts | tee "$OUTPUT_DIR/trigger.log"
printf 'scenario=E2E.P0.099\nmethod=playwright-fullstack-real-backend\ntrigger_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/trigger.env"
echo "trigger: ok"
