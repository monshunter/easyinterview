#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-061-debrief-get-isolation"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.061 RUNNER go test"
  cd "$REPO_ROOT/backend"
  go test -v ./internal/store/debrief -run 'TestStoreGetDebrief' -count=1
  go test -v ./internal/debrief -run 'TestServiceGetDebrief_ProvenanceWireOnly' -count=1
  go test -v ./internal/api/debriefs -run 'TestGetDebrief' -count=1
} | tee "$OUTPUT_DIR/trigger.log"
