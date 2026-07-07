#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-102-auth-gated-interview-routes"

mkdir -p "$OUTPUT_DIR"

run_step() {
  local name="$1"
  shift
  echo "=== $name ==="
  "$@"
}

(
  cd "$REPO_ROOT"
  echo "SCENARIO_RUNNER=E2E.P0.102"
  run_step ui-design-contract \
    node --test ui-design/ui-design-contract.test.mjs
  run_step frontend-auth-gate \
    pnpm --filter @easyinterview/frontend test \
      src/app/screens/home/HomeRecentMocks.test.tsx \
      src/app/screens/home/HomeAuthGate.test.tsx \
      src/app/AppAuthDispatch.test.tsx
  run_step backend-session-policy \
    bash -c 'cd backend && go test -v ./internal/auth -run TestSessionPolicyClassifiesPublicOptionalAndProtectedOperations -count=1'
  run_step backend-route-middleware \
    bash -c 'cd backend && go test -v ./cmd/api -run '"'"'TestBuildAPIHandlerMounts(TargetJobRoutes|UploadPresign|ResumeRoutes|PracticeRoutes|ReportRoutes|JobRoute)BehindSessionMiddleware|TestBuildAPIHandlerDoesNotMountNonCurrentDebriefOrProfileRoutes|TestJDMatchRoutesRemainUnmountedPerD17'"'"' -count=1'
) | tee "$OUTPUT_DIR/trigger.log"
