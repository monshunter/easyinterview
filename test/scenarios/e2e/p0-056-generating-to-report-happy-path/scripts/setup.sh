#!/usr/bin/env bash
set -euo pipefail

ROOT="$(git -C "$(dirname "${BASH_SOURCE[0]}")" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-056-generating-to-report-happy-path"
OWNER_EVIDENCE="$ROOT/.test-output/e2e/p0-047-practice-text-loop-complete-and-generating-handoff/completion-backend-evidence.json"

mkdir -p "$OUT"
rm -f "$OUT"/*.log "$OUT/setup.env" "$OUT/backend-evidence.json"
test -s "$OWNER_EVIDENCE" || {
  echo "E2E.P0.056: run E2E.P0.047 first; completion owner evidence is missing" >&2
  exit 1
}

jq -e '
  keys == ["command","database","markers","result","scenarioId","schemaVersion","tests"] and
  .schemaVersion == "practice-completion-evidence.v1" and
  .scenarioId == "E2E.P0.047" and
  .result == "PASS" and
  (.tests | length) == 3 and
  (.tests | all(.status == "PASS")) and
  (.markers | sort) == [
    "REPORT_CONTEXT_REPLAY_PASS",
    "REPORT_CONTEXT_SNAPSHOT_PASS",
    "ZERO_ANSWER_COMPLETION_REJECTED_PASS"
  ] and
  .database.snapshotSchemaVersion == "report-context.v1" and
  .database.snapshotReplayEqual == true and
  .database.mismatchSideEffectCount == 0
' "$OWNER_EVIDENCE" >/dev/null

RUN_CORRELATION_ID="$(python3 -c 'import uuid; print(uuid.uuid4())')"
printf 'scenario=E2E.P0.056\nsetup_at=%s\nrun_correlation_id=%s\nconsumed_owner_schema=practice-completion-evidence.v1\n' \
  "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$RUN_CORRELATION_ID" > "$OUT/setup.env"
