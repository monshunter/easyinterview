#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
RUN_ID="e2e-p0-099-$(date -u '+%Y%m%dT%H%M%SZ')-$$"

mkdir -p "$OUT/screenshots"
rm -f "$OUT/evidence.md" "$OUT/result.json" "$OUT/trigger.log" "$OUT/trigger.env"
printf 'scenario=E2E.P0.099\nRUN_ID=%s\nsetup_at=%s\n' "$RUN_ID" "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" > "$OUT/setup.env"

(cd "$ROOT" && test/scenarios/env-verify.sh >/dev/null)
echo "setup: ok run_id=$RUN_ID"
