#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-056-generating-to-report-happy-path"
LOG="$OUT/trigger.log"
BACKEND_LOG="$OUT/backend.log"
OWNER_EVIDENCE="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/completion-backend-evidence.json"
ARTIFACT="$OUT/backend-evidence.json"
BACKEND_COMMAND="cd backend && go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0056ReportBackendEvidence$' -count=1 -v"

test -s "$LOG"
test -s "$BACKEND_LOG"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.056

for frontend_file in \
  preflight.test.ts \
  useReportGenerationPoll.test.tsx \
  GeneratingScreen.test.tsx \
  ConversationReport.test.tsx; do
  grep -Fq "$frontend_file" "$LOG" || {
    echo "E2E.P0.056: $frontend_file did not run" >&2
    exit 1
  }
done

test "$(grep -Fc -- '=== RUN   TestE2EP0056ReportBackendEvidence' "$BACKEND_LOG")" -eq 3
test "$(grep -Fc -- '--- PASS: TestE2EP0056ReportBackendEvidence' "$BACKEND_LOG")" -eq 3
for package in internal/review internal/store/review internal/api/reports; do
  grep -Eq "^ok[[:space:]]+github.com/monshunter/easyinterview/backend/$package([[:space:]]|$)" "$BACKEND_LOG"
done
for marker in \
  REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS \
  REPORT_DIRECT_READY_PASS \
  REPORT_FROZEN_CONTEXT_READ_PASS \
  REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS; do
  grep -Fq "$marker" "$BACKEND_LOG"
done
for database_assertion in \
  direct_ready_status=ready \
  frozen_context_read_equal=true \
  legacy_identifier_count=0; do
  grep -Fq "$database_assertion" "$BACKEND_LOG"
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
  echo "E2E.P0.056: failing or empty runner evidence found" >&2
  exit 1
fi
if grep -Eiq 'raw_(cookie|jd|resume|transcript|prompt|output)[=:]|session_cookie=|jd_text=|resume_text=|transcript_text=|prompt_body=|model_output=' "$LOG"; then
  echo "E2E.P0.056: raw sensitive content marker found" >&2
  exit 1
fi

jq -n --arg command "$BACKEND_COMMAND" '{
  schemaVersion: "report-backend-evidence.v1",
  scenarioId: "E2E.P0.056",
  command: $command,
  tests: [
    {package: "internal/review", name: "TestE2EP0056ReportBackendEvidence", status: "PASS"},
    {package: "internal/store/review", name: "TestE2EP0056ReportBackendEvidence", status: "PASS"},
    {package: "internal/api/reports", name: "TestE2EP0056ReportBackendEvidence", status: "PASS"}
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
    "REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS",
    "REPORT_DIRECT_READY_PASS",
    "REPORT_FROZEN_CONTEXT_READ_PASS",
    "REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS"
  ],
  database: {
    directReadyStatus: "ready",
    frozenContextReadEqual: true,
    legacyIdentifierCount: 0
  },
  result: "PASS"
}' > "$ARTIFACT"

jq -e --arg command "$BACKEND_COMMAND" '
  keys == ["command","consumedOwnerEvidence","database","markers","result","scenarioId","schemaVersion","tests"] and
  .schemaVersion == "report-backend-evidence.v1" and
  .scenarioId == "E2E.P0.056" and
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
    "REPORT_COMPLETION_OWNER_EVIDENCE_CONSUMED_PASS",
    "REPORT_DIRECT_READY_PASS",
    "REPORT_FROZEN_CONTEXT_READ_PASS",
    "REPORT_REVIEW_LEGACY_IDENTIFIER_NEGATIVE_PASS"
  ] and
  .database.directReadyStatus == "ready" and
  .database.frozenContextReadEqual == true and
  .database.legacyIdentifierCount == 0
' "$ARTIFACT" >/dev/null
if grep -Eiq 'cookie|jdText|resumeText|transcript|promptBody|rawOutput|sourceMessageSeqNos' "$ARTIFACT"; then
  echo "E2E.P0.056: sensitive/internal key leaked into redacted artifact" >&2
  exit 1
fi

echo "E2E.P0.056 PASS"
