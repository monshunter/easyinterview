#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-064-debrief-privacy-legacy"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.064 RUNNER go test + lint"
  cd "$REPO_ROOT/backend"
  go test -v ./internal/store/debrief -run 'TestOutboxPayload_NoRawText|TestGenerateHandler_OutboxPayloadSchema|TestCreateDebrief_OutboxPayloadSchema' -count=1
  go test -v ./internal/debrief -run 'TestAuditEvents_NoRawText|TestAITaskRunsWritten|TestAuditEventsWritten' -count=1
  cd "$REPO_ROOT"
  python3 scripts/lint/backend_debrief_legacy.py --phase all
  make validate-fixtures
} | tee "$OUTPUT_DIR/trigger.log"
