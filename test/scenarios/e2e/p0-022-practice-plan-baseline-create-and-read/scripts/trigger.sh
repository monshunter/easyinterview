#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-022-practice-plan-baseline-create-and-read"
mkdir -p "$OUT"
{
  echo 'E2E.P0.022 real practice handler and repository gate'
  cd "$ROOT"
  go test -v ./backend/internal/api/practice ./backend/internal/practice ./backend/internal/store/practice -run 'TestCreatePracticePlanMapsOnlyCurrentFields|TestCreatePracticePlanPassesOnlyConversationPlanFields|TestSQLRepositoryCreatePlanUsesConversationColumns' -count=1
} | tee "$OUT/trigger.log"
