#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
LOG="$OUT/trigger.log"
RESULT="$OUT/result.json"

test -s "$LOG"
test -s "$RESULT"
test "$(jq -r '.status' "$RESULT")" = "PASS"
test "$(jq -r '.scenario' "$RESULT")" = "E2E.P0.099"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|0 tests' "$LOG"

for marker in \
  "TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns" \
  "TestPersistReportUsesPostgresTextArrayForRetryFocus" \
  "TestUpdateFeedbackReportStatusAllowsGeneratingRetry"; do
  grep -Fq -- "--- PASS: $marker" "$LOG"
done

! rg -n 'appendSessionEvent|/practice/sessions/\{sessionId\}/events|currentQuestion|questionCount|题[目号]|本轮题目' \
  "$ROOT/frontend/src/app/screens/practice" \
  "$ROOT/backend/internal/api/practice" \
  "$ROOT/backend/internal/store/practice"

echo "verify: ok"
