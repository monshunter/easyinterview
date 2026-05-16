#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-060-debrief-create-worker-happy"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "E2E.P0.060 RUNNER go test" "$LOG_FILE"
grep -Fq "TestCreateDebrief_HappyResponse" "$LOG_FILE"
grep -Fq "TestGenerateHandler_HappyResolution" "$LOG_FILE"
grep -Fq "TestStoreUpdateDebriefCompleted_HappyTransaction" "$LOG_FILE"
grep -Eq '^PASS$' "$LOG_FILE"
python3 "$REPO_ROOT/scripts/lint/backend_debrief_legacy.py" --phase all
echo "E2E.P0.060 PASS"
