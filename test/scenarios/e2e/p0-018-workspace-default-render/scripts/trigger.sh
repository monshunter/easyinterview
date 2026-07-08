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
    src/app/screens/workspace/WorkspaceHeader.test.tsx \
    src/app/screens/workspace/WorkspaceHandoff.test.tsx \
    src/app/screens/workspace/WorkspaceModalIntegration.test.tsx \
    src/app/screens/workspace/modals/PlanSwitcherModal.test.tsx \
    src/app/screens/workspace/modals/ResumePickerModal.test.tsx \
    src/app/screens/workspace/modals/useModalA11y.test.tsx
) | tee "$OUTPUT_DIR/trigger.log"
