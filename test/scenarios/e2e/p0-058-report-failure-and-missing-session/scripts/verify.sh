#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-058-report-failure-and-missing-session"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}"
grep -Fq 'E2E.P0.058: validating focused failure contracts' "$LOG_FILE" || { echo "E2E.P0.058: focused failure preflight did not run" >&2; exit 1; }
grep -Fq 'preflight.test.ts' "$LOG_FILE" || { echo "E2E.P0.058: owner preflight did not pass" >&2; exit 1; }
grep -Fq 'ReportFailureState.test.tsx' "$LOG_FILE" || { echo "E2E.P0.058: failure-state test did not pass" >&2; exit 1; }
grep -Fq 'ReportMissingSessionState.test.tsx' "$LOG_FILE" || { echo "E2E.P0.058: missing-state test did not pass" >&2; exit 1; }
grep -Fq 'useFeedbackReport.test.tsx' "$LOG_FILE" || { echo "E2E.P0.058: report hook test did not pass" >&2; exit 1; }
grep -Fq 'ConversationReport.test.tsx' "$LOG_FILE" || { echo "E2E.P0.058: conversation report test did not pass" >&2; exit 1; }
grep -Fq 'useReportGenerationPoll.test.tsx' "$LOG_FILE" || { echo "E2E.P0.058: poll failure test did not pass" >&2; exit 1; }

# AI_* enum is covered + REPORT_NOT_FOUND has separate copy.
grep -Fq 'AI_PROVIDER_TIMEOUT' "$REPO_ROOT/frontend/src/app/i18n/locales/zh.ts" || { echo "E2E.P0.058: zh missing AI_PROVIDER_TIMEOUT" >&2; exit 1; }
grep -Fq 'failureState.notFound.title' "$REPO_ROOT/frontend/src/app/i18n/locales/en.ts" || { echo "E2E.P0.058: en missing notFound title" >&2; exit 1; }
