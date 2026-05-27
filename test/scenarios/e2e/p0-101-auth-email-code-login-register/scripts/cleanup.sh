#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-101-auth-email-code-login-register"
SETUP_ENV="$OUTPUT_DIR/setup.env"
DEV_STACK_ENV="$REPO_ROOT/deploy/dev-stack/.env"

if [ -s "$SETUP_ENV" ]; then
  # shellcheck disable=SC1090
  . "$SETUP_ENV"
else
  AUTH_EMAIL=""
  RUN_ID="unknown"
fi

if [ -s "$DEV_STACK_ENV" ]; then
  set -a
  # shellcheck disable=SC1090
  . "$DEV_STACK_ENV"
  set +a
fi

PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"

if [ -n "${AUTH_EMAIL:-}" ]; then
  psql "$PG_DSN" >/dev/null <<SQL
with scenario_emails(email) as (
  values ('$AUTH_EMAIL')
),
scenario_users as (
  select id from users where email in (select email from scenario_emails)
),
scenario_challenges as (
  select id from auth_challenges where email in (select email from scenario_emails)
)
delete from audit_events
where user_id in (select id from scenario_users);

with scenario_challenges as (
  select id from auth_challenges
  where email = '$AUTH_EMAIL'
)
delete from async_jobs
where resource_type = 'auth_challenge'
  and resource_id in (select id from scenario_challenges);

delete from idempotency_records
where user_id in (
  select id from users where email = '$AUTH_EMAIL'
);

delete from sessions
where user_id in (
  select id from users where email = '$AUTH_EMAIL'
);

delete from auth_challenges
where email = '$AUTH_EMAIL';

delete from users
where email = '$AUTH_EMAIL';
SQL
fi

mkdir -p "$OUTPUT_DIR"
printf 'scenario=E2E.P0.101\nrun_id=%s\ncleanup_at=%s\nshared_environment_cleanup=not_run_by_scenario_cleanup\n' \
  "${RUN_ID:-unknown}" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/cleanup.env"
echo "cleanup: ok"
