#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-007-cascaded-voice-turn"
mkdir -p "$OUTPUT_DIR"
exec > >(tee "$OUTPUT_DIR/trigger.log") 2>&1

frontend_status=0
backend_status=0

echo "RUNNER frontend-vitest E2E.P0.007"
(
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx \
    src/app/screens/practice/PracticeScreen.test.tsx \
    src/app/screens/practice/__tests__/practiceModeSwitch.test.tsx
) || frontend_status=$?

echo "RUNNER backend-go-test E2E.P0.007"
(
  cd "$REPO_ROOT/backend"
  go test -v ./internal/practice ./internal/store/practice ./internal/api/practice ./cmd/api \
    -run 'TestCreatePracticeVoiceTurn|TestSQLRepositoryRecordPracticeVoiceTurn|TestE2EP0007PracticeVoiceTurnHTTPRoute' \
    -count=1
) || backend_status=$?

echo "frontend_status=$frontend_status"
echo "backend_status=$backend_status"
if [ "$frontend_status" -ne 0 ] || [ "$backend_status" -ne 0 ]; then
  exit 1
fi
