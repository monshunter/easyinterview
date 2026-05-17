#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-074-resume-confirm-master-and-version-reads"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.074 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER make validate-fixtures' "$LOG"
  grep -q 'validate-fixtures: OK' "$LOG"
  grep -q 'RUNNER go test cmd/api resume version HTTP scenarios' "$LOG"
  grep -q 'TestResumeConfirmStructuredMasterHTTPScenario' "$LOG"
  grep -q 'TestResumeVersionReadHTTPScenario' "$LOG"
  grep -q 'TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' "$LOG"
  grep -q 'RUNNER go test resume handler confirm and version parity' "$LOG"
  grep -q 'TestConfirmStructuredMaster' "$LOG"
  grep -q 'TestResumeVersionReadFixtureParity' "$LOG"
  grep -q 'RUNNER go test resume service version reads' "$LOG"
  grep -q 'TestConfirmStructuredMasterCreatesStructuredMasterVersion' "$LOG"
  grep -q 'TestGetAndListResumeVersions' "$LOG"
  grep -q 'TestResumeVersionReadMapsStoreErrors' "$LOG"
  grep -q 'RUNNER go test resume store unit reads' "$LOG"
  grep -q 'TestCreateStructuredMasterFromAssetInsertsReadyAssetMaster' "$LOG"
  grep -q 'TestGetVersionByIDScopesUser' "$LOG"
  grep -q 'TestListVersionsByAssetScopesAssetAndPaginates' "$LOG"
  grep -q 'RUNNER go test resume store live integration' "$LOG"
  grep -q 'TestStructuredMasterUniqueCrossUserReadinessAndSoftDelete' "$LOG"
  grep -q 'TestResumeVersionListPaginationCrossUserAndCursor' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT/backend"
  go test ./internal/resume/handler -run TestResumeVersionReadFixtureParity -count=1
  cd "$ROOT"
  if rg -n 'inline|rewrite|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/rewrite/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|Checkout reliability|suggested bullet text' "$OUT"; then
    echo "ERROR: private resume or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "fixture parity: confirmResumeStructuredMaster/getResumeVersion/listResumeVersions"
  echo "DB state: partial unique index, parse-not-ready, cross-user isolation, cursor pagination"
  echo "privacy: no raw resume or suggestion text in scenario evidence"
} | tee "$OUT/verify.log"
