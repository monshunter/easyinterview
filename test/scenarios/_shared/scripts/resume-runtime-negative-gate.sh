#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
cd "$ROOT"

run_negative_scan() {
  local label="$1"
  local pattern="$2"
  local matches
  local status

  if matches="$(rg -n -i "$pattern" backend/internal/resume --glob '!**/*_test.go' --glob '!**/verify.sh' 2>&1)"; then
    printf '%s\n' "$matches"
    printf 'ERROR: %s found\n' "$label" >&2
    exit 1
  else
    status=$?
    if ((status != 1)); then
      printf '%s\n' "$matches" >&2
      printf 'ERROR: %s scan failed\n' "$label" >&2
      exit "$status"
    fi
  fi
}

run_negative_scan \
  "out-of-scope inline/rewrite/mirror Resume mode vocabulary" \
  '(tailor|mode).*(inline|rewrite|mirror)|(inline|rewrite|mirror).*(tailor|mode)'
run_negative_scan \
  "out-of-scope mistakes/growth/drill Resume module vocabulary" \
  'mistakes|growth|drill|inline-debrief-record'
