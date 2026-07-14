#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery"
SETUP_ENV="$OUTPUT_DIR/setup.env"
DATABASE_STATE="$OUTPUT_DIR/isolated-database.env"
CLEANUP_LOG="$OUTPUT_DIR/cleanup.log"
CLEANUP_HELPER="$SCRIPT_DIR/isolated-postgres-cleanup.sh"

# shellcheck disable=SC1090
. "$CLEANUP_HELPER"

if [ ! -s "$SETUP_ENV" ] && [ ! -s "$DATABASE_STATE" ]; then
  rm -rf "$OUTPUT_DIR/playwright"
  exit 0
fi

run_id="$(sed -n 's/^run_id=//p' "$SETUP_ENV" 2>/dev/null | head -n 1 || true)"
state_run_id="$(sed -n 's/^run_id=//p' "$DATABASE_STATE" 2>/dev/null | head -n 1 || true)"
isolated_database_name="$(sed -n 's/^isolated_database_name=//p' "$DATABASE_STATE" 2>/dev/null | head -n 1 || true)"
if [ -z "$run_id" ]; then
  run_id="$state_run_id"
fi
: "${run_id:?cleanup: persisted run_id is missing}"

expected_database_name="$(practice_database_name_for_run_id "$run_id")"
if [ -n "$state_run_id" ] && [ "$state_run_id" != "$run_id" ]; then
  echo "cleanup: database state run_id does not match setup.env" >&2
  exit 1
fi
if [ -n "$isolated_database_name" ] && [ "$isolated_database_name" != "$expected_database_name" ]; then
  echo "cleanup: persisted database name does not match run_id" >&2
  exit 1
fi
isolated_database_name="$expected_database_name"
practice_write_database_state "$DATABASE_STATE" "$run_id" "$isolated_database_name" 1

if [ -z "${DATABASE_URL:-}" ] && [ -f "$REPO_ROOT/deploy/dev-stack/.env" ]; then
  DATABASE_URL="$(sed -n 's/^DATABASE_URL=//p' "$REPO_ROOT/deploy/dev-stack/.env" | head -n 1)"
fi
: "${DATABASE_URL:?cleanup: DATABASE_URL must supply PostgreSQL server credentials}"
for command_name in python3 dropdb psql; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "cleanup: required command is unavailable: $command_name" >&2
    exit 1
  fi
done

admin_database_url="$(practice_admin_database_url "$DATABASE_URL")"
practice_cleanup_isolated_database \
  "$admin_database_url" \
  "$isolated_database_name" \
  "$DATABASE_STATE" \
  "$run_id" \
  "$CLEANUP_LOG"

rm -f "$SETUP_ENV" "$DATABASE_STATE"
rm -rf "$OUTPUT_DIR/playwright"
