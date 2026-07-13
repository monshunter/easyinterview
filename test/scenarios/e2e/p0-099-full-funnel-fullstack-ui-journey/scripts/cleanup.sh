#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey"
VALIDATOR="$ROOT/test/scenarios/e2e/p0-099-full-funnel-fullstack-ui-journey/scripts/validate_evidence.py"
mkdir -p "$OUT"

PRIVACY_CLEAN=1
python3 "$VALIDATOR" --sanitize-output "$OUT" >/dev/null || PRIVACY_CLEAN=0
RESULT_STATE="$(jq -r '.result // ""' "$OUT/result.json" 2>/dev/null || true)"
RUN_ID="$(sed -n 's/^RUN_ID=//p' "$OUT/setup.env" 2>/dev/null || true)"
if [ "$PRIVACY_CLEAN" -eq 1 ] && [ "$RESULT_STATE" = "PASS" ] && python3 "$VALIDATOR" --output-dir "$OUT" --run-id "$RUN_ID" >/dev/null; then
  python3 "$VALIDATOR" --sanitize-output "$OUT" >/dev/null
  EVIDENCE_RETENTION="retained"
else
  python3 "$VALIDATOR" --sanitize-output "$OUT" --failed >/dev/null
  EVIDENCE_RETENTION="deleted"
fi

printf 'scenario=E2E.P0.099\ncleanup_at=%s\nshared_environment_cleanup=not_run_by_scenario_cleanup\nredacted_evidence=%s\n' \
  "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$EVIDENCE_RETENTION" > "$OUT/cleanup.env"
echo "cleanup: shared environment retained; redacted evidence=$EVIDENCE_RETENTION"
