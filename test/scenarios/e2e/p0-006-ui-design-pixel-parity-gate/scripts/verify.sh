#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-006-ui-design-pixel-parity-gate"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq '[0-9]+ passed' "$LOG_FILE"
if grep -Eq '[0-9]+ failed' "$LOG_FILE"; then
  echo "[verify] trigger.log reports failed tests" >&2
  exit 1
fi
# Project markers must appear so we know both viewport profiles ran.
grep -Fq "[desktop]" "$LOG_FILE"
grep -Fq "[mobile]" "$LOG_FILE"
# Spec markers — assert the current parity suite was exercised.
for spec in \
  debrief.spec.ts \
  generating.spec.ts \
  home.spec.ts \
  layout.spec.ts \
  parse.spec.ts \
  practice.spec.ts \
  report.spec.ts \
  resume-workshop-branch-rewrites-edit.spec.ts \
  resume-workshop-create.spec.ts \
  resume-workshop.spec.ts \
  screens.spec.ts \
  screenshot.spec.ts \
  topbar.spec.ts \
  workspace.spec.ts; do
  grep -Fq "tests/pixel-parity/$spec" "$LOG_FILE"
done

# Negative: trigger.log must not mention retired entries.
for forbidden in \
  'route-welcome' \
  'route-jd_match' \
  'route-company_intel' \
  'topbar-nav-jd_match' \
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
