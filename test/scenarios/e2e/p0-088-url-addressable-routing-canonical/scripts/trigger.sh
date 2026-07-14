#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-088-url-addressable-routing-canonical"

mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  python3 "$SCRIPT_DIR/source_contract_test.py"
  pnpm --filter @easyinterview/frontend test \
    src/app/outOfScopeRouteNegative.test.ts \
    src/app/AppRoutingHistory.test.tsx \
    src/app/routeUrl.test.ts \
    src/app/topbar/TopBar.test.tsx \
    src/app/scenarios/p0-088-url-addressable-routing-canonical.test.tsx \
    --reporter=verbose
) 2>&1 | tee "$OUTPUT_DIR/trigger.log"
