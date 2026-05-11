#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-032-dev-mock-auth-state-and-user-menu"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-032-dev-mock-auth-state-and-user-menu.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +1 passed \(1\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

for required in \
  'dev mock unauthenticated login avatar dropdown profile settings logout' \
  'Alice Example' \
  'ali***@example.com' \
  'topbar-user-chip' \
  'topbar-user-avatar'; do
  if ! grep -Fq "$required" "$LOG_FILE"; then
    echo "missing auth-state menu evidence: $required" >&2
    exit 1
  fi
done

for forbidden in \
  'topbar-user-inline' \
  'legacy inline user menu' \
  'ui-design/src/data'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden auth-state evidence leaked: $forbidden" >&2
    exit 1
  fi
done
