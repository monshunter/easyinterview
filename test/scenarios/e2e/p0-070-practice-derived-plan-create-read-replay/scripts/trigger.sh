#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; OUT="$ROOT/.test-output/e2e/p0-070-practice-derived-plan-create-read-replay"; mkdir -p "$OUT"
{
  cd "$ROOT"
  go test -v ./backend/internal/practice -run '^TestCreateDerivedPracticePlanPassesReportSourceAndCompetencyFocus$' -count=1
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" \
    go test -v -tags=integration ./backend/internal/store/practice -run '^TestSQLRepositoryIntegration_CreatePlanProjectsCanonicalRoundLedger$' -count=1
} | tee "$OUT/trigger.log"
