#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_ID="$(basename "$(dirname "$SCRIPT_DIR")")"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/$SCENARIO_ID"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq '^[[:space:]]*RUN[[:space:]]+v[0-9]' "$LOG_FILE" || {
  echo "$SCENARIO_ID: vitest runner marker missing" >&2
  exit 1
}
if grep -Eiq 'No test files found|No tests found|No test suite found|No test cases found' "$LOG_FILE"; then
  echo "$SCENARIO_ID: no-test marker found" >&2
  exit 1
fi
if grep -Eq '^[[:space:]]*Test Files[[:space:]].*failed|^[[:space:]]*Tests[[:space:]].*failed' "$LOG_FILE"; then
  echo "$SCENARIO_ID: failing vitest summary found" >&2
  exit 1
fi
grep -Eq '^[[:space:]]*Test Files[[:space:]]+[1-9][0-9]*[[:space:]]+passed' "$LOG_FILE" || {
  echo "$SCENARIO_ID: no passing test files" >&2
  exit 1
}
grep -Eq '^[[:space:]]*Tests[[:space:]]+[1-9][0-9]*[[:space:]]+passed' "$LOG_FILE" || {
  echo "$SCENARIO_ID: no passing tests" >&2
  exit 1
}
for spec in \
  ResumeCreateFlow.test.tsx \
  UploadTab.test.tsx \
  PreviewStage.test.tsx \
  useResumePresignUpload.test.tsx \
  useResumeRegistration.test.tsx \
  ParsingStage.test.tsx \
  CreateFlowNonCurrentNegative.test.ts; do
  grep -Fq "$spec" "$LOG_FILE" || {
    echo "$SCENARIO_ID: spec $spec did not run" >&2
    exit 1
  }
done
echo "$SCENARIO_ID PASS"
