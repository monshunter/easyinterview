#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery"
SETUP_ENV="$OUT/setup.env"
DATABASE_STATE="$OUT/isolated-database.env"
CLEANUP_HELPER="$SCRIPT_DIR/isolated-postgres-cleanup.sh"
SOURCE_MANIFEST="$ROOT/test/scenarios/e2e/practice-source-fingerprint-paths.json"

if [ ! -s "$SETUP_ENV" ]; then
  echo "trigger: missing setup.env; run scripts/setup.sh first" >&2
  exit 1
fi

# shellcheck disable=SC1090
. "$SETUP_ENV"
: "${run_id:?setup.env is missing run_id}"
# shellcheck disable=SC1090
. "$CLEANUP_HELPER"

if [ -z "${DATABASE_URL:-}" ] && [ -f "$ROOT/deploy/dev-stack/.env" ]; then
  DATABASE_URL="$(sed -n 's/^DATABASE_URL=//p' "$ROOT/deploy/dev-stack/.env" | head -n 1)"
fi
: "${DATABASE_URL:?DATABASE_URL must supply PostgreSQL server credentials}"

for command_name in python3 createdb dropdb psql; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "trigger: required command is unavailable: $command_name" >&2
    exit 1
  fi
done

SOURCE_DATABASE_URL="$DATABASE_URL"
ISOLATED_DATABASE_NAME="$(practice_database_name_for_run_id "$run_id")"
ADMIN_DATABASE_URL="$(practice_admin_database_url "$SOURCE_DATABASE_URL")"
ISOLATED_DATABASE_URL="$(python3 - "$SOURCE_DATABASE_URL" "$ISOLATED_DATABASE_NAME" <<'PY'
import sys
from urllib.parse import quote, urlsplit, urlunsplit

source = urlsplit(sys.argv[1])
if source.scheme not in {"postgres", "postgresql"} or not source.netloc:
    raise SystemExit("trigger: DATABASE_URL must be a postgres:// or postgresql:// URL")
database = quote(sys.argv[2], safe="")
print(urlunsplit((source.scheme, source.netloc, f"/{database}", source.query, "")))
PY
)"

rm -rf "$OUT/playwright" "$OUT/screenshots"
mkdir -p "$OUT/screenshots"
rm -f "$OUT"/*.log "$OUT/result.json" "$OUT/source-fingerprint.json" "$OUT/source-fingerprint.verify.json"

printf 'SCENARIO_RUNNER=E2E.P0.046\nRUN_ID=%s\n' "$run_id" \
  | tee "$OUT/scenario-start.log" \
  | tee "$OUT/trigger.log"

if [ -s "$DATABASE_STATE" ]; then
  state_run_id="$(sed -n 's/^run_id=//p' "$DATABASE_STATE" | head -n 1)"
  state_database_name="$(sed -n 's/^isolated_database_name=//p' "$DATABASE_STATE" | head -n 1)"
  if [ "$state_run_id" != "$run_id" ] || [ "$state_database_name" != "$ISOLATED_DATABASE_NAME" ]; then
    echo "trigger: persisted isolated database identity does not match setup.env" >&2
    exit 1
  fi
  practice_cleanup_isolated_database \
    "$ADMIN_DATABASE_URL" "$ISOLATED_DATABASE_NAME" "$DATABASE_STATE" "$run_id" "$OUT/trigger.log"
fi
practice_write_database_state "$DATABASE_STATE" "$run_id" "$ISOLATED_DATABASE_NAME" 1

cleanup_on_exit() {
  local status=$?
  trap - EXIT INT TERM
  if ! practice_cleanup_isolated_database \
    "$ADMIN_DATABASE_URL" "$ISOLATED_DATABASE_NAME" "$DATABASE_STATE" "$run_id" "$OUT/trigger.log"; then
    if [ "$status" -eq 0 ]; then
      status=1
    fi
  fi
  exit "$status"
}

trap cleanup_on_exit EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

PGCONNECT_TIMEOUT=5 createdb \
  --maintenance-db="$ADMIN_DATABASE_URL" "$ISOLATED_DATABASE_NAME"
export DATABASE_URL="$ISOLATED_DATABASE_URL"

python3 "$ROOT/test/scenarios/_shared/scripts/capture-source-fingerprint.py" \
  --repo-root "$ROOT" \
  --output "$OUT/source-fingerprint.json" \
  --source-paths-from "$SOURCE_MANIFEST"

copy_screenshot() {
  local name="$1"
  local source
  local metadata_name="${name%.png}.metadata.json"
  local metadata_source
  source="$(find "$OUT/playwright" -type f -name "$name" -print -quit)"
  if [ -z "$source" ]; then
    echo "trigger: missing Playwright screenshot $name" >&2
    return 1
  fi
  metadata_source="$(find "$OUT/playwright" -type f -name "$metadata_name" -print -quit)"
  if [ -z "$metadata_source" ]; then
    echo "trigger: missing Playwright screenshot metadata $metadata_name" >&2
    return 1
  fi
  cp "$source" "$OUT/screenshots/$name"
  cp "$metadata_source" "$OUT/screenshots/$metadata_name"
}

(
  cd "$ROOT/backend"
  go run ./cmd/migrate \
    --migrations-dir "$ROOT/migrations" \
    --backfill-manifest "$ROOT/migrations/backfill/manifest.yaml" \
    up
  go run ./cmd/migrate \
    --migrations-dir "$ROOT/migrations" \
    --backfill-manifest "$ROOT/migrations/backfill/manifest.yaml" \
    status
) 2>&1 | tee "$OUT/migrations.log" | tee -a "$OUT/trigger.log"
echo 'PRACTICE_ISOLATED_POSTGRES_MIGRATIONS_PASS' \
  | tee "$OUT/migrations-finish.log" \
  | tee -a "$OUT/trigger.log"

"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-gate.sh" "$ROOT" \
  2>&1 | tee "$OUT/real-backend.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/frontend"
  pnpm exec vitest run \
    src/app/screens/practice/PracticeScreen.test.tsx \
    src/app/screens/practice/PracticeI18n.test.ts \
    src/app/screens/practice/hooks/usePracticeMessages.test.tsx \
    src/app/screens/practice/hooks/usePracticeSessionLoader.test.tsx \
    src/app/screens/practice/hooks/useCompletePracticeSession.test.tsx \
    --reporter=verbose
) 2>&1 | tee "$OUT/frontend-contract.log" | tee -a "$OUT/trigger.log"
echo 'PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS timeout_ms=95000 same_id=true stale_read_directions=2 missing_id_fail_locked=true read_failure_fail_locked=true' \
  | tee "$OUT/frontend-recovery-finish.log" \
  | tee -a "$OUT/trigger.log"

(
  cd "$ROOT"
  pnpm --filter @easyinterview/frontend build
) 2>&1 | tee "$OUT/frontend-build.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/backend"
  go test ./internal/api/practice ./internal/practice ./internal/store/practice \
    -run '^(TestSendPracticeMessageMapsConflictAndIsolationErrors|TestSendPracticeMessageProviderFailureKeepsReservationUncommitted|TestSendPracticeMessagePersistsRetryableFailureWithDetachedBoundedContext|TestSendPracticeMessagePersistsTerminalFailure|TestSendPracticeMessageCommitFailurePersistsRetryableStateWithDetachedBoundedContext|TestSendPracticeMessageCommitFailureReturnsFinalizationError|TestSendPracticeMessageFailsClosedWithoutResumeContextAndSkipsAI|TestSendPracticeMessageExactReplayReturnsOriginalResultWithoutAICall|TestSendPracticeMessageMapsClientMismatchAndCrossUserAccess|TestSendPracticeMessagePendingSameIDDoesNotCallAI|TestSQLRepositoryGetSessionReturnsUserReplyRecoveryStateOnly|TestSQLRepositoryReservePracticeMessageRetriesOnlyRetryableFailure|TestSQLRepositoryReservePracticeMessageRejectsPendingAndTerminalSameID|TestSQLRepositoryReservePracticeMessageRejectsNewMessageWhileReplyPending|TestSQLRepositoryFailPracticeMessageTransitionsPendingAtomically)$' \
    -count=1 -v
) 2>&1 | tee "$OUT/failure-recovery-contract.log" | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/backend"
  go test -tags integration ./internal/store/practice \
    -run '^(TestIntegrationPracticeReplyStateRecovery|TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce|TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce|TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration|TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery)$' \
    -count=1 -v
) 2>&1 | tee "$OUT/failure-recovery-postgresql.log" | tee -a "$OUT/trigger.log"
echo 'PRACTICE_PENDING_LEASE_RECOVERY_PASS lease_seconds=90 exact_boundary=true expired_same_id_generation=2' \
  | tee "$OUT/postgresql-lease-finish.log" \
  | tee -a "$OUT/trigger.log"
echo 'PRACTICE_STALE_GENERATION_FENCED_PASS stale_generation=1 current_generation=2 stale_writes=0' \
  | tee -a "$OUT/postgresql-lease-finish.log" \
  | tee -a "$OUT/trigger.log"
echo 'PRACTICE_CONCURRENT_RESERVATION_PASS new_ids=one_winner same_id=one_winner expired_same_id=one_generation_advance' \
  | tee -a "$OUT/postgresql-lease-finish.log" \
  | tee -a "$OUT/trigger.log"

(
  cd "$ROOT/frontend"
  CI=1 pnpm exec playwright test tests/pixel-parity/practice.spec.ts \
    --grep 'retryable failure exposes one row-local retry and preserves the next draft|terminal failure has no retry escape hatch and keeps the interview locked' \
    --project=desktop \
    --project=mobile \
    --workers=1 \
    --retries=0 \
    --reporter=list \
    --output="$OUT/playwright"
) 2>&1 | tee "$OUT/playwright.log" | tee -a "$OUT/trigger.log"

for screenshot in \
  practice-retryable-failed-desktop.png \
  practice-retryable-failed-mobile.png \
  practice-terminal-failed-desktop.png \
  practice-terminal-failed-mobile.png; do
  copy_screenshot "$screenshot"
done
echo 'PRACTICE_P0046_SCREENSHOT_CAPTURE_PASS viewports=1440x900,390x844 states=retryable-failed,terminal-failed' \
  | tee "$OUT/scenario-finish.log" \
  | tee -a "$OUT/trigger.log"
echo 'PRACTICE_TERMINAL_PLAN_RECOVERY_PASS route=parse target_job_id_only=true workspace=false plan_id=false row_retry=false' \
  | tee -a "$OUT/scenario-finish.log" \
  | tee -a "$OUT/trigger.log"

practice_cleanup_isolated_database \
  "$ADMIN_DATABASE_URL" "$ISOLATED_DATABASE_NAME" "$DATABASE_STATE" "$run_id" "$OUT/trigger.log"
trap - EXIT INT TERM
