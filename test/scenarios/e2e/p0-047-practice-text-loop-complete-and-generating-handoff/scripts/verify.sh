#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.047: no passing test files found" >&2; exit 1; }
grep -Fq 'useCompletePracticeSession.test.tsx' "$LOG_FILE" || { echo "E2E.P0.047: useCompletePracticeSession.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceHandoffParams.test.ts' "$LOG_FILE" || { echo "E2E.P0.047: practiceHandoffParams.test.ts did not run" >&2; exit 1; }
grep -Fq 'completePracticeSessionBody.test.tsx' "$LOG_FILE" || { echo "E2E.P0.047: completePracticeSessionBody.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceCompletion.test.tsx' "$LOG_FILE" || { echo "E2E.P0.047: practiceCompletion.test.tsx did not run" >&2; exit 1; }
if rg -n '\bgetFeedbackReport\b|\bcreatePracticeVoiceTurn\b' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.047: out-of-scope generated client method called from practice runtime" >&2
  exit 1
fi
echo "E2E.P0.047 PASS"
