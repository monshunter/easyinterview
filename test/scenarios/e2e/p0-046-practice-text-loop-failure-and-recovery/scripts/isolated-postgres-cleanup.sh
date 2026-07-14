#!/usr/bin/env bash
# Shared by trigger.sh and cleanup.sh so interrupted runs use the same
# run-bound database identity and residual check.

practice_database_name_for_run_id() {
  local run_id="$1"
  local compact
  if [[ ! "$run_id" =~ ^[0-9a-fA-F-]+$ ]]; then
    echo "practice cleanup: invalid run_id" >&2
    return 1
  fi
  compact="${run_id//-/}"
  if [[ ! "$compact" =~ ^[0-9a-fA-F]{32}$ ]]; then
    echo "practice cleanup: run_id must be a UUID" >&2
    return 1
  fi
  printf 'ei_p0046_%s\n' "${compact,,}"
}

practice_admin_database_url() {
  local source_database_url="$1"
  python3 - "$source_database_url" postgres <<'PY'
import sys
from urllib.parse import quote, urlsplit, urlunsplit

source = urlsplit(sys.argv[1])
if source.scheme not in {"postgres", "postgresql"} or not source.netloc:
    raise SystemExit("practice cleanup: DATABASE_URL must be a postgres:// or postgresql:// URL")
database = quote(sys.argv[2], safe="")
print(urlunsplit((source.scheme, source.netloc, f"/{database}", source.query, "")))
PY
}

practice_write_database_state() {
  local state_file="$1"
  local run_id="$2"
  local database_name="$3"
  local residual="$4"
  local temporary="${state_file}.tmp"
  umask 077
  printf 'run_id=%s\nisolated_database_name=%s\nresidual=%s\n' \
    "$run_id" "$database_name" "$residual" > "$temporary"
  mv "$temporary" "$state_file"
}

practice_database_residual() {
  local admin_database_url="$1"
  local database_name="$2"
  if [[ ! "$database_name" =~ ^ei_p0046_[0-9a-f]{32}$ ]]; then
    echo "practice cleanup: invalid isolated database name" >&2
    return 1
  fi
  PGCONNECT_TIMEOUT=5 psql \
    --dbname="$admin_database_url" \
    --no-align \
    --tuples-only \
    --set=ON_ERROR_STOP=1 \
    --command="SELECT count(*) FROM pg_database WHERE datname = '$database_name';"
}

practice_cleanup_isolated_database() {
  local admin_database_url="$1"
  local database_name="$2"
  local state_file="$3"
  local run_id="$4"
  local log_file="$5"
  local attempt
  local residual

  for attempt in 1 2 3; do
    if PGCONNECT_TIMEOUT=5 dropdb --if-exists --force \
      --maintenance-db="$admin_database_url" "$database_name" >/dev/null 2>&1; then
      if residual="$(practice_database_residual "$admin_database_url" "$database_name" 2>/dev/null)"; then
        residual="${residual//[[:space:]]/}"
        if [ "$residual" = "0" ]; then
          practice_write_database_state "$state_file" "$run_id" "$database_name" 0
          echo "PRACTICE_ISOLATED_POSTGRES_CLEANUP_PASS residual=0 attempts=$attempt database=$database_name" \
            | tee -a "$log_file"
          return 0
        fi
      fi
    fi
    practice_write_database_state "$state_file" "$run_id" "$database_name" 1
    echo "PRACTICE_ISOLATED_POSTGRES_CLEANUP_RETRY residual=1 attempt=$attempt database=$database_name" \
      | tee -a "$log_file" >&2
  done

  echo "PRACTICE_ISOLATED_POSTGRES_CLEANUP_FAIL residual=1 attempts=3 database=$database_name" \
    | tee -a "$log_file" >&2
  return 1
}
