#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-004-app-shell-language-switch"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-004-app-shell-language-switch.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +1 passed \(1\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

for required in \
  'Home' \
  'Mock Interview' \
  'Resume' \
  'Sign in' \
  'language dropdown' \
  'Accept-Language: en'; do
  if ! grep -Fq "$required" "$LOG_FILE"; then
    echo "missing language-switch evidence: $required" >&2
    exit 1
  fi
done

for forbidden in \
  'route-welcome' \
  'topbar-nav-mistakes' \
  'topbar-nav-growth' \
  'topbar-nav-drill' \
  'topbar-nav-voice' \
  'topbar-nav-jd_match' \
  'topbar-register' \
  'ui-design/src/data'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden legacy/prototype evidence leaked: $forbidden" >&2
    exit 1
  fi
done
