#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-024-practice-session-ai-failure-retry/trigger.log"
grep -Fq -- '--- PASS: TestStartPracticeSessionAIErrorFailsReservationWithoutOpeningMessage' "$LOG"
grep -Fq -- '--- PASS: TestStartPracticeSessionFailsClosedWithoutResumeContextAndSkipsAI' "$LOG"
! grep -Eq -- '--- SKIP:|\[no tests to run\]|no tests to run|--- FAIL:|^FAIL($|[[:space:]])' "$LOG"
echo 'E2E.P0.024 PASS'
