#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/trigger.log"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.047
grep -Fq 'PracticeScreen.test.tsx' "$LOG"; grep -Fq 'useCompletePracticeSession.test.tsx' "$LOG"; grep -Fq 'GeneratingScreen.test.tsx' "$LOG"
grep -Fq -- '--- PASS: TestSQLRepositoryCommitPracticeMessageRejectsClosedSession' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG"
echo 'E2E.P0.047 PASS'
