#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-057-replay-cta-paths-a-and-b"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  echo "E2E.P0.057: validating direct-start owner contract"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/report/__tests__/preflight.test.ts \
    src/app/auth/__tests__/pendingActionReplayPractice.test.ts \
    src/app/interview-context/roundAssumptions.test.ts \
    src/app/screens/report/__tests__/useReportContextData.test.tsx \
    src/app/screens/report/__tests__/ReplayCta.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
