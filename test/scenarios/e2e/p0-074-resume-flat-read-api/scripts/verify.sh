#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-074-resume-flat-read-api"
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
  grep -q 'RUNNER go test cmd/api non-current resume route gate' "$LOG"
  grep -q 'TestResumeVersionRoutesRemainUnmountedPerD20' "$LOG"
  grep -q 'TestGeneratedRouteCatalogHasNoResumeVersionOperations' "$LOG"
  grep -q 'TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' "$LOG"
  grep -q 'RUNNER go test resume handler flat reads' "$LOG"
  grep -q 'TestGetResumeFixtureParity' "$LOG"
  grep -q 'TestListResumesFixtureParity' "$LOG"
  grep -q 'RUNNER go test resume service flat reads' "$LOG"
  grep -q 'TestGetAndListResumesMapStoreRecordsWithUserScope' "$LOG"
  grep -q 'RUNNER go test resume store flat reads' "$LOG"
  grep -q 'TestGetScopesUserAndMapsStructuredProfile' "$LOG"
  grep -q 'TestRepositoryExposesFlatResumeMethods' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations' -count=1
  cd "$ROOT"
  if rg -n 'inline|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: non-current inline/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: non-current mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|Checkout reliability|suggested bullet text' "$OUT"; then
    echo "ERROR: private resume or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "fixture parity: getResume/listResumes"
  echo "non-current routes: resume route family returns 404 and is absent from generated catalog"
  echo "privacy: no raw resume or suggestion text in scenario evidence"
} | tee "$OUT/verify.log"
