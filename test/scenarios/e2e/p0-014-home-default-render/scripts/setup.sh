#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-014-home-default-render"
mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR/trigger.log"
run_id="$(uuidgen | tr '[:upper:]' '[:lower:]')"
printf 'scenario=E2E.P0.014\nrun_id=%s\nsetup_at=%s\n' \
  "$run_id" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/setup.env"
