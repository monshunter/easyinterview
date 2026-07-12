#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
mkdir -p "$OUT"

{
  echo "E2E.P0.098 current conversation funnel contract composition"
  cd "$ROOT"
  go test -v ./backend/cmd/api -run 'TestE2EP0OperationMatrixPreflight|TestBuildResumeRuntimeWrapsParseAIWithObservability' -count=1
  go test -v ./backend/internal/api/practice ./backend/internal/practice ./backend/internal/store/practice -run 'TestCreatePracticePlanMapsOnlyCurrentFields|TestCreatePracticePlanPassesOnlyConversationPlanFields|TestSQLRepositoryCreatePlanUsesConversationColumns|TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns|TestSendPracticeMessage' -count=1
  go test -v ./backend/internal/review ./backend/internal/store/review -run 'TestGenerateReportUsesOneConversationLevelAICall|TestPersistReportUsesPostgresTextArrayForRetryFocus|TestUpdateFeedbackReportStatusAllowsGeneratingRetry' -count=1
} | tee "$OUT/trigger.log"

printf 'scenario=E2E.P0.098\nmethod=current-contract-composition\ntrigger_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/trigger.env"
echo "trigger: ok"
