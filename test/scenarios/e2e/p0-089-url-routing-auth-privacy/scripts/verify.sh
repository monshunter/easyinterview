#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-089-url-routing-auth-privacy"
LOG_FILE="$OUTPUT_DIR/trigger.log"

test -s "$LOG_FILE"
grep -Fq "src/app/scenarios/p0-089-url-routing-auth-privacy.test.tsx" "$LOG_FILE"
grep -Eq 'Tests +4 passed \(4\)' "$LOG_FILE"
grep -Eq 'Test Files +1 passed \(1\)' "$LOG_FILE"

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
