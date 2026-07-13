#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"
VALIDATOR="$SCRIPT_DIR/validate_reliability.py"
RUN_ID="e2e-p0-100-$(date -u '+%Y%m%dT%H%M%SZ')-$$"

mkdir -p "$(dirname "$OUTPUT_DIR")"
python3 "$VALIDATOR" --sanitize-output "$OUTPUT_DIR" --sanitize-stage setup >/dev/null
mkdir -p "$OUTPUT_DIR"

{
  echo "scenario=E2E.P0.100"
  echo "mode=hybrid-real-provider-reliability"
  echo "RUN_ID=$RUN_ID"
  echo "phase=setup"
  echo "setup_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "runner=agent-first"
  echo "env_entrypoint=test/scenarios/env-verify.sh"
  echo "manifest_file=$OUTPUT_DIR/reliability-manifest.json"
} > "$OUTPUT_DIR/setup.env"

"$REPO_ROOT/test/scenarios/env-verify.sh" 2>&1 | tee "$OUTPUT_DIR/setup.log"

for rel_path in \
  README.md \
  checklist.md \
  data/seed-input.md \
  data/expected-outcome.md \
  scripts/run_live_reliability.py \
  scripts/validate_reliability.py \
  scripts/trigger.sh \
  scripts/verify.sh \
  scripts/cleanup.sh; do
  test -s "$REPO_ROOT/test/scenarios/e2e/p0-100-real-provider-full-funnel-hybrid/$rel_path"
done

echo "setup: ok run_id=$RUN_ID"
