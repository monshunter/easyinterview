#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-057-replay-cta-paths-a-and-b"
LOG="$OUT/trigger.log"

test -s "$LOG"
"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.057
grep -Fq 'E2E.P0.057: validating closed derived-plan requests and one fresh session' "$LOG"
for frontend_file in \
  preflight.test.ts \
  pendingActionReplayPractice.test.ts \
  buildCreatePlanRequest.test.ts \
  startPractice.test.ts \
  ConversationReport.test.tsx \
  ReplayCta.test.tsx; do
  grep -Fq "$frontend_file" "$LOG" || {
    echo "E2E.P0.057: $frontend_file did not run" >&2
    exit 1
  }
done
for frontend_assertion in \
  'creates a closed retry_current_round request without client focus or identity' \
  'creates retry-current-round from only goal + sourceReportId and trusts server-derived context' \
  'uses the server-derived non-contiguous next round without sending round identity' \
  'uses the first action only for CTA visual priority and keeps an empty replay focus valid' \
  'fails closed when non-empty replay focus is not backed by a needs-work dimension and same-code issue' \
  'replay sends only goal + sourceReportId and starts the backend-derived plan' \
  'next-round sends only goal + sourceReportId when frozen context allows it' \
  'locks both CTAs synchronously and creates at most one plan'; do
  grep -Fq "$frontend_assertion" "$LOG"
done

for runtime_file in \
  "$ROOT/frontend/src/app/interview-context/buildCreatePlanRequest.ts" \
  "$ROOT/frontend/src/app/interview-context/startPractice.ts" \
  "$ROOT/frontend/src/app/screens/report/handoff.ts" \
  "$ROOT/frontend/src/app/screens/report/useReplayCtaHandlers.ts"; do
  if rg -n 'retryFocusCompetencyCodes|focusCompetencyCodes|retryFocusDimensionCodes|focusDimensionCodes|evidenceGaps|evidenceGap|answerText|questionText|promptHash' "$runtime_file"; then
    echo "E2E.P0.057: client focus/evidence/raw-text authority leaked into $runtime_file" >&2
    exit 1
  fi
done

grep -Fq 'return { goal, sourceReportId };' "$ROOT/frontend/src/app/interview-context/buildCreatePlanRequest.ts"
grep -Fq 'startPracticeFromParams(runtime.client, params, lang)' "$ROOT/frontend/src/app/screens/report/useReplayCtaHandlers.ts"
grep -Fq 'navigate({ name: "practice", params: started.params })' "$ROOT/frontend/src/app/screens/report/useReplayCtaHandlers.ts"
if rg -n 'ROUND_ORDER|DEFAULT_NEXT_ROUND|inferNextRoundId|useReportContextData' \
    "$ROOT/frontend/src/app/screens/report" \
    "$ROOT/frontend/src/app/interview-context" \
    -g '!*.test.*'; then
  echo "E2E.P0.057: mutable/fallback next-round authority leaked into runtime" >&2
  exit 1
fi
if grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|\[no tests to run\]' "$LOG"; then
  echo "E2E.P0.057: failing or empty runner evidence found" >&2
  exit 1
fi

echo "E2E.P0.057 PASS"
