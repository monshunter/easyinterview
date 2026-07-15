#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-report-generating-live-ui"
SETUP="$OUT/setup.env"
MANIFEST="$OUT/manifest.json"
NAVIGATION="$OUT/conversation-navigation.json"
LIVE_CAPTURE="$OUT/live-capture.json"
RESULT="$OUT/result.json"
VALIDATOR="$ROOT/test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/validate_evidence.py"
LIVE_CAPTURE_RUNNER="$ROOT/test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/capture_live_evidence.py"
MANUAL_AUDIT="$OUT/manual-visual-audit.json"
BACKEND_LOG="$ROOT/.test-output/local-dev/backend.log"
DEV_ENV="$ROOT/deploy/dev-stack/.env"
EVIDENCE_RETAINABLE=0
LIVE_SESSION_COOKIE="${P0_099_SESSION_COOKIE:-}"
unset P0_099_SESSION_COOKIE
LIVE_DATABASE_URL="${P0_099_DATABASE_URL:-}"
unset P0_099_DATABASE_URL

cleanup_untrusted_evidence() {
  local status="$?"
  local result_state
  trap - EXIT
  result_state="$(jq -r '.result // ""' "$RESULT" 2>/dev/null || true)"
  if [ "$EVIDENCE_RETAINABLE" -eq 1 ] && { [ "$result_state" = "PASS" ] || [ "$result_state" = "MANUAL_REQUIRED" ]; }; then
    if ! python3 "$VALIDATOR" --sanitize-output "$OUT" >/dev/null; then
      python3 "$VALIDATOR" --sanitize-output "$OUT" --failed >/dev/null || exit 1
      exit 1
    fi
  else
    python3 "$VALIDATOR" --sanitize-output "$OUT" --failed >/dev/null || exit 1
  fi
  exit "$status"
}
trap cleanup_untrusted_evidence EXIT

write_result() {
  local result="$1"
  local reason="$2"
  python3 - "$RESULT" "$result" "$reason" "${RUN_ID:-}" <<'PY'
import json
import sys
from pathlib import Path

Path(sys.argv[1]).write_text(json.dumps({
    "scenario_id": "E2E.P0.099",
    "suite_id": "e2e",
    "mode": "hybrid",
    "result": sys.argv[2],
    "reason": sys.argv[3],
    "run_id": sys.argv[4],
}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY
}

write_nonpass_result() {
  python3 "$VALIDATOR" --sanitize-output "$OUT" --failed >/dev/null
  write_result "$1" "$2"
}

if [ ! -s "$SETUP" ]; then
  echo "trigger: missing setup.env" >&2
  exit 1
fi
RUN_ID="$(sed -n 's/^RUN_ID=//p' "$SETUP")"
if [ -z "$RUN_ID" ]; then
  echo "trigger: setup.env has no RUN_ID" >&2
  exit 1
fi
BACKEND_LOG_START_BYTES="$(sed -n 's/^backend_log_start_bytes=//p' "$SETUP")"
if ! [[ "$BACKEND_LOG_START_BYTES" =~ ^[0-9]+$ ]]; then
  echo "trigger: setup.env has no backend_log_start_bytes" >&2
  exit 1
fi

exec > >(tee "$OUT/trigger.log") 2>&1

echo "SCENARIO_RUNNER=E2E.P0.099"
echo "SCENARIO_MODE=hybrid"
echo "RUN_ID=$RUN_ID"
API_HOST_PORT="$(sed -n 's/^API_HOST_PORT=//p' "$DEV_ENV" | head -n 1)"
API_HOST_PORT="${API_HOST_PORT:-8080}"
LIVE_API_BASE_URL="${P0_099_API_BASE_URL:-http://127.0.0.1:${API_HOST_PORT}/api/v1}"
if [ -z "$LIVE_DATABASE_URL" ]; then
  LIVE_DATABASE_URL="$(
    set -a
    # shellcheck disable=SC1090
    . "$DEV_ENV"
    printf '%s' "${DATABASE_URL:-}"
  )"
fi
set +e
P0_099_DATABASE_URL="$LIVE_DATABASE_URL" P0_099_SESSION_COOKIE="$LIVE_SESSION_COOKIE" python3 "$LIVE_CAPTURE_RUNNER" \
  --manifest "$MANIFEST" \
  --output "$LIVE_CAPTURE" \
  --run-id "$RUN_ID" \
  --api-base-url "$LIVE_API_BASE_URL" \
  --navigation "$NAVIGATION" \
  --bind-manifest
LIVE_CAPTURE_STATUS="$?"
set -e
unset LIVE_DATABASE_URL LIVE_SESSION_COOKIE

cd "$ROOT"
test/scenarios/env-verify.sh

if python3 - "$BACKEND_LOG" "$BACKEND_LOG_START_BYTES" <<'PY'
import sys
from pathlib import Path

path = Path(sys.argv[1])
offset = int(sys.argv[2])
with path.open("rb") as handle:
    handle.seek(offset)
    current_run = handle.read()
if b"AI_RAW_OUTPUT_DEBUG_BEGIN" in current_run or b"AI_RAW_OUTPUT_DEBUG_END" in current_run:
    raise SystemExit(1)
PY
then
  echo "P0_099_CURRENT_RUN_RAW_DEBUG_ABSENT_PASS"
else
  write_nonpass_result "FAIL" "current backend run emitted forbidden AI raw output markers"
  exit 0
fi

if ! python3 "$VALIDATOR" --sanitize-output "$OUT" >/dev/null; then
  write_nonpass_result "FAIL" "sensitive evidence was deleted before result evaluation"
  exit 0
fi
if [ ! -s "$MANIFEST" ]; then
  echo "MANUAL_REQUIRED exact six-image manifest is not present"
  write_nonpass_result "MANUAL_REQUIRED" "awaiting current exact six-image browser evidence"
  exit 0
fi
LIVE_CAPTURE_RESULT="$(jq -r '.result // ""' "$LIVE_CAPTURE" 2>/dev/null || true)"
LIVE_CAPTURE_REASON="$(jq -r '.reason_code // ""' "$LIVE_CAPTURE" 2>/dev/null || true)"
case "$LIVE_CAPTURE_RESULT:$LIVE_CAPTURE_STATUS" in
  PASS:0)
    ;;
  MANUAL_REQUIRED:2)
    if [ "$LIVE_CAPTURE_REASON" = "conversation_navigation_missing" ]; then
      echo "MANUAL_REQUIRED bounded report conversation navigation artifact is not present"
      write_nonpass_result "MANUAL_REQUIRED" "awaiting bounded report conversation navigation browser evidence"
    else
      echo "MANUAL_REQUIRED independent live HTTP capture unavailable reason=$LIVE_CAPTURE_REASON"
      write_nonpass_result "MANUAL_REQUIRED" "awaiting independent authenticated live HTTP capture"
    fi
    exit 0
    ;;
  *)
    write_nonpass_result "FAIL" "independent live HTTP capture failed or was malformed"
    exit 0
    ;;
esac
if python3 "$VALIDATOR" --output-dir "$OUT" --run-id "$RUN_ID" --automated-only; then
  echo "P0_099_AUTOMATED_EVIDENCE_PASS"
  echo "P0_099_LIVE_CAPTURE_BOUND_PASS"
  echo "P0_099_CONVERSATION_NAVIGATION_BOUND_PASS"
  EVIDENCE_RETAINABLE=1
  write_result "MANUAL_REQUIRED" "awaiting exact-six manual visual audit"
  echo "MANUAL_REQUIRED awaiting exact-six manual visual audit file=$MANUAL_AUDIT"
else
  write_nonpass_result "FAIL" "present screenshot evidence violated the exact-six or privacy contract"
fi

printf 'scenario=E2E.P0.099\nmethod=hybrid-real-browser\nRUN_ID=%s\ntrigger_at=%s\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/trigger.env"
echo "trigger: complete result=$(jq -r '.result' "$RESULT")"
