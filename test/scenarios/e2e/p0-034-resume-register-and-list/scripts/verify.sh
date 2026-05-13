#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-034-resume-register-and-list"
mkdir -p "$OUT"

{
  echo "E2E.P0.034 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$OUT/trigger.log"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$OUT/trigger.log"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'TestResumeRegisterListHTTPScenario' "$OUT/trigger.log"
  grep -q 'TestResumeRegisterListHTTPValidationScenario' "$OUT/trigger.log"
  grep -q 'TestRegisterResumeFixtureParity' "$OUT/trigger.log"
  grep -q 'TestRegisterResumeValidationErrorsReturnUnprocessableEntity' "$OUT/trigger.log"
  grep -q 'TestGetResumeFixtureParity' "$OUT/trigger.log"
  grep -q 'TestListResumesFixtureParity' "$OUT/trigger.log"
  grep -q 'TestListResumesInvalidCursorReturnsUnprocessableEntity' "$OUT/trigger.log"
  grep -q 'TestRegisterFileObjectRejectsMissingObjectAndIllegalStates' "$OUT/trigger.log"
  grep -q 'TestCreateWithParseJobRollsBackWhenJobInsertFails' "$OUT/trigger.log"
  grep -q 'TestListCursorPagination' "$OUT/trigger.log"
  grep -q 'TestResumeAssetsIntegrationCRUDStateIsolationPaginationAndRollback' "$OUT/trigger.log"
  cd "$ROOT/backend"
  go test ./internal/resume/handler -run 'Test(RegisterResumeFixtureParity|GetResumeFixtureParity|ListResumesFixtureParity)' -count=1
  cd "$ROOT"
  if rg -n 'inline|rewrite|mirror' backend/internal/resume backend/cmd/api/resume_http_scenario_test.go --glob '!**/verify.sh'; then
    exit 1
  fi
  if rg -n 'mistake|growth|drill' backend/internal/resume; then
    exit 1
  fi
  if rg -n 'Private resume body|Checkout reliability' "$OUT"; then
    echo "ERROR: private resume content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "fixture parity: registerResume/getResume/listResumes"
  echo "DB state machine: asset/job atomic create, rollback, cursor pagination, cross-user isolation"
  echo "upload handoff: RegisterFileObject missing object and size mismatch rejection"
} | tee "$OUT/verify.log"
