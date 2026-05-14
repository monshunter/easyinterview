#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/practice/PracticeScreen.test.tsx \
    src/app/screens/practice/hooks/usePracticeSessionLoader.test.tsx \
    src/app/screens/practice/hooks/usePracticeEvents.test.tsx \
    src/app/screens/practice/components/AssistantActionRenderer.test.tsx \
    src/app/screens/practice/__tests__/PracticeScreenIntegration.test.tsx \
    src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx \
    src/app/screens/practice/__tests__/idempotencyContract.test.tsx \
    src/app/screens/practice/__tests__/appendSessionEventBody.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
