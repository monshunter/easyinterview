#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
LOG="$ROOT/.test-output/e2e/p0-007-practice-voice-disabled-fail-closed/trigger.log"
test -s "$LOG"
grep -Fq -- '--- PASS: TestCreatePracticeVoiceTurnFailsClosed' "$LOG"
grep -Fq 'PracticeScreen.test.tsx' "$LOG"
grep -Fq 'disabled' "$ROOT/frontend/src/app/screens/practice/components/TopBar.tsx"
! rg -n 'PracticePhoneSurface|usePracticeVoiceTurn|phoneVad' "$ROOT/frontend/src/app/screens/practice" -g '!*.test.*'
echo 'E2E.P0.007 PASS'
