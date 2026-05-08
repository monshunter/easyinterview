#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-015-jd-import-and-parse"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/home/JDAssistModal.test.tsx \
    src/app/screens/home/HomeImport.test.tsx \
    src/app/screens/home/HomeAuthGate.test.tsx \
    src/app/screens/parse/ParseScreen.test.tsx \
    src/app/screens/parse/ParseFlow.test.tsx \
    src/app/screens/parse/ParseFailedState.test.tsx \
    src/app/screens/parse/ParseEdit.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
