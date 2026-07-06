#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-015-jd-import-and-parse"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq 'VITE_EI_API_MODE=real' "$LOG_FILE"
grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1' "$LOG_FILE"
grep -Fq 'targetJob.realApiMode.test.ts' "$LOG_FILE"
grep -Fq "JDAssistModal.test.tsx" "$LOG_FILE"
grep -Fq "HomeResumeSelection.test.tsx" "$LOG_FILE"
grep -Fq "HomeImport.test.tsx" "$LOG_FILE"
grep -Fq "HomeAuthGate.test.tsx" "$LOG_FILE"
grep -Fq "ParseScreen.test.tsx" "$LOG_FILE"
grep -Fq "ParseFlow.test.tsx" "$LOG_FILE"
grep -Fq "ParseFailedState.test.tsx" "$LOG_FILE"
grep -Fq "ParseEdit.test.tsx" "$LOG_FILE"
grep -Fq "tests/pixel-parity/parse.spec.ts" "$LOG_FILE"
grep -Fq "ready target job response keeps ui-design loading demo before preview" "$LOG_FILE"
grep -Fq "E2E.P0.015 ready-response loading browser gate screenshotBytes=" "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
grep -Eq '[0-9]+ passed' "$LOG_FILE"
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

# Source-level negative gate: frontend implementation must not hard-code
# provider/model/prompt assumptions while rendering home/import/parse.
if grep -R --exclude='*.test.ts' --exclude='*.test.tsx' -E 'claude|haiku|prompt@|prompt\.registry|promptRegistry|provider\.key|providerKey|AIClient|LLM\.endpoint' \
  "$REPO_ROOT/frontend/src/app/screens/home" \
  "$REPO_ROOT/frontend/src/app/screens/parse"; then
  echo "AI provider or prompt assumption leaked in frontend source" >&2
  exit 1
fi

for forbidden in '粘贴 JD，或继续最近一次模拟面试。每一次练习都绑定具体岗位，而不是泛用题库。' '解析并确认面试'; do
  if grep -R --exclude='*.test.ts' --exclude='*.test.tsx' -F "$forbidden" \
    "$REPO_ROOT/frontend/src/app/screens/home" \
    "$REPO_ROOT/frontend/src/app/i18n" \
    "$REPO_ROOT/ui-design/src/screen-home.jsx"; then
    echo "retired Home copy remains in source: $forbidden" >&2
    exit 1
  fi
done
