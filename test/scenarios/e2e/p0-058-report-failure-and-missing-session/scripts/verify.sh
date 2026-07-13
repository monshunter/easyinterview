#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-058-report-failure-and-missing-session"
LOG="$OUT/trigger.log"
BACKEND_LOG="$OUT/backend.log"
OWNER_EVIDENCE="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/completion-backend-evidence.json"
ARTIFACT="$OUT/backend-evidence.json"
BACKEND_COMMAND="cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0058ReportFailureBackendEvidence$' -count=1 -v"

test -s "$LOG"
test -s "$BACKEND_LOG"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.058

for frontend_file in \
  preflight.test.ts \
  ReportFailureState.test.tsx \
  ReportMissingSessionState.test.tsx \
  useFeedbackReport.test.tsx \
  ConversationReport.test.tsx \
  useReportGenerationPoll.test.tsx \
  GeneratingScreen.test.tsx; do
  grep -Fq "$frontend_file" "$LOG" || {
    echo "E2E.P0.058: $frontend_file did not run" >&2
    exit 1
  }
done
for frontend_assertion in \
  'missing reportId stays in error and never fetches' \
  'surfaces an exhausted network check separately from a resource timeout and allows checking again' \
  'terminal report failures hide reload and keep the back action' \
  'renders REPORT_CONTEXT_TOO_LARGE as a back-only terminal state with actionable shorter-input copy' \
  'fails closed when non-empty replay focus is not backed by a needs-work dimension and same-code issue'; do
  grep -Fq "$frontend_assertion" "$LOG"
done

test "$(grep -Fxc -- '=== RUN   TestE2EP0058ReportFailureBackendEvidence' "$BACKEND_LOG")" -eq 3
test "$(grep -Ec -- '^--- PASS: TestE2EP0058ReportFailureBackendEvidence \(' "$BACKEND_LOG")" -eq 3
for package in internal/review internal/store/review internal/api/reports; do
  grep -Eq "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/$package([[:space:]]|$)" "$BACKEND_LOG"
done
for marker in \
  REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS \
  REPORT_CONTEXT_TOO_LARGE_PASS \
  REPORT_OUTPUT_RETRY_PASS \
  REPORT_FOUR_INVALID_FAIL_CLOSED_PASS \
  REPORT_ACTION_RETRY_RESET_PASS \
  REPORT_RETRY_LAYER_SEPARATION_PASS; do
  grep -Fq "$marker" "$BACKEND_LOG"
done
for database_assertion in \
  context_mismatch_fail_closed=true \
  context_too_large_status=failed \
  four_invalid_status=failed \
  failed_ready_columns_empty=true; do
  grep -Fq "$database_assertion" "$BACKEND_LOG"
done
for runtime_assertion in \
  context_too_large_provider_calls=0 \
  output_retry_provider_calls=2 \
  four_invalid_provider_calls=4 \
  first_action_call_count=4 \
  second_action_initial_attempt=1 \
  retry_state_destroyed_after_action=true \
  action_retry_schedule=10s,20s,40s \
  async_attempts_affect_product_attempt=false \
  attempt_four_terminal=true; do
  grep -Fq "$runtime_assertion" "$BACKEND_LOG"
done

jq -e '
  .schemaVersion == "practice-completion-evidence.v1" and
  .scenarioId == "E2E.P0.047" and
  .result == "PASS" and
  (.markers | sort) == [
    "REPORT_CONTEXT_REPLAY_PASS",
    "REPORT_CONTEXT_SNAPSHOT_PASS",
    "ZERO_ANSWER_COMPLETION_REJECTED_PASS"
  ]
' "$OWNER_EVIDENCE" >/dev/null

if grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]' "$LOG"; then
  echo "E2E.P0.058: failing or empty runner evidence found" >&2
  exit 1
fi
if grep -Eiq 'raw_(cookie|jd|resume|transcript|prompt|output)[=:]|session_cookie=|jd_text=|resume_text=|transcript_text=|prompt_body=|model_output=' "$LOG"; then
  echo "E2E.P0.058: raw sensitive content marker found" >&2
  exit 1
fi

jq -n --arg command "$BACKEND_COMMAND" '{
  schemaVersion: "report-backend-evidence.v3",
  scenarioId: "E2E.P0.058",
  command: $command,
  tests: [
    {package: "internal/review", name: "TestE2EP0058ReportFailureBackendEvidence", status: "PASS"},
    {package: "internal/store/review", name: "TestE2EP0058ReportFailureBackendEvidence", status: "PASS"},
    {package: "internal/api/reports", name: "TestE2EP0058ReportFailureBackendEvidence", status: "PASS"}
  ],
  consumedOwnerEvidence: {
    schemaVersion: "practice-completion-evidence.v1",
    scenarioId: "E2E.P0.047",
    result: "PASS",
    zeroAnswerCompletionRejected: true,
    reportContextSnapshot: true,
    reportContextReplay: true
  },
  markers: [
    "REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS",
    "REPORT_CONTEXT_TOO_LARGE_PASS",
    "REPORT_OUTPUT_RETRY_PASS",
    "REPORT_FOUR_INVALID_FAIL_CLOSED_PASS",
    "REPORT_ACTION_RETRY_RESET_PASS",
    "REPORT_RETRY_LAYER_SEPARATION_PASS"
  ],
  database: {
    contextMismatchFailClosed: true,
    contextTooLargeStatus: "failed",
    fourInvalidStatus: "failed",
    failedReadyColumnsEmpty: true
  },
  runtime: {
    contextTooLargeProviderCalls: 0,
    outputRetryProviderCalls: 2,
    fourInvalidProviderCalls: 4,
    firstActionCallCount: 4,
    secondActionInitialAttempt: 1,
    retryStateDestroyedAfterAction: true,
    actionRetryScheduleSeconds: [10, 20, 40],
    asyncAttemptsAffectProductAttempt: false,
    attemptFourTerminal: true
  },
  result: "PASS"
}' > "$ARTIFACT"

jq -e --arg command "$BACKEND_COMMAND" '
  keys == ["command","consumedOwnerEvidence","database","markers","result","runtime","scenarioId","schemaVersion","tests"] and
  .schemaVersion == "report-backend-evidence.v3" and
  .scenarioId == "E2E.P0.058" and
  .command == $command and
  .result == "PASS" and
  (.tests | length) == 3 and
  (.tests | all(.status == "PASS")) and
  ([.tests[].package] | sort) == ["internal/api/reports","internal/review","internal/store/review"] and
  .consumedOwnerEvidence.schemaVersion == "practice-completion-evidence.v1" and
  .consumedOwnerEvidence.result == "PASS" and
  .consumedOwnerEvidence.zeroAnswerCompletionRejected == true and
  .consumedOwnerEvidence.reportContextSnapshot == true and
  .consumedOwnerEvidence.reportContextReplay == true and
  (.markers | sort) == [
    "REPORT_ACTION_RETRY_RESET_PASS",
    "REPORT_CONTEXT_MISMATCH_FAIL_CLOSED_PASS",
    "REPORT_CONTEXT_TOO_LARGE_PASS",
    "REPORT_FOUR_INVALID_FAIL_CLOSED_PASS",
    "REPORT_OUTPUT_RETRY_PASS",
    "REPORT_RETRY_LAYER_SEPARATION_PASS"
  ] and
  .database.contextMismatchFailClosed == true and
  .database.contextTooLargeStatus == "failed" and
  .database.fourInvalidStatus == "failed" and
  .database.failedReadyColumnsEmpty == true and
  (.database | keys) == [
    "contextMismatchFailClosed",
    "contextTooLargeStatus",
    "failedReadyColumnsEmpty",
    "fourInvalidStatus"
  ] and
  .runtime.contextTooLargeProviderCalls == 0 and
  .runtime.outputRetryProviderCalls == 2 and
  .runtime.fourInvalidProviderCalls == 4 and
  .runtime.firstActionCallCount == 4 and
  .runtime.secondActionInitialAttempt == 1 and
  .runtime.retryStateDestroyedAfterAction == true and
  .runtime.actionRetryScheduleSeconds == [10,20,40] and
  .runtime.asyncAttemptsAffectProductAttempt == false and
  .runtime.attemptFourTerminal == true and
  (.runtime | keys) == [
    "actionRetryScheduleSeconds",
    "asyncAttemptsAffectProductAttempt",
    "attemptFourTerminal",
    "contextTooLargeProviderCalls",
    "firstActionCallCount",
    "fourInvalidProviderCalls",
    "outputRetryProviderCalls",
    "retryStateDestroyedAfterAction",
    "secondActionInitialAttempt"
  ]
' "$ARTIFACT" >/dev/null
if grep -Eiq 'cookie|jdText|resumeText|transcript|promptBody|rawOutput|sourceMessageSeqNos' "$ARTIFACT"; then
  echo "E2E.P0.058: sensitive/internal key leaked into redacted artifact" >&2
  exit 1
fi

echo "E2E.P0.058 PASS"
