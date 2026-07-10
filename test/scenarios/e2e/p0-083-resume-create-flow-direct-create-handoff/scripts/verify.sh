#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
for spec in \
  ResumeCreateFlow.test.tsx \
  UploadTab.test.tsx \
  CreateFlowIntegration.test.tsx \
  ResumeWorkshopAuthGate.test.tsx; do
  grep -Fq "$spec" "$LOG_FILE" || {
    echo "$SCENARIO_ID: spec $spec did not run" >&2
    exit 1
  }
done
# Sanity-check current CTA/direct create branches exercised.
for term in Home pendingAction; do
  grep -Fq "$term" "$LOG_FILE" || {
    echo "$SCENARIO_ID: branch $term was not run" >&2
    exit 1
  }
done
echo "$SCENARIO_ID PASS"
