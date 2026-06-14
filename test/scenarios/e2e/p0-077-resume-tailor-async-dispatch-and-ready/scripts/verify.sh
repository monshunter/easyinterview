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
  grep -q 'RUNNER go test resume store unit tailor run' "$LOG"
  grep -q 'TestCreateTailorRunInsertsAsyncJobWithResumePayload' "$LOG"
  grep -q 'TestCompleteTailorRunSuccessWritesResultAndOutbox' "$LOG"
  grep -q 'RUNNER go test cmd/api resume tailor drainer ready' "$LOG"
  grep -q 'TestResumeTailorDrainerHTTPScenario' "$LOG"
  grep -q 'RUNNER go test resume jobs tailor ready' "$LOG"
  grep -q 'TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT/backend"
  go test ./internal/resume/handler -run TestResumeTailorFixtureParity -count=1
  cd "$ROOT"
  if rg -n 'inline|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/mirror vocabulary found"
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
  echo "fixture parity: requestResumeTailor + getResumeTailorRun"
  echo "DB state: queued flat resume tailor ai_task_run, status transitions, ready suggestions, ready-only completed outbox"
  echo "privacy: async dispatch, tailor endpoint, drainer, and completed outbox evidence contains IDs/status/provenance only"
} | tee "$OUT/verify.log"
