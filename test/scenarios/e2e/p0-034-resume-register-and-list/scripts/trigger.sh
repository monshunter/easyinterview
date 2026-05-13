#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-034-resume-register-and-list"
mkdir -p "$OUT"

{
  echo "E2E.P0.034 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT/backend"
  go test ./cmd/api -run TestResumeRegisterListHTTPScenario -count=1 -v
  go test ./internal/resume/handler -run 'Test(RegisterResumeFixtureParity|GetResumeFixtureParity|ListResumesFixtureParity|RegisterSourceType|RegisterIdempotency|GetResumeNotFoundAndCrossUserReturns404|ListResumesPassesPaginationAndUserScope)' -count=1 -v
  go test ./internal/resume/store -run 'Test(CreateWithParseJobInsertsAssetAndJobAtomically|CreateWithParseJobRollsBackWhenJobInsertFails|ParseStatusTransition|ListCursorPagination|CompleteParseSuccessWritesReadyStateAndCompletedOutboxAtomically|CompleteParseFailureMarksFailedWithoutCompletedOutbox)' -count=1 -v
  go test ./internal/upload/service -run 'TestRegisterFileObject(MarksPendingUploadedAfterObjectExists|RejectsMissingObjectAndIllegalStates)' -count=1 -v
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestResumeAssetsIntegrationCRUDStateIsolationPaginationAndRollback -count=1 -v
} | tee "$OUT/trigger.log"
