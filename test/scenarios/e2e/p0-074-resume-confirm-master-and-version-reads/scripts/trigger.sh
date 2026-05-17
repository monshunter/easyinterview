#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-074-resume-confirm-master-and-version-reads"
mkdir -p "$OUT"

{
  echo "E2E.P0.074 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "RUNNER make validate-fixtures"
  cd "$ROOT"
  make validate-fixtures
  echo "RUNNER go test cmd/api resume version HTTP scenarios"
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeConfirmStructuredMasterHTTPScenario|TestResumeVersionReadHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v
  echo "RUNNER go test resume handler confirm and version parity"
  go test ./internal/resume/handler -run 'Test(ConfirmStructuredMaster|ResumeVersionReadFixtureParity|GetResumeVersion|ListResumeVersions)' -count=1 -v
  echo "RUNNER go test resume service version reads"
  go test ./internal/resume -run 'Test(ConfirmStructuredMaster|GetAndListResumeVersions|ResumeVersionReadMapsStoreErrors)' -count=1 -v
  echo "RUNNER go test resume store unit reads"
  go test ./internal/resume/store -run 'Test(CreateStructuredMaster|GetVersionByID|ListVersionsByAsset)' -count=1 -v
  echo "RUNNER go test resume store live integration"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run 'TestStructuredMasterUnique|TestResumeVersionListPagination' -count=1 -v
} | tee "$OUT/trigger.log"
