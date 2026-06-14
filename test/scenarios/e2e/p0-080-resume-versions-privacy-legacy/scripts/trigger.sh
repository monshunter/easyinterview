#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.." && pwd)"
OUT="$ROOT/.test-output/e2e/p0-080-resume-versions-privacy-legacy"
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
  DATABASE_URL="${DATABASE_URL:-postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable}" go test ./internal/resume/store -tags=integration -run TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox -count=1 -v
  echo "RUNNER go test cmd/api resume tailor drainer privacy"
  go test ./cmd/api -run 'TestResumeTailorDrainerHTTPScenario|TestResumeTailorDrainerFailureScenario' -count=1 -v
  cd "$ROOT"
  echo "RUNNER rg retired inline rewrite mirror"
  if rg -n 'inline|mirror' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired inline/mirror vocabulary found"
    exit 1
  fi
  echo "evidence retired_inline_mirror=0"
  echo "RUNNER rg retired mistakes growth drill"
  if rg -n 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume --glob '!**/verify.sh'; then
    echo "ERROR: retired mistakes/growth/drill vocabulary found"
    exit 1
  fi
  echo "evidence retired_mistakes_growth_drill=0"
  echo "evidence outbox_payload=ids_mode_status_only"
  echo "evidence ai_task_runs=no_prompt_or_raw_response"
  echo "evidence audit_metadata=no_prompt_or_response_body"
} | tee "$OUT/trigger.log"
