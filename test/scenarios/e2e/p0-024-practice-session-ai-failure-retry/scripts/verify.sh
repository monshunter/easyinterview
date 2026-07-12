#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-024-practice-session-ai-failure-retry/trigger.log"
grep -Fq -- '--- PASS: TestStartPracticeSessionAIErrorFailsReservationWithoutOpeningMessage' "$LOG"
echo 'E2E.P0.024 PASS'
