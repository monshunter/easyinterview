#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-075-resume-update-flat-fields-and-ik"
mkdir -p "$OUT"

{
  echo "E2E.P0.075 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "RUNNER make validate-fixtures"
  cd "$ROOT"
  make validate-fixtures
  echo "RUNNER go test cmd/api flat resume update HTTP scenario"
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeRegisterListHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v
  echo "RUNNER go test resume handler update and fixture parity"
  go test ./internal/resume/handler -run 'TestUpdateResume(OverwritesEditableFields|ValidationAndErrors|RequiresIdempotencyKey|FixtureParity)' -count=1 -v
  echo "RUNNER go test resume service update"
  go test ./internal/resume -run 'TestUpdateResume(OverwritesAndStripsProvenance|ValidationAndStoreErrors)' -count=1 -v
  echo "RUNNER go test resume store unit update"
  go test ./internal/resume/store -run 'Test(UpdateResumeOverwritesProfileAndScopesUser|UpdateResumeNotFound|RepositoryExposesFlatResumeMethods)' -count=1 -v
} | tee "$OUT/trigger.log"
