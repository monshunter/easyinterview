#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
LOG="$ROOT/.test-output/e2e/p0-022-practice-plan-baseline-create-and-read/trigger.log"
test -s "$LOG"
for name in TestCreatePracticePlanMapsOnlyCurrentFields TestCreatePracticePlanPassesOnlyConversationPlanFields TestSQLRepositoryCreatePlanUsesConversationColumns TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger; do grep -Fq -- "--- PASS: $name" "$LOG"; done
grep -Fq -- 'canonical-round-prompt-context=PASS' "$LOG"
grep -Fq -- 'equal-duration-next-round=PASS' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|DATABASE_URL is required' "$LOG"
echo 'E2E.P0.022 PASS'
