#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-069-debrief-pixel-parity-and-legacy-negative"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.069 RUNNER pnpm vitest"
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test -- --run \
    src/app/i18n/__tests__/debriefI18nCoverage.test.ts \
    src/app/screens/debrief/__tests__/privacyBoundary.test.ts \
    src/api/devMockClient.test.ts
} | tee "$OUTPUT_DIR/trigger.log"
{
  echo "E2E.P0.069 RUNNER playwright debrief pixel parity"
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend build
  pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/debrief.spec.ts
} | tee -a "$OUTPUT_DIR/trigger.log"
echo "E2E.P0.069 LEGACY GREP" | tee -a "$OUTPUT_DIR/trigger.log"
python3 "$REPO_ROOT/scripts/lint/frontend_debrief_legacy.py" \
    --repo-root "$REPO_ROOT" --phase 8.12 | tee -a "$OUTPUT_DIR/trigger.log"
{
  # Use --exclude to skip this trigger.sh file itself - it carries the
  # negative-assertion pattern that would otherwise be a false positive.
  if grep -RInE --exclude="trigger.sh" \
      "experience_library|star_editor|drill_builder|mistakes_book|growth_center|report_timeline" \
      "$REPO_ROOT/test/scenarios/e2e/p0-065-debrief-default-render-and-pickers" \
      "$REPO_ROOT/test/scenarios/e2e/p0-066-debrief-text-suggestions-and-submit" \
      "$REPO_ROOT/test/scenarios/e2e/p0-067-debrief-polling-happy-and-analysis" \
      "$REPO_ROOT/test/scenarios/e2e/p0-068-debrief-failure-and-handoff" \
      "$REPO_ROOT/test/scenarios/e2e/p0-069-debrief-pixel-parity-and-legacy-negative" \
      2>/dev/null; then
    echo "ERROR: retired vocabulary found in scenario tree"
    exit 1
  fi
  echo "SCENARIO TREE LEGACY GREP CLEAN"
} | tee -a "$OUTPUT_DIR/trigger.log"
