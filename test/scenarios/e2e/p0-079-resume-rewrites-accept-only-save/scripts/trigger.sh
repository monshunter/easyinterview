#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-079-resume-rewrites-accept-only-save"
mkdir -p "$OUT"

{
  echo "E2E.P0.079 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT"
  echo "RUNNER make validate-fixtures D-20 flat resume fixtures"
  make validate-fixtures
  cd "$ROOT/backend"
  echo "RUNNER go test cmd/api non-current suggestion routes"
  go test ./cmd/api -run 'TestResumeVersionRoutesRemainUnmountedPerD20|TestGeneratedRouteCatalogHasNoResumeVersionOperations' -count=1 -v
  echo "RUNNER go test handler flat save fixture parity"
  go test ./internal/resume/handler -run 'Test(UpdateResumeFixtureParity|DuplicateResumeFixtureParity|ResumeTailorFixtureParity)' -count=1 -v
  cd "$ROOT"
  echo "RUNNER frontend vitest rewrites accept-only save flow"
  pnpm --filter @easyinterview/frontend exec vitest run --reporter=verbose \
    src/app/screens/resume-workshop/tabs/ResumeRewritesTab.test.tsx \
    src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx
  echo "evidence non_current_accept_reject_routes=gone"
  echo "evidence rewrites_accept_only=true"
  echo "evidence save_paths=updateResume_or_duplicateResume"
} | tee "$OUT/trigger.log"
