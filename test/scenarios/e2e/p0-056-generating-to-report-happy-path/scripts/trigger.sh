#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-056-generating-to-report-happy-path"
OWNER_EVIDENCE="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/completion-backend-evidence.json"
mkdir -p "$OUT"

if [[ -z "${DATABASE_URL:-}" && -f "$ROOT/deploy/dev-stack/.env" ]]; then
  DATABASE_URL="$(sed -n 's/^DATABASE_URL=//p' "$ROOT/deploy/dev-stack/.env" | head -n 1)"
fi
: "${DATABASE_URL:?DATABASE_URL is required}"
export DATABASE_URL
export PRACTICE_COMPLETION_EVIDENCE_PATH="$OWNER_EVIDENCE"

{
  "$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT"
  (
    cd "$ROOT/frontend"
    pnpm exec vitest run \
      src/app/screens/report/__tests__/preflight.test.ts \
      src/app/screens/generating/__tests__/useReportGenerationPoll.test.tsx \
      src/app/screens/generating/__tests__/GeneratingScreen.test.tsx \
      src/app/screens/report/__tests__/ConversationReport.test.tsx \
      --reporter=verbose
  ) | tee "$OUT/frontend.log"
  echo "E2E.P0.056: running exact backend evidence command"
  (
    cd "$ROOT/backend"
    go test ./internal/review ./internal/store/review ./internal/api/reports -run '^TestE2EP0056ReportBackendEvidence$' -count=1 -v
  ) | tee "$OUT/backend.log"
} 2>&1 | tee "$OUT/trigger.log"
