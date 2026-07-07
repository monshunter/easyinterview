#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "$SCENARIO_ID: no passing test files" >&2; exit 1; }
grep -Fq 'App.test.tsx' "$LOG_FILE" || { echo "$SCENARIO_ID: App route hydration test did not run" >&2; exit 1; }
grep -Fq 'useWorkspacePracticePlan.test.tsx' "$LOG_FILE" || { echo "$SCENARIO_ID: practice plan refresh test did not run" >&2; exit 1; }
non_current_mode="debrief""_replay"
if rg -n "questionText|${non_current_mode}|JD original|resume body" \
  "$REPO_ROOT/frontend/src/app/screens/workspace" \
  "$REPO_ROOT/frontend/src/app/interview-context" \
  -g '!*.test.tsx'; then
  echo "$SCENARIO_ID: forbidden workspace runtime privacy field leaked" >&2
  exit 1
fi
echo "$SCENARIO_ID PASS"
