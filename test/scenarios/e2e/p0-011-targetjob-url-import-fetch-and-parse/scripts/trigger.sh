#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-011-targetjob-url-import-fetch-and-parse"

mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/targetjob ./internal/targetjob/urlfetch -run 'TestE2EP0011URLImportFetchAndParse|TestParseExecutor_URLFetchBodyIsPersistedAndParsed|TestFetch_' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
