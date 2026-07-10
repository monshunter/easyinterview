#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-079-resume-rewrites-accept-only-save"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.079 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER make validate-fixtures D-20 flat resume fixtures' "$LOG"
  grep -q 'validate-fixtures: OK' "$LOG"
  grep -q 'RUNNER go test cmd/api out-of-scope suggestion routes' "$LOG"
  grep -q 'TestResumeVersionRoutesRemainUnmountedPerD20' "$LOG"
  grep -q 'TestGeneratedRouteCatalogHasNoResumeVersionOperations' "$LOG"
  grep -q 'RUNNER go test handler flat save fixture parity' "$LOG"
  grep -q 'TestUpdateResumeFixtureParity' "$LOG"
  grep -q 'TestDuplicateResumeFixtureParity' "$LOG"
  grep -q 'TestResumeTailorFixtureParity' "$LOG"
  grep -q 'RUNNER frontend vitest read-only detail negative flow' "$LOG"
  grep -q 'ResumeDetailView.test.tsx' "$LOG"
  grep -q 'out_of_scope_accept_reject_routes=gone' "$LOG"
  grep -q 'detail_rewrites_edit_surface=gone' "$LOG"
  grep -q 'backend_flat_save_fixtures=updateResume_or_duplicateResume' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  cd "$ROOT"
  if rg -n -i '(tailor|mode).*(inline|rewrite|mirror)|(inline|rewrite|mirror).*(tailor|mode)' backend/internal/resume --glob '!**/*_test.go' --glob '!**/verify.sh'; then
    echo "ERROR: out-of-scope inline/rewrite/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/*_test.go' --glob '!**/verify.sh'; then
    echo "ERROR: out-of-scope mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|raw resume text|full suggested bullet text|match_summary' "$OUT"; then
    echo "ERROR: private resume or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "terminal: accept/reject route family out-of-scope by D-20"
  echo "frontend: detail rewrites/edit surface is absent"
} | tee "$OUT/verify.log"
