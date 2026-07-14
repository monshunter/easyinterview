#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
SETUP_ENV="$OUT/setup.env"
RESULT_FILE="$OUT/result.json"
mkdir -p "$OUT"

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

# shellcheck disable=SC1090
. "$SETUP_ENV"

export EI_P0_098_FRONTEND_ORIGIN="$FRONTEND_ORIGIN"
export EI_P0_098_API_BASE_URL="$API_BASE_URL"
export EI_P0_098_MAILPIT_BASE_URL="$MAILPIT_BASE_URL"
export EI_P0_098_AUTH_EMAIL="$AUTH_EMAIL"
export EI_P0_098_RESUME_ID="$RESUME_ID"
export EI_P0_098_TARGET_JOB_ID="$TARGET_JOB_ID"
export EI_P0_098_ROUND_ONE_SESSION_ID="$ROUND_ONE_SESSION_ID"
export EI_PLAYWRIGHT_OUTPUT_DIR="$OUT/playwright"
rm -rf "$OUT/playwright"
rm -f "$OUT/trigger.log" "$OUT/trigger.env" "$RESULT_FILE"

{
  echo "E2E.P0.098 persisted interview round journey"
  echo "SCENARIO_RUNNER=E2E.P0.098"
  echo "PLAYWRIGHT_SPEC=frontend/tests/e2e/practice-progress-refresh.spec.ts"
  echo "PLAYWRIGHT_CONFIG=frontend/playwright.auth-email-code.config.ts"
  cd "$ROOT"
  "$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT"
  make validate-fixtures
  go test -v ./backend/cmd/api -run 'TestE2EP0OperationMatrixPreflight|TestBuildResumeRuntimeWrapsParseAIWithObservability' -count=1
  go test -v ./backend/internal/api/practice ./backend/internal/practice ./backend/internal/store/practice -run 'TestCreatePracticePlanMapsOnlyCurrentFields|TestCreatePracticePlanPassesOnlyConversationPlanFields|TestSQLRepositoryCreatePlanUsesConversationColumns|TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns|TestSQLRepositoryCompleteSessionReplayDoesNotAppendSecondCompletedFact|TestSendPracticeMessage' -count=1
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" \
    go test -v -tags=integration ./backend/internal/store/practice -run '^TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger$' -count=1
  go test -v ./backend/internal/targetjob -run 'TestService_ListTargetJobs_ProjectsCanonicalPracticeProgressIndependentOfLifecycleStatus|TestService_GetTargetJob_HidesCompletedFactsAfterFirstCanonicalGap|TestService_GetTargetJob_ProjectsFirstRoundAndAllCompleted|TestHandler_GetAndListTargetJobs_ReturnPracticeProgressWithWireParity|TestSQLStore_ListTargetJobsForUser_LoadsPageScopedPracticeLedgerFactsInOneQuery' -count=1
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" \
    go test -v -tags=integration ./backend/internal/targetjob -run '^TestSQLStoreIntegration_PracticeProgressProjectionPersistsAcrossGetAndList$' -count=1
  go test -v ./backend/internal/review ./backend/internal/store/review -run 'TestGenerateReportPersistsDirectModelSemanticsAndActualProvenance|TestPersistReportUsesPostgresTextArrayForRetryFocus|TestUpdateFeedbackReportStatusAllowsGeneratingRetry' -count=1
  pnpm --filter @easyinterview/frontend test \
    src/app/routeUrl.test.ts \
    src/app/interview-context/roundAssumptions.test.ts \
    src/app/interview-context/startPractice.test.ts \
    src/app/navigation/interviewContext.test.ts \
    src/app/scope.test.ts \
    src/app/screens/home/MockInterviewCard.test.tsx \
    src/app/screens/home/HomeRecentMocks.test.tsx \
    src/app/screens/workspace/WorkspaceEmptyState.test.tsx \
    src/app/screens/parse/ParseFlow.test.tsx \
    src/app/screens/parse/ParseRoundStates.test.tsx \
    src/app/screens/parse/ParseResumeBinding.test.tsx \
    src/app/screens/report/__tests__/ReplayCta.test.tsx \
    --reporter=verbose
  pnpm --filter @easyinterview/frontend test \
    src/app/App.test.tsx \
    --reporter=verbose \
    -t 'renders a target-scoped workspace|uses only targetJobId as workspace detail authority'
  node --test ui-design/ui-design-contract.test.mjs
  pnpm --filter @easyinterview/frontend exec playwright test \
    --config=playwright.auth-email-code.config.ts \
    --reporter=list \
    --workers=1 \
    practice-progress-refresh.spec.ts
} 2>&1 | tee "$OUT/trigger.log"

python3 - "$RESULT_FILE" "$RUN_ID" "$OUT" <<'PY'
import json
import sys
from pathlib import Path

Path(sys.argv[1]).write_text(
    json.dumps(
        {
            "scenario_id": "E2E.P0.098",
            "suite_id": "e2e",
            "mode": "automated",
            "result": "PASS",
            "run_id": sys.argv[2],
            "output_dir": sys.argv[3],
            "live_browser_gate": True,
        },
        ensure_ascii=False,
        indent=2,
    )
    + "\n",
    encoding="utf-8",
)
PY

printf 'scenario=E2E.P0.098\nrun_id=%s\nmethod=postgres-ledger-frontend-projection-and-live-browser-refresh\ntrigger_at=%s\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/trigger.env"
echo "trigger: ok"
