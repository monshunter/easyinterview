#!/usr/bin/env bash
set -euo pipefail
umask 077

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-report-generating-live-ui"
RUN_ID="e2e-p0-099-$(date -u '+%Y%m%dT%H%M%SZ')-$$"
VALIDATOR="$ROOT/test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/validate_evidence.py"
DEV_ENV="$ROOT/deploy/dev-stack/.env"
BACKEND_PID="$ROOT/.test-output/local-dev/backend.pid"

python3 "$VALIDATOR" --sanitize-output "$OUT" --setup >/dev/null
mkdir -p "$OUT/screenshots"
rm -f "$OUT/manifest.json" "$OUT/conversation-navigation.json" "$OUT/live-capture.json" "$OUT/manual-visual-audit.json" "$OUT/result.json" "$OUT/trigger.log" "$OUT/trigger.env" "$OUT/evidence.md"

(cd "$ROOT" && test/scenarios/env-verify.sh >/dev/null)
test -s "$DEV_ENV"
set -a
# shellcheck disable=SC1090
. "$DEV_ENV"
set +a
if [ "${AI_DEBUG_CAPTURE_RAW_IO:-}" != "true" ]; then
  echo "setup: P0.099 requires AI_DEBUG_CAPTURE_RAW_IO=true in the active dev-stack env" >&2
  exit 1
fi
if [ -z "${AI_DEBUG_RAW_IO_PATH:-}" ]; then
  echo "setup: P0.099 requires a dedicated AI raw I/O path" >&2
  exit 1
fi

# Match the backend's ConfigDir-parent anchor. Resolve only after rejecting
# every existing symlink component; the target must be regular-or-absent and
# its realpath must stay outside this scenario evidence directory.
python3 - "$ROOT" "$OUT" "$AI_DEBUG_RAW_IO_PATH" <<'PY'
import os
import stat
import sys
from pathlib import Path

root = Path(sys.argv[1])
evidence = Path(sys.argv[2])
configured = Path(sys.argv[3])
config_dir = (root / "config").resolve(strict=True)
config_parent = config_dir.parent
candidate = configured if configured.is_absolute() else config_parent / configured
candidate = Path(os.path.abspath(candidate))

current = Path(candidate.anchor)
for component in candidate.parts[1:]:
    current /= component
    try:
        mode = os.lstat(current).st_mode
    except FileNotFoundError:
        continue
    if stat.S_ISLNK(mode):
        raise SystemExit("setup: AI raw path contains a symlink component")

try:
    target_mode = os.lstat(candidate).st_mode
except FileNotFoundError:
    target_mode = None
if target_mode is not None and not stat.S_ISREG(target_mode):
    raise SystemExit("setup: AI raw target must be a regular file or absent")

candidate_real = Path(os.path.realpath(candidate))
evidence_real = evidence.resolve(strict=True)
if os.path.ismount(candidate.parent):
    raise SystemExit("setup: AI raw parent must not be a filesystem or volume root")
try:
    inside_evidence = os.path.commonpath((candidate_real, evidence_real)) == str(evidence_real)
except ValueError:
    inside_evidence = False
if inside_evidence:
    raise SystemExit("setup: AI raw realpath must remain outside scenario evidence")
PY

test -s "$BACKEND_PID"
BACKEND_PROCESS_ID="$(cat "$BACKEND_PID")"
if ! kill -0 "$BACKEND_PROCESS_ID" >/dev/null 2>&1; then
  echo "setup: backend pid is not running" >&2
  exit 1
fi
if [ "$DEV_ENV" -nt "$BACKEND_PID" ]; then
  echo "setup: backend must be redeployed after changing AI raw capture config" >&2
  exit 1
fi
SETUP_AT="$(python3 - <<'PY'
from datetime import datetime, timezone

print(datetime.now(timezone.utc).isoformat(timespec="microseconds").replace("+00:00", "Z"))
PY
)"
printf 'scenario=E2E.P0.099\nRUN_ID=%s\nsetup_at=%s\nai_raw_capture=true\nbackend_pid=%s\n' \
  "$RUN_ID" "$SETUP_AT" "$BACKEND_PROCESS_ID" > "$OUT/setup.env"
chmod 600 "$OUT/setup.env"
echo "setup: ok run_id=$RUN_ID"
