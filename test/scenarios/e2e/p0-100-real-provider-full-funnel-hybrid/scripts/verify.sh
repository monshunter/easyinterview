#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
LOG="$OUTPUT_DIR/trigger.log"
EVIDENCE_FILE="$OUTPUT_DIR/evidence.md"
RESULT_FILE="$OUTPUT_DIR/result.json"

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
      echo "verify: sensitive marker leaked into evidence: $forbidden" >&2
      return 1
    fi
  done
}

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

RESULT="$(python3 - "$RESULT_FILE" <<'PY'
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
print(payload.get("result"))
PY
)"

if [ "$RESULT" = "PASS" ]; then
  if [ ! -s "$EVIDENCE_FILE" ]; then
    echo "verify: PASS requires evidence.md" >&2
    exit 1
  fi
  scan_evidence_redline "$EVIDENCE_FILE"
  if ! grep -q -- "PASS redacted evidence present" "$LOG"; then
    echo "verify: PASS requires trigger redacted evidence marker" >&2
    exit 1
  fi
fi

echo "verify: ok"
