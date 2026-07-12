#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
REPO_ROOT="$ROOT"
OUT="$ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
SCENARIO_DIR="$ROOT/test/scenarios/e2e/p0-098-full-funnel-import-to-next-round-journey"
SETUP_ENV="$OUT/setup.env"

if [ -s "$SETUP_ENV" ]; then
  # shellcheck disable=SC1090
  . "$SETUP_ENV"
else
  RUN_ID="unknown"
  USER_ID="019f6098-0000-7000-8000-000000000001"
fi

# shellcheck disable=SC1090
. "$ROOT/test/scenarios/_shared/scripts/local-dev-runtime.sh"
load_dev_stack_env
PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"

psql "$PG_DSN" -v ON_ERROR_STOP=1 -q \
  -f "$SCENARIO_DIR/data/live-round-refresh-cleanup.sql"
remaining="$(psql "$PG_DSN" -tAc "
select
  (select count(*) from users where id='$USER_ID')
  + (select count(*) from auth_challenges where email='p0-098-live-round-refresh@example.test')
  + (select count(*) from resumes where id='019f6098-0000-7000-8000-000000000002')
  + (select count(*) from target_jobs where id='019f6098-0000-7000-8000-000000000003')
  + (select count(*) from practice_plans where user_id='$USER_ID')
  + (select count(*) from practice_sessions where user_id='$USER_ID')
  + (select count(*) from async_jobs where payload->>'sessionId'='019f6098-0000-7000-8000-000000000020')
  + (select count(*) from outbox_events where aggregate_id='019f6098-0000-7000-8000-000000000020')
  + (select count(*) from audit_events where actor_id='$USER_ID')
  + (select count(*) from ai_task_runs where user_id='$USER_ID');
")"
if [ "$remaining" != "0" ]; then
  echo "cleanup: live round refresh resources still exist (count=$remaining)" >&2
  exit 1
fi

mkdir -p "$OUT"
printf 'scenario=E2E.P0.098\nrun_id=%s\ncleanup_at=%s\nshared_environment_cleanup=not_run_by_scenario_cleanup\nlive_seed_remaining=0\n' \
  "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/cleanup.env"
echo "cleanup: ok"
