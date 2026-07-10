#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-016-parse-confirm-to-workspace"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}" "targetJob.realApiMode.test.ts"
grep -Fq "ParseScreen.test.tsx" "$LOG_FILE"
grep -Fq "ParseEdit.test.tsx" "$LOG_FILE"
grep -Fq "ParseAuthGate.test.tsx" "$LOG_FILE"
grep -Fq "ParseResumeBinding.test.tsx" "$LOG_FILE"
grep -Fq "MockInterviewCard.test.tsx" "$LOG_FILE"
grep -Fq "HomeRecentMocks.test.tsx" "$LOG_FILE"
grep -Fq "navigation/interviewContext.test.ts" "$LOG_FILE"
grep -Fq "renders MiniRoundRail labels from target-job structured interview rounds" "$LOG_FILE"
grep -Fq "derives route round context through target-job round assumptions" "$LOG_FILE"
grep -Fq "does not inherit route resumeId when the saved TargetJob lacks one" "$LOG_FILE"
grep -Fq "starts interview directly from parse with the saved resumeId and no target patch" "$LOG_FILE"
grep -Fq "tests/pixel-parity/parse.spec.ts" "$LOG_FILE"
grep -Fq "readonly plan detail exposes only direct start with bound resume context" "$LOG_FILE"
grep -Fq "start interview hands off directly to practice with bound resume" "$LOG_FILE"
grep -Fq "Frontend architecture screen · 45m" "$LOG_FILE"
grep -Fq "Hiring manager impact interview · 50m" "$LOG_FILE"
grep -Fq "Collaboration and operating style · 40m" "$LOG_FILE"
grep -Fq "E2E.P0.016 parse readonly-detail browser gate resumeId=01918fa0-0000-7000-8000-000000001000 screenshotBytes=" "$LOG_FILE"
grep -Fq "E2E.P0.016 parse start-interview direct browser gate resumeId=01918fa0-0000-7000-8000-000000001000 route=practice noUpdateTargetJob=true" "$LOG_FILE"
# Verify: removed success-detail controls do not appear as positive markers
for forbidden in 'parse-action-save-plan' 'parse-action-cancel' 'parse-action-reparse' 'parse-resume-picker-toggle' 'parse-resume-picker'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "removed parse control found in test output: $forbidden" >&2
    exit 1
  fi
done
for forbidden in 'resume-unbound' 'workspace-missing-resume' 'autoStart browser gate'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "out-of-scope success marker found in test output: $forbidden" >&2
    exit 1
  fi
done
for forbidden in 'Technical Round 1' 'R1 Phone Screen' 'interviewHypotheses'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "out-of-scope round marker found in test output: $forbidden" >&2
    exit 1
  fi
done
