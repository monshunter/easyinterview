#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-068-debrief-failure-and-handoff"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.068 RUNNER pnpm vitest"
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test -- --run \
    src/app/screens/debrief \
    src/app/interview-context/InterviewContext.test.tsx
} | tee "$OUTPUT_DIR/trigger.log"
echo "E2E.P0.068 HANDOFF GREP" | tee -a "$OUTPUT_DIR/trigger.log"
if grep -RIn "createPracticePlan\|startPracticeSession" \
    "$REPO_ROOT/frontend/src/app/screens/debrief" 2>/dev/null; then
  echo "ERROR: forbidden direct call found" | tee -a "$OUTPUT_DIR/trigger.log"
  exit 1
fi
echo "DEBRIEF HANDOFF GREP CLEAN" | tee -a "$OUTPUT_DIR/trigger.log"
