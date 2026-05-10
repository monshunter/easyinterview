#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-025-jd-match-profile-and-agent-status"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run \
    src/app/screens/jd_match/JDMatchScreen.test.tsx \
    src/app/screens/jd_match/JDMatchScreen.dataDriven.test.tsx \
    src/app/screens/jd_match/JDMatchScreen.fetchBehavior.test.tsx \
    src/app/screens/jd_match/JDMatchScreen.placeholderRemoved.test.tsx \
    src/app/screens/jd_match/useJobMatchProfile.test.tsx \
    src/app/screens/jd_match/useAgentScanStatus.test.tsx \
    src/app/screens/jd_match/JDMatchAuthGate.test.tsx \
    src/app/screens/jd_match/JDMatchAutoResume.test.tsx
  pnpm --filter @easyinterview/frontend build
  pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/jd_match.spec.ts \
    -g "Responsive geometry|dark mode"
) | tee "$OUTPUT_DIR/trigger.log"
