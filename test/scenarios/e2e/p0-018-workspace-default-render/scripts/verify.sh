#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-018-workspace-default-render"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.018: no passing test files found" >&2; exit 1; }
grep -Fq 'WorkspaceModalIntegration.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: modal integration test did not run" >&2; exit 1; }
grep -Fq 'PlanSwitcherModal.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: plan switcher test did not run" >&2; exit 1; }
grep -Fq 'ResumePickerModal.test.tsx' "$LOG_FILE" || { echo "E2E.P0.018: resume picker test did not run" >&2; exit 1; }
testid_count="$(
  rg -o 'data-testid=' \
    "$REPO_ROOT/frontend/src/app/screens/workspace/WorkspaceScreen.tsx" \
    "$REPO_ROOT/frontend/src/app/screens/workspace/modals/PlanSwitcherModal.tsx" \
    "$REPO_ROOT/frontend/src/app/screens/workspace/modals/ResumePickerModal.tsx" \
    | wc -l | tr -d ' '
)"
if [ "$testid_count" -lt 20 ]; then
  echo "E2E.P0.018: expected >=20 workspace runtime testids, got $testid_count" >&2
  exit 1
fi
if rg -n 'practice-mode-card-|growth-center|drill-builder|mistake-queue' "$REPO_ROOT/frontend/src/app/screens/workspace" -g '!*.test.tsx'; then
  echo "E2E.P0.018: forbidden legacy runtime testid leaked" >&2
  exit 1
fi
if rg -n '\.listResumes\(' "$REPO_ROOT/frontend/src/app/screens/workspace" -g '!*.test.tsx'; then
  echo "E2E.P0.018: workspace runtime calls listResumes" >&2
  exit 1
fi
echo "E2E.P0.018 PASS"
