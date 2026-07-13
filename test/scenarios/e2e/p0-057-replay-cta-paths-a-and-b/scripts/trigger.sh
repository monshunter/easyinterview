#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-057-replay-cta-paths-a-and-b"
mkdir -p "$OUT"

{
  "$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT"
  echo "E2E.P0.057: validating closed derived-plan requests and one fresh session"
  (
    cd "$ROOT/frontend"
    pnpm exec vitest run \
      src/app/screens/report/__tests__/preflight.test.ts \
      src/app/auth/__tests__/pendingActionReplayPractice.test.ts \
      src/app/interview-context/buildCreatePlanRequest.test.ts \
      src/app/interview-context/startPractice.test.ts \
      src/app/screens/report/__tests__/ConversationReport.test.tsx \
      src/app/screens/report/__tests__/ReplayCta.test.tsx \
      --reporter=verbose
  )
} 2>&1 | tee "$OUT/trigger.log"
