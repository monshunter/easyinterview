#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-045-practice-text-loop-mode-policy-display/trigger.log"; P="$ROOT/frontend/src/app/screens/practice"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.045
grep -Fq -- '--- PASS: TestCreatePracticeVoiceTurnFailsClosed' "$LOG"
grep -Fq 'buildCreatePlanRequest.test.ts' "$LOG"
grep -Fq 'startPractice.test.ts' "$LOG"
grep -Fq 'PracticeScreen.test.tsx' "$LOG"
grep -Fq 'ui-design-contract.test.mjs' "$LOG"
grep -Fq 'expect(body.timeBudgetMinutes).toBe(60)' "$ROOT/frontend/src/app/interview-context/buildCreatePlanRequest.test.ts"
grep -Fq '/ 60:00' "$ROOT/frontend/src/app/screens/practice/PracticeScreen.test.tsx"
grep -Fq '/ --:--' "$ROOT/frontend/src/app/screens/practice/PracticeScreen.test.tsx"
! rg -n '25:00|budget="25:00"' "$P" -g '!*.test.*'
grep -Fq 'disabled' "$P/components/TopBar.tsx"
! rg -n 'SessionMap|QuestionCard|HintBanner|PracticePhoneSurface|currentTurn|questionAssessments' "$P" -g '!*.test.*'
echo 'E2E.P0.045 PASS'
