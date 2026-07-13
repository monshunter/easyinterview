#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
LOG="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"
MANIFEST="$OUTPUT_DIR/reliability-manifest.json"
AGENT_AUDIT="$OUTPUT_DIR/independent-agent-audit.json"
VALIDATOR="$SCRIPT_DIR/validate_reliability.py"
EVIDENCE_VALIDATED=0

cleanup_untrusted_evidence() {
  local status="$?"
  trap - EXIT
  if [ "$EVIDENCE_VALIDATED" -eq 1 ]; then
    python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage pass >/dev/null || {
      python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage failed >/dev/null || exit 1
      exit 1
    }
  else
    python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage failed >/dev/null || exit 1
  fi
  exit "$status"
}
trap cleanup_untrusted_evidence EXIT

test -s "$LOG"
test -s "$RESULT_FILE"
RESULT="$(jq -r '.result' "$RESULT_FILE")"
SANITIZE_STAGE="failed"
if [ "$RESULT" = "PASS" ]; then
  SANITIZE_STAGE="pass"
fi
if ! python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage "$SANITIZE_STAGE" >/dev/null; then
  echo "verify: forbidden raw/browser artifact persisted" >&2
  echo "verify: secret/cookie material leaked into P0.100 output" >&2
  exit 1
fi
for marker in \
  "SCENARIO_RUNNER=E2E.P0.100" \
  "SCENARIO_MODE=hybrid-real-provider-reliability" \
  "TRUST_BOUNDARY=review.BuildReportPromptMessages" \
  "P0_100_OWNER_MARKERS_PASS" \
  "P0_100_REGISTERED_EVAL_PATH_PASS"; do
  grep -Fq -- "$marker" "$LOG"
done
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run|0 tests' "$LOG"
python3 "$VALIDATOR" --runner-log "$LOG" >/dev/null

case "$RESULT" in
  PASS)
    python3 "$VALIDATOR" \
      --manifest "$MANIFEST" \
      --agent-audit "$AGENT_AUDIT" \
      --run-id "$(jq -r '.run_id' "$RESULT_FILE")" >/dev/null
    grep -Fq -- "P0_100_REPORT_RELIABILITY_PASS" "$LOG"
    test -s "$MANIFEST"
    test -s "$AGENT_AUDIT"
    python3 - "$MANIFEST" "$AGENT_AUDIT" <<'PY'
import sys
from pathlib import Path

for raw in sys.argv[1:]:
    path = Path(raw)
    if path.stat().st_mode & 0o777 != 0o600:
        raise SystemExit(f"verify: {path.name} must have mode 0600")
PY
    python3 - "$(jq -r '.run_id' "$RESULT_FILE")" <<'PY'
import sys
import tempfile
from pathlib import Path

matches = list(Path(tempfile.gettempdir()).glob(f"easyinterview-p0-100-review-{sys.argv[1]}-*"))
if matches:
    raise SystemExit("verify: current raw Agent review packet directory was not deleted")
PY
    EVIDENCE_VALIDATED=1
    ;;
  MANUAL_REQUIRED)
    grep -Fq -- "MANUAL_REQUIRED" "$LOG"
    ;;
  FAIL)
    echo "verify: P0.100 real-provider reliability failed" >&2
    exit 1
    ;;
  *)
    echo "verify: unsupported hybrid result $RESULT" >&2
    exit 1
    ;;
esac

echo "verify: ok result=$RESULT"
