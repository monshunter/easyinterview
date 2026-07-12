#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-051-practice-assistance-stale-contract-negative"
mkdir -p "$OUT"
(
  cd "$ROOT"
  go test -v ./backend/internal/practice -run '^TestSendPracticeMessageUsesOrdinaryConversationHistory$' -count=1
  python3 scripts/lint/backend_practice_out_of_scope.py --repo-root . --phase all
) | tee "$OUT/trigger.log"
