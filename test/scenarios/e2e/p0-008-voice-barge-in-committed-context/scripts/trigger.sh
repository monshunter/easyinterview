#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-008-voice-barge-in-committed-context"
mkdir -p "$OUTPUT_DIR"
exec > >(tee "$OUTPUT_DIR/trigger.log") 2>&1

frontend_status=0
backend_status=0

echo "RUNNER frontend-vitest E2E.P0.008"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx \
    src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx
) || frontend_status=$?

echo "RUNNER backend-go-test E2E.P0.008"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/practice ./internal/store/practice ./internal/api/practice \
    -run 'TestBuildCommittedVoiceContext|TestSQLRepositoryLoadCommittedVoiceContextBuildsFromLatestVoiceTurnEvents|TestCreatePracticeVoiceTurnLoadsCommittedContextFromStoredPlaybackEvents|TestVoiceQuestionTemplateInjectsCommittedContextWithoutUnplayedDraft|TestSessionEventServiceRoutesVoicePlaybackEvents|TestSessionEventServiceRejectsMalformedVoicePlaybackEvent|TestAppendSessionEventReturns200ForSupportedKinds|TestAppendSessionEventRejectsIdempotencyKeyHeader' \
    -count=1
) || backend_status=$?

echo "frontend_status=$frontend_status"
echo "backend_status=$backend_status"
if [ "$frontend_status" -ne 0 ] || [ "$backend_status" -ne 0 ]; then
  exit 1
fi
