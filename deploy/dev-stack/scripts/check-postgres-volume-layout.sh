#!/bin/sh
# Owned by docs/spec/local-dev-stack/plans/001-bootstrap (Postgres 18 volume guard).
# Read-only preflight for incompatible PostgreSQL data layouts before dev-up.

set -eu

VOLUME_NAME="${POSTGRES_VOLUME_NAME:-easyinterview-pg-data}"
GUARD_IMAGE="${POSTGRES_VOLUME_GUARD_IMAGE:-postgres:18-alpine}"

fail() {
  printf '%s\n' "$1" >&2
  printf '%s\n' "[dev-stack] Postgres 18 expects ${VOLUME_NAME} mounted at /var/lib/postgresql with PGDATA=/var/lib/postgresql/18/docker." >&2
  printf '%s\n' "[dev-stack] Existing local data was preserved. Back it up if needed, then run DEV_RESET_FORCE=1 make dev-reset to recreate the dev volume." >&2
  exit 1
}

check_path() {
  root=$1
  if [ -s "${root}/PG_VERSION" ]; then
    fail "[dev-stack] incompatible Postgres volume '${VOLUME_NAME}': found legacy database files at volume root."
  fi
  if [ -s "${root}/data/PG_VERSION" ]; then
    fail "[dev-stack] incompatible Postgres volume '${VOLUME_NAME}': found legacy database files at /var/lib/postgresql/data."
  fi
  if [ -d "${root}/18" ] && [ ! -s "${root}/18/docker/PG_VERSION" ]; then
    fail "[dev-stack] incomplete Postgres 18 volume '${VOLUME_NAME}': found /var/lib/postgresql/18 without a valid /18/docker database."
  fi
}

if [ -n "${DEV_STACK_POSTGRES_VOLUME_PATH:-}" ]; then
  check_path "${DEV_STACK_POSTGRES_VOLUME_PATH}"
  exit 0
fi

if ! docker volume inspect "${VOLUME_NAME}" >/dev/null 2>&1; then
  exit 0
fi

docker run --rm \
  -v "${VOLUME_NAME}:/var/lib/postgresql:ro" \
  -e DEV_STACK_POSTGRES_VOLUME_PATH=/var/lib/postgresql \
  -e POSTGRES_VOLUME_NAME="${VOLUME_NAME}" \
  --entrypoint /bin/sh \
  "${GUARD_IMAGE}" \
  -c "$(sed '1,5d' "$0")"
