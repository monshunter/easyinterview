#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-044-practice-text-loop-assisted-happy-path"
LOG="$OUT/trigger.log"
SETUP_ENV="$OUT/setup.env"
RESULT_FILE="$OUT/result.json"
SOURCE_FINGERPRINT="$OUT/source-fingerprint.json"
VERIFY_FINGERPRINT="$OUT/source-fingerprint.verify.json"

test -s "$LOG"
test -s "$SETUP_ENV"
test -s "$SOURCE_FINGERPRINT"

# shellcheck disable=SC1090
. "$SETUP_ENV"
: "${run_id:?setup.env is missing run_id}"
: "${setup_epoch:?setup.env is missing setup_epoch}"
grep -Fq "RUN_ID=$run_id" "$LOG"

"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.044

for frontend_test in \
  PracticeScreen.test.tsx \
  PracticeI18n.test.ts \
  Transcript.test.tsx \
  usePracticeMessages.test.tsx \
  usePracticeSessionLoader.test.tsx; do
  grep -Fq "$frontend_test" "$LOG"
done

for frontend_marker in \
  'appends the user row and clears/locks the composer synchronously while the interviewer is thinking' \
  'renders persisted assistant and user text through one semantic GFM message body' \
  'preserves last same-session unresolved facts when a refresh read fails' \
  'rehydrates a server pending row, shows thinking, polls server truth, and never resends it'; do
  grep -Fq -- "$frontend_marker" "$OUT/frontend-contract.log"
done

for pass_marker in \
  '--- PASS: TestSendPracticeMessageReturnsConversationMessages' \
  '--- PASS: TestSendPracticeMessageUsesOrdinaryConversationHistory' \
  '--- PASS: TestSendPracticeMessagePendingSameIDDoesNotCallAI' \
  '--- PASS: TestSQLRepositoryGetSessionReturnsUserReplyRecoveryStateOnly' \
  '--- PASS: TestSQLRepositoryGetSessionKeepsPendingBeforeLeaseBoundary' \
  '--- PASS: TestSQLRepositoryReservePracticeMessageRetriesOnlyRetryableFailure' \
  '--- PASS: TestSQLRepositoryCommitPracticeMessageInsertsReplyAndCompletesUserAtomically'; do
  grep -Fq -- "$pass_marker" "$LOG"
done

grep -Fq 'PRACTICE_P0044_SCREENSHOT_CAPTURE_PASS viewports=1440x900,390x844 states=immediate-pending,persisted-pending,markdown-gfm' "$LOG"
grep -Fq 'PRACTICE_IMMEDIATE_PENDING_PASS user_row=immediate composer_locked=true thinking=true' "$LOG"
grep -Fq 'PRACTICE_PERSISTED_PENDING_PASS reload=true message_posts=0 lease_before_expiry=true lease_seconds=90' "$LOG"
grep -Fq 'PRACTICE_SAFE_GFM_PROJECTION_PASS roles=user,assistant semantic=true prototype_parity=true mobile_local_overflow=true document_overflow=0' "$LOG"

for browser_marker in \
  '[desktop]' \
  '[mobile]' \
  'renders one full-width chat with no structured-question surfaces' \
  'user and assistant GFM keep prototype typography with only local pre/table overflow' \
  'new user input is visible before the reply and locks the composer' \
  'reloads a persisted pending reply, keeps all actions locked, and sends zero POSTs'; do
  grep -Fq -- "$browser_marker" "$OUT/playwright.log"
done
grep -Eq '(^|[[:space:]])8 passed([[:space:]]|$)' "$OUT/playwright.log"

if grep -Eq -- '--- SKIP:|\[no tests to run\]|no tests to run|No tests found|--- FAIL:|^FAIL($|[[:space:]])|^[[:space:]]*[0-9]+ failed([[:space:]]|$)' "$LOG"; then
  echo "verify: skip, no-op, or failure marker found" >&2
  exit 1
fi

for forbidden in \
  '完整简历' \
  '完整 JD' \
  '我负责了迁移。' \
  'reply-state@example.test' \
  'ei_session=' \
  'SESSION_COOKIE_SECRET' \
  'AUTH_CHALLENGE_TOKEN_PEPPER'; do
  if grep -Fq -- "$forbidden" "$LOG"; then
    echo "verify: private value leaked into trigger log: $forbidden" >&2
    exit 1
  fi
done

python3 "$ROOT/test/scenarios/_shared/scripts/capture-source-fingerprint.py" \
  --repo-root "$ROOT" \
  --output "$VERIFY_FINGERPRINT" \
  --source-paths-from "$SOURCE_FINGERPRINT" >/dev/null
if ! cmp -s "$SOURCE_FINGERPRINT" "$VERIFY_FINGERPRINT"; then
  echo "verify: source fingerprint changed after screenshot capture" >&2
  exit 1
fi
rm -f "$VERIFY_FINGERPRINT"

python3 - "$OUT/screenshots" "$RESULT_FILE" "$run_id" "$SOURCE_FINGERPRINT" "$setup_epoch" "$LOG" <<'PY'
import hashlib
import json
import struct
import sys
from pathlib import Path

screenshot_dir = Path(sys.argv[1])
result_file = Path(sys.argv[2])
run_id = sys.argv[3]
source_fingerprint_path = Path(sys.argv[4])
setup_epoch = int(sys.argv[5])
trigger_log = Path(sys.argv[6])
source_fingerprint = json.loads(source_fingerprint_path.read_text(encoding="utf-8"))
for artifact in (source_fingerprint_path, trigger_log):
    if artifact.stat().st_mtime < setup_epoch:
        raise SystemExit(f"verify: stale evidence artifact: {artifact.name}")
expected = {
    "practice-immediate-pending-desktop.png": ("immediate-pending", [1440, 900], 1, [1440, 900]),
    "practice-immediate-pending-mobile.png": ("immediate-pending", [390, 844], 3, [1170, 2532]),
    "practice-persisted-pending-desktop.png": ("persisted-pending", [1440, 900], 1, [1440, 900]),
    "practice-persisted-pending-mobile.png": ("persisted-pending", [390, 844], 3, [1170, 2532]),
    "practice-markdown-gfm-desktop.png": ("markdown-gfm", [1440, 900], 1, [1440, 900]),
    "practice-markdown-gfm-mobile.png": ("markdown-gfm", [390, 844], 3, [1170, 2532]),
}
screenshots = []
for name, (state, css_viewport, dpr, png_size) in expected.items():
    path = screenshot_dir / name
    metadata_path = screenshot_dir / f"{path.stem}.metadata.json"
    data = path.read_bytes() if path.is_file() else b""
    if len(data) <= 10_000 or data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"verify: missing or invalid PNG evidence: {name}")
    if path.stat().st_mtime < setup_epoch:
        raise SystemExit(f"verify: stale evidence artifact: {name}")
    if not metadata_path.is_file():
        raise SystemExit(f"verify: missing screenshot metadata: {metadata_path.name}")
    if metadata_path.stat().st_mtime < setup_epoch:
        raise SystemExit(f"verify: stale evidence artifact: {metadata_path.name}")
    metadata = json.loads(metadata_path.read_text(encoding="utf-8"))
    if metadata.get("screenshot_file") != name:
        raise SystemExit(f"verify: screenshot metadata file mismatch: {name}")
    if metadata["css_viewport"] != css_viewport:
        raise SystemExit(f"verify: CSS viewport mismatch: {name}")
    if metadata["device_scale_factor"] != dpr:
        raise SystemExit(f"verify: DPR mismatch: {name}")
    width, height = struct.unpack(">II", data[16:24])
    if [width, height] != png_size:
        raise SystemExit(
            f"verify: {name} is {width}x{height}; expected {png_size[0]}x{png_size[1]}"
        )
    screenshots.append(
        {
            "state": state,
            "file": f"screenshots/{name}",
            "css_viewport": metadata["css_viewport"],
            "device_scale_factor": metadata["device_scale_factor"],
            "png_size": [width, height],
            "sha256": hashlib.sha256(data).hexdigest(),
        }
    )

payload = {
    "scenario_id": "E2E.P0.044",
    "suite_id": "e2e",
    "mode": "automated",
    "result": "PASS",
    "run_id": run_id,
    "source_fingerprint": source_fingerprint,
    "markers": [
        "PRACTICE_IMMEDIATE_PENDING_PASS",
        "PRACTICE_PERSISTED_PENDING_PASS",
        "PRACTICE_SAFE_GFM_PROJECTION_PASS",
        "PRACTICE_EVIDENCE_FINGERPRINT_PASS",
    ],
    "evidence": {
        "api_projection_test": "TestSendPracticeMessageReturnsConversationMessages",
        "repository_projection_test": "TestSQLRepositoryGetSessionReturnsUserReplyRecoveryStateOnly",
        "pending_reload_unit_test": "rehydrates a server pending row, shows thinking, polls server truth, and never resends it",
        "pending_reload_browser_test": "reloads a persisted pending reply, keeps all actions locked, and sends zero POSTs",
        "markdown_projection_test": "renders persisted assistant and user text through one semantic GFM message body",
        "markdown_browser_test": "user and assistant GFM keep prototype typography with only local pre/table overflow",
        "pending_reload_message_posts": 0,
        "screenshots": screenshots,
    },
}
result_file.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY

echo 'PRACTICE_EVIDENCE_FINGERPRINT_PASS scenario=E2E.P0.044 screenshots=6 source=current'
echo 'PRACTICE_P0044_VIEWPORT_EVIDENCE_PASS css=1440x900,390x844 png=1440x900,1170x2532'
echo 'E2E.P0.044 PASS'
