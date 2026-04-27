#!/bin/sh
# Owned by docs/spec/local-dev-stack/plans/001-bootstrap (1.2).
# Idempotent MinIO bucket bootstrap — runs inside the minio-init container.
set -eu

MINIO_ALIAS="${MINIO_ALIAS:-local}"
MINIO_ENDPOINT="${MINIO_ENDPOINT:-http://minio-dev:9000}"
MINIO_ROOT_USER="${MINIO_ROOT_USER:-dev-access-key}"
MINIO_ROOT_PASSWORD="${MINIO_ROOT_PASSWORD:-dev-secret-key}"
OBJECT_STORAGE_BUCKET="${OBJECT_STORAGE_BUCKET:-easyinterview-dev}"

echo "[minio-init] aliasing ${MINIO_ALIAS} -> ${MINIO_ENDPOINT}"
mc alias set "${MINIO_ALIAS}" "${MINIO_ENDPOINT}" "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}" >/dev/null

# Wait briefly for server readiness even though depends_on is service_healthy,
# in case the alias registration races a freshly restarted minio.
attempts=0
until mc ready "${MINIO_ALIAS}" >/dev/null 2>&1; do
  attempts=$((attempts + 1))
  if [ "${attempts}" -ge 12 ]; then
    echo "[minio-init] minio not ready after ${attempts} retries" >&2
    exit 1
  fi
  sleep 1
done

if mc ls "${MINIO_ALIAS}/${OBJECT_STORAGE_BUCKET}" >/dev/null 2>&1; then
  echo "[minio-init] bucket ${OBJECT_STORAGE_BUCKET} already present"
else
  echo "[minio-init] creating bucket ${OBJECT_STORAGE_BUCKET}"
  mc mb "${MINIO_ALIAS}/${OBJECT_STORAGE_BUCKET}"
fi

echo "[minio-init] done"
