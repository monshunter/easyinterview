#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-010-targetjob-text-import-parse-ready"

mkdir -p "$OUTPUT_DIR"

if [[ "${ALLOW_TARGETJOB_PACKAGE_PROXY:-}" != "1" ]]; then
  {
    echo "Blocked: E2E.P0.010 is proxy-only right now."
    echo "The package-level go test below is TDD support, not valid BDD evidence."
    echo "Complete cmd/api target_import drainer + F3/A3/urlfetch runtime wiring and replace this trigger with auth -> HTTP API -> drainer execution."
    echo "Set ALLOW_TARGETJOB_PACKAGE_PROXY=1 only when intentionally running the old focused test proxy."
  } | tee "$OUTPUT_DIR/trigger.log"
  exit 2
fi

(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/targetjob -run 'TestE2EP0010TextImportParseReady|TestSQLStore_ImportTargetJob_DedupeReturnsExistingActiveRunnerJob|TestService_UpdateTargetJob_DedupeHitBypassesLaterStateTransition' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
