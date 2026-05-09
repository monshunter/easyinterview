#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/workspace/WorkspaceHandoff.test.tsx \
    src/app/screens/workspace/WorkspaceScreen.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
