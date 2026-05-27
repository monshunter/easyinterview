#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
LOCAL_ENV="$OUTPUT_DIR/dev-real.env"
EVIDENCE_FILE="$OUTPUT_DIR/evidence.md"
RESULT_FILE="$OUTPUT_DIR/result.json"

mkdir -p "$OUTPUT_DIR"

write_result() {
  local result="$1"
  local reason="$2"
  python3 - "$RESULT_FILE" "$result" "$reason" "$OUTPUT_DIR" <<'PY'
import json
import sys
from pathlib import Path

result_file = Path(sys.argv[1])
payload = {
    "scenario_id": "E2E.P0.100",
    "suite_id": "e2e",
    "mode": "hybrid",
    "result": sys.argv[2],
    "reason": sys.argv[3],
    "output_dir": sys.argv[4],
}
result_file.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY
}

{
  echo "SCENARIO_RUNNER=E2E.P0.100"
  echo "SCENARIO_MODE=hybrid"
  echo "EXECUTOR_ORDER=ai-agent-then-human"
  echo "SCENARIO_DIR=$SCENARIO_DIR"
  echo "OUTPUT_DIR=$OUTPUT_DIR"
  echo "ENV_TEMPLATE=$SCENARIO_DIR/env-template/dev-real.env.example"
  echo "LOCAL_ENV=$LOCAL_ENV"
  echo "EVIDENCE_FILE=$EVIDENCE_FILE"

  if [ ! -s "$LOCAL_ENV" ]; then
    echo "MANUAL_REQUIRED missing local real-provider env file"
    echo "Copy env-template/dev-real.env.example to $LOCAL_ENV and fill local secrets."
    write_result "MANUAL_REQUIRED" "missing local real-provider env file"
    exit 0
  fi

  set -a
  # shellcheck disable=SC1090
  . "$LOCAL_ENV"
  set +a

  missing=0
  for key in SESSION_COOKIE_SECRET AUTH_CHALLENGE_TOKEN_PEPPER AI_PROVIDER_BASE_URL AI_PROVIDER_API_KEY VITE_EI_API_MODE VITE_EI_API_BASE_URL; do
    if [ -z "${!key:-}" ]; then
      echo "MANUAL_REQUIRED missing required env: $key"
      missing=1
    fi
  done

  if [ "${VITE_EI_API_MODE:-}" != "real" ]; then
    echo "MANUAL_REQUIRED VITE_EI_API_MODE must be real"
    missing=1
  fi

  if [ "$missing" -ne 0 ]; then
    write_result "MANUAL_REQUIRED" "local real-provider env is incomplete"
    exit 0
  fi

  if [ -s "$EVIDENCE_FILE" ]; then
    for marker in "provider" "profile" "model" "task-run"; do
      if ! grep -qi -- "$marker" "$EVIDENCE_FILE"; then
        echo "MANUAL_REQUIRED evidence.md missing marker: $marker"
        write_result "MANUAL_REQUIRED" "redacted evidence is incomplete"
        exit 0
      fi
    done
    echo "PASS redacted evidence present"
    write_result "PASS" "redacted real-provider evidence present"
  else
    echo "MANUAL_REQUIRED real browser/provider journey evidence is not present yet"
    write_result "MANUAL_REQUIRED" "awaiting browser or human execution evidence"
  fi
} 2>&1 | tee "$OUTPUT_DIR/trigger.log"

echo "trigger: ok"
