#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery/trigger.log"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.046
grep -Fq -- '--- PASS: TestSendPracticeMessageUsesOrdinaryConversationHistory' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG"
echo 'E2E.P0.046 PASS'
