#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-010-targetjob-text-import-parse-ready"

mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/targetjob -run 'TestE2EP0010TextImportParseReady|TestSQLStore_ImportTargetJob_DedupeReturnsExistingActiveRunnerJob|TestService_UpdateTargetJob_DedupeHitBypassesLaterStateTransition' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
