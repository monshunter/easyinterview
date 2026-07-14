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
  useResumePresignUpload.test.tsx \
  useResumeRegistration.test.tsx \
  CreateFlowScopeNegative.test.ts; do
  grep -Fq "$spec" "$LOG_FILE" || {
    echo "$SCENARIO_ID: spec $spec did not run" >&2
    exit 1
  }
done
grep -Fq "submits pasted content with a neutral source title" "$LOG_FILE" || {
  echo "$SCENARIO_ID: neutral paste title regression test did not run" >&2
  exit 1
}
for boundary_assertion in \
  "accepts exact UTF-8 paste bytes from runtime config" \
  "rejects paste limit+1 before register and preserves the draft" \
  "rejects limit+1 before presign using the runtime 10 MiB resume upload ceiling"; do
  grep -Fq "$boundary_assertion" "$LOG_FILE" || {
    echo "$SCENARIO_ID: boundary assertion did not run: $boundary_assertion" >&2
    exit 1
  }
done
echo "$SCENARIO_ID PASS"
