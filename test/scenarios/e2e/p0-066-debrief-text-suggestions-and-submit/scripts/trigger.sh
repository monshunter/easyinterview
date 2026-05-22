#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-066-debrief-text-suggestions-and-submit"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.066 RUNNER pnpm vitest"
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run --reporter=verbose \
    src/app/screens/debrief/DebriefScreen.test.tsx \
    src/app/screens/debrief/DebriefPickerRegression.test.tsx \
    src/app/screens/debrief/__tests__/privacyBoundary.test.ts \
    src/app/interview-context/InterviewContext.test.tsx \
    src/app/auth/pendingAction.test.ts
} | tee "$OUTPUT_DIR/trigger.log"
