#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-026-practice-observability-and-privacy-redlines"
LOG_FILE="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"
RUN_ID="${TEST_RUN_ID:-practice-observability-$(date -u '+%Y%m%dT%H%M%SZ')}"
RUN_DIR="${TEST_OUTPUT_DIR:-$REPO_ROOT/.test-output}/runs/$RUN_ID/e2e/E2E.P0.026"

test -s "$LOG_FILE"
for marker in \
  '--- PASS: TestPracticeOutboxPayloadContainsOnlyLifecycleData' \
  '--- PASS: TestPrivacy_NoPlaintextLeaksAnywhere' \
  '--- PASS: TestDecorator_ReportMetricLabelsExcludeProvenanceAndRawModelID' \
  '--- PASS: TestGenerateReportUsesOneConversationLevelAICall'; do
  grep -Fq -- "$marker" "$LOG_FILE"
done
if grep -Fq 'no tests to run' "$LOG_FILE"; then
  echo "E2E.P0.026: focused gate matched no tests" >&2
  exit 1
fi

for forbidden in 'question_text' 'answer_text' 'hint_text' 'prompt body' 'response body' 'provider secret' 'sk-test'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden scenario evidence leaked: $forbidden" >&2
    exit 1
  fi
done

python3 "$REPO_ROOT/scripts/lint/backend_practice_out_of_scope.py" --repo-root "$REPO_ROOT" --phase all

mkdir -p "$RUN_DIR"
cp "$LOG_FILE" "$RUN_DIR/trigger.log"
printf '{"scenario":"E2E.P0.026","status":"passed","method":"focused-privacy-observability","validBddEvidence":true,"verifiedAt":"%s","evidence":"%s","snapshot":"lifecycle-only outbox plaintext redaction metric allowlist and conversation-report privacy"}\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$RUN_DIR/trigger.log" > "$RUN_DIR/result.json"
cp "$RUN_DIR/result.json" "$RESULT_FILE"
