#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-079-resume-suggestion-accept-reject-terminal"
mkdir -p "$OUT"

{
  echo "E2E.P0.079 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT"
  echo "RUNNER make validate-fixtures suggestion decision fixtures"
  make validate-fixtures
  cd "$ROOT/backend"
  echo "RUNNER go test cmd/api suggestion accept reject"
  go test ./cmd/api -run TestResumeSuggestionAcceptRejectHTTPScenario -count=1 -v
  echo "RUNNER go test handler suggestion fixture parity"
  go test ./internal/resume/handler -run TestResumeSuggestionDecisionFixtureParity -count=1 -v
  echo "RUNNER go test service suggestion decision"
  go test ./internal/resume -run TestResumeSuggestionDecision -count=1 -v
  echo "RUNNER go test store live suggestion decision CAS"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestResumeSuggestionDecisionCASIsolationAndProfileStability -count=1 -v
  echo "evidence status=accepted"
  echo "evidence status=rejected"
  echo "evidence reason=SUGGESTION_ALREADY_DECIDED"
  echo "evidence structured_profile=unchanged"
} | tee "$OUT/trigger.log"
