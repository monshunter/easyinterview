#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; OUT="$ROOT/.test-output/e2e/p0-045-practice-text-loop-mode-policy-display"; mkdir -p "$OUT"
{
  cd "$ROOT"
  "$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT"
  node --test ui-design/ui-design-contract.test.mjs
  pnpm --filter @easyinterview/frontend test src/app/screens/practice/PracticeScreen.test.tsx src/app/App.test.tsx src/app/routeUrl.test.ts
  go test -v ./backend/internal/api/practice -run '^TestCreatePracticeVoiceTurnFailsClosed$' -count=1
} | tee "$OUT/trigger.log"
