#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery"
DATABASE_STATE="$OUTPUT_DIR/isolated-database.env"

mkdir -p "$OUTPUT_DIR"
if [ -s "$DATABASE_STATE" ]; then
  echo "setup: isolated database state remains; run scripts/cleanup.sh first" >&2
  exit 1
fi
rm -rf "$OUTPUT_DIR/playwright" "$OUTPUT_DIR/screenshots"
rm -f "$OUTPUT_DIR"/*.log "$OUTPUT_DIR/result.json" "$OUTPUT_DIR/setup.env" \
  "$OUTPUT_DIR/source-fingerprint.json" "$OUTPUT_DIR/source-fingerprint.verify.json"

RUN_ID="$(python3 -c 'import uuid; print(uuid.uuid4())')"
SETUP_EPOCH="$(date '+%s')"
printf 'scenario=E2E.P0.046\nrun_id=%s\nsetup_at=%s\nsetup_epoch=%s\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$SETUP_EPOCH" > "$OUTPUT_DIR/setup.env"
