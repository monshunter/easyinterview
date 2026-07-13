#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
DEV_STACK_ENV="$REPO_ROOT/deploy/dev-stack/.env"
SETUP_ENV="$OUTPUT_DIR/setup.env"
MANIFEST="$OUTPUT_DIR/reliability-manifest.json"
AGENT_AUDIT="$OUTPUT_DIR/independent-agent-audit.json"
RESULT_FILE="$OUTPUT_DIR/result.json"
EVALKIT="$REPO_ROOT/backend/bin/evalkit"

cleanup_untrusted_evidence() {
  local status="$?"
  local result_state
  trap - EXIT
  result_state="$(jq -r '.result // ""' "$RESULT_FILE" 2>/dev/null || true)"
  if [ "${EVIDENCE_VALIDATED:-0}" -eq 1 ] && [ "$result_state" = "PASS" ]; then
    if ! python3 "$SCRIPT_DIR/validate_reliability.py" --sanitize-output "$OUTPUT_DIR" --sanitize-stage pass >/dev/null; then
      python3 "$SCRIPT_DIR/validate_reliability.py" --sanitize-output "$OUTPUT_DIR" --sanitize-stage failed >/dev/null || exit 1
      exit 1
    fi
  else
    python3 "$SCRIPT_DIR/validate_reliability.py" --sanitize-output "$OUTPUT_DIR" --sanitize-stage failed >/dev/null || exit 1
  fi
  exit "$status"
}

write_result() {
  local result="$1"
  local reason="$2"
  python3 - "$RESULT_FILE" "$result" "$reason" "${RUN_ID:-}" <<'PY'
import json
import sys
from pathlib import Path

Path(sys.argv[1]).write_text(json.dumps({
    "scenario_id": "E2E.P0.100",
    "suite_id": "e2e",
    "mode": "hybrid-real-provider-reliability",
    "result": sys.argv[2],
    "reason": sys.argv[3],
    "run_id": sys.argv[4],
}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY
}

write_nonpass_result() {
  python3 "$SCRIPT_DIR/validate_reliability.py" --sanitize-output "$OUTPUT_DIR" --sanitize-stage failed >/dev/null
  write_result "$1" "$2"
}

has_real_provider_config() {
  [ "${APP_ENV:-}" = "dev" ] && \
    [ -n "${AI_PROVIDER_BASE_URL:-}" ] && \
    [ -n "${AI_PROVIDER_API_KEY:-}" ] && \
    [ "${AI_PROVIDER_API_KEY:-}" != "replace-me" ] && \
    [ "${AI_PROVIDER_API_KEY:-}" != "changeme" ]
}

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env" >&2
  exit 1
fi
RUN_ID="$(awk -F= '$1 == "RUN_ID" {print substr($0, index($0, "=") + 1)}' "$SETUP_ENV")"
if [ -z "$RUN_ID" ]; then
  echo "trigger: setup.env missing RUN_ID" >&2
  exit 1
fi

{
  EVIDENCE_VALIDATED=0
  trap cleanup_untrusted_evidence EXIT
  echo "SCENARIO_RUNNER=E2E.P0.100"
  echo "SCENARIO_MODE=hybrid-real-provider-reliability"
  echo "TRUST_BOUNDARY=review.BuildReportPromptMessages"
  echo "RUN_ID=$RUN_ID"
  cd "$REPO_ROOT"
  test/scenarios/env-verify.sh

  grep -F "<!-- verified:" docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/checklist.md | grep -Fq "REPORT_RUBRIC_V020_PASS"
  grep -F "<!-- verified:" docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/checklist.md | grep -Fq "REPORT_CONTEXT_AWARE_EVAL_PASS"
  grep -F "<!-- verified:" docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/checklist.md | grep -Fq "REPORT_PROMPT_V020_PASS"
  echo "P0_100_OWNER_MARKERS_PASS"

  go test -v ./backend/internal/ai/registry -run \
    'TestV020ActivationOwnerMarkersReady|TestReportGenerateConversationContractPreflight|TestReportGenerateGroundedCandidateContractPreflight' -count=1
  echo "P0_100_REGISTRY_TESTS_PASS"
  go test ./backend/cmd/evalkit ./backend/internal/eval -count=1
  echo "P0_100_EVAL_PACKAGES_PASS"
  go build -o "$EVALKIT" ./backend/cmd/evalkit
  echo "P0_100_EVALKIT_BUILD_PASS"
  "$EVALKIT" drift-check
  echo "P0_100_EVALKIT_DRIFT_PASS"
  echo "P0_100_REGISTERED_EVAL_PATH_PASS"

  if [ -s "$DEV_STACK_ENV" ]; then
    set -a
    # shellcheck disable=SC1090
    . "$DEV_STACK_ENV"
    set +a
  fi
  export AI_DEBUG_PRINT_RAW_OUTPUT=false

  if [ ! -s "$MANIFEST" ]; then
    if [ "${P0_100_RUN_LIVE:-0}" != "1" ]; then
      echo "MANUAL_REQUIRED set P0_100_RUN_LIVE=1 to authorize real provider sampling"
      write_nonpass_result "MANUAL_REQUIRED" "awaiting explicit real provider sampling opt-in"
      exit 0
    fi
    if ! has_real_provider_config; then
      echo "MANUAL_REQUIRED real provider configuration is unavailable"
      write_nonpass_result "MANUAL_REQUIRED" "real provider configuration is unavailable"
      exit 0
    fi
    if ! python3 "$SCRIPT_DIR/run_live_reliability.py" \
      --repo-root "$REPO_ROOT" \
      --evalkit "$EVALKIT" \
      --output-dir "$OUTPUT_DIR" \
      --agent-audit "$AGENT_AUDIT" \
      --run-id "$RUN_ID"; then
      write_nonpass_result "FAIL" "real provider generation or context-aware judge gate failed"
      exit 0
    fi
  fi

  if ! python3 "$SCRIPT_DIR/validate_reliability.py" --sanitize-output "$OUTPUT_DIR" --sanitize-stage pass >/dev/null; then
    write_nonpass_result "FAIL" "sensitive evidence was deleted before result evaluation"
    exit 0
  fi
  if python3 "$SCRIPT_DIR/validate_reliability.py" \
    --manifest "$MANIFEST" \
    --agent-audit "$AGENT_AUDIT" \
    --run-id "$RUN_ID"; then
    EVIDENCE_VALIDATED=1
    write_result "PASS" "five-case grounded report reliability gate passed"
  else
    write_nonpass_result "FAIL" "current reliability manifest violated threshold, isolation, causal, or privacy gates"
  fi
} 2>&1 | tee "$OUTPUT_DIR/trigger.log"

echo "trigger: complete result=$(jq -r '.result' "$RESULT_FILE")"
