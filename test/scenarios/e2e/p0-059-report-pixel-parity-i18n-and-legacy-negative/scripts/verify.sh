#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-059-report-pixel-parity-i18n-and-legacy-negative"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE" || { echo "E2E.P0.059: i18n + legacy negative tests did not pass" >&2; exit 1; }
grep -Fq 'frontend-report-dashboard legacy lint OK' "$LOG_FILE" || { echo "E2E.P0.059: legacy lint script did not succeed" >&2; exit 1; }
grep -Fq 'E2E.P0.059: running Playwright pixel parity' "$LOG_FILE" || { echo "E2E.P0.059: Playwright pixel parity did not run" >&2; exit 1; }
grep -Fq 'tests/pixel-parity/generating.spec.ts' "$LOG_FILE" || { echo "E2E.P0.059: generating pixel parity spec was not executed" >&2; exit 1; }
grep -Fq 'tests/pixel-parity/report.spec.ts' "$LOG_FILE" || { echo "E2E.P0.059: report pixel parity spec was not executed" >&2; exit 1; }
awk '
  /E2E\.P0\.059: running Playwright pixel parity/ { in_playwright = 1 }
  in_playwright && /^[[:space:]]*[0-9]+ passed/ { passed = 1 }
  END { exit passed ? 0 : 1 }
' "$LOG_FILE" || { echo "E2E.P0.059: Playwright pixel parity pass marker missing" >&2; exit 1; }
grep -Fq 'E2E.P0.059: Playwright pixel parity complete' "$LOG_FILE" || { echo "E2E.P0.059: Playwright pixel parity did not complete" >&2; exit 1; }
