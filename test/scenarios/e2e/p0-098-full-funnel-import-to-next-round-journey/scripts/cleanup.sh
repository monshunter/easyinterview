#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-098-full-funnel-import-to-next-round-journey"
PG_DSN="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}"

mkdir -p "$OUTPUT_DIR"

psql "$PG_DSN" >/dev/null <<'SQL'
with stale_users as (
  select id from users
  where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
),
owned_resources as (
  select id from resumes where user_id in (select id from stale_users)
  union select id from target_jobs where user_id in (select id from stale_users)
  union select id from practice_plans where user_id in (select id from stale_users)
  union select id from practice_sessions where user_id in (select id from stale_users)
  union select id from feedback_reports where user_id in (select id from stale_users)
)
delete from outbox_events where aggregate_id in (select id from owned_resources);

with stale_users as (
  select id from users
  where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
),
owned_resources as (
  select id from resumes where user_id in (select id from stale_users)
  union select id from target_jobs where user_id in (select id from stale_users)
  union select id from practice_plans where user_id in (select id from stale_users)
  union select id from practice_sessions where user_id in (select id from stale_users)
  union select id from feedback_reports where user_id in (select id from stale_users)
)
delete from async_jobs where resource_id in (select id from owned_resources);

with stale_users as (
  select id from users
  where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
),
owned_resources as (
  select id from resumes where user_id in (select id from stale_users)
  union select id from target_jobs where user_id in (select id from stale_users)
  union select id from practice_plans where user_id in (select id from stale_users)
  union select id from practice_sessions where user_id in (select id from stale_users)
  union select id from feedback_reports where user_id in (select id from stale_users)
)
delete from ai_task_runs
where user_id in (select id from stale_users)
   or resource_id in (select id from owned_resources);

delete from idempotency_records
where user_id in (
  select id from users
  where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
);

delete from auth_challenges
where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
   or user_id in (
     select id from users
     where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
   );

delete from sessions
where user_id in (
  select id from users
  where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com')
);

delete from users
where email in ('full-funnel-journey@example.com', 'full-funnel-seed@example.com');
SQL

printf 'scenario=E2E.P0.098\ncleanup_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUTPUT_DIR/cleanup.env"
echo "cleanup: ok"
