#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
LOCALES_DIR="$REPO_ROOT/frontend/src/app/i18n/locales"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" \
  "$LOG_FILE" \
  "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'practiceSessionLost.test.tsx' "$LOG_FILE" || { echo "E2E.P0.046: practiceSessionLost.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceErrors.test.tsx' "$LOG_FILE" || { echo "E2E.P0.046: practiceErrors.test.tsx did not run" >&2; exit 1; }
grep -Fq 'practiceVoiceTurn.test.tsx' "$LOG_FILE" || { echo "E2E.P0.046: practiceVoiceTurn.test.tsx did not run" >&2; exit 1; }
grep -Fq 'useCompletePracticeSession.test.tsx' "$LOG_FILE" || { echo "E2E.P0.046: useCompletePracticeSession.test.tsx did not run" >&2; exit 1; }
grep -Fq 'session_wait retains the answer without a transcript duplicate and a new submit mints a new clientEventId' "$PRACTICE_DIR/__tests__/practiceErrors.test.tsx" || { echo "E2E.P0.046: session_wait retained-input/new-ID evidence missing" >&2; exit 1; }
grep -Fq 'not.toBe(first.clientEventId)' "$PRACTICE_DIR/__tests__/practiceErrors.test.tsx" || { echo "E2E.P0.046: session_wait new clientEventId assertion missing" >&2; exit 1; }
grep -Fq 'not.toHaveTextContent(answer)' "$PRACTICE_DIR/__tests__/practiceErrors.test.tsx" || { echo "E2E.P0.046: session_wait duplicate-transcript assertion missing" >&2; exit 1; }
grep -Fq 'keeps the same session and localizes a double-invalid chat failure' "$PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.046: localized same-session voice failure evidence missing" >&2; exit 1; }
grep -Fq 'createPracticeVoiceTurn: "chat-output-invalid"' "$PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.046: chat-output-invalid scenario evidence missing" >&2; exit 1; }
grep -Fq 'messageKey: "practice.errors.aiOutputInvalid"' "$PRACTICE_DIR/PracticeScreen.tsx" || { echo "E2E.P0.046: AI_OUTPUT_INVALID is not mapped to the localized practice error" >&2; exit 1; }
grep -Fq 'RUNNER backend-go-test E2E.P0.046' "$LOG_FILE" || { echo "E2E.P0.046: backend runner marker missing" >&2; exit 1; }
grep -Fq '=== RUN   TestAppendSessionEventSecondInvalidQuestionReturnsSessionWaitWithoutAdvancingTurn' "$LOG_FILE" || { echo "E2E.P0.046: double-invalid session_wait backend test did not run" >&2; exit 1; }
for key in aiTimeout aiOutputInvalid network sessionConflict unknown retry backToWorkspace; do
  grep -q "\"practice.errors.${key}\":" "$LOCALES_DIR/zh.ts" || { echo "E2E.P0.046: missing zh practice.errors.${key}" >&2; exit 1; }
  grep -q "\"practice.errors.${key}\":" "$LOCALES_DIR/en.ts" || { echo "E2E.P0.046: missing en practice.errors.${key}" >&2; exit 1; }
done
if rg -n 'AI_PROVIDER_API_KEY|AI_PROVIDER_BASE_URL|prompt-registry|provider-registry|AIClient' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.046: practice runtime references LLM provider plumbing directly" >&2
  exit 1
fi
echo "E2E.P0.046 PASS"
