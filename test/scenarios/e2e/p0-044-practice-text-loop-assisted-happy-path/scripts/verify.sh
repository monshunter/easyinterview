#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path/trigger.log"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.044
grep -Fq 'PracticeScreen.test.tsx' "$LOG"
grep -Fq -- '--- PASS: TestSendPracticeMessageReturnsConversationMessages' "$LOG"
grep -Fq -- '--- PASS: TestSendPracticeMessageUsesOrdinaryConversationHistory' "$LOG"
grep -Fq -- '--- PASS: TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage' "$LOG"
! grep -Eq -- '--- SKIP:|\[no tests to run\]|no tests to run|--- FAIL:|^FAIL($|[[:space:]])' "$LOG"
echo 'E2E.P0.044 PASS'
