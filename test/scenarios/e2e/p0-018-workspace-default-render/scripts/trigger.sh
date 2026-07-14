#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-018-workspace-default-render"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/topbar/TopBar.test.tsx \
    src/app/topbar/TopBarVisual.test.tsx \
    src/app/scenarios/p0-004-app-shell-language-switch.test.tsx \
    src/app/App.test.tsx \
    src/app/screens/workspace/WorkspaceScreen.test.tsx \
    src/app/screens/workspace/WorkspaceEmptyState.test.tsx \
    src/app/screens/workspace/hooks/useWorkspaceTargetJobs.test.tsx \
    src/app/screens/parse/ParseFlow.test.tsx \
    src/app/screens/parse/ParseResumeBinding.test.tsx \
    src/app/screens/parse/ParseRoundStates.test.tsx
  cd "$REPO_ROOT/frontend"
  CI=1 COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm exec playwright test \
    tests/pixel-parity/workspace.spec.ts \
    tests/pixel-parity/parse.spec.ts \
    --grep 'ready plan card opens workspace detail without Parse animation or route-side mutation|workspace detail round states match the UI truth at desktop and mobile' \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUTPUT_DIR/playwright"
) | tee "$OUTPUT_DIR/trigger.log"
