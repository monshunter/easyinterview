#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; OUT="$ROOT/.test-output/e2e/p0-024-practice-session-ai-failure-retry"; mkdir -p "$OUT"
{ cd "$ROOT"; go test -v ./backend/internal/practice -run '^TestStartPracticeSessionAIErrorFailsReservationWithoutOpeningMessage$' -count=1; } | tee "$OUT/trigger.log"
