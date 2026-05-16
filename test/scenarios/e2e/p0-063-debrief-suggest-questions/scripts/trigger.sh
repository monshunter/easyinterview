#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-063-debrief-suggest-questions"
mkdir -p "$OUTPUT_DIR"
{
  echo "E2E.P0.063 RUNNER go test"
  cd "$REPO_ROOT/backend"
  go test -v ./internal/debrief -run 'TestServiceSuggestQuestions_Happy|TestServiceSuggestQuestions_CrossUserTargetJob_403|TestServiceSuggestQuestions_F3ResolveFailed|TestServiceSuggestQuestions_A3Timeout|TestServiceSuggestQuestions_ParseFailed|TestAITaskRunsWritten|TestAuditEventsWritten|TestAuditEvents_NoRawText' -count=1
  go test -v ./internal/api/debriefs -run 'TestSuggestDebriefQuestions_CountBoundary|TestSuggestDebriefQuestions_Unauthenticated_401' -count=1
} | tee "$OUTPUT_DIR/trigger.log"
