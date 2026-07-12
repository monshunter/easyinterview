#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-058-report-failure-and-missing-session"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  echo "E2E.P0.058: validating focused failure contracts"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/report/__tests__/preflight.test.ts \
    src/app/screens/report/__tests__/ReportFailureState.test.tsx \
    src/app/screens/report/__tests__/ReportMissingSessionState.test.tsx \
    src/app/screens/report/__tests__/useFeedbackReport.test.tsx \
    src/app/screens/report/__tests__/ConversationReport.test.tsx \
    src/app/screens/generating/__tests__/useReportGenerationPoll.test.tsx
  go test -v ./backend/internal/review -run 'TestGenerateReportRejectsInvalidDimensionScoresBeforePersistence' -count=1
) | tee "$OUTPUT_DIR/trigger.log"
