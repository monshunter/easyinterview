#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff"
mkdir -p "$OUT"

DATABASE_URL="${DATABASE_URL:-$(sed -n 's/^DATABASE_URL=//p' "$ROOT/deploy/dev-stack/.env" | head -n 1)}"
: "${DATABASE_URL:?DATABASE_URL is required}"
export DATABASE_URL

{
  "$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT"
  (
    cd "$ROOT/frontend"
    pnpm exec vitest run \
      src/app/screens/practice/PracticeScreen.test.tsx \
      src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx \
      --reporter=verbose
  ) | tee "$OUT/frontend-disabled-reason.log"
  (
    cd "$ROOT/backend"
    go test ./internal/api/practice ./internal/practice ./internal/store/practice \
      -run '^(TestE2EP0047RejectsZeroAnswerCompletion|TestE2EP0047FreezesReportContext|TestE2EP0047CompletionReplayPreservesReportContext)$' \
      -count=1 -v
  ) | tee "$OUT/completion-owner.log"
  (
    cd "$ROOT/backend"
    go test -tags integration ./internal/store/practice \
      -run '^TestIntegrationE2EP0047RejectsZeroAnswerCompletion$' \
      -count=1 -v
  ) | tee "$OUT/completion-database.log"
  (
    cd "$ROOT/backend"
    go test ./internal/store/practice \
      -run '^TestSQLRepositoryCommitPracticeMessageRejectsClosedSession$' \
      -count=1 -v
  ) | tee "$OUT/completion-regression.log"
} 2>&1 | tee "$OUT/trigger.log"
