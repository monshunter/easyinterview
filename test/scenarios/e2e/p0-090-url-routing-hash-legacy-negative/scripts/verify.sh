#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-090-url-routing-hash-legacy-negative"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-090-url-routing-hash-legacy-negative.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +10 passed \(10\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

# Retired-route grep: ROUTE_TO_PATH must not enumerate retired aliases.
ROUTE_FILE="$REPO_ROOT/frontend/src/app/routeUrl.ts"
test -s "$ROUTE_FILE"
for retired in '/voice' '/welcome' '/growth' '/mistakes' '/drill' '/followup' '/experiences' '/star' '/onboarding' '/debrief' '/profile'; do
  if grep -E "^\\s*[a-zA-Z_]+: \"$retired\"," "$ROUTE_FILE" >/dev/null; then
    echo "routeUrl.ROUTE_TO_PATH leaked retired path: $retired" >&2
    exit 1
  fi
done

# Retired-screen grep: no standalone screen file for retired aliases.
SCREEN_DIR="$REPO_ROOT/frontend/src/app/screens"
for legacy in 'welcome' 'growth' 'mistakes' 'drill' 'followup' 'experiences' 'star' 'onboarding' 'debrief' 'profile'; do
  if find "$SCREEN_DIR" -mindepth 1 -maxdepth 2 -type d -iname "$legacy" 2>/dev/null | grep -q .; then
    echo "retired screen directory still present: $legacy" >&2
    exit 1
  fi
done
