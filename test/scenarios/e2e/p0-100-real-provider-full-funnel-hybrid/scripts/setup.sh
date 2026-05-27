#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"

mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR/setup.log" "$OUTPUT_DIR/trigger.log" "$OUTPUT_DIR/result.json"

{
  echo "scenario=E2E.P0.100"
  echo "mode=hybrid"
  echo "phase=setup"
  echo "setup_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "runner=agent-first"
  echo "env_entrypoint=test/scenarios/env-setup.sh --with-migrations"
} > "$OUTPUT_DIR/setup.env"

"$REPO_ROOT/test/scenarios/env-setup.sh" --with-migrations 2>&1 | tee "$OUTPUT_DIR/setup.log"

for rel_path in \
  README.md \
  checklist.md \
  env-template/dev-real.env.example \
  data/account.md \
  data/seed-input.md \
  data/expected-outcome.md \
  data/jd-backend-engineer.zh.md \
  data/jd-backend-engineer.en.md \
  data/resume-backend-engineer.zh.md \
  data/resume-backend-engineer.en.md \
  data/answer-sample-backend-engineer.zh.md \
  data/answer-sample-backend-engineer.en.md \
  data/expected-observations.md; do
  test -s "$SCENARIO_DIR/$rel_path"
done

echo "setup: ok"
