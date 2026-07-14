#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-089-url-routing-auth-privacy"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "source_contract_test.py" "$SCRIPT_DIR/trigger.sh"
grep -Fq "Ran 3 tests" "$LOG_FILE"
grep -Fq "OK" "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx" "$LOG_FILE"
grep -Fq "restores an unauthenticated Reports deep link with targetJobId only" "$LOG_FILE"
grep -Fq "navigate(reports) keeps only targetJobId" "$LOG_FILE"
grep -Fq "navigate(workspace) with raw markers drops every marker" "$LOG_FILE"
grep -Fq "renders a target-scoped workspace as read-only detail with one getTargetJob" "$LOG_FILE"
grep -Fq "retains targetJobId as the sole workspace detail locator" "$LOG_FILE"
grep -Fq "auth/login direct open for Reports keeps targetJobId" "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"

# Verify scenario design carries raw markers in test sources so the
# allowlist drop is actually exercised (negative grep target identifiers).
SOURCE_FILE="$REPO_ROOT/frontend/src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx"
test -s "$SOURCE_FILE"
for marker in \
  'RAW_JD_TEXT_2c1a' \
  'RAW_GUIDED_ANSWER_7b6f' \
  'RAW_AI_PROMPT_0412' \
  'AUTH_SECRET_TOKEN_3745' \
  'AUTH_PASSWORD_4856'; do
  grep -Fq "$marker" "$SOURCE_FILE"
done

# Trigger output must NOT echo raw markers (test runner does not log them
# either, but the negative grep guards against future leaks).
for marker in \
  'RAW_JD_TEXT_2c1a' \
  'RAW_GUIDED_ANSWER_7b6f' \
  'RAW_AI_PROMPT_0412' \
  'AUTH_SECRET_TOKEN_3745' \
  'AUTH_PASSWORD_4856'; do
  if grep -Fq "$marker" "$LOG_FILE"; then
    echo "raw privacy marker leaked into scenario trigger log: $marker" >&2
    exit 1
  fi
done
