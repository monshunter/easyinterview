#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-014-home-default-render"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/targetJob.realApiMode.test.ts
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/home/HomeScreen.test.tsx \
    src/app/screens/home/HomeResumeSelection.test.tsx \
    src/app/screens/home/HomeRecentMocks.test.tsx \
    src/app/screens/home/MockInterviewCard.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
