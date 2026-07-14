#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-report-generating-live-ui"
LOG="$OUT/trigger.log"
RESULT="$OUT/result.json"
SETUP="$OUT/setup.env"
VALIDATOR="$ROOT/test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/validate_evidence.py"
MANUAL_AUDIT="$OUT/manual-visual-audit.json"
EVIDENCE_RETAINABLE=0

write_result() {
  local result="$1"
  local reason="$2"
  local run_id="$3"
  python3 - "$RESULT" "$result" "$reason" "$run_id" <<'PY'
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

cleanup_untrusted_evidence() {
  local status="$?"
  trap - EXIT
  if [ "$EVIDENCE_RETAINABLE" -eq 1 ]; then
    python3 "$VALIDATOR" --sanitize-output "$OUT" >/dev/null || {
      python3 "$VALIDATOR" --sanitize-output "$OUT" --failed >/dev/null || exit 1
      exit 1
    }
  else
    python3 "$VALIDATOR" --sanitize-output "$OUT" --failed >/dev/null || exit 1
  fi
  exit "$status"
}
trap cleanup_untrusted_evidence EXIT

test -s "$LOG"
test -s "$RESULT"
python3 "$VALIDATOR" --sanitize-output "$OUT" >/dev/null
for marker in "SCENARIO_RUNNER=E2E.P0.099" "SCENARIO_MODE=hybrid"; do
  grep -Fq -- "$marker" "$LOG"
done
! grep -Eq -- 'independent live HTTP capture failed|current backend run emitted forbidden AI raw output markers' "$LOG"

RESULT_STATE="$(jq -r '.result' "$RESULT")"
case "$RESULT_STATE" in
  PASS)
    RUN_ID="$(sed -n 's/^RUN_ID=//p' "$SETUP")"
    python3 "$VALIDATOR" --output-dir "$OUT" --run-id "$RUN_ID" >/dev/null
    grep -Fq -- "P0_099_SIX_SCREENSHOT_PASS" "$LOG"
    grep -Fq -- "P0_099_CURRENT_RUN_RAW_DEBUG_ABSENT_PASS" "$LOG"
    grep -Fq -- "P0_099_LIVE_CAPTURE_PASS reports=3 privacy=redacted" "$LOG"
    grep -Fq -- "P0_099_LIVE_CAPTURE_BOUND_PASS" "$LOG"
    grep -Fq -- "P0_099_MANUAL_VISUAL_AUDIT_BOUND_PASS" "$LOG"
    EVIDENCE_RETAINABLE=1
    ;;
  MANUAL_REQUIRED)
    RESULT_REASON="$(jq -r '.reason // ""' "$RESULT")"
    if [ "$RESULT_REASON" = "awaiting exact-six manual visual audit" ]; then
      RUN_ID="$(sed -n 's/^RUN_ID=//p' "$SETUP")"
      if ! python3 "$VALIDATOR" --output-dir "$OUT" --run-id "$RUN_ID" --automated-only >/dev/null; then
        write_result "FAIL" "automated evidence no longer satisfies the exact-six contract" "$RUN_ID"
        echo "verify: automated evidence no longer satisfies the exact-six contract" >&2
        exit 1
      fi
      EVIDENCE_RETAINABLE=1
      if [ ! -s "$MANUAL_AUDIT" ]; then
        echo "verify: MANUAL_REQUIRED awaiting exact-six manual visual audit"
      elif python3 "$VALIDATOR" --output-dir "$OUT" --run-id "$RUN_ID" >/dev/null; then
        echo "P0_099_MANUAL_VISUAL_AUDIT_BOUND_PASS" | tee -a "$LOG"
        write_result "PASS" "current exact six-image report evidence passed" "$RUN_ID"
        RESULT_STATE="PASS"
      else
        EVIDENCE_RETAINABLE=0
        write_result "FAIL" "manual visual audit violated the exact-six contract" "$RUN_ID"
        echo "verify: manual visual audit violated the exact-six contract" >&2
        exit 1
      fi
    else
      grep -Eq -- 'MANUAL_REQUIRED (exact six-image manifest is not present|independent live HTTP capture unavailable)' "$LOG"
    fi
    ;;
  FAIL)
    echo "verify: present P0.099 evidence failed validation" >&2
    exit 1
    ;;
  *)
    echo "verify: unsupported result $RESULT_STATE" >&2
    exit 1
    ;;
esac

echo "verify: ok result=$RESULT_STATE"
