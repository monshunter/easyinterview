#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx \
    src/app/screens/practice/utils/practiceHandoffParams.test.ts \
    src/app/screens/practice/__tests__/completePracticeSessionBody.test.tsx \
    src/app/screens/practice/__tests__/practicePrivacy.test.tsx \
    src/app/screens/practice/__tests__/practiceCompletion.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
