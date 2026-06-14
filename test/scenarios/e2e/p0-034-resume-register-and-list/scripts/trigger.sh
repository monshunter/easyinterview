#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-034-resume-register-and-list"
mkdir -p "$OUT"

{
  echo "E2E.P0.034 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT/backend"
  go test ./cmd/api -run 'TestResumeRegisterListHTTPScenario|TestResumeRegisterListHTTPValidationScenario' -count=1 -v
  go test ./internal/resume/handler -run 'Test(RegisterResumeFixtureParity|RegisterResumeValidationErrorsReturnUnprocessableEntity|GetResumeFixtureParity|ListResumesFixtureParity|RegisterSourceType|RegisterIdempotency|GetResumeNotFoundAndCrossUserReturns404|ListResumesPassesPaginationAndUserScope|ListResumesInvalidCursorReturnsUnprocessableEntity)' -count=1 -v
  go test ./internal/resume/store -run 'Test(CreateWithParseJobInsertsResumeAndJobAtomically|CreateWithParseJobRollsBackWhenJobInsertFails|ParseStatusTransition|ListCursorPagination|CompleteParseSuccessWritesReadyStateProfileAndCompletedOutboxAtomically|CompleteParseFailureMarksFailedWithoutCompletedOutbox)' -count=1 -v
  go test ./internal/upload/service -run 'TestRegisterFileObject(MarksPendingUploadedAfterObjectExists|RejectsMissingObjectAndIllegalStates)' -count=1 -v
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestResumesIntegrationCRUDStateIsolationPaginationAndRollback -count=1 -v
} | tee "$OUT/trigger.log"
