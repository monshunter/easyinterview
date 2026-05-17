#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-079-resume-suggestion-accept-reject-terminal"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.079 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER make validate-fixtures suggestion decision fixtures' "$LOG"
  grep -q 'validate-fixtures: OK' "$LOG"
  grep -q 'RUNNER go test cmd/api suggestion accept reject' "$LOG"
  grep -q 'TestResumeSuggestionAcceptRejectHTTPScenario' "$LOG"
  grep -q 'RUNNER go test handler suggestion fixture parity' "$LOG"
  grep -q 'TestResumeSuggestionDecisionFixtureParity' "$LOG"
  grep -q 'RUNNER go test service suggestion decision' "$LOG"
  grep -q 'TestResumeSuggestionDecisionRoutesAcceptRejectToStore' "$LOG"
  grep -q 'RUNNER go test store live suggestion decision CAS' "$LOG"
  grep -q 'TestResumeSuggestionDecisionCASIsolationAndProfileStability' "$LOG"
  grep -q 'status=accepted' "$LOG"
  grep -q 'status=rejected' "$LOG"
  grep -q 'SUGGESTION_ALREADY_DECIDED' "$LOG"
  grep -q 'structured_profile=unchanged' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT"
  if rg -n 'inline|rewrite|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/rewrite/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|raw resume text|full suggested bullet text|match_summary' "$OUT"; then
    echo "ERROR: private resume or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "terminal: accept + reject + already-decided 409"
  echo "isolation: cross-user 404"
  echo "profile: structured_profile unchanged"
} | tee "$OUT/verify.log"
