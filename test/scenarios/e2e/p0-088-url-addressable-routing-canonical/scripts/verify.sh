#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-088-url-addressable-routing-canonical"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "source_contract_test.py" "$SCRIPT_DIR/trigger.sh"
grep -Fq "Ran 2 tests" "$LOG_FILE"
grep -Fq "OK" "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-088-url-addressable-routing-canonical.test.tsx" "$LOG_FILE"
grep -Fq "src/app/outOfScopeRouteNegative.test.ts" "$LOG_FILE"
grep -Fq "direct-open /reports with hostile legacy params" "$LOG_FILE"
grep -Fq "reload preserves the canonical Reports target context" "$LOG_FILE"
grep -Fq "back/forward restores Reports with targetJobId only" "$LOG_FILE"
grep -Fq "direct-opens Reports with targetJobId only and keeps chrome visible" "$LOG_FILE"
grep -Fq "replaces an untrusted Reports deep link with workspace without adding a back-loop" "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"

for forbidden in \
  'route-welcome' \
  'route-voice' \
  'topbar-nav-mistakes' \
  'topbar-nav-growth' \
  'topbar-nav-drill' \
  'topbar-nav-voice' \
  '/voice?' \
  '/welcome?' \
  '/growth?' \
  '/mistakes?' \
  '/drill?' \
  '/debrief?' \
  '/profile?' \
  'debrief-screen' \
  'route-profile' \
  'ui-design/src/data'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden out-of-scope entry leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done
