#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" \
  "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}" "clientFactory.test.ts"

for frontend_file in \
  targetJobReports.test.ts \
  ReportsScreen.test.tsx \
  reportContract.test.ts \
  ConversationReport.test.tsx \
  GeneratingBackNavigation.test.tsx \
  GeneratingScreen.test.tsx \
  preflight.test.ts \
  reportDashboardI18nCoverage.test.ts \
  outOfScopeNegative.test.ts; do
  grep -Fq "$frontend_file" "$LOG_FILE" || { echo "E2E.P0.059: $frontend_file did not run" >&2; exit 1; }
done

grep -Fq 'frontend-report-dashboard out-of-scope lint OK' "$LOG_FILE"
grep -Fq 'only consumer=frontend/src/app/screens/reports/ReportsScreen.tsx' "$LOG_FILE"
grep -Fq '5 passed' "$LOG_FILE"

for browser_spec in reports.spec.ts report.spec.ts generating.spec.ts; do
  grep -Fq "$browser_spec" "$LOG_FILE" || { echo "E2E.P0.059: $browser_spec did not run" >&2; exit 1; }
done
grep -Fq 'current-plan reports ready state matches the UI truth' "$LOG_FILE"
grep -Fq 'reports loading empty error latest-ready and mismatch states match the UI truth' "$LOG_FILE"
grep -Fq 'currentPlanIsolation=true' "$LOG_FILE"
grep -Fq 'currentLatestOnly=true' "$LOG_FILE"
grep -Fq 'backTarget=workspace-detail' "$LOG_FILE"
grep -Fq 'changedRatio=' "$LOG_FILE"
grep -Fq 'E2E.P0.059: Playwright pixel parity complete' "$LOG_FILE"

awk '
  /E2E\.P0\.059: running ReportsScreen, Report, and Generating Playwright pixel parity/ { in_playwright = 1 }
  in_playwright && /^[[:space:]]*[0-9]+ passed/ { passed = 1 }
  END { exit passed ? 0 : 1 }
' "$LOG_FILE" || { echo "E2E.P0.059: Playwright pass marker missing" >&2; exit 1; }

if grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|[[:space:]][1-9][0-9]* failed' "$LOG_FILE"; then
  echo "E2E.P0.059: failing or empty runner evidence found" >&2
  exit 1
fi
