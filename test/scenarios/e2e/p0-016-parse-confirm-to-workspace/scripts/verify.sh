#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-016-parse-confirm-to-workspace"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" \
  "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}" "targetJob.realApiMode.test.ts"

for frontend_file in \
  targetJob.realApiMode.test.ts \
  ParseReports.test.tsx \
  ParseScreen.test.tsx \
  ParseFlow.test.tsx \
  ParseEdit.test.tsx \
  ParseAuthGate.test.tsx \
  ParseResumeBinding.test.tsx \
  ParseRoundStates.test.tsx \
  MockInterviewCard.test.tsx \
  HomeRecentMocks.test.tsx \
  interviewContext.test.ts \
  routeUrl.test.ts \
  TopBar.test.tsx; do
  grep -Fq "$frontend_file" "$LOG_FILE" || { echo "E2E.P0.016: $frontend_file did not run" >&2; exit 1; }
done

for title in \
  'workspace detail exposes only direct start with bound resume context' \
  'workspace detail round states match the UI truth at desktop and mobile' \
  'workspace plan-detail reports entry matches the UI truth and stays report-list-free' \
  'workspace start interview hands off directly to practice with bound resume'; do
  grep -Fq "$title" "$LOG_FILE" || { echo "E2E.P0.016: Playwright title missing: $title" >&2; exit 1; }
done

grep -Fq '8 passed' "$LOG_FILE"
grep -Fq 'reportListRequestsBeforeClick=0' "$LOG_FILE"
grep -Fq 'topbarReportsEntry=0' "$LOG_FILE"
grep -Fq 'embeddedReports=0' "$LOG_FILE"
grep -Fq 'sectionReportsAccepted=false' "$LOG_FILE"
grep -Fq 'changedRatio=' "$LOG_FILE"
grep -Fq 'sequence=done,current,pending distinctBackgrounds=3 distinctBorders=3' "$LOG_FILE"
grep -Fq 'route=practice noUpdateTargetJob=true' "$LOG_FILE"

if grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]|[[:space:]][1-9][0-9]* failed' "$LOG_FILE"; then
  echo "E2E.P0.016: failing or empty runner evidence found" >&2
  exit 1
fi
