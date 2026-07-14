#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
LOG="$OUT/trigger.log"
RESULT_FILE="$OUT/result.json"
test -s "$LOG"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.098

for marker in \
  "TestE2EP0OperationMatrixPreflight" \
  "TestBuildResumeRuntimeWrapsParseAIWithObservability" \
  "TestSQLRepositoryCreatePlanUsesConversationColumns" \
  "TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns" \
  "TestSQLRepositoryCompleteSessionReplayDoesNotAppendSecondCompletedFact" \
  "TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger" \
  "TestService_GetTargetJob_HidesCompletedFactsAfterFirstCanonicalGap" \
  "TestSQLStoreIntegration_PracticeProgressProjectionPersistsAcrossGetAndList" \
  "TestGenerateReportPersistsDirectModelSemanticsAndActualProvenance" \
  "TestPersistReportUsesPostgresTextArrayForRetryFocus" \
  "TestUpdateFeedbackReportStatusAllowsGeneratingRetry"; do
  grep -Fq -- "--- PASS: $marker" "$LOG"
done

for marker in \
  "target-resume-binding-and-provenance=PASS" \
  "canonical-round-type-case-sensitive=PASS" \
  "canonical-round-prompt-context=PASS" \
  "start-and-send-bound-resume-fail-closed=PASS" \
  "completed-ledger-successor=PASS" \
  "non-contiguous-successor=PASS" \
  "round-budget-mismatch=PASS" \
  "all-rounds-complete-fail-closed=PASS" \
  "wrong-resume-completion-ignored=PASS" \
  "persisted-first-to-next=PASS" \
  "target-report-status-independent=PASS" \
  "out-of-order-gap-hidden=PASS" \
  "non-contiguous-round-1-2-4=PASS" \
  "get-list-first-next-final-parity=PASS" \
  "App.test.tsx" \
  "routeUrl.test.ts" \
  "roundAssumptions.test.ts" \
  "startPractice.test.ts" \
  "scope.test.ts" \
  "MockInterviewCard.test.tsx" \
  "HomeRecentMocks.test.tsx" \
  "WorkspaceEmptyState.test.tsx" \
  "ParseFlow.test.tsx" \
  "ParseRoundStates.test.tsx" \
  "ParseResumeBinding.test.tsx" \
  "ReplayCta.test.tsx" \
  "prototype round progress is backend-projected and never inferred from lifecycle text" \
  "prototype does not persist interview business progress in browser state" \
  "SCENARIO_RUNNER=E2E.P0.098" \
  "PLAYWRIGHT_SPEC=frontend/tests/e2e/practice-progress-refresh.spec.ts" \
  "PLAYWRIGHT_CONFIG=frontend/playwright.auth-email-code.config.ts" \
  "practice-progress-refresh.spec.ts" \
  "E2E.P0.098 live completion API PASS completionStatus=202 persistedFact=session_completed" \
  "E2E.P0.098 workspace refresh PASS states=done,current,pending currentRound=round-2-technical currentRoundSequence=2" \
  "E2E.P0.098 home and workspace detail refresh PASS homeStates=done,current,pending detailCurrentRound=round-2-technical detailCurrentRoundSequence=2" \
  "E2E.P0.098 ready cards direct detail PASS sources=workspace,home route=/workspace?targetJobId=019f6098-0000-7000-8000-000000000003 perVisitGetTargetJob=1 importTargetJob=0 parsePoll=0" \
  "E2E.P0.098 workspace detail refresh PASS states=done,current,pending labels=已进行,即将进行,未进行 visualStyles=distinct" \
  "E2E.P0.098 next plan POST PASS requestRoundId=round-2-technical responseRoundId=round-2-technical responseRoundSequence=2 persistedRoundSequence=2" \
  "E2E.P0.098 session start interception PASS realPlanCreate=true aiSessionStart=intercepted" \
  "1 passed"; do
  grep -Fq -- "$marker" "$LOG"
done

grep -Fq 'validate-fixtures: OK' "$LOG"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|--- SKIP:|no tests to run|\[no tests to run\]' "$LOG"
! rg -n 'appendSessionEvent|/practice/sessions/\{sessionId\}/events|client_event_id|practice_turns' \
  "$ROOT/backend/cmd/api/e2e_p0_operation_matrix_preflight_test.go" \
  "$ROOT/backend/internal/api/practice" \
  "$ROOT/backend/internal/store/practice" \
  "$ROOT/backend/internal/review" \
  "$ROOT/backend/internal/store/review"
! rg -n 'roundIndexFromTargetJobStatus|getHomeRoundIndex|getCurrentRoundIndex' \
  "$ROOT/frontend/src" "$ROOT/ui-design/src"
! rg -n 'sequence[[:space:]]*!==[[:space:]]*index[[:space:]]*\+[[:space:]]*1' \
  "$ROOT/frontend/src/app/interview-context" "$ROOT/ui-design/src"
! rg -n 'toHaveURL\(/\\/parse|home and parse refresh|parseCurrentRound' \
  "$ROOT/frontend/tests/e2e/practice-progress-refresh.spec.ts"

if grep -Eq -- '(mailCode|code|token)=[0-9]{6}|ei_session=|SESSION_COOKIE_SECRET|AUTH_CHALLENGE_TOKEN_PEPPER' "$LOG"; then
  echo "verify: credential or raw email code leaked into trigger.log" >&2
  exit 1
fi

python3 - "$RESULT_FILE" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
if not path.is_file():
    raise SystemExit("verify: missing result.json")
payload = json.loads(path.read_text(encoding="utf-8"))
expected = {
    "scenario_id": "E2E.P0.098",
    "suite_id": "e2e",
    "mode": "automated",
    "result": "PASS",
    "live_browser_gate": True,
}
for key, value in expected.items():
    if payload.get(key) != value:
        raise SystemExit(f"verify: result.json {key} mismatch")
if not payload.get("run_id"):
    raise SystemExit("verify: result.json missing run_id")
PY

test -d "$OUT/playwright"

echo "verify: ok"
