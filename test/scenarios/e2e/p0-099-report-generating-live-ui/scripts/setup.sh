#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-report-generating-live-ui"
RUN_ID="e2e-p0-099-$(date -u '+%Y%m%dT%H%M%SZ')-$$"
VALIDATOR="$ROOT/test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/validate_evidence.py"
DEV_ENV="$ROOT/deploy/dev-stack/.env"
BACKEND_LOG="$ROOT/.test-output/local-dev/backend.log"
BACKEND_PID="$ROOT/.test-output/local-dev/backend.pid"

python3 "$VALIDATOR" --sanitize-output "$OUT" --setup >/dev/null
mkdir -p "$OUT/screenshots"
rm -f "$OUT/manifest.json" "$OUT/live-capture.json" "$OUT/manual-visual-audit.json" "$OUT/result.json" "$OUT/trigger.log" "$OUT/trigger.env" "$OUT/evidence.md"

(cd "$ROOT" && test/scenarios/env-verify.sh >/dev/null)
test -s "$DEV_ENV"
set -a
# shellcheck disable=SC1090
. "$DEV_ENV"
set +a
if [ "${AI_DEBUG_PRINT_RAW_OUTPUT:-}" != "false" ]; then
  echo "setup: P0.099 requires AI raw debug disabled in the active dev-stack env" >&2
  exit 1
fi
test -s "$BACKEND_LOG"
test -s "$BACKEND_PID"
BACKEND_PROCESS_ID="$(cat "$BACKEND_PID")"
if ! kill -0 "$BACKEND_PROCESS_ID" >/dev/null 2>&1; then
  echo "setup: backend pid is not running" >&2
  exit 1
fi
if [ "$DEV_ENV" -nt "$BACKEND_PID" ]; then
  echo "setup: backend must be redeployed after disabling AI raw debug" >&2
  exit 1
fi
BACKEND_LOG_START_BYTES="$(wc -c < "$BACKEND_LOG" | tr -d '[:space:]')"
SETUP_AT="$(python3 - <<'PY'
from datetime import datetime, timezone

print(datetime.now(timezone.utc).isoformat(timespec="microseconds").replace("+00:00", "Z"))
PY
)"
printf 'scenario=E2E.P0.099\nRUN_ID=%s\nsetup_at=%s\nai_raw_debug=false\nbackend_pid=%s\nbackend_log_start_bytes=%s\n' \
  "$RUN_ID" "$SETUP_AT" "$BACKEND_PROCESS_ID" "$BACKEND_LOG_START_BYTES" > "$OUT/setup.env"
chmod 600 "$OUT/setup.env"
echo "setup: ok run_id=$RUN_ID"
