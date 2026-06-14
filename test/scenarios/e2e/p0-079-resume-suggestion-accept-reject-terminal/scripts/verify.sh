#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-079-resume-suggestion-accept-reject-terminal"
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
  grep -q 'RUNNER go test cmd/api retired suggestion routes' "$LOG"
  grep -q 'TestResumeVersionRoutesAreGonePerD20' "$LOG"
  grep -q 'TestGeneratedRouteCatalogHasNoResumeVersionOperations' "$LOG"
  grep -q 'RUNNER go test handler flat save fixture parity' "$LOG"
  grep -q 'TestUpdateResumeFixtureParity' "$LOG"
  grep -q 'TestDuplicateResumeFixtureParity' "$LOG"
  grep -q 'TestResumeTailorFixtureParity' "$LOG"
  grep -q 'RUNNER frontend vitest rewrites accept-only save flow' "$LOG"
  grep -q 'ResumeRewritesTab.test.tsx' "$LOG"
  grep -q 'ResumeDetailView.test.tsx' "$LOG"
  grep -q 'retired_accept_reject_routes=gone' "$LOG"
  grep -q 'rewrites_accept_only=true' "$LOG"
  grep -q 'save_paths=updateResume_or_duplicateResume' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/handler([[:space:]]|$)' "$LOG"
  cd "$ROOT"
  if rg -n 'inline|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'Private resume body|secret-response|raw resume text|full suggested bullet text|match_summary' "$OUT"; then
    echo "ERROR: private resume or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "terminal: accept/reject route family retired by D-20"
  echo "frontend: rewrites are accept-only and saved through updateResume/duplicateResume"
} | tee "$OUT/verify.log"
