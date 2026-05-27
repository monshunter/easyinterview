#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
LOG="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"

if [ ! -s "$LOG" ]; then
  echo "verify: missing trigger.log" >&2
  exit 1
fi

for marker in \
  "SCENARIO_RUNNER=E2E.P0.100" \
  "SCENARIO_MODE=hybrid" \
  "EXECUTOR_ORDER=ai-agent-then-human"; do
  if ! grep -q -- "$marker" "$LOG"; then
    echo "verify: missing marker $marker" >&2
    exit 1
  fi
done

for forbidden in \
  "AI_PROVIDER_API_KEY=" \
  "SESSION_COOKIE_SECRET=" \
  "AUTH_CHALLENGE_TOKEN_PEPPER=" \
  "ei_session=" \
  "auth/email/verify?token=" \
  "prompt body" \
  "response body"; do
  if grep -Fq -- "$forbidden" "$LOG"; then
    echo "verify: sensitive marker leaked into trigger log: $forbidden" >&2
    exit 1
  fi
done

python3 - "$RESULT_FILE" <<'PY'
import json
import sys
from pathlib import Path

result_file = Path(sys.argv[1])
if not result_file.is_file():
    raise SystemExit("verify: missing result.json")

payload = json.loads(result_file.read_text(encoding="utf-8"))
if payload.get("scenario_id") != "E2E.P0.100":
    raise SystemExit("verify: wrong scenario_id")
if payload.get("mode") != "hybrid":
    raise SystemExit("verify: wrong mode")
if payload.get("result") not in {"PASS", "MANUAL_REQUIRED"}:
    raise SystemExit("verify: hybrid result must be PASS or MANUAL_REQUIRED")
PY

echo "verify: ok"
