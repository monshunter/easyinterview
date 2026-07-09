#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-045-practice-text-loop-mode-policy-display"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/practice/hooks/usePracticeAssistance.test.ts \
    src/app/screens/practice/__tests__/practiceGoalParity.test.tsx \
    src/app/screens/practice/__tests__/practiceHints.test.tsx \
    src/app/screens/practice/__tests__/practicePauseResume.test.tsx \
    src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx \
    src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx \
    src/app/screens/practice/__tests__/SessionMap.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
