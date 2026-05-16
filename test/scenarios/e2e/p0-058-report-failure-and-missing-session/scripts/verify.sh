#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-058-report-failure-and-missing-session"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE" || { echo "E2E.P0.058: no passing test files" >&2; exit 1; }
grep -Fq 'ReportFailureState.test.tsx' "$LOG_FILE"
grep -Fq 'ReportMissingSessionState.test.tsx' "$LOG_FILE"
grep -Fq 'useFeedbackReport.test.tsx' "$LOG_FILE"

# AI_* enum is covered + REPORT_NOT_FOUND has separate copy.
grep -Fq 'AI_PROVIDER_TIMEOUT' "$REPO_ROOT/frontend/src/app/i18n/locales/zh.ts" || { echo "E2E.P0.058: zh missing AI_PROVIDER_TIMEOUT" >&2; exit 1; }
grep -Fq 'failureState.notFound.title' "$REPO_ROOT/frontend/src/app/i18n/locales/en.ts" || { echo "E2E.P0.058: en missing notFound title" >&2; exit 1; }
