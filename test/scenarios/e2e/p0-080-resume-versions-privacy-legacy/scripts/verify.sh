#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-080-resume-versions-privacy-legacy"
LOG="$OUT/trigger.log"
mkdir -p "$OUT"

{
  echo "E2E.P0.080 verify"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  test -s "$LOG"
  if grep -E -- '--- SKIP:|\\[no tests to run\\]|no tests to run' "$LOG"; then
    echo "ERROR: skipped or no-op focused gate detected"
    exit 1
  fi
  grep -q 'RUNNER go test resume jobs privacy regression' "$LOG"
  grep -q 'TestOutboxPrivacyForTailorCompletedEvent' "$LOG"
  grep -q 'TestAiTaskRunsPrivacyForTailorDrainer' "$LOG"
  grep -q 'TestAuditPrivacyForTailorDrainer' "$LOG"
  grep -q 'RUNNER go test resume jobs tailor ready payload privacy' "$LOG"
  grep -q 'TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox' "$LOG"
  grep -q 'RUNNER go test resume store live ready-only outbox privacy' "$LOG"
  grep -q 'TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox' "$LOG"
  grep -q 'RUNNER go test cmd/api resume tailor drainer privacy' "$LOG"
  grep -q 'TestResumeTailorDrainerHTTPScenario' "$LOG"
  grep -q 'TestResumeTailorDrainerFailureScenario' "$LOG"
  grep -q 'retired_inline_rewrite_mirror=0' "$LOG"
  grep -q 'retired_mistakes_growth_drill=0' "$LOG"
  grep -q 'outbox_payload=ids_mode_status_only' "$LOG"
  grep -q 'ai_task_runs=no_prompt_or_raw_response' "$LOG"
  grep -q 'audit_metadata=no_prompt_or_response_body' "$LOG"
  grep -Eq '^PASS$' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/jobs([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/internal/resume/store([[:space:]]|$)' "$LOG"
  grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG"
  cd "$ROOT"
  if rg -n 'inline|rewrite|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/rewrite/mirror vocabulary found"
    exit 1
  fi
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  if rg -n 'PRIVATE_RESUME_SUMMARY|PRIVATE_STRUCTURED_PROFILE|PRIVATE_JD_CONTEXT|PRIVATE_TARGET_TITLE|PRIVATE_PROMPT_BODY|PRIVATE_ORIGINAL_BULLET|PRIVATE_MATCH_SUMMARY|PRIVATE_MODEL_RAW_RESPONSE|PRIVATE_SUGGESTED_BULLET|PRIVATE_SUGGESTION_REASON|raw resume text|match_summary|suggested bullet text|prompt body|model raw response' "$LOG"; then
    echo "ERROR: private resume, JD, prompt, model response, match summary, or suggestion content leaked into scenario evidence"
    exit 1
  fi
  echo "method=cmd-api-http"
  echo "privacy: outbox, ai_task_runs, and audit metadata keep payloads redacted"
  echo "legacy-negative: retired resume runtime vocabulary remains absent"
} | tee "$OUT/verify.log"
