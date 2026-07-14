#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
REPO_ROOT="$ROOT"
OUT="$ROOT/.test-output/e2e/p0-098-practice-completion-progress-refresh"
SCENARIO_DIR="$ROOT/test/scenarios/e2e/p0-098-practice-completion-progress-refresh"
RUN_ID="e2e-p0-098-$(date -u '+%Y%m%d%H%M%S')-$$"
AUTH_EMAIL="p0-098-live-round-refresh@example.test"
USER_ID="019f6098-0000-7000-8000-000000000001"
RESUME_ID="019f6098-0000-7000-8000-000000000002"
TARGET_JOB_ID="019f6098-0000-7000-8000-000000000003"
ROUND_ONE_SESSION_ID="019f6098-0000-7000-8000-000000000020"

# shellcheck disable=SC1090
. "$ROOT/test/scenarios/_shared/scripts/local-dev-runtime.sh"
load_dev_stack_env

PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"
FRONTEND_ORIGIN="http://127.0.0.1:$(frontend_port)"
API_BASE_URL="http://127.0.0.1:$(api_port)/api/v1"
MAILPIT_BASE_URL="http://127.0.0.1:$(mailpit_web_port)"

if [ "${VITE_EI_API_MODE:-}" != "real" ]; then
  echo "setup: E2E.P0.098 requires VITE_EI_API_MODE=real" >&2
  exit 1
fi
if [ "${VITE_EI_API_BASE_URL:-}" != "$API_BASE_URL" ]; then
  echo "setup: E2E.P0.098 requires the frontend API base to match the live backend" >&2
  exit 1
fi

mkdir -p "$OUT"
rm -rf "$OUT/playwright"
rm -f \
  "$OUT/setup.env" \
  "$OUT/setup.log" \
  "$OUT/trigger.log" \
  "$OUT/trigger.env" \
  "$OUT/result.json" \
  "$OUT/cleanup.env"

for rel_path in \
  README.md \
  data/seed-input.md \
  data/expected-outcome.md \
  data/live-round-refresh-seed.sql \
  data/live-round-refresh-cleanup.sql \
  scripts/setup.sh \
  scripts/trigger.sh \
  scripts/verify.sh \
  scripts/cleanup.sh; do
  test -s "$SCENARIO_DIR/$rel_path"
done

ROUND_COLUMN_COUNT="$(psql "$PG_DSN" -tAc "
select count(*)
from information_schema.columns
where table_schema = 'public'
  and table_name = 'practice_plans'
  and column_name in ('round_id', 'round_sequence');
")"
if [ "$ROUND_COLUMN_COUNT" != "2" ]; then
  echo "setup: migration 000017 round identity is not applied" >&2
  exit 1
fi

{
  echo "SCENARIO_RUNNER=E2E.P0.098"
  echo "RUN_ID=$RUN_ID"
  psql "$PG_DSN" -v ON_ERROR_STOP=1 -q \
    -f "$SCENARIO_DIR/data/live-round-refresh-cleanup.sql"
  psql "$PG_DSN" -v ON_ERROR_STOP=1 -q \
    -f "$SCENARIO_DIR/data/live-round-refresh-seed.sql"
  test "$(psql "$PG_DSN" -tAc "select count(*) from users where id='$USER_ID'")" = "1"
  test "$(psql "$PG_DSN" -tAc "select count(*) from practice_sessions where id='$ROUND_ONE_SESSION_ID' and status='waiting_user_input'")" = "1"
  test "$(psql "$PG_DSN" -tAc "select count(*) from practice_messages u join practice_messages a on a.reply_to_message_id=u.id and a.role='assistant' where u.session_id='$ROUND_ONE_SESSION_ID' and u.role='user' and u.reply_status='complete'")" = "1"
  curl -fsS --max-time 5 "$FRONTEND_ORIGIN/" >/dev/null
  curl -fsS --max-time 5 "$API_BASE_URL/runtime-config" >/dev/null
  curl -fsS --max-time 5 "$MAILPIT_BASE_URL/readyz" >/dev/null
  echo "VITE_EI_API_MODE=real"
  echo "VITE_EI_API_BASE_URL=$VITE_EI_API_BASE_URL"
  echo "live_seed_user=$USER_ID"
  echo "live_seed_target=$TARGET_JOB_ID"
  echo "live_seed_session=$ROUND_ONE_SESSION_ID"
  echo "live_seed_answered_turns=1"
  echo "setup: ok"
} 2>&1 | tee "$OUT/setup.log"

{
  echo "scenario=E2E.P0.098"
  echo "RUN_ID=$RUN_ID"
  echo "AUTH_EMAIL=$AUTH_EMAIL"
  echo "USER_ID=$USER_ID"
  echo "RESUME_ID=$RESUME_ID"
  echo "TARGET_JOB_ID=$TARGET_JOB_ID"
  echo "ROUND_ONE_SESSION_ID=$ROUND_ONE_SESSION_ID"
  echo "FRONTEND_ORIGIN=$FRONTEND_ORIGIN"
  echo "API_BASE_URL=$API_BASE_URL"
  echo "MAILPIT_BASE_URL=$MAILPIT_BASE_URL"
  echo "setup_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
} > "$OUT/setup.env"
