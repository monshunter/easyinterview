#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-070-practice-derived-plan-create-read-replay/trigger.log"
grep -Fq -- '=== RUN   TestCreateDerivedPracticePlanIdempotencyMismatchHasNoSecondInsertOrLeak' "$LOG"
grep -Fq -- '--- PASS: TestCreateDerivedPracticePlanIdempotencyMismatchHasNoSecondInsertOrLeak' "$LOG"
grep -Fq -- '=== RUN   TestPracticeChatV020CandidateUsesSemanticFocus' "$LOG"
grep -Fq -- '--- PASS: TestPracticeChatV020CandidateUsesSemanticFocus' "$LOG"
grep -Fq -- '=== RUN   TestE2EP0070PracticeDerivedPlanCreateReadReplay' "$LOG"
grep -Fq -- '--- PASS: TestE2EP0070PracticeDerivedPlanCreateReadReplay' "$LOG"
for marker in REPORT_GENERIC_RETRY_PASS REPORT_DERIVED_FOCUS_PASS REPORT_DERIVED_SEMANTIC_PROMPT_PASS REPORT_NEXT_EMPTY_FOCUS_PASS REPORT_DERIVED_IDEMPOTENCY_PASS REPORT_DERIVED_POSTGRES_PASS; do grep -Fq -- "$marker" "$LOG"; done
grep -Eq -- '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/api/practice' "$LOG"
grep -Eq -- '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/practice' "$LOG"
grep -Eq -- '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/store/practice' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|DATABASE_URL is required' "$LOG"
F3_CHECKLIST="$ROOT/docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/checklist.md"
grep -F '<!-- verified:' "$F3_CHECKLIST" | grep -Fq 'PRACTICE_SEMANTIC_FOCUS_PROMPT_V020_PASS'
echo 'F3_PRACTICE_SEMANTIC_FOCUS_MARKER_CONSUMED_PASS'
echo 'E2E.P0.070 PASS'
