#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-026-jd-match-confirm-interview-handoff"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec vitest run \
    src/app/screens/jd_match/RecommendedConfirmInterview.test.tsx
  bash test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/setup.sh
  bash test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/trigger.sh
  bash test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/verify.sh
  bash test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/cleanup.sh
  echo "P0.015 regression PASS"
  bash test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/setup.sh
  bash test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/trigger.sh
  bash test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/verify.sh
  bash test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/cleanup.sh
  echo "P0.016 regression PASS"
) | tee "$OUTPUT_DIR/trigger.log"
