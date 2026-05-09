#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "$SCENARIO_ID: no passing test files" >&2; exit 1; }
for forbidden in 'getCompanyIntel' 'getFeedbackReport' 'listResumes' 'recentSessions' 'questionText' 'console.log'; do
  if grep -Fiq "$forbidden" "$LOG_FILE"; then echo "$SCENARIO_ID: forbidden $forbidden in log (may be test name - check)" >&2; fi
done
echo "$SCENARIO_ID PASS"
