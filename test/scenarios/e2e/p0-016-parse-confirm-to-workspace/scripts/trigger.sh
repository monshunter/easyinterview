#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-016-parse-confirm-to-workspace"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/targetJob.realApiMode.test.ts
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend test \
    src/app/screens/parse/ParseEdit.test.tsx \
    src/app/screens/parse/ParseAuthGate.test.tsx \
    src/app/screens/parse/ParseResumeBinding.test.tsx
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend build
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/parse.spec.ts \
    --grep "save plan navigates|start interview hands off"
) | tee "$OUTPUT_DIR/trigger.log"
