#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-066-debrief-text-suggestions-and-submit"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "E2E.P0.066 RUNNER pnpm vitest" "$LOG_FILE"
grep -Eq "Tests[[:space:]]+[0-9]+ passed" "$LOG_FILE"
grep -Eq "InterviewContext.test.tsx" "$LOG_FILE"
grep -Eq "privacyBoundary.test.ts" "$LOG_FILE"
! grep -Eq "no tests" "$LOG_FILE"
! grep -Eq "Tests[[:space:]]+[0-9]+ failed" "$LOG_FILE"
python3 "$REPO_ROOT/scripts/lint/frontend_debrief_legacy.py" --repo-root "$REPO_ROOT" --phase 8.9 >/dev/null
echo "E2E.P0.066 PASS"
