#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-045-practice-text-loop-strict-and-debrief-display"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.045: no passing test files found" >&2; exit 1; }
grep -Fq 'usePracticeAssistance.test.ts' "$LOG_FILE" || { echo "E2E.P0.045: usePracticeAssistance.test.ts did not run" >&2; exit 1; }
grep -Fq 'practiceGoalParity.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceGoalParity.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceHints.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceHints.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceStrictToggleLocked.test.tsx' "$LOG_FILE" || { echo "E2E.P0.045: practiceStrictToggleLocked.test.tsx did not run" >&2; exit 1; }
if rg -n "practiceMode\s*[=:]\s*['\"]debrief['\"]" "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: legacy practiceMode='debrief' literal leaked" >&2
  exit 1
fi
if rg -n '切到语音|Switch to voice' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.045: legacy mode-switch copy leaked" >&2
  exit 1
fi
# usePracticeEvents must NOT set the Idempotency-Key header on the request.
# Detect any header assignment that reaches the wire.
if rg -n '"Idempotency-Key"\s*:|idempotencyKey\s*:|setIdempotencyKey\b|opts\.idempotencyKey' "$PRACTICE_DIR/hooks/usePracticeEvents.ts"; then
  echo "E2E.P0.045: usePracticeEvents must NOT set Idempotency-Key on appendSessionEvent" >&2
  exit 1
fi
echo "E2E.P0.045 PASS"
