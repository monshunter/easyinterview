#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; OUT="$ROOT/.test-output/e2e/p0-072-practice-derived-source-isolation-privacy"; mkdir -p "$OUT"
{ cd "$ROOT"; go test -v ./backend/internal/practice -run '^TestDerivedPracticePlanRequiresSourceReport$' -count=1; } | tee "$OUT/trigger.log"
