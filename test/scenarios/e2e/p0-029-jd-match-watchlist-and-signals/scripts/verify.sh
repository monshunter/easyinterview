#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-029-jd-match-watchlist-and-signals"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
grep -Fq 'VITE_EI_API_MODE=real' "$LOG_FILE"
grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1' "$LOG_FILE"

required_specs=(
  'jdMatch.realApiMode.test.ts'
  'WatchlistTab.test.tsx'
  'WatchlistChevron.test.tsx'
  'WatchlistEmpty.test.tsx'
  'MarketSignals.test.tsx'
  'MarketSignalsPartial.test.tsx'
  'WatchlistPrivacy.test.tsx'
)
for spec in "${required_specs[@]}"; do
  if ! grep -Fq "$spec" "$LOG_FILE"; then
    echo "missing required spec in trigger log: $spec" >&2
    exit 1
  fi
done
