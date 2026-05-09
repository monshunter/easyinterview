#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-018-workspace-default-render"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.018: no passing test files found" >&2; exit 1; }
grep -q "testid" "$LOG_FILE" || echo "E2E.P0.018: warning - no testid assertions in log" >&2
for forbidden in 'practice-mode-card-' 'growth-center' 'drill-builder' 'mistake-queue'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then echo "E2E.P0.018: forbidden legacy testid $forbidden leaked" >&2; exit 1; fi
done
echo "E2E.P0.018 PASS"
