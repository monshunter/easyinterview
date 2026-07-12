#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
LOG="$ROOT/.test-output/e2e/p0-051-practice-assistance-stale-contract-negative/trigger.log"
grep -Fq -- '--- PASS: TestSendPracticeMessageUsesOrdinaryConversationHistory' "$LOG"
grep -Fq 'backend_practice_out_of_scope all: OK' "$LOG"
if grep -Fq 'no tests to run' "$LOG"; then
  echo 'E2E.P0.051: focused gate matched no tests' >&2
  exit 1
fi
