#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-063-debrief-suggest-questions"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.063 RUNNER go test"
  cd "$REPO_ROOT/backend"
  go test -v ./internal/store/debrief -run 'TestStoreGetSuggestionContext_TargetJobScoped|TestStoreGetSuggestionContext_LoadsPracticeSessionSummary|TestStoreGetSuggestionContext_CrossUserSessionNotFound|TestStoreGetSuggestionContext_LoadsResumeStructuredProfile|TestStoreGetSuggestionContext_CrossUserResumeNotFound' -count=1
  go test -v ./internal/debrief -run 'TestServiceSuggestQuestions_Happy|TestServiceSuggestQuestions_SessionContextInPrompt|TestServiceSuggestQuestions_ResumeContextInPrompt|TestServiceSuggestQuestions_CrossUserTargetJob_403|TestServiceSuggestQuestions_F3ResolveFailed|TestServiceSuggestQuestions_A3Timeout|TestServiceSuggestQuestions_ParseFailed|TestAITaskRunsWritten|TestAuditEventsWritten|TestAuditEvents_NoRawText' -count=1
  go test -v ./internal/api/debriefs -run 'TestSuggestDebriefQuestions_MapsSessionIDToService|TestSuggestDebriefQuestions_MapsResumeIDToService|TestSuggestDebriefQuestions_CountBoundary|TestSuggestDebriefQuestions_Unauthenticated_401' -count=1
  go test -v ./cmd/api -run 'TestBuildAPIHandlerMountsDebriefSuggestQuestionsBehindSessionMiddleware|TestBuildDebriefRoutesWiresHandlerAndIdempotency' -count=1
  cd "$REPO_ROOT"
  echo "E2E.P0.063 FIXTURE validate-fixtures"
  make validate-fixtures
  echo "E2E.P0.063 FIXTURE parity PASS"
  echo "E2E.P0.063 sessionId backend contract PASS"
  echo "E2E.P0.063 resumeId backend contract PASS"
} | tee "$OUTPUT_DIR/trigger.log"
