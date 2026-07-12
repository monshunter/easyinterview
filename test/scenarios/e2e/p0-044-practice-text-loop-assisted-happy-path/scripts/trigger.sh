#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; OUT="$ROOT/.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path"; mkdir -p "$OUT"
{
  cd "$ROOT"
  "$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT"
  pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx
  go test -v ./backend/internal/api/practice ./backend/internal/practice ./backend/internal/store/practice -run 'TestSendPracticeMessageReturnsConversationMessages|TestSendPracticeMessageUsesOrdinaryConversationHistory|TestSQLRepositoryReservePracticeMessageRetriesPendingUserMessage' -count=1
} | tee "$OUT/trigger.log"
