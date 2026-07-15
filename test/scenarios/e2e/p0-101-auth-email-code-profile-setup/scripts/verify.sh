#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-101-auth-email-code-profile-setup"
SETUP_ENV="$OUTPUT_DIR/setup.env"
LOG="$OUTPUT_DIR/trigger.log"
RESULT_FILE="$OUTPUT_DIR/result.json"

if [ ! -s "$SETUP_ENV" ] || [ ! -s "$LOG" ]; then
  echo "verify: missing setup.env or trigger.log" >&2
  exit 1
fi

# shellcheck disable=SC1090
. "$SETUP_ENV"
AUTH_EMAIL_URLENCODED="$(python3 -c 'import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1], safe=""))' "$AUTH_EMAIL")"

for marker in \
  "SCENARIO_RUNNER=E2E.P0.101" \
  "E2E_TRANSPORT=host-run-real-frontend-backend" \
  "PLAYWRIGHT_SPEC=frontend/tests/e2e/auth-email-code.spec.ts" \
  "PLAYWRIGHT_CONFIG=frontend/playwright.auth-email-code.config.ts" \
  "E2E.P0.101 first-login-profile-setup email-code flow PASS" \
  "E2E.P0.101 cross-browser-relogin-profile-setup email-code flow PASS" \
  "E2E.P0.101 logout-relogin-profile-setup email-code flow PASS" \
  "E2E.P0.101 existing-email-login email-code flow PASS" \
  "E2E.P0.101 profile-required gates PASS" \
  "refresh=profile-setup" \
  "deepLink=profile-setup" \
  "crossBrowser=profile-setup" \
  "logoutRelogin=profile-setup" \
  "authStartBodyKeys=email" \
  "authRegisterLivePage=absent" \
  "topbarRegister=absent" \
  "settingsEntry=single-gear" \
  "settingsAccount=runtime-full-email" \
  "settingsLegacySurfaces=absent" \
  "settingsMountedGetMe=0" \
  "deleteMeRequests=0" \
  "E2E.P0.101 auth email-code same-email lifecycle passed" \
  "mailCode=<redacted>" \
  "email=<redacted-synthetic>" \
  "meStatus=200" \
  "profileCompletionRequired=true" \
  "profileCompletionRequired=false" \
  "consoleErrors=0" \
  "pageErrors=0" \
  "httpFailures=0" \
  "1 passed"; do
  if ! grep -Fq -- "$marker" "$LOG"; then
    echo "verify: missing marker $marker" >&2
    exit 1
  fi
done

if grep -Eq -- "0 passed|0 tests|No tests found|skipped|failed|timed out|Error:" "$LOG"; then
  echo "verify: scenario log contains skip/failure marker" >&2
  exit 1
fi

for forbidden in \
  "http://127.0.0.1:8080/api/v1/auth/email/verify" \
  "auth/verify?token=" \
  "purpose=signup" \
  "purpose=login" \
  '"purpose"' \
  '"displayName"' \
  "ei_session=" \
  "SESSION_COOKIE_SECRET" \
  "AUTH_CHALLENGE_TOKEN_PEPPER"; do
  if grep -Fq -- "$forbidden" "$LOG"; then
    echo "verify: sensitive or wrong-link marker leaked into trigger log: $forbidden" >&2
    exit 1
  fi
done

if grep -Eq -- "auth-email-code-[^[:space:]]+@example\.test" "$LOG"; then
  echo "verify: synthetic email leaked into trigger log" >&2
  exit 1
fi

for forbidden_email in "$AUTH_EMAIL" "$AUTH_EMAIL_URLENCODED"; do
  if grep -Fq -- "$forbidden_email" "$LOG"; then
    echo "verify: current-run email leaked into trigger log" >&2
    exit 1
  fi
done

if grep -Eq -- "(token|code|mailCode)=[0-9]{6}" "$LOG"; then
  echo "verify: raw email verification code leaked into trigger log" >&2
  exit 1
fi

python3 - "$RESULT_FILE" <<'PY'
import json
import sys
from pathlib import Path

result_file = Path(sys.argv[1])
if not result_file.is_file():
    raise SystemExit("verify: missing result.json")
payload = json.loads(result_file.read_text(encoding="utf-8"))
if payload.get("scenario_id") != "E2E.P0.101":
    raise SystemExit("verify: wrong scenario_id")
if payload.get("suite_id") != "e2e":
    raise SystemExit("verify: wrong suite_id")
if payload.get("mode") != "automated":
    raise SystemExit("verify: wrong mode")
if payload.get("result") != "PASS":
    raise SystemExit("verify: scenario result is not PASS")
if not payload.get("run_id"):
    raise SystemExit("verify: missing run_id")
PY

if [ ! -d "$OUTPUT_DIR/playwright" ]; then
  echo "verify: missing Playwright output directory" >&2
  exit 1
fi

echo "verify: ok"
