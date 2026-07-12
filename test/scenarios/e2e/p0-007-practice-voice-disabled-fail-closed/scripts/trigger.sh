#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-007-practice-voice-disabled-fail-closed"
mkdir -p "$OUT"
{
  echo "E2E.P0.007 voice-disabled contract"
  cd "$ROOT"
  pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx
  go test -v ./backend/internal/api/practice -run '^TestCreatePracticeVoiceTurnFailsClosed$' -count=1
} | tee "$OUT/trigger.log"
