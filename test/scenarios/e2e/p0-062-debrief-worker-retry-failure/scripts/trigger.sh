#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-062-debrief-worker-retry-failure"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.062 RUNNER go test"
  cd "$REPO_ROOT/backend"
  go test -v ./internal/debrief -run 'TestGenerateHandler_F3ResolveFailed|TestGenerateHandler_A3Timeout|TestGenerateHandler_ParseEmpty|TestGenerateHandler_PermanentFailAt5Attempts' -count=1
  go test -v ./internal/targetjob -run 'TestRetryPolicy_BackoffBelowMax|TestRetryPolicy_PermanentFailAtMax' -count=1
} | tee "$OUTPUT_DIR/trigger.log"
