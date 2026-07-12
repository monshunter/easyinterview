#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-072-practice-derived-source-isolation-privacy/trigger.log"
grep -Fq -- '--- PASS: TestDerivedPracticePlanRequiresSourceReport' "$LOG"
grep -Fq -- '--- PASS: TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger' "$LOG"
grep -Fq -- 'stale-source-and-round-budget-mismatch=PASS' "$LOG"
grep -Fq -- 'all-rounds-complete-fail-closed=PASS' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|DATABASE_URL is required' "$LOG"
echo 'E2E.P0.072 PASS'
