#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
DEV_STACK_ENV="$REPO_ROOT/deploy/dev-stack/.env"
EVIDENCE_FILE="$OUTPUT_DIR/evidence.md"
SETUP_ENV="$OUTPUT_DIR/setup.env"
RESULT_FILE="$OUTPUT_DIR/result.json"

mkdir -p "$OUTPUT_DIR"

write_result() {
  local result="$1"
  local reason="$2"
  python3 - "$RESULT_FILE" "$result" "$reason" "$OUTPUT_DIR" <<'PY'
import json
import os
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
if os.environ.get("RUN_ID"):
    payload["run_id"] = os.environ["RUN_ID"]
result_file.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY
}

scan_evidence_redline() {
  local file="$1"
  local forbidden
  for forbidden in \
    "AI_PROVIDER_API_KEY" \
    "SESSION_COOKIE_SECRET" \
    "AUTH_CHALLENGE_TOKEN_PEPPER" \
    "ei_session=" \
    "auth/email/verify\\?token=" \
    "prompt body" \
    "response body" \
    "prompt:" \
    "response:" \
    "provider response" \
    "sk-[A-Za-z0-9_-]{12,}"; do
    if grep -Eiq -- "$forbidden" "$file"; then
      echo "MANUAL_REQUIRED evidence contains forbidden marker: $forbidden"
      return 1
    fi
  done
}

{
  echo "SCENARIO_RUNNER=E2E.P0.100"
  echo "SCENARIO_MODE=hybrid"
  echo "EXECUTOR_ORDER=ai-agent-then-human"
  echo "SCENARIO_DIR=$SCENARIO_DIR"
  echo "OUTPUT_DIR=$OUTPUT_DIR"
  echo "ENV_SOURCE=deploy/dev-stack/.env"
  echo "EVIDENCE_FILE=$EVIDENCE_FILE"

  if [ ! -s "$SETUP_ENV" ]; then
    echo "MANUAL_REQUIRED missing setup.env; run scripts/setup.sh before trigger.sh"
    write_result "MANUAL_REQUIRED" "missing setup.env run marker"
    exit 0
  fi

  RUN_ID="$(awk -F= '$1 == "RUN_ID" {print substr($0, index($0, "=") + 1)}' "$SETUP_ENV")"
  if [ -z "$RUN_ID" ]; then
    echo "MANUAL_REQUIRED setup.env missing RUN_ID"
    write_result "MANUAL_REQUIRED" "missing setup run marker"
    exit 0
  fi
  export RUN_ID
  echo "RUN_ID=$RUN_ID"

  if [ ! -s "$DEV_STACK_ENV" ]; then
    echo "MANUAL_REQUIRED missing deploy/dev-stack/.env"
    echo "Run test/scenarios/env-setup.sh or copy deploy/dev-stack/.env.example to deploy/dev-stack/.env and fill local secrets."
    write_result "MANUAL_REQUIRED" "missing deploy/dev-stack/.env"
    exit 0
  fi

  set -a
  # shellcheck disable=SC1090
  . "$DEV_STACK_ENV"
  set +a

  missing=0
  for key in APP_ENV DATABASE_URL SESSION_COOKIE_SECRET AUTH_CHALLENGE_TOKEN_PEPPER EMAIL_PROVIDER AI_PROVIDER_BASE_URL AI_PROVIDER_API_KEY AI_DEBUG_PRINT_RAW_OUTPUT VITE_EI_API_MODE VITE_EI_API_BASE_URL; do
    if [ -z "${!key:-}" ]; then
      echo "MANUAL_REQUIRED missing required env: $key"
      missing=1
    fi
  done

  if [ "${APP_ENV:-}" != "dev" ]; then
    echo "MANUAL_REQUIRED APP_ENV must be dev"
    missing=1
  fi

  if [ "${EMAIL_PROVIDER:-}" != "mailpit" ]; then
    echo "MANUAL_REQUIRED EMAIL_PROVIDER must be mailpit"
    missing=1
  fi

  if [ "${VITE_EI_API_MODE:-}" != "real" ]; then
    echo "MANUAL_REQUIRED VITE_EI_API_MODE must be real"
    missing=1
  fi

  if [ "${AI_DEBUG_PRINT_RAW_OUTPUT:-}" != "true" ]; then
    echo "MANUAL_REQUIRED AI_DEBUG_PRINT_RAW_OUTPUT must be true for local real-provider debug"
    missing=1
  fi

  if [ "$missing" -ne 0 ]; then
    write_result "MANUAL_REQUIRED" "deploy/dev-stack/.env is incomplete"
    exit 0
  fi

  if [ -s "$EVIDENCE_FILE" ]; then
    for marker in "provider" "profile" "model" "task-run" "run_id"; do
      if ! grep -qi -- "$marker" "$EVIDENCE_FILE"; then
        echo "MANUAL_REQUIRED evidence.md missing marker: $marker"
        write_result "MANUAL_REQUIRED" "redacted evidence is incomplete"
        exit 0
      fi
    done
    if ! grep -Fq -- "$RUN_ID" "$EVIDENCE_FILE"; then
      echo "MANUAL_REQUIRED evidence run_id does not match setup run"
      write_result "MANUAL_REQUIRED" "redacted evidence is stale"
      exit 0
    fi
    if ! scan_evidence_redline "$EVIDENCE_FILE"; then
      write_result "MANUAL_REQUIRED" "redacted evidence contains forbidden marker"
      exit 0
    fi
    echo "PASS redacted evidence present"
    write_result "PASS" "redacted real-provider evidence present"
  else
    echo "MANUAL_REQUIRED real browser/provider journey evidence is not present yet"
    write_result "MANUAL_REQUIRED" "awaiting browser or human execution evidence"
  fi
} 2>&1 | tee "$OUTPUT_DIR/trigger.log"

echo "trigger: ok"
