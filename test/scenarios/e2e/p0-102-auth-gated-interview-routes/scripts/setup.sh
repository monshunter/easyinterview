#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-102-auth-gated-interview-routes"

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

cat > "$OUTPUT_DIR/setup.json" <<'JSON'
{"scenario":"E2E.P0.102","status":"ready","isolation":"local runner"}
JSON
