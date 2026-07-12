#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-023-practice-session-start-and-opening-message/trigger.log"
grep -Fq -- '--- PASS: TestStartPracticeSessionCreatesOpeningAssistantMessage' "$LOG"
grep -Fq -- '--- PASS: TestSQLRepositoryGetSessionReturnsOrderedMessages' "$LOG"
echo 'E2E.P0.023 PASS'
