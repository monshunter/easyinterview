#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-059-report-pixel-parity-i18n-and-legacy-negative"
mkdir -p "$OUTPUT_DIR"
(
  cd "$REPO_ROOT"
  "$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend test \
    src/app/i18n/__tests__/reportDashboardI18nCoverage.test.ts \
    src/app/screens/report/__tests__/legacyNegative.test.ts \
    src/app/screens/generating/__tests__/legacyNegative.test.ts
  python3 scripts/lint/frontend_report_dashboard_legacy.py --repo-root . --phase E2E.P0.059
  python3 -m pytest scripts/lint/frontend_report_dashboard_legacy_test.py -q
  echo "E2E.P0.059: building frontend before pixel parity"
  pnpm --filter @easyinterview/frontend build
  echo "E2E.P0.059: running Playwright pixel parity"
  pnpm --filter @easyinterview/frontend test:pixel-parity -- \
    tests/pixel-parity/generating.spec.ts \
    tests/pixel-parity/report.spec.ts
  echo "E2E.P0.059: Playwright pixel parity complete"
) | tee "$OUTPUT_DIR/trigger.log"
