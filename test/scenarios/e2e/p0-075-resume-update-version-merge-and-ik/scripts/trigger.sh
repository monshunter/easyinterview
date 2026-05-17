#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-075-resume-update-version-merge-and-ik"
mkdir -p "$OUT"

{
  echo "E2E.P0.075 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "RUNNER make validate-fixtures"
  cd "$ROOT"
  make validate-fixtures
  echo "RUNNER go test cmd/api resume update HTTP scenario"
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeUpdateVersionHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v
  echo "RUNNER go test resume handler update and fixture parity"
  go test ./internal/resume/handler -run 'TestUpdateResumeVersion|TestUpdateResumeVersionFixtureParity' -count=1 -v
  echo "RUNNER go test resume service update"
  go test ./internal/resume -run 'TestUpdateResumeVersion' -count=1 -v
  echo "RUNNER go test resume store unit update"
  go test ./internal/resume/store -run 'TestUpdateVersionPatch|TestRepositoryExposesResumeAssetMethods' -count=1 -v
  echo "RUNNER go test resume store live update integration"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run 'TestResumeVersionUpdatePatch|TestStructuredMasterUnique|TestResumeVersionListPagination' -count=1 -v
} | tee "$OUT/trigger.log"
