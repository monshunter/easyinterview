#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-064-debrief-privacy-legacy"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "E2E.P0.064 RUNNER go test + lint" "$LOG_FILE"
grep -Fq "TestOutboxPayload_NoRawText" "$LOG_FILE"
grep -Fq "TestAuditEvents_NoRawText" "$LOG_FILE"
grep -Fq "OK: backend-debrief legacy terms absent from runtime surfaces" "$LOG_FILE"
grep -Fq "validate-fixtures: OK" "$LOG_FILE"
if grep -F "__SECRET_RAW_TEXT__" "$LOG_FILE" | grep -Ev 'TestOutboxPayload_NoRawText|TestAuditEvents_NoRawText|seed-input|README|trigger.sh|verify.sh'; then
  echo "E2E.P0.064: raw marker leaked into runner output" >&2
  exit 1
fi
echo "E2E.P0.064 PASS"
