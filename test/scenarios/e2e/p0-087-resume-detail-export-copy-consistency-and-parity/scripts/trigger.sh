#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  echo "## vitest-jsdom"
  pnpm --filter @easyinterview/frontend exec vitest run --reporter=verbose \
    src/app/screens/resume-workshop/components/ResumeDetailExport.test.tsx \
    src/app/screens/resume-workshop/components/ResumeDetailFixtureParity.test.tsx \
    src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx \
    src/app/screens/resume-workshop/tabs/ResumeRewritesTab.test.tsx \
    src/app/screens/resume-workshop/tabs/ResumeEditTab.test.tsx
  echo "## frontend-build"
  pnpm --filter @easyinterview/frontend build
  echo "## playwright-pixel-parity-axe"
  pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts
) | tee "$OUTPUT_DIR/trigger.log"
