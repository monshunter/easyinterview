#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-014-home-default-render"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"
grep -Fq 'VITE_EI_API_MODE=real' "$LOG_FILE"
grep -Fq 'VITE_EI_API_BASE_URL=http://localhost:8080/api/v1' "$LOG_FILE"
grep -Fq 'targetJob.realApiMode.test.ts' "$LOG_FILE"
grep -Fq "HomeScreen.test.tsx" "$LOG_FILE"
grep -Fq "HomeResumeSelection.test.tsx" "$LOG_FILE"
grep -Fq "HomeRecentMocks.test.tsx" "$LOG_FILE"
grep -Fq "MockInterviewCard.test.tsx" "$LOG_FILE"
grep -Eq 'Test Files +[0-9]+ passed' "$LOG_FILE"
grep -Eq 'Tests +[0-9]+ passed' "$LOG_FILE"
# Negative: legacy testids not in output
for forbidden in 'route-welcome' 'topbar-nav-mistakes' 'topbar-nav-growth' 'topbar-nav-drill' 'topbar-nav-voice' 'home-pasted-success' 'home-mocked-recent' 'jdmatch-card-' 'jdmatch-market-signal-'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "forbidden legacy entry leaked: $forbidden" >&2
    exit 1
  fi
done

for forbidden in '粘贴 JD，或继续最近一次模拟面试。每一次练习都绑定具体岗位，而不是泛用题库。' '解析并确认面试'; do
  if grep -R --exclude='*.test.ts' --exclude='*.test.tsx' -F "$forbidden" \
    "$REPO_ROOT/frontend/src/app/screens/home" \
    "$REPO_ROOT/frontend/src/app/i18n" \
    "$REPO_ROOT/ui-design/src/screen-home.jsx"; then
    echo "retired Home copy remains in source: $forbidden" >&2
    exit 1
  fi
done
