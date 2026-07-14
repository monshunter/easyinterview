#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-090-url-routing-hash-out-of-scope-negative"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "source_contract_test.py" "$SCRIPT_DIR/trigger.sh"
grep -Fq "Ran 2 tests" "$LOG_FILE"
grep -Fq "OK" "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-090-url-routing-hash-out-of-scope-negative.test.tsx" "$LOG_FILE"
grep -Fq "bootstrap rewrites to target-scoped detail" "$LOG_FILE"
grep -Fq "renders a target-scoped workspace as read-only detail with one getTargetJob" "$LOG_FILE"
grep -Fq "Reports hash bootstrap keeps targetJobId only" "$LOG_FILE"
grep -Fq "legacy Parse report params are stripped" "$LOG_FILE"
grep -Fq "SPA fallback explicitly serves the known /reports path" "$LOG_FILE"
grep -Fq "canonicalizes a Reports hash to targetJobId-only /reports" "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"

# Out-of-scope-route grep: ROUTE_TO_PATH must not enumerate out-of-scope aliases.
ROUTE_FILE="$REPO_ROOT/frontend/src/app/routeUrl.ts"
test -s "$ROUTE_FILE"
for out_of_scope_path in '/voice' '/welcome' '/growth' '/mistakes' '/drill' '/followup' '/experiences' '/star' '/onboarding' '/debrief' '/profile'; do
  if grep -E "^\\s*[a-zA-Z_]+: \"$out_of_scope_path\"," "$ROUTE_FILE" >/dev/null; then
    echo "routeUrl.ROUTE_TO_PATH leaked out-of-scope path: $out_of_scope_path" >&2
    exit 1
  fi
done

# Out-of-scope-screen grep: no standalone screen file for out-of-scope aliases.
SCREEN_DIR="$REPO_ROOT/frontend/src/app/screens"
for out_of_scope_alias in 'welcome' 'growth' 'mistakes' 'drill' 'followup' 'experiences' 'star' 'onboarding' 'debrief' 'profile'; do
  if find "$SCREEN_DIR" -mindepth 1 -maxdepth 2 -type d -iname "$out_of_scope_alias" 2>/dev/null | grep -q .; then
    echo "out-of-scope screen directory still present: $out_of_scope_alias" >&2
    exit 1
  fi
done
