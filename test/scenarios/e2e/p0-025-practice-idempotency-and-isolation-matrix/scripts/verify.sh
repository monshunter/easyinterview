#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-025-practice-idempotency-and-isolation-matrix"
LOG_FILE="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"
RUN_ID="${TEST_RUN_ID:-practice-idempotency-$(date -u '+%Y%m%dT%H%M%SZ')}"
RUN_DIR="${TEST_OUTPUT_DIR:-$REPO_ROOT/.test-output}/runs/$RUN_ID/e2e/E2E.P0.025"

test -s "$LOG_FILE"
grep -Fq -- '--- PASS: TestE2EP0025PracticeIdempotencyAndIsolationMatrix' "$LOG_FILE"
grep -Eq 'ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api' "$LOG_FILE"

for forbidden in 'prompt body' 'response body' 'provider secret' 'sk-test'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden scenario evidence leaked: $forbidden" >&2
    exit 1
  fi
done

mkdir -p "$RUN_DIR"
cp "$LOG_FILE" "$RUN_DIR/trigger.log"
printf '{"scenario":"E2E.P0.025","status":"passed","method":"cmd-api-http","validBddEvidence":true,"verifiedAt":"%s","evidence":"%s","snapshot":"in-process practice store asserted idempotency replay mismatch isolation and active-plan conflict"}\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$RUN_DIR/trigger.log" > "$RUN_DIR/result.json"
cp "$RUN_DIR/result.json" "$RESULT_FILE"
