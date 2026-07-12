#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
LOG="$ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey/trigger.log"
test -s "$LOG"

for marker in \
  "TestE2EP0OperationMatrixPreflight" \
  "TestBuildResumeRuntimeWrapsParseAIWithObservability" \
  "TestSQLRepositoryCreatePlanUsesConversationColumns" \
  "TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns" \
  "TestGenerateReportUsesOneConversationLevelAICall" \
  "TestPersistReportUsesPostgresTextArrayForRetryFocus" \
  "TestUpdateFeedbackReportStatusAllowsGeneratingRetry"; do
  grep -Fq -- "--- PASS: $marker" "$LOG"
done

! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]' "$LOG"
! rg -n 'appendSessionEvent|/practice/sessions/\{sessionId\}/events|client_event_id|practice_turns' \
  "$ROOT/backend/cmd/api/e2e_p0_operation_matrix_preflight_test.go" \
  "$ROOT/backend/internal/api/practice" \
  "$ROOT/backend/internal/store/practice" \
  "$ROOT/backend/internal/review" \
  "$ROOT/backend/internal/store/review"

echo "verify: ok"
