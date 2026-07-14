#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-102-auth-gated-interview-routes"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"

grep -Fq "SCENARIO_RUNNER=E2E.P0.102" "$LOG_FILE"
grep -Fq "=== ui-design-contract ===" "$LOG_FILE"
grep -Fq "Home recent mock interviews are signed-in only" "$LOG_FILE"
grep -Fq "=== frontend-auth-and-single-flight ===" "$LOG_FILE"
grep -Fq "src/app/screens/home/HomeRecentMocks.test.tsx" "$LOG_FILE"
grep -Fq "src/app/screens/home/HomeAuthGate.test.tsx" "$LOG_FILE"
grep -Fq "src/app/AppAuthDispatch.test.tsx" "$LOG_FILE"
grep -Fq "src/api/generatedClient.test.ts" "$LOG_FILE"
grep -Fq "src/app/runtime/AppRuntimeProvider.test.tsx" "$LOG_FILE"
grep -Fq "src/app/screens/home/useRecentTargetJobs.test.tsx" "$LOG_FILE"
grep -Fq "src/app/screens/workspace/hooks/useWorkspaceTargetJobs.test.tsx" "$LOG_FILE"
grep -Fq "src/app/screens/parse/ParseFlow.test.tsx" "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx" "$LOG_FILE"
grep -Fq "E2E.P0.015 Parse StrictMode transport PASS initial=1 tick1=2 tick2=3" "$LOG_FILE"
grep -Fq "E2E.P0.036 summary-only list/detail transport PASS summaryFields=9 listResumes=1 getResumeBeforeOpen=0 getResumeAfterOpen=1" "$LOG_FILE"
grep -Fq "E2E.P0.036 list rejection retry transport PASS initialRejected=1 retrySucceeded=2" "$LOG_FILE"
grep -Eq 'Test Files +9 passed \(9\)' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed \([0-9]+\)' "$LOG_FILE"
grep -Fq "=== backend-session-policy ===" "$LOG_FILE"
grep -Fq -- "--- PASS: TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations" "$LOG_FILE"
grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/auth' "$LOG_FILE"
grep -Fq "=== backend-route-middleware ===" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerMountsTargetJobRoutesBehindSessionMiddleware" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerMountsUploadPresignBehindSessionMiddleware" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerMountsResumeRoutesBehindSessionMiddleware" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerMountsPracticeRoutesBehindSessionMiddleware" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerDoesNotMountOutOfScopeDebriefOrProfileRoutes" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerMountsReportRoutesBehindSessionMiddleware" "$LOG_FILE"
grep -Fq -- "--- PASS: TestBuildAPIHandlerMountsJobRouteBehindSessionMiddleware" "$LOG_FILE"
grep -Fq -- "--- PASS: TestJDMatchRoutesRemainUnmountedPerD17" "$LOG_FILE"
grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api' "$LOG_FILE"

for forbidden in \
  "--- FAIL:" \
  "--- SKIP:" \
  "FAIL" \
  "SKIP" \
  "no tests to run" \
  "[no tests to run]" \
  "testing: warning: no tests to run" \
  "missing fixture for operationId: listTargetJobs" \
  "Open debrief" \
  "pendingRoute: \"debrief\"" \
  "pendingRoute=debrief" \
  "expected document not to contain element"; do
  if grep -Fq -- "$forbidden" "$LOG_FILE"; then
    echo "forbidden marker leaked into scenario evidence: $forbidden" >&2
    exit 1
  fi
done

cat > "$OUTPUT_DIR/result.json" <<'JSON'
{"scenario_id":"E2E.P0.102","suite_id":"e2e","mode":"automated","result":"PASS","status":"passed"}
JSON
