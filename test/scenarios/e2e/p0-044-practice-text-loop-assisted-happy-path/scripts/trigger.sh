#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path"
SETUP_ENV="$OUT/setup.env"
SOURCE_MANIFEST="$ROOT/test/scenarios/e2e/practice-source-fingerprint-paths.json"

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

# shellcheck disable=SC1090
. "$SETUP_ENV"

rm -rf "$OUT/playwright" "$OUT/screenshots"
mkdir -p "$OUT/screenshots"
rm -f "$OUT"/*.log "$OUT/result.json" "$OUT/source-fingerprint.json" "$OUT/source-fingerprint.verify.json"

python3 "$ROOT/test/scenarios/_shared/scripts/capture-source-fingerprint.py" \
  --repo-root "$ROOT" \
  --output "$OUT/source-fingerprint.json" \
  --source-paths-from "$SOURCE_MANIFEST"

copy_screenshot() {
  local name="$1"
  local source
  local metadata_name="${name%.png}.metadata.json"
  local metadata_source
  source="$(find "$OUT/playwright" -type f -name "$name" -print -quit)"
  if [ -z "$source" ]; then
    echo "trigger: missing Playwright screenshot $name" >&2
    return 1
  fi
  metadata_source="$(find "$OUT/playwright" -type f -name "$metadata_name" -print -quit)"
  if [ -z "$metadata_source" ]; then
    echo "trigger: missing Playwright screenshot metadata $metadata_name" >&2
    return 1
  fi
  cp "$source" "$OUT/screenshots/$name"
  cp "$metadata_source" "$OUT/screenshots/$metadata_name"
}

printf 'SCENARIO_RUNNER=E2E.P0.044\nRUN_ID=%s\n' "$run_id" \
  | tee "$OUT/scenario-start.log" \
  | tee "$OUT/trigger.log"

"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT" \
  2>&1 | tee "$OUT/real-backend.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/frontend"
  pnpm exec vitest run \
    src/app/screens/practice/PracticeScreen.test.tsx \
    src/app/screens/practice/PracticeI18n.test.ts \
    src/app/screens/practice/components/Transcript.test.tsx \
    src/app/screens/practice/hooks/usePracticeMessages.test.tsx \
    src/app/screens/practice/hooks/usePracticeSessionLoader.test.tsx \
    --reporter=verbose
) 2>&1 | tee "$OUT/frontend-contract.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT"
  pnpm --filter @easyinterview/frontend build
) 2>&1 | tee "$OUT/frontend-build.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/backend"
  go test ./internal/api/practice ./internal/practice ./internal/store/practice \
    -run '^(TestSendPracticeMessageReturnsConversationMessages|TestSendPracticeMessageUsesOrdinaryConversationHistory|TestSendPracticeMessagePendingSameIDDoesNotCallAI|TestSQLRepositoryGetSessionReturnsUserReplyRecoveryStateOnly|TestSQLRepositoryGetSessionKeepsPendingBeforeLeaseBoundary|TestSQLRepositoryReservePracticeMessageRetriesOnlyRetryableFailure|TestSQLRepositoryCommitPracticeMessageInsertsReplyAndCompletesUserAtomically)$' \
    -count=1 -v
) 2>&1 | tee "$OUT/reply-state-contract.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/frontend"
  CI=1 pnpm exec playwright test tests/pixel-parity/practice.spec.ts \
    --grep 'renders one full-width chat with no structured-question surfaces|user and assistant GFM keep prototype typography with only local pre/table overflow|new user input is visible before the reply and locks the composer|reloads a persisted pending reply, keeps all actions locked, and sends zero POSTs' \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUT/playwright"
) 2>&1 | tee "$OUT/playwright.log" | tee -a "$OUT/trigger.log"

for screenshot in \
  practice-immediate-pending-desktop.png \
  practice-immediate-pending-mobile.png \
  practice-persisted-pending-desktop.png \
  practice-persisted-pending-mobile.png \
  practice-markdown-gfm-desktop.png \
  practice-markdown-gfm-mobile.png; do
  copy_screenshot "$screenshot"
done
echo 'PRACTICE_P0044_SCREENSHOT_CAPTURE_PASS viewports=1440x900,390x844 states=immediate-pending,persisted-pending,markdown-gfm' \
  | tee "$OUT/scenario-finish.log" \
  | tee -a "$OUT/trigger.log"
echo 'PRACTICE_IMMEDIATE_PENDING_PASS user_row=immediate composer_locked=true thinking=true' \
  | tee -a "$OUT/scenario-finish.log" \
  | tee -a "$OUT/trigger.log"
echo 'PRACTICE_PERSISTED_PENDING_PASS reload=true message_posts=0 lease_before_expiry=true lease_seconds=90' \
  | tee -a "$OUT/scenario-finish.log" \
  | tee -a "$OUT/trigger.log"
echo 'PRACTICE_SAFE_GFM_PROJECTION_PASS roles=user,assistant semantic=true prototype_parity=true mobile_local_overflow=true document_overflow=0' \
  | tee -a "$OUT/scenario-finish.log" \
  | tee -a "$OUT/trigger.log"
