#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-063-debrief-suggest-questions"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "E2E.P0.063 RUNNER go test" "$LOG_FILE"
grep -Fq "TestServiceSuggestQuestions_Happy" "$LOG_FILE"
grep -Fq "TestServiceSuggestQuestions_A3Timeout" "$LOG_FILE"
grep -Fq "TestSuggestDebriefQuestions_CountBoundary" "$LOG_FILE"
grep -Eq '^PASS$' "$LOG_FILE"
echo "E2E.P0.063 PASS"
