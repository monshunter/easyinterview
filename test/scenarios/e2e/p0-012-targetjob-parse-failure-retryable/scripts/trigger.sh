#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-012-targetjob-parse-failure-retryable"

mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/targetjob -run 'TestE2EP0012ParseFailureRetryableAndNonRetryable|TestParseExecutor_(AIProviderTimeout_Retryable|AIOutputInvalid_NonRetryable|F3FailClosed|URLFetchUnavailable_Retryable|URLFetchInvalid_NonRetryable|BlankJDText_SourceInvalid)' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
