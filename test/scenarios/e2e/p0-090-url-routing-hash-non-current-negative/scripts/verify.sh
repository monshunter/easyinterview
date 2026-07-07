#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-090-url-routing-hash-non-current-negative"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-090-url-routing-hash-non-current-negative.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +10 passed \(10\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

# Non-current-route grep: ROUTE_TO_PATH must not enumerate non-current aliases.
ROUTE_FILE="$REPO_ROOT/frontend/src/app/routeUrl.ts"
test -s "$ROUTE_FILE"
for non_current_path in '/voice' '/welcome' '/growth' '/mistakes' '/drill' '/followup' '/experiences' '/star' '/onboarding' '/debrief' '/profile'; do
  if grep -E "^\\s*[a-zA-Z_]+: \"$non_current_path\"," "$ROUTE_FILE" >/dev/null; then
    echo "routeUrl.ROUTE_TO_PATH leaked non-current path: $non_current_path" >&2
    exit 1
  fi
done

# Non-current-screen grep: no standalone screen file for non-current aliases.
SCREEN_DIR="$REPO_ROOT/frontend/src/app/screens"
for non_current_alias in 'welcome' 'growth' 'mistakes' 'drill' 'followup' 'experiences' 'star' 'onboarding' 'debrief' 'profile'; do
  if find "$SCREEN_DIR" -mindepth 1 -maxdepth 2 -type d -iname "$non_current_alias" 2>/dev/null | grep -q .; then
    echo "non-current screen directory still present: $non_current_alias" >&2
    exit 1
  fi
done
