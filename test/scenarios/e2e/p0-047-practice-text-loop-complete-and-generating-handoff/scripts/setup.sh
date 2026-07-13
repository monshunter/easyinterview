#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
mkdir -p "$OUTPUT_DIR"
RUN_CORRELATION_ID="$(python3 -c 'import uuid; print(uuid.uuid4())')"
printf 'scenario=E2E.P0.047\nsetup_at=%s\nrun_correlation_id=%s\n' \
  "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$RUN_CORRELATION_ID" > "$OUTPUT_DIR/setup.env"
