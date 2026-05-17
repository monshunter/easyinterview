#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.077 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER make validate-fixtures' "$LOG"
  grep -q 'validate-fixtures: OK' "$LOG"
  grep -q 'RUNNER go test cmd/api branch ai_select dispatch' "$LOG"
  grep -q 'TestResumeBranchVersionHTTPScenario' "$LOG"
  grep -q 'RUNNER go test resume handler branch fixture parity' "$LOG"
  grep -q 'TestBranchResumeVersionFixtureParity' "$LOG"
  grep -q 'ai-select-202-with-job' "$LOG"
  grep -q 'RUNNER go test resume service branch ai_select' "$LOG"
  grep -q 'TestBranchResumeVersionRoutesSeedStrategies' "$LOG"
  grep -q 'RUNNER go test resume store live branch dispatch integration' "$LOG"
  grep -q 'TestBranchVersionInsertStrategiesCrossUserAndRollback' "$LOG"
  grep -q 'RUNNER go test cmd/api resume tailor endpoints' "$LOG"
  grep -q 'TestResumeTailorEndpointsHTTPScenario' "$LOG"
  grep -q 'RUNNER go test resume handler tailor fixture parity' "$LOG"
  grep -q 'TestResumeTailorFixtureParity' "$LOG"
  grep -q 'request_idempotency_replay' "$LOG"
  grep -q 'get_queued' "$LOG"
  grep -q 'get_generating' "$LOG"
  grep -q 'get_failed' "$LOG"
  grep -q 'RUNNER go test resume service request get tailor' "$LOG"
  grep -q 'TestRequestResumeTailorCreatesQueuedRunAndJob' "$LOG"
  grep -q 'TestGetResumeTailorRunMapsStatusesAndErrors' "$LOG"
  grep -q 'RUNNER go test resume store live tailor run integration' "$LOG"
  grep -q 'TestResumeTailorRunStoreStateTransitionsIsolationAndClaim' "$LOG"
  grep -q 'RUNNER go test cmd/api resume tailor drainer ready' "$LOG"
  grep -q 'TestResumeTailorDrainerHTTPScenario' "$LOG"
  grep -q 'RUNNER go test resume jobs tailor ready' "$LOG"
  grep -q 'TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox' "$LOG"
  grep -q 'RUNNER go test resume store live tailor ready outbox integration' "$LOG"
  grep -q 'TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT/backend"
  go test ./internal/resume/handler -run TestBranchResumeVersionFixtureParity -count=1
  go test ./internal/resume/handler -run TestResumeTailorFixtureParity -count=1
  cd "$ROOT"
  if rg -n 'inline|rewrite|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/rewrite/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|suggested bullet text|raw resume text|match_summary' "$OUT"; then
    echo "ERROR: private resume, suggestion, or match-summary content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "fixture parity: branchResumeVersion ai-select-202-with-job + requestResumeTailor + getResumeTailorRun"
  echo "DB state: provisional version, queued resume_tailor_run, queued resume_tailor async_job, status transitions, concurrent claim, rollback, ready suggestions, ready-only completed outbox"
  echo "privacy: async dispatch, tailor endpoint, drainer, and completed outbox evidence contains IDs/status/provenance only"
} | tee "$OUT/verify.log"
