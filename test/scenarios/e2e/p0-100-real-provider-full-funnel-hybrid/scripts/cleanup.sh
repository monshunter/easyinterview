#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-100-real-provider-full-funnel-hybrid"

mkdir -p "$OUTPUT_DIR"

{
  echo "scenario=E2E.P0.100"
  echo "mode=hybrid"
  echo "cleanup_at=$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  echo "cleanup_contract=data/account.md"
  echo "shared_environment_cleanup=not_run_by_scenario_cleanup"
} > "$OUTPUT_DIR/cleanup.env"

if [ -n "${P0_100_SESSION_COOKIE_VALUE:-}" ]; then
  curl -fsS -X DELETE "http://127.0.0.1:8080/api/v1/me" \
    -H "Cookie: ei_session=${P0_100_SESSION_COOKIE_VALUE}" \
    -H "Idempotency-Key: manual-uat-full-funnel-cleanup-$(date +%Y%m%d%H%M%S)" \
    > "$OUTPUT_DIR/cleanup-response.json"
  echo "cleanup: product privacy delete requested"
else
  echo "cleanup: manual cleanup required via data/account.md when real browser session exists"
fi
