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
echo "E2E.P0.068 HANDOFF SESSION GATE" | tee -a "$OUTPUT_DIR/trigger.log"
if ! grep -RIn "createPracticePlan" \
    "$REPO_ROOT/frontend/src/app/screens/debrief/DebriefScreen.tsx" >/dev/null 2>&1; then
  echo "ERROR: debrief replay CTA does not create a practice plan" | tee -a "$OUTPUT_DIR/trigger.log"
  exit 1
fi
if ! grep -RIn "startPracticeSession" \
    "$REPO_ROOT/frontend/src/app/screens/debrief/DebriefScreen.tsx" >/dev/null 2>&1; then
  echo "ERROR: debrief replay CTA does not start a practice session" | tee -a "$OUTPUT_DIR/trigger.log"
  exit 1
fi
if ! grep -RIn "sourceDebriefId" \
    "$REPO_ROOT/frontend/src/app/screens/debrief/DebriefScreen.tsx" >/dev/null 2>&1; then
  echo "ERROR: debrief replay plan is missing sourceDebriefId" | tee -a "$OUTPUT_DIR/trigger.log"
  exit 1
fi
echo "DEBRIEF HANDOFF SESSION GATE OK" | tee -a "$OUTPUT_DIR/trigger.log"
