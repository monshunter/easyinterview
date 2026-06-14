#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-035-resume-parse-async-job-lifecycle"
mkdir -p "$OUT"

{
  echo "E2E.P0.035 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeParseDrainerHTTPScenario|TestResumeParseDrainerRetryableFailureScenario' -count=1 -v
  go test ./cmd/api -run TestBuildResumeRuntimeWiresRoutesDrainerAndDeterministicAI -count=1 -v
  go test ./internal/resume/jobs -run 'TestParseHandler(UsesTwoSourceInputsAndWritesReadyOutbox|FailurePathsMarkFailedAndSkipCompletedOutbox|RetriesFailedAssetBackToProcessing|ObservedAIWritesResumeTaskRunColumns|PIIRedlineForLogsAuditTaskRunsAndOutbox)' -count=1 -v
  go test ./internal/resume/store -run 'Test(CompleteParseSuccessWritesReadyStateProfileAndCompletedOutboxAtomically|CompleteParseFailureMarksFailedWithoutCompletedOutbox|ParseStatusTransition)' -count=1 -v
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestResumesIntegrationCRUDStateIsolationPaginationAndRollback -count=1 -v
} | tee "$OUT/trigger.log"
