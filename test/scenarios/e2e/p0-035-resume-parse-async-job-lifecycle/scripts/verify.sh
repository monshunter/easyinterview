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
  if grep -E -- '--- FAIL:|^FAIL([[:space:]]|$)' "$OUT/trigger.log"; then
    echo "ERROR: failed test evidence detected"
    exit 1
  fi
  required_tests=(
    TestResumeParseRunnerHTTPScenario
    TestResumeParseRunnerRetryableFailureScenario
    TestBuildResumeRuntimeWiresRoutesRunnerAndDeterministicAI
    TestCatalogKeepsResumeParseOutputBudget
    TestParseHandlerRejectsDOCXUploadText
    TestParseHandlerRejectsUnreadablePDFText
    TestParseHandlerExtractsReadableUploadText
    TestParseHandlerUsesTwoSourceInputsAndWritesReadyOutbox
    TestParseHandlerFailurePathsMarkFailedAndSkipCompletedOutbox
    TestParseHandlerRetriesFailedAssetBackToProcessing
    TestParseHandlerObservedAIWritesResumeTaskRunColumns
    TestParseHandlerPIIRedlineForLogsAuditTaskRunsAndOutbox
    TestParseHandlerPreservesInlineHeadingWordsInSourceSnapshot
    TestParseHandlerPreservesLongInputTailWithStructuredOnlyResponse
    TestParseHandlerRejectsLengthFinishReasonAndPreservesSourceSnapshot
    TestParseHandlerUsesConfiguredExtractedTextByteLimitBeforeAI
    TestCreateWithParseJobKeepsDisplayNameUnsetUntilParseReady
    TestCompleteParseSuccessWritesReadyStateProfileDisplayNameAndCompletedOutboxAtomically
    TestCompleteParseFailureCanPersistExtractedTextSnapshot
    TestCompleteParseFailureMarksFailedWithoutCompletedOutbox
    TestResumesIntegrationCRUDStateIsolationPaginationAndRollback
  )
  for test_name in "${required_tests[@]}"; do
    if ! grep -Fq -- "--- PASS: $test_name" "$OUT/trigger.log"; then
      echo "ERROR: missing PASS evidence for $test_name"
      exit 1
    fi
  done
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
  echo "resume.parse boundary: configured extracted-text exact/limit+1 inputs are constructed in memory and overflow is rejected before AI"
} | tee "$OUT/verify.log"
