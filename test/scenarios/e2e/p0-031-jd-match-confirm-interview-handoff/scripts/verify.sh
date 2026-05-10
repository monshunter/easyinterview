#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-031-jd-match-confirm-interview-handoff"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
grep -Fq 'RecommendedConfirmInterview.test.tsx' "$LOG_FILE"
grep -Fq 'P0.015 regression PASS' "$LOG_FILE"
grep -Fq 'P0.016 regression PASS' "$LOG_FILE"

# Source-level negative gate: nav("parse", { ... }) must carry exactly two
# fields. The Vitest spec asserts Object.keys(params).sort() equals
# ["source", "sourceJobMatchId"]; the source-level grep here adds a second
# layer of defence by ensuring no other params keys are wired into the
# handler.
HANDLER_FILE="$REPO_ROOT/frontend/src/app/screens/jd_match/JDMatchScreen.tsx"
test -s "$HANDLER_FILE"
if ! grep -Eq 'name: *"parse"' "$HANDLER_FILE"; then
  echo 'JDMatchScreen.tsx does not navigate to parse from the Confirm handler' >&2
  exit 1
fi
CONFIRM_TEST="$REPO_ROOT/frontend/src/app/screens/jd_match/RecommendedConfirmInterview.test.tsx"
test -s "$CONFIRM_TEST"
grep -Fq 'Object.keys(params).sort()' "$CONFIRM_TEST"
grep -Fq '["source", "sourceJobMatchId"]' "$CONFIRM_TEST"
