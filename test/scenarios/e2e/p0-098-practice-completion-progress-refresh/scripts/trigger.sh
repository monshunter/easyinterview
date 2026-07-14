#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-098-practice-completion-progress-refresh"
SETUP_ENV="$OUT/setup.env"
RESULT_FILE="$OUT/result.json"
mkdir -p "$OUT"

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

# shellcheck disable=SC1090
. "$SETUP_ENV"

export EI_P0_098_FRONTEND_ORIGIN="$FRONTEND_ORIGIN"
export EI_P0_098_API_BASE_URL="$API_BASE_URL"
export EI_P0_098_MAILPIT_BASE_URL="$MAILPIT_BASE_URL"
export EI_P0_098_AUTH_EMAIL="$AUTH_EMAIL"
export EI_P0_098_TARGET_JOB_ID="$TARGET_JOB_ID"
export EI_P0_098_ROUND_ONE_SESSION_ID="$ROUND_ONE_SESSION_ID"
export EI_PLAYWRIGHT_OUTPUT_DIR="$OUT/playwright"
rm -rf "$OUT/playwright"
rm -f "$OUT/trigger.log" "$OUT/trigger.env" "$RESULT_FILE"

{
  echo "E2E.P0.098 persisted interview round journey"
  echo "SCENARIO_RUNNER=E2E.P0.098"
  echo "E2E_TRANSPORT=host-run-real-frontend-backend"
  echo "PLAYWRIGHT_SPEC=frontend/tests/e2e/practice-progress-refresh.spec.ts"
  echo "PLAYWRIGHT_CONFIG=frontend/playwright.auth-email-code.config.ts"
  cd "$ROOT"
  pnpm --filter @easyinterview/frontend exec playwright test \
    --config=playwright.auth-email-code.config.ts \
    --reporter=list \
    --workers=1 \
    practice-progress-refresh.spec.ts
} 2>&1 | tee "$OUT/trigger.log"

python3 - "$RESULT_FILE" "$RUN_ID" "$OUT" <<'PY'
import json
import sys
from pathlib import Path

Path(sys.argv[1]).write_text(
    json.dumps(
        {
            "scenario_id": "E2E.P0.098",
            "suite_id": "e2e",
            "mode": "automated",
            "result": "PASS",
            "run_id": sys.argv[2],
            "output_dir": sys.argv[3],
            "live_browser_gate": True,
        },
        ensure_ascii=False,
        indent=2,
    )
    + "\n",
    encoding="utf-8",
)
PY

printf 'scenario=E2E.P0.098\nrun_id=%s\nmethod=host-run-real-browser-and-http\ntrigger_at=%s\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/trigger.env"
echo "trigger: ok"
