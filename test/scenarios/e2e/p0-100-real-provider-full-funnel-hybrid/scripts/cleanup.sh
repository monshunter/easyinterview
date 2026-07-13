#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
VALIDATOR="$SCRIPT_DIR/validate_reliability.py"
mkdir -p "$OUTPUT_DIR"

RESULT_STATE="$(jq -r '.result // ""' "$OUTPUT_DIR/result.json" 2>/dev/null || true)"
RUN_ID="$(jq -r '.run_id // ""' "$OUTPUT_DIR/result.json" 2>/dev/null || true)"
SANITIZE_STAGE="failed"
if [ "$RESULT_STATE" = "PASS" ]; then
  SANITIZE_STAGE="pass"
fi
PRIVACY_CLEAN=1
python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage "$SANITIZE_STAGE" >/dev/null || PRIVACY_CLEAN=0
if [ "$PRIVACY_CLEAN" -eq 1 ] && [ "$RESULT_STATE" = "PASS" ] && python3 "$VALIDATOR" \
  --manifest "$OUTPUT_DIR/reliability-manifest.json" \
  --agent-audit "$OUTPUT_DIR/independent-agent-audit.json" \
  --run-id "$RUN_ID" >/dev/null; then
  python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage pass >/dev/null
  EVIDENCE_RETENTION="retained"
else
  python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage failed >/dev/null
  EVIDENCE_RETENTION="deleted"
fi

{
  echo "scenario=E2E.P0.100"
  echo "mode=hybrid-real-provider-reliability"
  echo "cleanup_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "shared_environment_cleanup=not_run_by_scenario_cleanup"
  echo "raw_temp_artifacts=absent"
  echo "redacted_manifest=$EVIDENCE_RETENTION"
} > "$OUTPUT_DIR/cleanup.env"

echo "cleanup: shared environment unchanged; redacted reliability evidence=$EVIDENCE_RETENTION"
