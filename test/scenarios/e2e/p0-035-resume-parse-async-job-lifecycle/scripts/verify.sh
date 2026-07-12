#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-035-resume-parse-async-job-lifecycle"
mkdir -p "$OUT"

{
  echo "E2E.P0.035 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$OUT/trigger.log"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$OUT/trigger.log"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'TestResumeParseRunnerHTTPScenario' "$OUT/trigger.log"
  grep -q 'TestResumeParseRunnerRetryableFailureScenario' "$OUT/trigger.log"
  grep -q 'TestBuildResumeRuntimeWiresRoutesRunnerAndDeterministicAI' "$OUT/trigger.log"
  grep -q 'TestCatalogKeepsResumeParseOutputBudget' "$OUT/trigger.log"
  grep -q 'TestParseHandlerRejectsDOCXUploadText' "$OUT/trigger.log"
  grep -q 'TestParseHandlerRejectsUnreadablePDFText' "$OUT/trigger.log"
  grep -q 'TestParseHandlerExtractsReadableUploadText' "$OUT/trigger.log"
  grep -q 'TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox' "$OUT/trigger.log"
  grep -q 'TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox' "$OUT/trigger.log"
  grep -q 'TestParseHandlerRetriesFailedAssetBackToProcessing' "$OUT/trigger.log"
  grep -q 'TestParseHandlerObservedAIWritesResumeTaskRunColumns' "$OUT/trigger.log"
  grep -q 'TestParseHandlerPIIRedlineForLogsAuditTaskRunsAndOutbox' "$OUT/trigger.log"
  grep -q 'TestParseHandlerPreservesLongInputTailWithStructuredOnlyResponse' "$OUT/trigger.log"
  grep -q 'TestParseHandlerRejectsLengthFinishReasonAndPreservesSourceSnapshot' "$OUT/trigger.log"
  grep -q 'TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady' "$OUT/trigger.log"
  grep -q 'TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically' "$OUT/trigger.log"
  grep -q 'TestCompleteParseFailureCanPersistExtractedTextSnapshot' "$OUT/trigger.log"
  grep -q 'TestCompleteParseFailureMarksFailedWithoutCompletedOutbox' "$OUT/trigger.log"
  grep -q 'TestResumesIntegrationCRUDStateIsolationPaginationAndRollback' "$OUT/trigger.log"
  cd "$ROOT/backend"
  go test ./cmd/api -run TestResumeParseRunnerHTTPScenario -count=1
  go test ./internal/resume/jobs -run 'TestParseHandlerPIIRedlineForLogsAuditTaskRunsAndOutbox' -count=1
  cd "$ROOT"
  if rg -n -i '(tailor|mode).*(inline|mirror)|(inline|mirror).*(tailor|mode)' backend/internal/resume backend/cmd/api/resume_parse_runner_scenario_test.go --glob '!**/verify.sh'; then
    exit 1
  fi
  if rg -n 'failed_retryable' backend/internal/resume backend/cmd/api/resume_parse_runner_scenario_test.go; then
    exit 1
  fi
  if rg -n 'Private resume body|Ada Lovelace|secret-response' "$OUT"; then
    echo "ERROR: private resume content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-runner"
  echo "parse status: queued -> processing -> ready or failed"
  echo "observability: AI task run typed columns covered"
  echo "resume.parse.default: long-resume output budget >= 8192 covered"
  echo "resume.parse input: full long-resume tail marker preserved in prompt and deterministic snapshot"
  echo "resume.parse output: structured-only response covered; finish_reason=length fails closed"
  echo "outbox: ready-only resume.parse.completed payload covered"
  echo "resumes.structured_profile: ready-state persistence covered by integration gate"
  echo "resumes.display_name: queued null + LLM-derived ready-state name + failed-with-snapshot fallback covered by parse/store/runner kernel gates"
  echo "upload parsed_text_snapshot: deterministic full PDF/Markdown/text extraction covered; DOCX and unreadable PDF fallback rejected before AI"
} | tee "$OUT/verify.log"
