#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-070-practice-derived-plan-create-read-replay/trigger.log"
grep -Fq -- '--- PASS: TestCreateDerivedPracticePlanPassesReportSourceAndCompetencyFocus' "$LOG"
grep -Fq -- '--- PASS: TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger' "$LOG"
for marker in retry-source-round=PASS equal-duration-next-round=PASS non-contiguous-successor=PASS; do grep -Fq -- "$marker" "$LOG"; done
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|DATABASE_URL is required' "$LOG"
echo 'E2E.P0.070 PASS'
