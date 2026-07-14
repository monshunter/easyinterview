#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-098-practice-completion-progress-refresh"
LOG="$OUT/trigger.log"
RESULT_FILE="$OUT/result.json"
test -s "$LOG"

for marker in \
  "SCENARIO_RUNNER=E2E.P0.098" \
  "E2E_TRANSPORT=host-run-real-frontend-backend" \
  "PLAYWRIGHT_SPEC=frontend/tests/e2e/practice-progress-refresh.spec.ts" \
  "PLAYWRIGHT_CONFIG=frontend/playwright.auth-email-code.config.ts" \
  "practice-progress-refresh.spec.ts" \
  "E2E.P0.098 live completion API PASS completionStatus=202 persistedFact=session_completed" \
  "E2E.P0.098 workspace refresh PASS states=done,current,pending currentRound=round-2-technical currentRoundSequence=2" \
  "E2E.P0.098 home and workspace detail refresh PASS homeStates=done,current,pending detailCurrentRound=round-2-technical detailCurrentRoundSequence=2" \
  "E2E.P0.098 ready cards direct detail PASS sources=workspace,home route=/workspace?targetJobId=019f6098-0000-7000-8000-000000000003 perVisitGetTargetJob=1 importTargetJob=0 parsePoll=0" \
  "E2E.P0.098 workspace detail refresh PASS states=done,current,pending labels=已进行,即将进行,未进行 visualStyles=distinct" \
  "1 passed"; do
  grep -Fq -- "$marker" "$LOG"
done

! grep -Eq -- '^[[:space:]]*[0-9]+ failed|Error:|Timed out|no tests found' "$LOG"

if grep -Eq -- '(mailCode|code|token)=[0-9]{6}|ei_session=|SESSION_COOKIE_SECRET|AUTH_CHALLENGE_TOKEN_PEPPER' "$LOG"; then
  echo "verify: credential or raw email code leaked into trigger.log" >&2
  exit 1
fi

python3 - "$RESULT_FILE" <<'PY'
import json
import sys
from pathlib import Path

path = Path(sys.argv[1])
if not path.is_file():
    raise SystemExit("verify: missing result.json")
payload = json.loads(path.read_text(encoding="utf-8"))
expected = {
    "scenario_id": "E2E.P0.098",
    "suite_id": "e2e",
    "mode": "automated",
    "result": "PASS",
    "live_browser_gate": True,
}
for key, value in expected.items():
    if payload.get(key) != value:
        raise SystemExit(f"verify: result.json {key} mismatch")
if not payload.get("run_id"):
    raise SystemExit("verify: result.json missing run_id")
PY

test -d "$OUT/playwright"

echo "verify: ok"
