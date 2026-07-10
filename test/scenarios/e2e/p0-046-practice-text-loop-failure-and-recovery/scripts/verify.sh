#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
LOCALES_DIR="$REPO_ROOT/frontend/src/app/i18n/locales"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'practiceSessionLost.test.tsx' "$LOG_FILE" || { echo "E2E.P0.046: practiceSessionLost.test.tsx did not run" >&2; exit 1; }
grep -Fq 'useCompletePracticeSession.test.tsx' "$LOG_FILE" || { echo "E2E.P0.046: useCompletePracticeSession.test.tsx did not run" >&2; exit 1; }
for key in aiTimeout network sessionConflict unknown retry backToWorkspace; do
  grep -q "\"practice.errors.${key}\":" "$LOCALES_DIR/zh.ts" || { echo "E2E.P0.046: missing zh practice.errors.${key}" >&2; exit 1; }
  grep -q "\"practice.errors.${key}\":" "$LOCALES_DIR/en.ts" || { echo "E2E.P0.046: missing en practice.errors.${key}" >&2; exit 1; }
done
if rg -n 'AI_PROVIDER_API_KEY|AI_PROVIDER_BASE_URL|prompt-registry|provider-registry|AIClient' "$PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**'; then
  echo "E2E.P0.046: practice runtime references LLM provider plumbing directly" >&2
  exit 1
fi
echo "E2E.P0.046 PASS"
