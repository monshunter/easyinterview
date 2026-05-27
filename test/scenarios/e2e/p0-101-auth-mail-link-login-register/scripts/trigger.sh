#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-101-auth-mail-link-login-register"
SETUP_ENV="$OUTPUT_DIR/setup.env"
LOG="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

# shellcheck disable=SC1090
. "$SETUP_ENV"

mkdir -p "$OUTPUT_DIR"
rm -rf "$OUTPUT_DIR/playwright"
rm -f "$LOG" "$RESULT_FILE" "$OUTPUT_DIR/trigger.env"

export EI_AUTH_MAIL_LINK_FRONTEND_ORIGIN="$FRONTEND_ORIGIN"
export EI_AUTH_MAIL_LINK_API_BASE_URL="$API_BASE_URL"
export EI_AUTH_MAIL_LINK_MAILPIT_BASE_URL="$MAILPIT_BASE_URL"
export EI_AUTH_MAIL_LINK_LOGIN_EMAIL="$LOGIN_EMAIL"
export EI_AUTH_MAIL_LINK_REGISTER_EMAIL="$REGISTER_EMAIL"
export EI_PLAYWRIGHT_OUTPUT_DIR="$OUTPUT_DIR/playwright"

{
  echo "SCENARIO_RUNNER=E2E.P0.101"
  echo "RUN_ID=$RUN_ID"
  echo "PLAYWRIGHT_SPEC=frontend/tests/e2e/auth-mail-link.spec.ts"
  echo "PLAYWRIGHT_CONFIG=frontend/playwright.auth-mail-link.config.ts"
  echo "FRONTEND_ORIGIN=$FRONTEND_ORIGIN"
  echo "API_BASE_URL=$API_BASE_URL"
  echo "MAILPIT_BASE_URL=$MAILPIT_BASE_URL"
  echo "EMAILS=$LOGIN_EMAIL,$REGISTER_EMAIL"
  cd "$REPO_ROOT"
  pnpm --filter @easyinterview/frontend exec playwright test \
    --config=playwright.auth-mail-link.config.ts \
    --reporter=list \
    --workers=1 \
    auth-mail-link.spec.ts
} 2>&1 | tee "$LOG"

python3 - "$RESULT_FILE" "$RUN_ID" "$OUTPUT_DIR" <<'PY'
import json
import sys
from pathlib import Path

Path(sys.argv[1]).write_text(
    json.dumps(
        {
            "scenario_id": "E2E.P0.101",
            "suite_id": "e2e",
            "mode": "automated",
            "result": "PASS",
            "run_id": sys.argv[2],
            "output_dir": sys.argv[3],
        },
        ensure_ascii=False,
        indent=2,
    )
    + "\n",
    encoding="utf-8",
)
PY

printf 'scenario=E2E.P0.101\nrun_id=%s\nmethod=playwright-host-run-mail-link\ntrigger_at=%s\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/trigger.env"
echo "trigger: ok"
