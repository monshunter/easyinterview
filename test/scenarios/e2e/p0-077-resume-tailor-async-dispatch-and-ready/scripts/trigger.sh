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
  echo "RUNNER go test cmd/api resume tailor endpoints"
  cd "$ROOT/backend"
  go test ./cmd/api -run TestResumeTailorEndpointsHTTPScenario -count=1 -v
  echo "RUNNER go test resume handler tailor fixture parity"
  go test ./internal/resume/handler -run TestResumeTailorFixtureParity -count=1 -v
  echo "RUNNER go test resume service request get tailor"
  go test ./internal/resume -run 'TestRequestResumeTailor|TestGetResumeTailorRun' -count=1 -v
  echo "RUNNER go test resume store unit tailor run"
  go test ./internal/resume/store -run 'Test(CreateTailorRunInsertsAsyncJobWithResumePayload|CreateTailorRunWithoutTargetJobSkipsTargetCheck|CreateTailorRunResumeNotFound|GetTailorRunMapsStatusFromAsyncJob|GetTailorRunCrossUserNotFound|CompleteTailorRunSuccessWritesResultAndOutbox)' -count=1 -v
  echo "RUNNER go test cmd/api resume tailor runner kernel ready"
  go test ./cmd/api -run TestResumeTailorRunnerHTTPScenario -count=1 -v
  echo "RUNNER go test resume jobs tailor ready"
  go test ./internal/resume/jobs -run TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox -count=1 -v
} | tee "$OUT/trigger.log"
