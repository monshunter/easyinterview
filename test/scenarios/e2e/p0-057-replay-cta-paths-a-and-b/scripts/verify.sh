#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-057-replay-cta-paths-a-and-b"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'E2E.P0.057: validating direct-start owner contract' "$LOG_FILE" || { echo "E2E.P0.057: direct-start owner preflight did not run" >&2; exit 1; }
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE" || { echo "E2E.P0.057: no passing test files" >&2; exit 1; }
grep -Fq 'preflight.test.ts' "$LOG_FILE" || { echo "E2E.P0.057: report owner preflight did not pass" >&2; exit 1; }
grep -Fq 'pendingActionReplayPractice.test.ts' "$LOG_FILE" || { echo "E2E.P0.057: pendingAction replay test did not run" >&2; exit 1; }
grep -Fq 'ReplayCta.test.tsx' "$LOG_FILE" || { echo "E2E.P0.057: ReplayCta test did not run" >&2; exit 1; }
grep -Fq 'TestReplayCtaPathA_AuthenticatedDirectStartPractice' \
  "$REPO_ROOT/frontend/src/app/screens/report/__tests__/ReplayCta.test.tsx" || { echo "E2E.P0.057: replay direct-start assertion is missing" >&2; exit 1; }
grep -Fq 'TestNextRoundCta_DirectStartPractice' \
  "$REPO_ROOT/frontend/src/app/screens/report/__tests__/ReplayCta.test.tsx" || { echo "E2E.P0.057: next-round direct-start assertion is missing" >&2; exit 1; }
grep -Fq 'startPracticeFromParams(runtime.client, params, lang)' \
  "$REPO_ROOT/frontend/src/app/screens/report/useReplayCtaHandlers.ts" || { echo "E2E.P0.057: replay CTA does not use the shared direct-start helper" >&2; exit 1; }
grep -Fq 'navigate({ name: "practice", params: started.params })' \
  "$REPO_ROOT/frontend/src/app/screens/report/useReplayCtaHandlers.ts" || { echo "E2E.P0.057: replay CTA does not navigate directly to practice" >&2; exit 1; }
grep -Fq 'route: "report"' \
  "$REPO_ROOT/frontend/src/app/auth/__tests__/pendingActionReplayPractice.test.ts" || { echo "E2E.P0.057: replay pending action does not restore the report route" >&2; exit 1; }
grep -Fq 'client.createPracticePlan(' \
  "$REPO_ROOT/frontend/src/app/interview-context/startPractice.ts" || { echo "E2E.P0.057: direct-start helper does not create a derived plan" >&2; exit 1; }
grep -Fq 'client.startPracticeSession(' \
  "$REPO_ROOT/frontend/src/app/interview-context/startPractice.ts" || { echo "E2E.P0.057: direct-start helper does not start a fresh session" >&2; exit 1; }

# Replay CTA payload must not assign raw text literals into payload keys.
# Match key-style usages (`answerText:`, `answerText =`, `"answerText"`).
if grep -RnE '("answerText"|''answerText''|answerText\s*[:=]|"questionText"|''questionText''|questionText\s*[:=]|promptHash\s*[:=]|"promptHash")' \
    "$REPO_ROOT/frontend/src/app/screens/report" \
    --include='*.tsx' --include='*.ts' --exclude-dir=__tests__; then
  echo "E2E.P0.057: raw text leaked into replay payload code" >&2
  exit 1
fi

# practiceGoal values surface in both paths.
grep -Fq 'retry_current_round' "$REPO_ROOT/frontend/src/app/screens/report/handoff.ts"
grep -Fq 'next_round' "$REPO_ROOT/frontend/src/app/screens/report/handoff.ts"
