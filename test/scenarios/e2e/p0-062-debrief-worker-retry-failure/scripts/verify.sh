#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-062-debrief-worker-retry-failure"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "E2E.P0.062 RUNNER go test" "$LOG_FILE"
grep -Fq "TestGenerateHandler_A3Timeout" "$LOG_FILE"
grep -Fq "TestGenerateHandler_PermanentFailAt5Attempts" "$LOG_FILE"
grep -Fq "TestRetryPolicy_BackoffBelowMax" "$LOG_FILE"
grep -Fq "TestRetryPolicy_PermanentFailAtMax" "$LOG_FILE"
grep -Eq '^PASS$' "$LOG_FILE"
echo "E2E.P0.062 PASS"
