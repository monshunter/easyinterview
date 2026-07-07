#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-001-default-home-shell"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-001-default-home-shell.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +1 passed \(1\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

for forbidden in \
  'route-welcome' \
  'topbar-nav-mistakes' \
  'topbar-nav-growth' \
  'topbar-nav-drill' \
  'topbar-nav-voice' \
  'topbar-nav-debrief' \
  'topbar-nav-profile' \
  'topbar-user-profile'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden non-current entry leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done
