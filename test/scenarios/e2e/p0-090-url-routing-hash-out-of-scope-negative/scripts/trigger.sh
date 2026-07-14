#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-090-url-routing-hash-out-of-scope-negative"

mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  python3 "$SCRIPT_DIR/source_contract_test.py"
  pnpm --filter @easyinterview/frontend test \
    src/app/bootstrapRoute.test.ts \
    src/app/routeUrl.test.ts \
    src/app/spaFallback.test.ts \
    src/app/topbar/TopBar.test.tsx \
    src/app/scenarios/p0-090-url-routing-hash-out-of-scope-negative.test.tsx \
    --reporter=verbose
  pnpm --filter @easyinterview/frontend test \
    src/app/App.test.tsx \
    --reporter=verbose \
    -t 'renders a target-scoped workspace|uses only targetJobId as workspace detail authority'
) 2>&1 | tee "$OUTPUT_DIR/trigger.log"
