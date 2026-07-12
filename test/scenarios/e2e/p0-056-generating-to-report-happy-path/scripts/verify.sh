#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-056-generating-to-report-happy-path"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'preflight.test.ts' "$LOG_FILE" || { echo "E2E.P0.056: preflight test did not run" >&2; exit 1; }
grep -Fq 'useReportGenerationPoll.test.tsx' "$LOG_FILE" || { echo "E2E.P0.056: poll hook test did not run" >&2; exit 1; }
grep -Fq 'GeneratingScreen.test.tsx' "$LOG_FILE" || { echo "E2E.P0.056: GeneratingScreen test did not run" >&2; exit 1; }
grep -Fq 'ConversationReport.test.tsx' "$LOG_FILE" || { echo "E2E.P0.056: conversation report test did not run" >&2; exit 1; }
grep -Fq -- '--- PASS: TestReadinessFromContentUsesCandidateScoreBoundaries' "$LOG_FILE" || { echo "E2E.P0.056: readiness boundary test did not run" >&2; exit 1; }
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG_FILE"

# Testid coverage in implementation (excluding __tests__).
testid_count=$(grep -RoE 'data-testid="(generating-|report-)' \
  "$REPO_ROOT/frontend/src/app/screens/generating" \
  "$REPO_ROOT/frontend/src/app/screens/report" \
  --include='*.tsx' --include='*.ts' \
  --exclude-dir=__tests__ | wc -l | tr -d ' ')
if [ "$testid_count" -lt 15 ]; then
  echo "E2E.P0.056: expected >=15 generating-/report- testids, got $testid_count" >&2
  exit 1
fi

# Out-of-scope vocabulary guard.
python3 "$REPO_ROOT/scripts/lint/frontend_report_dashboard_out_of_scope.py" --repo-root "$REPO_ROOT" --phase E2E.P0.056

# listTargetJobReports must not appear in implementation code.
if grep -RnE 'listTargetJobReports' "$REPO_ROOT/frontend/src/app/screens/generating" "$REPO_ROOT/frontend/src/app/screens/report" --include='*.tsx' --include='*.ts' --exclude-dir=__tests__; then
  echo "E2E.P0.056: listTargetJobReports leaked into implementation" >&2
  exit 1
fi
