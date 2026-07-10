#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-005-app-shell-visual-system-smoke"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-005-app-shell-visual-system-smoke.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +8 passed \(8\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

# Negative: scenario evidence must not surface out-of-scope-module testid leakage.
for forbidden in \
  'route-welcome' \
  'topbar-nav-mistakes' \
  'topbar-nav-growth' \
  'topbar-nav-drill' \
  'topbar-nav-voice'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden out-of-scope entry leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done
