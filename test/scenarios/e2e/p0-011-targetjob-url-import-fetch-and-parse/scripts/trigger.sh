#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-011-targetjob-url-import-fetch-and-parse"

mkdir -p "$OUTPUT_DIR"

if [[ "${ALLOW_TARGETJOB_PACKAGE_PROXY:-}" != "1" ]]; then
  {
    echo "Blocked: E2E.P0.011 is proxy-only right now."
    echo "The package-level go test below is TDD support, not valid BDD evidence."
    echo "Complete cmd/api target_import drainer + F3/A3/urlfetch runtime wiring and replace this trigger with auth -> HTTP API -> drainer execution."
    echo "Set ALLOW_TARGETJOB_PACKAGE_PROXY=1 only when intentionally running the old focused test proxy."
  } | tee "$OUTPUT_DIR/trigger.log"
  exit 2
fi

(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/targetjob ./internal/targetjob/urlfetch -run 'TestE2EP0011URLImportFetchAndParse|TestParseExecutor_URLFetchBodyIsPersistedAndParsed|TestFetch_' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
