#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-015-jd-import-and-parse"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/targetJob.realApiMode.test.ts
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/home/JDAssistModal.test.tsx \
    src/app/screens/home/HomeImport.test.tsx \
    src/app/screens/home/HomeAuthGate.test.tsx \
    src/app/screens/parse/ParseScreen.test.tsx \
    src/app/screens/parse/ParseFlow.test.tsx \
    src/app/screens/parse/ParseFailedState.test.tsx \
    src/app/screens/parse/ParseEdit.test.tsx
  pnpm --filter @easyinterview/frontend build
  pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/parse.spec.ts \
    --grep "ready target job response keeps ui-design loading demo before preview"
) | tee "$OUTPUT_DIR/trigger.log"
