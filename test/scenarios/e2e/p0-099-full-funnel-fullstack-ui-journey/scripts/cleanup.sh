#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
mkdir -p "$OUT"
printf 'scenario=E2E.P0.099\ncleanup_at=%s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/cleanup.env"
echo "cleanup: retained shared environment and redacted evidence"
