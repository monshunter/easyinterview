#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready"
mkdir -p "$OUT"

{
  echo "E2E.P0.077 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  echo "RUNNER make validate-fixtures"
  cd "$ROOT"
  make validate-fixtures
  echo "RUNNER go test cmd/api branch ai_select dispatch"
  cd "$ROOT/backend"
  go test ./cmd/api -run TestResumeBranchVersionHTTPScenario -count=1 -v
  echo "RUNNER go test resume handler branch fixture parity"
  go test ./internal/resume/handler -run TestBranchResumeVersionFixtureParity -count=1 -v
  echo "RUNNER go test resume service branch ai_select"
  go test ./internal/resume -run TestBranchResumeVersionRoutesSeedStrategies -count=1 -v
  echo "RUNNER go test resume store live branch dispatch integration"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestBranchVersion -count=1 -v
  echo "RUNNER go test cmd/api resume tailor endpoints"
  go test ./cmd/api -run TestResumeTailorEndpointsHTTPScenario -count=1 -v
  echo "RUNNER go test resume handler tailor fixture parity"
  go test ./internal/resume/handler -run TestResumeTailorFixtureParity -count=1 -v
  echo "RUNNER go test resume service request get tailor"
  go test ./internal/resume -run 'TestRequestResumeTailor|TestGetResumeTailorRun' -count=1 -v
  echo "RUNNER go test resume store live tailor run integration"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestResumeTailorRunStore -count=1 -v
} | tee "$OUT/trigger.log"
