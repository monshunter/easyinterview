#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-074-resume-flat-read-api"
mkdir -p "$OUT"

{
  echo "E2E.P0.074 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "RUNNER make validate-fixtures"
  cd "$ROOT"
  make validate-fixtures
  echo "RUNNER go test cmd/api non-current resume route gate"
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v
  echo "RUNNER go test resume handler flat reads"
  go test ./internal/resume/handler -run 'Test(GetResume|GetResumeFixtureParity|GetResumeNotFoundAndCrossUserReturns404|ListResumesFixtureParity|ListResumesInvalidCursorReturnsUnprocessableEntity)' -count=1 -v
  echo "RUNNER go test resume service flat reads"
  go test ./internal/resume -run 'TestGetAndListResumesMapStoreRecordsWithUserScope|TestGetResumeMapsStoreNotFound' -count=1 -v
  echo "RUNNER go test resume store flat reads"
  go test ./internal/resume/store -run 'Test(GetScopesUserAndMapsStructuredProfile|ListCursorPagination|RepositoryExposesFlatResumeMethods)' -count=1 -v
} | tee "$OUT/trigger.log"
