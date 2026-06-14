#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-076-resume-branch-version-sync-paths"
mkdir -p "$OUT"

{
  echo "E2E.P0.076 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "RUNNER make validate-fixtures"
  cd "$ROOT"
  make validate-fixtures
  echo "RUNNER go test cmd/api flat resume duplicate HTTP scenario"
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeRegisterListHTTPScenario|TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware' -count=1 -v
  echo "RUNNER go test resume handler duplicate and fixture parity"
  go test ./internal/resume/handler -run 'TestDuplicateResume(Returns201|AllowsEmptyBody|ValidationAndErrors|RequiresIdempotencyKey|FixtureParity)' -count=1 -v
  echo "RUNNER go test resume service duplicate"
  go test ./internal/resume -run 'TestDuplicateResume(AllocatesNewIDAndAppliesProfile|ValidationAndStoreErrors)' -count=1 -v
  echo "RUNNER go test resume store unit duplicate"
  go test ./internal/resume/store -run 'Test(DuplicateResumeCopiesSourceSnapshotAndAppliesProfile|DuplicateResumeSourceNotFoundRollsBack|RepositoryExposesFlatResumeMethods)' -count=1 -v
} | tee "$OUT/trigger.log"
