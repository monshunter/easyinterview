#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-022-practice-plan-baseline-create-and-read"
LOG_FILE="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"
RUN_ID="${TEST_RUN_ID:-practice-plan-$(date -u '+%Y%m%dT%H%M%SZ')}"
RUN_DIR="${TEST_OUTPUT_DIR:-$REPO_ROOT/.test-output}/runs/$RUN_ID/e2e/E2E.P0.022"
non_current_replay_value='lega''cy debrief replay value'
non_current_mode_literal="debrief""_replay"

test -s "$LOG_FILE"
grep -Fq -- '--- PASS: TestE2EP0022PracticePlanBaselineCreateAndRead' "$LOG_FILE"
grep -Eq 'ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api' "$LOG_FILE"

for forbidden in \
  'question_text' \
  'answer_text' \
  'hint_text' \
  'prompt body' \
  'response body' \
  "$non_current_replay_value"; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden scenario evidence leaked: $forbidden" >&2
    exit 1
  fi
done

if rg -n "${non_current_replay_value}|${non_current_mode_literal}" \
  "$REPO_ROOT/shared" \
  "$REPO_ROOT/openapi" \
  "$REPO_ROOT/backend/internal/shared" \
  "$REPO_ROOT/backend/internal/api/generated" \
  "$REPO_ROOT/backend/internal/practice" \
  "$REPO_ROOT/backend/internal/api/practice" \
  "$REPO_ROOT/backend/internal/store/practice" \
  "$REPO_ROOT/backend/internal/middleware/idempotency"; then
  echo "PracticeMode non-current literal must not exist in generated/runtime surfaces" >&2
  exit 1
fi

mkdir -p "$RUN_DIR"
cp "$LOG_FILE" "$RUN_DIR/trigger.log"
printf '{"scenario":"E2E.P0.022","status":"passed","method":"cmd-api-http","validBddEvidence":true,"verifiedAt":"%s","evidence":"%s","snapshot":"in-process practice store asserted plan/audit/idempotency/cross-user isolation"}\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$RUN_DIR/trigger.log" > "$RUN_DIR/result.json"
cp "$RUN_DIR/result.json" "$RESULT_FILE"
