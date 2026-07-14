#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-016-parse-confirm-to-workspace"

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"
RUN_ID="$(python3 -c 'import uuid; print(uuid.uuid4())')"
printf 'scenario=E2E.P0.016\nrun_id=%s\nsetup_at=%s\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/setup.env"
rm -f "$OUTPUT_DIR/trigger.log"
