#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-076-resume-branch-version-sync-paths"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.076 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER make validate-fixtures' "$LOG"
  grep -q 'validate-fixtures: OK' "$LOG"
  grep -q 'RUNNER go test cmd/api branch version HTTP scenario' "$LOG"
  grep -q 'TestResumeBranchVersionHTTPScenario' "$LOG"
  grep -q 'RUNNER go test resume handler branch and fixture parity' "$LOG"
  grep -q 'TestBranchResumeVersionFixtureParity' "$LOG"
  grep -q 'RUNNER go test resume service branch' "$LOG"
  grep -q 'TestBranchResumeVersionRoutesSeedStrategies' "$LOG"
  grep -q 'RUNNER go test resume store unit branch' "$LOG"
  grep -q 'TestRepositoryExposesResumeAssetMethods' "$LOG"
  grep -q 'RUNNER go test resume store live branch integration' "$LOG"
  grep -q 'TestBranchVersionInsertStrategiesCrossUserAndRollback' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT/backend"
  go test ./internal/resume/handler -run TestBranchResumeVersionFixtureParity -count=1
  cd "$ROOT"
  if rg -n 'inline|rewrite|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/rewrite/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|suggested bullet text|raw resume text' "$OUT"; then
    echo "ERROR: private resume or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "fixture parity: branchResumeVersion sync scenarios"
  echo "DB state: copy_master provenance reset, blank profile, cross-user target isolation, rollback"
  echo "privacy: no raw resume or suggestion text in scenario evidence"
} | tee "$OUT/verify.log"
