#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-030-jd-match-profile-and-agent-status"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/jdMatch.realApiMode.test.ts
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
