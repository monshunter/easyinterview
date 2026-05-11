#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-006-ui-design-pixel-parity-gate"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq '110 passed' "$LOG_FILE"
if grep -Eq '[0-9]+ failed' "$LOG_FILE"; then
  echo "[verify] trigger.log reports failed tests" >&2
  exit 1
fi
# Project markers must appear so we know both viewport profiles ran.
grep -Fq "[desktop]" "$LOG_FILE"
grep -Fq "[mobile]" "$LOG_FILE"
# Spec markers — assert all four parity specs were exercised.
grep -Fq "tests/pixel-parity/topbar.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/screens.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/layout.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/screenshot.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/home.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/parse.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/jd_match.spec.ts" "$LOG_FILE"
grep -Fq "tests/pixel-parity/workspace.spec.ts" "$LOG_FILE"

# Negative: trigger.log must not mention retired entries.
for forbidden in \
  'route-welcome' \
  'topbar-nav-mistakes' \
  'topbar-nav-growth' \
  'topbar-nav-drill' \
  'topbar-nav-voice'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    # The retired-entries spec asserts these locators have count=0; if any of
    # them shows up in a failure trace we fail the gate hard.
    if grep -F "$forbidden" "$LOG_FILE" | grep -Eq 'Expected|Received|Error'; then
      echo "[verify] retired entry $forbidden leaked into a failing trace" >&2
      exit 1
    fi
  fi
done
