#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="${1:-$(git rev-parse --show-toplevel)}"
cd "$REPO_ROOT"

REAL_API_MODE="${VITE_EI_API_MODE:-real}"
REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"

printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" pnpm --filter @easyinterview/frontend exec vitest run \
  src/api/clientFactory.test.ts
