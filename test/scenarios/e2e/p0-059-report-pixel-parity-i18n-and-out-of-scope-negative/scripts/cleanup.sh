#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative"
LOG="$OUT/trigger.log"

if test -s "$LOG" \
  && grep -Fq 'E2E.P0.059: Playwright pixel parity complete' "$LOG" \
  && ! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|[[:space:]][1-9][0-9]* failed' "$LOG"; then
  rm -rf "$OUT" "$ROOT/frontend/.playwright-output"
else
  rm -f "$OUT/setup.env"
fi
