#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-009-voice-provider-failure-fallback"
mkdir -p "$OUTPUT_DIR"
exec > >(tee "$OUTPUT_DIR/trigger.log") 2>&1

frontend_status=0
backend_status=0
a3_status=0

echo "RUNNER frontend-vitest E2E.P0.009"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx \
    src/api/devMockClient.test.ts
) || frontend_status=$?

echo "RUNNER backend-practice-go-test E2E.P0.009"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/practice \
    -run 'TestCreatePracticeVoiceTurnStopsWhenSTTFails|TestCreatePracticeVoiceTurnStopsWhenChatFailsBeforeTTS|TestCreatePracticeVoiceTurnSecondLanguageMismatchSkipsTTSAndPersistence|TestCreatePracticeVoiceTurnReturnsTranscriptAndAssistantTextWhenTTSFails|TestCreatePracticeVoiceTurnPersistsBusinessTextOutsideAIMetadata' \
    -count=1
) || backend_status=$?

echo "RUNNER a3-go-test E2E.P0.009"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/ai/aiclient ./internal/ai/aiclient/profile \
    -run 'TestTranscribe_RealtimeProfileFailsClosed|TestSynthesize_UnsupportedCapabilityFailsClosedWithSharedError|TestSynthesize_DisabledProfileFailsClosedWithSharedError|TestTrackedCatalogCoversF3AndProductUICapabilityProfiles' \
    -count=1
) || a3_status=$?

echo "frontend_status=$frontend_status"
echo "backend_status=$backend_status"
echo "a3_status=$a3_status"
if [ "$frontend_status" -ne 0 ] || [ "$backend_status" -ne 0 ] || [ "$a3_status" -ne 0 ]; then
  exit 1
fi
