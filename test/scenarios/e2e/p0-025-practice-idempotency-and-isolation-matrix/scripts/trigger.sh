#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-025-practice-idempotency-and-isolation-matrix"

mkdir -p "$OUTPUT_DIR"

(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/practice -run 'TestSendPracticeMessage(ExactReplayReturnsOriginalResultWithoutAICall|MapsClientMismatchAndCrossUserAccess|PendingSameIDDoesNotCallAI)' -count=1
  go test -v ./internal/store/practice -run 'TestSQLRepositoryReservePracticeMessage(RetriesOnlyRetryableFailure|RejectsPendingAndTerminalSameID|RejectsNewMessageWhileReplyPending)' -count=1
  go test -v ./internal/api/practice -run 'TestSendPracticeMessageMapsConflictAndIsolationErrors' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
