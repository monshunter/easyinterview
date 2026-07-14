#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
LOG="$OUT/trigger.log"
OWNER_LOG="$OUT/completion-owner.log"
DATABASE_LOG="$OUT/completion-database.log"
FRONTEND_LOG="$OUT/frontend-completion-contract.log"
ARTIFACT="$OUT/completion-backend-evidence.json"
OWNER_COMMAND="cd backend && go test ./internal/api/practice ./internal/practice ./internal/store/practice -run '^(TestE2EP0047RejectsZeroAnswerCompletion|TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' -count=1 -v"

"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.047
grep -Fq 'ZERO_ANSWER_FINISH_DISABLED_PASS' "$FRONTEND_LOG"
grep -Fq 'keeps Finish disabled while the latest committed candidate message still awaits an assistant reply' "$FRONTEND_LOG"
grep -Fq 'P0.047 hands completion to Generating with reportId as the only URL and history locator' "$FRONTEND_LOG"
grep -Fq 'retry of the same complete reuses the same Idempotency-Key' "$FRONTEND_LOG"

for test_name in \
  TestE2EP0047RejectsZeroAnswerCompletion \
  TestE2EP0047FreezesReportContext \
  TestE2EP0047CompletionReplayPreservesReportContext; do
  grep -Fq -- "=== RUN   $test_name" "$OWNER_LOG"
done
grep -Fq -- '--- PASS: TestE2EP0047RejectsZeroAnswerCompletion' "$OWNER_LOG"
grep -Fq -- '--- PASS: TestE2EP0047FreezesReportContext' "$OWNER_LOG"
grep -Fq -- '--- PASS: TestE2EP0047CompletionReplayPreservesReportContext' "$OWNER_LOG"
for marker in \
  ZERO_ANSWER_COMPLETION_REJECTED_PASS \
  REPORT_CONTEXT_SNAPSHOT_PASS \
  REPORT_CONTEXT_REPLAY_PASS; do
  grep -Fq "$marker" "$OWNER_LOG"
  grep -Fq "$marker" "$DATABASE_LOG"
done
grep -Fq -- '--- PASS: TestIntegrationE2EP0047RejectsZeroAnswerCompletion' "$DATABASE_LOG"
for database_assertion in \
  zero_answer_side_effect_count=0 \
  pending_reply_side_effect_count=0 \
  snapshot_schema_version=report-context.v1 \
  concurrent_mutation_blocked=true \
  snapshot_replay_equal=true \
  mismatch_side_effect_count=0; do
  grep -Fq "$database_assertion" "$DATABASE_LOG"
done
grep -Fq -- '--- PASS: TestSQLRepositoryCommitPracticeMessageRejectsClosedSession' "$OUT/completion-regression.log"
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG"

jq -n --arg command "$OWNER_COMMAND" '{
  schemaVersion: "practice-completion-evidence.v1",
  scenarioId: "E2E.P0.047",
  command: $command,
  tests: [
    {name: "TestE2EP0047RejectsZeroAnswerCompletion", status: "PASS"},
    {name: "TestE2EP0047FreezesReportContext", status: "PASS"},
    {name: "TestE2EP0047CompletionReplayPreservesReportContext", status: "PASS"}
  ],
  markers: [
    "ZERO_ANSWER_COMPLETION_REJECTED_PASS",
    "REPORT_CONTEXT_SNAPSHOT_PASS",
    "REPORT_CONTEXT_REPLAY_PASS"
  ],
  database: {
    zeroAnswerSideEffectCount: 0,
    pendingReplySideEffectCount: 0,
    snapshotSchemaVersion: "report-context.v1",
    concurrentMutationBlocked: true,
    snapshotReplayEqual: true,
    mismatchSideEffectCount: 0
  },
  result: "PASS"
}' > "$ARTIFACT"

jq -e '
  keys == ["command","database","markers","result","scenarioId","schemaVersion","tests"] and
  .schemaVersion == "practice-completion-evidence.v1" and
  .scenarioId == "E2E.P0.047" and
  .result == "PASS" and
  (.tests | length) == 3 and
  (.tests | all(.status == "PASS")) and
  .database.zeroAnswerSideEffectCount == 0 and
  .database.pendingReplySideEffectCount == 0 and
  .database.snapshotSchemaVersion == "report-context.v1" and
  .database.concurrentMutationBlocked == true and
  .database.snapshotReplayEqual == true and
  .database.mismatchSideEffectCount == 0
' "$ARTIFACT" >/dev/null

echo 'E2E.P0.047 PASS'
