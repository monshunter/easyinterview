#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-016-parse-confirm-to-workspace"
SETUP_ENV="$OUTPUT_DIR/setup.env"

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

(
  cd "$REPO_ROOT"
  echo "SCENARIO_RUNNER=E2E.P0.016"
  python3 "$SCRIPT_DIR/source_contract_test.py" -v

  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" \
    VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" \
    COREPACK_ENABLE_DOWNLOAD_PROMPT=0 \
    corepack pnpm --filter @easyinterview/frontend exec vitest run \
      src/api/targetJob.realApiMode.test.ts \
      --reporter=verbose

  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend exec vitest run \
    src/app/screens/parse/ParseReports.test.tsx \
    src/app/screens/parse/ParseScreen.test.tsx \
    src/app/screens/parse/ParseFlow.test.tsx \
    src/app/screens/parse/ParseEdit.test.tsx \
    src/app/screens/parse/ParseAuthGate.test.tsx \
    src/app/screens/parse/ParseResumeBinding.test.tsx \
    src/app/screens/parse/ParseRoundStates.test.tsx \
    src/app/screens/home/MockInterviewCard.test.tsx \
    src/app/screens/home/HomeRecentMocks.test.tsx \
    src/app/navigation/interviewContext.test.ts \
    src/app/routeUrl.test.ts \
    src/app/topbar/TopBar.test.tsx \
    --reporter=verbose

  echo "E2E.P0.016: frontend build before browser parity"
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend build

  cd "$REPO_ROOT/frontend"
  CI=1 COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm exec playwright test \
    tests/pixel-parity/parse.spec.ts \
    --grep 'workspace detail exposes only direct start with bound resume context|workspace detail round states match the UI truth at desktop and mobile|workspace plan-detail reports entry matches the UI truth and stays report-list-free|workspace start interview hands off directly to practice with bound resume' \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUTPUT_DIR/playwright"
) 2>&1 | tee "$OUTPUT_DIR/trigger.log"
