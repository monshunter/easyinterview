#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-013-targetjob-manual-form-ready"

mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/targetjob -run 'TestE2EP0013ManualFormReady|TestService_ImportTargetJob_ManualFormSyncReady|TestSQLStore_ImportTargetJob_ManualFormSyncSucceededAndNoOutbox|TestSQLStore_ImportTargetJob_DedupeReturnsExistingActiveRunnerJob' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
