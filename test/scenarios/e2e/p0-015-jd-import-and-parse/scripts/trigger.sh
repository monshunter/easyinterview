#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-015-jd-import-and-parse"
SETUP_ENV="$OUTPUT_DIR/setup.env"
if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi
# shellcheck disable=SC1090
. "$SETUP_ENV"

rm -rf "$OUTPUT_DIR/playwright-home" "$OUTPUT_DIR/playwright-parse" "$OUTPUT_DIR/screenshots"
mkdir -p "$OUTPUT_DIR/screenshots"
rm -f "$OUTPUT_DIR/trigger.log" "$OUTPUT_DIR/result.json" \
  "$OUTPUT_DIR/source-fingerprint.json" "$OUTPUT_DIR/source-fingerprint.verify.json"

copy_screenshot() {
  local output_root="$1"
  local name="$2"
  local source
  source="$(find "$output_root" -type f -name "$name" -print -quit)"
  if [ -z "$source" ]; then
    echo "trigger: missing Playwright screenshot $name" >&2
    return 1
  fi
  cp "$source" "$OUTPUT_DIR/screenshots/$name"
}

python3 "$REPO_ROOT/test/scenarios/_shared/scripts/capture-source-fingerprint.py" \
  --repo-root "$REPO_ROOT" \
  --output "$OUTPUT_DIR/source-fingerprint.json" \
  --path frontend/src/app/screens/home \
  --path frontend/src/app/screens/parse \
  --path frontend/src/app/i18n \
  --path frontend/tests/pixel-parity/home.spec.ts \
  --path frontend/tests/pixel-parity/parse.spec.ts \
  --path frontend/tests/pixel-parity/report-parity-helpers.ts \
  --path openapi/openapi.yaml \
  --path openapi/fixtures \
  --path ui-design/src \
  --path docs/ui-design \
  --path test/scenarios/_shared/scripts/capture-source-fingerprint.py \
  --path test/scenarios/e2e/p0-015-jd-import-and-parse

(
  cd "$REPO_ROOT"
  REAL_API_MODE="${VITE_EI_API_MODE:-real}"
  REAL_API_BASE_URL="${VITE_EI_API_BASE_URL:-http://localhost:8080/api/v1}"
  printf 'VITE_EI_API_MODE=%s\nVITE_EI_API_BASE_URL=%s\n' "$REAL_API_MODE" "$REAL_API_BASE_URL"
  VITE_EI_API_MODE="$REAL_API_MODE" VITE_EI_API_BASE_URL="$REAL_API_BASE_URL" pnpm --filter @easyinterview/frontend exec vitest run \
    src/api/targetJob.realApiMode.test.ts
  pnpm --filter @easyinterview/frontend test \
    src/app/screens/home/HomeScreen.test.tsx \
    src/app/screens/home/HomeLayout.test.tsx \
    src/app/screens/home/HomeResumeSelection.test.tsx \
    src/app/screens/home/HomeImport.test.tsx \
    src/app/screens/home/HomeAuthGate.test.tsx \
    src/app/screens/home/pendingImportState.test.ts \
    src/app/screens/parse/ParseScreen.test.tsx \
    src/app/screens/parse/ParseFlow.test.tsx \
    src/app/screens/parse/ParseFailedState.test.tsx \
    src/app/screens/parse/ParseEdit.test.tsx
  pnpm --filter @easyinterview/frontend build
  pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/home.spec.ts \
    --grep "paste-only Home matches the UI truth and captures desktop/mobile evidence" \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUTPUT_DIR/playwright-home"
  pnpm --filter @easyinterview/frontend exec playwright test \
    tests/pixel-parity/parse.spec.ts \
    --grep "processing target job response keeps the loading demo free of internal metadata|parse loading matches the UI truth at desktop and mobile" \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUTPUT_DIR/playwright-parse"
) | tee "$OUTPUT_DIR/trigger.log"

for screenshot in home-formal-viewport-desktop.png home-formal-viewport-mobile.png; do
  copy_screenshot "$OUTPUT_DIR/playwright-home" "$screenshot"
done
for screenshot in parse-loading-formal-viewport-desktop.png parse-loading-formal-viewport-mobile.png; do
  copy_screenshot "$OUTPUT_DIR/playwright-parse" "$screenshot"
done
echo 'HOME_PARSE_P0015_VIEWPORT_SCREENSHOTS_CAPTURED css=1440x900,390x844 png=1440x900,1170x2532' \
  | tee -a "$OUTPUT_DIR/trigger.log"
