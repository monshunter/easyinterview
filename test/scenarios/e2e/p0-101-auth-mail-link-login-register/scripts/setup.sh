#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
SCENARIO_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-101-auth-mail-link-login-register"
RUN_ID="e2e-p0-101-$(date -u '+%Y%m%d%H%M%S')-$$"
LOGIN_EMAIL="auth-mail-link-login-${RUN_ID}@example.test"
REGISTER_EMAIL="auth-mail-link-register-${RUN_ID}@example.test"

mkdir -p "$OUTPUT_DIR"
rm -rf "$OUTPUT_DIR/playwright"
rm -f \
  "$OUTPUT_DIR/setup.env" \
  "$OUTPUT_DIR/setup.log" \
  "$OUTPUT_DIR/trigger.env" \
  "$OUTPUT_DIR/trigger.log" \
  "$OUTPUT_DIR/result.json" \
  "$OUTPUT_DIR/cleanup.env"

for rel_path in \
  README.md \
  data/seed-input.md \
  data/expected-outcome.md \
  scripts/setup.sh \
  scripts/trigger.sh \
  scripts/verify.sh \
  scripts/cleanup.sh; do
  test -s "$SCENARIO_DIR/$rel_path"
done

{
  echo "scenario=E2E.P0.101"
  echo "RUN_ID=$RUN_ID"
  echo "LOGIN_EMAIL=$LOGIN_EMAIL"
  echo "REGISTER_EMAIL=$REGISTER_EMAIL"
  echo "FRONTEND_ORIGIN=http://127.0.0.1:5173"
  echo "API_BASE_URL=http://127.0.0.1:8080/api/v1"
  echo "MAILPIT_BASE_URL=http://127.0.0.1:8025"
  echo "setup_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
} > "$OUTPUT_DIR/setup.env"

{
  echo "SCENARIO_RUNNER=E2E.P0.101"
  echo "RUN_ID=$RUN_ID"
  echo "frontend=http://127.0.0.1:5173"
  echo "backend=http://127.0.0.1:8080/api/v1"
  echo "mailpit=http://127.0.0.1:8025"
  curl -fsS --max-time 5 "http://127.0.0.1:5173/" >/dev/null
  curl -fsS --max-time 5 "http://127.0.0.1:8080/api/v1/runtime-config" >/dev/null
  curl -fsS --max-time 5 "http://127.0.0.1:8025/readyz" >/dev/null
  echo "setup: ok"
} 2>&1 | tee "$OUTPUT_DIR/setup.log"
