#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || {
  echo "$SCENARIO_ID: no passing test files" >&2
  exit 1
}
for spec in \
  ResumeCreateFlow.test.tsx \
  UploadTab.test.tsx \
  PasteGuidedTab.test.tsx \
  useResumePresignUpload.test.tsx \
  useResumeRegistration.test.tsx \
  ParsingStage.test.tsx \
  CreateFlowLegacyNegative.test.ts; do
  grep -Fq "$spec" "$LOG_FILE" || {
    echo "$SCENARIO_ID: spec $spec did not run" >&2
    exit 1
  }
done
echo "$SCENARIO_ID PASS"
