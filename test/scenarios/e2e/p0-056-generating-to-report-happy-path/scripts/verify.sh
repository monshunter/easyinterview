#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-056-generating-to-report-happy-path"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE" || { echo "E2E.P0.056: no passing test files in trigger log" >&2; exit 1; }
grep -Fq 'preflight.test.ts' "$LOG_FILE" || { echo "E2E.P0.056: preflight test did not run" >&2; exit 1; }
grep -Fq 'GeneratingScreen.test.tsx' "$LOG_FILE" || { echo "E2E.P0.056: GeneratingScreen test did not run" >&2; exit 1; }
grep -Fq 'ReportScreen.test.tsx' "$LOG_FILE" || { echo "E2E.P0.056: ReportScreen test did not run" >&2; exit 1; }
grep -Fq 'DetailSurface.test.tsx' "$LOG_FILE" || { echo "E2E.P0.056: DetailSurface test did not run" >&2; exit 1; }

# Testid coverage in implementation (excluding __tests__).
testid_count=$(grep -RoE 'data-testid="(generating-|report-)' \
  "$REPO_ROOT/frontend/src/app/screens/generating" \
  "$REPO_ROOT/frontend/src/app/screens/report" \
  --include='*.tsx' --include='*.ts' \
  --exclude-dir=__tests__ | wc -l | tr -d ' ')
if [ "$testid_count" -lt 30 ]; then
  echo "E2E.P0.056: expected >=30 generating-/report- testids, got $testid_count" >&2
  exit 1
fi

# Non-current vocabulary guard.
python3 "$REPO_ROOT/scripts/lint/frontend_report_dashboard_non_current.py" --repo-root "$REPO_ROOT" --phase E2E.P0.056

# listTargetJobReports must not appear in implementation code.
if grep -RnE 'listTargetJobReports' "$REPO_ROOT/frontend/src/app/screens/generating" "$REPO_ROOT/frontend/src/app/screens/report" --include='*.tsx' --include='*.ts' --exclude-dir=__tests__; then
  echo "E2E.P0.056: listTargetJobReports leaked into implementation" >&2
  exit 1
fi
