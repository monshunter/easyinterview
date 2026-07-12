#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-045-practice-text-loop-mode-policy-display/trigger.log"; P="$ROOT/frontend/src/app/screens/practice"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.045
grep -Fq -- '--- PASS: TestCreatePracticeVoiceTurnFailsClosed' "$LOG"
grep -Fq 'disabled' "$P/components/TopBar.tsx"
! rg -n 'SessionMap|QuestionCard|HintBanner|PracticePhoneSurface|currentTurn|questionAssessments' "$P" -g '!*.test.*'
echo 'E2E.P0.045 PASS'
