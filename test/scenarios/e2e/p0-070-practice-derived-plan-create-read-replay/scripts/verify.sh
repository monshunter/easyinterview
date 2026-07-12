#!/usr/bin/env bash
set -euo pipefail
ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"; LOG="$ROOT/.test-output/e2e/p0-070-practice-derived-plan-create-read-replay/trigger.log"
grep -Fq -- '--- PASS: TestCreateDerivedPracticePlanPassesReportSourceAndCompetencyFocus' "$LOG"
echo 'E2E.P0.070 PASS'
