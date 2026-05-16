#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-059-report-pixel-parity-i18n-and-legacy-negative"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE" || { echo "E2E.P0.059: i18n + legacy negative tests did not pass" >&2; exit 1; }
grep -Fq 'frontend-report-dashboard legacy lint OK' "$LOG_FILE" || { echo "E2E.P0.059: legacy lint script did not succeed" >&2; exit 1; }
grep -Fq 'passed' "$LOG_FILE"

# Confirm both pixel-parity spec files exist (Playwright runs are user-driven).
test -s "$REPO_ROOT/frontend/tests/pixel-parity/generating.spec.ts"
test -s "$REPO_ROOT/frontend/tests/pixel-parity/report.spec.ts"
