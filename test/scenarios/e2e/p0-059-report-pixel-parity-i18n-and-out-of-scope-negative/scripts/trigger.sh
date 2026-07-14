#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative"

if [ ! -s "$OUTPUT_DIR/setup.env" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

(
  cd "$REPO_ROOT"
  echo "E2E.P0.059: validating current-plan reports owner and browser evidence contract"
  python3 "$SCRIPT_DIR/script_contract_test.py" -v
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"

  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/targetJobReports.test.ts \
    src/app/screens/reports/__tests__/ReportsScreen.test.tsx \
    src/app/screens/report/__tests__/reportContract.test.ts \
    src/app/screens/report/__tests__/ConversationReport.test.tsx \
    src/app/screens/generating/__tests__/GeneratingBackNavigation.test.tsx \
    src/app/screens/generating/__tests__/GeneratingScreen.test.tsx \
    src/app/screens/report/__tests__/preflight.test.ts \
    src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts \
    src/app/screens/report/__tests__/outOfScopeNegative.test.ts \
    src/app/screens/generating/__tests__/outOfScopeNegative.test.ts \
    --reporter=verbose

  python3 scripts/lint/frontend_report_dashboard_out_of_scope.py --repo-root . --phase E2E.P0.059
  python3 -m pytest scripts/lint/frontend_report_dashboard_out_of_scope_test.py -q

  echo "E2E.P0.059: frontend build before pixel parity"
  COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend build

  echo "E2E.P0.059: running ReportsScreen, Report, and Generating Playwright pixel parity"
  echo "E2E.P0.059: viewports=1440x900,390x844; states=ready,loading,empty,error,latest-ready,mismatch; DOM/style/bbox; pixelmatch threshold 0.1; changed-pixel ratio <=0.5%"
  cd "$REPO_ROOT/frontend"
  CI=1 COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm exec playwright test \
    tests/pixel-parity/reports.spec.ts \
    tests/pixel-parity/report.spec.ts \
    tests/pixel-parity/generating.spec.ts \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUTPUT_DIR/playwright"
  echo "E2E.P0.059: Playwright pixel parity complete"
) 2>&1 | tee "$OUTPUT_DIR/trigger.log"
