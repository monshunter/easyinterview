#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-078-resume-tailor-failure-and-retry"
mkdir -p "$OUT"

{
  echo "E2E.P0.078 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT/backend"
  echo "RUNNER go test cmd/api resume tailor drainer failure"
  go test ./cmd/api -run TestResumeTailorDrainerFailureScenario -count=1 -v
  echo "evidence error_code=AI_PROVIDER_TIMEOUT"
  echo "evidence error_code=AI_OUTPUT_INVALID"
  echo "RUNNER go test resume jobs tailor failure"
  go test ./internal/resume/jobs -run TestTailorHandlerModeRoutingAndFailurePaths -count=1 -v
  echo "RUNNER go test resume store live ready-only outbox integration"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestCompleteTailorRunSuccessWritesResultAndOutbox -count=1 -v
} | tee "$OUT/trigger.log"
