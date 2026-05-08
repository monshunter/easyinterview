#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-015-jd-import-and-parse"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq "JDAssistModal.test.tsx" "$LOG_FILE"
grep -Fq "HomeImport.test.tsx" "$LOG_FILE"
grep -Fq "HomeAuthGate.test.tsx" "$LOG_FILE"
grep -Fq "ParseScreen.test.tsx" "$LOG_FILE"
grep -Fq "ParseFlow.test.tsx" "$LOG_FILE"
grep -Fq "ParseFailedState.test.tsx" "$LOG_FILE"
grep -Fq "ParseEdit.test.tsx" "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
# Privacy redline: JD raw text must not appear in trigger log
for forbidden in 'rawText:' 'raw_jd_text' 'sourceUrl:' 'console.log'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "privacy redline violation: $forbidden" >&2
    exit 1
  fi
done
# No AI provider/prompt registry references
for forbidden in 'prompt.registry' 'promptRegistry' 'provider.key' 'providerKey' 'AIClient' 'LLM.endpoint'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "AI provider reference leaked: $forbidden" >&2
    exit 1
  fi
done
