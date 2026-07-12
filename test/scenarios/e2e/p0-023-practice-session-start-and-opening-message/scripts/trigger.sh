#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-023-practice-session-start-and-opening-message"
mkdir -p "$OUT"
{ cd "$ROOT"; go test -v ./backend/internal/practice ./backend/internal/store/practice -run 'TestStartPracticeSessionCreatesOpeningAssistantMessage|TestSQLRepositoryGetSessionReturnsOrderedMessages' -count=1; } | tee "$OUT/trigger.log"
