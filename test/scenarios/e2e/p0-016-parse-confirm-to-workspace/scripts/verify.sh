#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-016-parse-confirm-to-workspace"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq 'VITE_EI_API_MODE=real' "$LOG_FILE"
grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1' "$LOG_FILE"
grep -Fq 'targetJob.realApiMode.test.ts' "$LOG_FILE"
grep -Fq "ParseEdit.test.tsx" "$LOG_FILE"
grep -Fq "ParseAuthGate.test.tsx" "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
# Verify: updateTargetJob body schema does NOT contain read-only fields
for forbidden in 'parse-basics-level' 'parse-basics-language'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "read-only field name found in test output: $forbidden" >&2
    exit 1
  fi
done
# Test must pass — grep for PASS marker
grep -q 'Tests.*passed' "$LOG_FILE"
