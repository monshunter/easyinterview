#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-015-jd-import-and-parse"
LOG_FILE="$OUTPUT_DIR/trigger.log"
SETUP_ENV="$OUTPUT_DIR/setup.env"
RESULT_FILE="$OUTPUT_DIR/result.json"
SOURCE_FINGERPRINT="$OUTPUT_DIR/source-fingerprint.json"
VERIFY_FINGERPRINT="$OUTPUT_DIR/source-fingerprint.verify.json"
test -s "$LOG_FILE"
test -s "$SETUP_ENV"
test -s "$SOURCE_FINGERPRINT"
# shellcheck disable=SC1090
. "$SETUP_ENV"
"$REPO_ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG_FILE" "${SCENARIO_ID:-$(basename "$OUTPUT_DIR")}" "targetJob.realApiMode.test.ts"
grep -Fq "HomeScreen.test.tsx" "$LOG_FILE"
grep -Fq "HomeLayout.test.tsx" "$LOG_FILE"
grep -Fq "HomeResumeSelection.test.tsx" "$LOG_FILE"
grep -Fq "HomeImport.test.tsx" "$LOG_FILE"
grep -Fq "HomeAuthGate.test.tsx" "$LOG_FILE"
grep -Fq "pendingImportState.test.ts" "$LOG_FILE"
grep -Fq "ParseScreen.test.tsx" "$LOG_FILE"
grep -Fq "ParseFlow.test.tsx" "$LOG_FILE"
grep -Fq "ParseFailedState.test.tsx" "$LOG_FILE"
grep -Fq "ParseEdit.test.tsx" "$LOG_FILE"
grep -Fq "tests/pixel-parity/home.spec.ts" "$LOG_FILE"
grep -Fq "paste-only Home matches the UI truth and captures desktop/mobile evidence" "$LOG_FILE"
grep -Fq "E2E.P0.014 home paste-only browser gate project=desktop viewport=1440x900 formalScreenshotBytes=" "$LOG_FILE"
grep -Fq "E2E.P0.014 home paste-only browser gate project=mobile viewport=390x844 formalScreenshotBytes=" "$LOG_FILE"
grep -Fq "tests/pixel-parity/parse.spec.ts" "$LOG_FILE"
grep -Fq "ready target job response keeps the loading demo free of internal metadata" "$LOG_FILE"
grep -Fq "parse loading matches the UI truth at desktop and mobile" "$LOG_FILE"
grep -Fq "E2E.P0.015 ready-response loading browser gate project=desktop viewport=1440x900 screenshotBytes=" "$LOG_FILE"
grep -Fq "E2E.P0.015 ready-response loading browser gate project=mobile viewport=390x844 screenshotBytes=" "$LOG_FILE"
grep -Fq "E2E.P0.015 parse loading parity browser gate project=desktop viewport=1440x900 formalScreenshotBytes=" "$LOG_FILE"
grep -Fq "E2E.P0.015 parse loading parity browser gate project=mobile viewport=390x844 formalScreenshotBytes=" "$LOG_FILE"
grep -Fq "HOME_PARSE_P0015_VIEWPORT_SCREENSHOTS_CAPTURED css=1440x900,390x844 png=1440x900,1170x2532" "$LOG_FILE"
# Privacy redline: JD raw text must not appear in trigger log
for forbidden in 'rawText:' 'raw_jd_text' 'sourceUrl:' 'console.log'; do
  if grep -Fq "$forbidden" "$LOG_FILE"; then
    echo "privacy redline violation: $forbidden" >&2
    exit 1
  fi
done

python3 "$REPO_ROOT/test/scenarios/_shared/scripts/capture-source-fingerprint.py" \
  --repo-root "$REPO_ROOT" \
  --output "$VERIFY_FINGERPRINT" \
  --source-paths-from "$SOURCE_FINGERPRINT" >/dev/null
if ! cmp -s "$SOURCE_FINGERPRINT" "$VERIFY_FINGERPRINT"; then
  echo "verify: source fingerprint changed after screenshot capture" >&2
  exit 1
fi
rm -f "$VERIFY_FINGERPRINT"

python3 - "$OUTPUT_DIR/screenshots" "$RESULT_FILE" "$run_id" "$SOURCE_FINGERPRINT" <<'PY'
import json
import struct
import sys
from pathlib import Path

screenshot_dir = Path(sys.argv[1])
result_file = Path(sys.argv[2])
run_id = sys.argv[3]
source_fingerprint = json.loads(Path(sys.argv[4]).read_text(encoding="utf-8"))
expected = {
    "home-formal-viewport-desktop.png": ("home-paste-only", [1440, 900], 1, [1440, 900]),
    "home-formal-viewport-mobile.png": ("home-paste-only", [390, 844], 3, [1170, 2532]),
    "parse-loading-formal-viewport-desktop.png": ("parse-loading", [1440, 900], 1, [1440, 900]),
    "parse-loading-formal-viewport-mobile.png": ("parse-loading", [390, 844], 3, [1170, 2532]),
}
screenshots = []
for name, (state, css_viewport, dpr, png_size) in expected.items():
    path = screenshot_dir / name
    data = path.read_bytes() if path.is_file() else b""
    if len(data) <= 10_000 or data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"verify: missing or invalid PNG evidence: {name}")
    width, height = struct.unpack(">II", data[16:24])
    if [width, height] != png_size:
        raise SystemExit(f"verify: {name} is {width}x{height}; expected {png_size}")
    screenshots.append({
        "state": state,
        "file": f"screenshots/{name}",
        "css_viewport": css_viewport,
        "device_scale_factor": dpr,
        "png_size": png_size,
    })

payload = {
    "scenario_id": "E2E.P0.015",
    "suite_id": "e2e",
    "mode": "automated",
    "result": "PASS",
    "run_id": run_id,
    "source_fingerprint": source_fingerprint,
    "evidence": {"screenshots": screenshots},
}
result_file.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY

echo 'HOME_PARSE_P0015_VIEWPORT_EVIDENCE_PASS css=1440x900,390x844 png=1440x900,1170x2532'
echo 'E2E.P0.015 PASS'

if grep -R --exclude='*.test.ts' --exclude='*.test.tsx' -E \
  'JDAssistModal|home-jd-source-controls|home-upload-trigger|home-jd-upload-trigger|target_job_attachment|source:[[:space:]]*\{[[:space:]]*type:[[:space:]]*"(url|file|manual_text)"' \
  "$REPO_ROOT/frontend/src/app/screens/home" \
  "$REPO_ROOT/ui-design/src/screen-home.jsx"; then
  echo "old JD intake surface remains in active Home source" >&2
  exit 1
fi
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
    echo "out-of-scope Home copy remains in source: $forbidden" >&2
    exit 1
  fi
done
