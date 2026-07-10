#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-078-resume-tailor-failure-and-retry"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.078 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER go test cmd/api resume tailor drainer failure' "$LOG"
  grep -q 'TestResumeTailorDrainerFailureScenario' "$LOG"
  grep -q 'RUNNER go test resume jobs tailor failure' "$LOG"
  grep -q 'TestTailorHandlerModeRoutingAndFailurePaths' "$LOG"
  grep -q 'RUNNER go test resume store live ready-only outbox integration' "$LOG"
  grep -q 'TestCompleteTailorRunSuccessWritesResultAndOutbox' "$LOG"
  grep -q 'AI_PROVIDER_TIMEOUT' "$LOG"
  grep -q 'AI_OUTPUT_INVALID' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/jobs([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT"
  if rg -n 'inline|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: out-of-scope inline/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: out-of-scope mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|suggested bullet text|raw resume text|match_summary' "$OUT"; then
    echo "ERROR: private resume, suggestion, or match-summary content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "failure: timeout retryable + output_invalid terminal + retry to ready"
  echo "privacy: failure evidence contains error codes and IDs only"
} | tee "$OUT/verify.log"
