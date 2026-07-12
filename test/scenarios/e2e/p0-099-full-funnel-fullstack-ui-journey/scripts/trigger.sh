#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
SETUP="$OUT/setup.env"
EVIDENCE="$OUT/evidence.md"
RESULT="$OUT/result.json"
test -s "$SETUP"
RUN_ID="$(sed -n 's/^RUN_ID=//p' "$SETUP")"

{
  echo "E2E.P0.099 real continuous-conversation browser acceptance"
  cd "$ROOT"
  test/scenarios/env-verify.sh
  go test -v ./backend/internal/store/practice ./backend/internal/store/review -run 'TestSQLRepositoryCompleteSessionUsesLifecycleOnlyEventColumns|TestPersistReportUsesPostgresTextArrayForRetryFocus|TestUpdateFeedbackReportStatusAllowsGeneratingRetry' -count=1
  pnpm --filter @easyinterview/frontend test -- src/app/screens/practice/PracticeScreen.test.tsx src/app/screens/report/__tests__/ConversationReport.test.tsx
} | tee "$OUT/trigger.log"

required_screenshots=(
  practice-conversation-desktop.png
  practice-conversation-mobile.png
  conversation-report-desktop.png
  conversation-report-mobile.png
)
for file in "${required_screenshots[@]}"; do
  test -s "$OUT/screenshots/$file"
done

for marker in \
  "run_id=$RUN_ID" \
  "environment=shared-host-run-real-provider" \
  "auth=mailpit-email-code" \
  "practice=continuous-conversation" \
  "voice=disabled" \
  "report=conversation-level-ready" \
  "database=ready" \
  "privacy=redacted"; do
  grep -Fq -- "$marker" "$EVIDENCE"
done

jq -n --arg run_id "$RUN_ID" --arg completed_at "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" \
  '{scenario:"E2E.P0.099",status:"PASS",run_id:$run_id,completed_at:$completed_at}' > "$RESULT"
printf 'scenario=E2E.P0.099\nmethod=hybrid-real-browser\nRUN_ID=%s\ntrigger_at=%s\n' "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/trigger.env"
echo "trigger: ok"
