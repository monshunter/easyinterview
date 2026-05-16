#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-061-debrief-get-isolation"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "E2E.P0.061 RUNNER go test" "$LOG_FILE"
grep -Fq "TestStoreGetDebrief_DraftPartial" "$LOG_FILE"
grep -Fq "TestGetDebrief_CrossUser404" "$LOG_FILE"
grep -Fq "TestGetDebrief_NotFound404" "$LOG_FILE"
grep -Eq '^PASS$' "$LOG_FILE"
echo "E2E.P0.061 PASS"
