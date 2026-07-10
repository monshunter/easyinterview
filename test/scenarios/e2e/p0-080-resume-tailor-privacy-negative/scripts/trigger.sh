#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-080-resume-tailor-privacy-negative"
mkdir -p "$OUT"

{
  echo "E2E.P0.080 trigger"
  date -u '+timestamp=%Y-%m-%dT%H:%M:%SZ'
  cd "$ROOT/backend"
  echo "RUNNER go test resume jobs privacy regression"
  go test ./internal/resume/jobs -run 'TestOutboxPrivacy|TestAuditPrivacy|TestAiTaskRunsPrivacy' -count=1 -v
  echo "RUNNER go test resume jobs tailor ready payload privacy"
  go test ./internal/resume/jobs -run TestTailorHandlerHappyPathWritesReadySuggestionsTaskRunAndPrivateOutbox -count=1 -v
  echo "RUNNER go test resume store live ready-only outbox privacy"
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestCompleteTailorRunSuccessWritesResultAndOutbox -count=1 -v
  echo "RUNNER go test cmd/api resume tailor runner kernel privacy"
  go test ./cmd/api -run 'TestResumeTailorRunnerHTTPScenario|TestResumeTailorRunnerFailureScenario' -count=1 -v
  cd "$ROOT"
  echo "RUNNER rg out-of-scope inline rewrite mirror"
  echo "RUNNER rg out-of-scope mistakes growth drill"
  "$ROOT/test/scenarios/_shared/scripts/resume-runtime-negative-gate.sh"
  echo "evidence out_of_scope_inline_rewrite_mirror=0"
  echo "evidence out_of_scope_mistakes_growth_drill=0"
  echo "evidence outbox_payload=ids_mode_status_only"
  echo "evidence ai_task_runs=no_prompt_or_raw_response"
  echo "evidence audit_metadata=no_prompt_or_response_body"
} | tee "$OUT/trigger.log"
