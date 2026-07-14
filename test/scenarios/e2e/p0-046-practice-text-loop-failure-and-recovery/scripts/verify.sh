#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUT="$ROOT/.test-output/e2e/p0-046-practice-text-loop-failure-and-recovery"
LOG="$OUT/trigger.log"
SETUP_ENV="$OUT/setup.env"
RESULT_FILE="$OUT/result.json"
SOURCE_FINGERPRINT="$OUT/source-fingerprint.json"
VERIFY_FINGERPRINT="$OUT/source-fingerprint.verify.json"
DATABASE_STATE="$OUT/isolated-database.env"

test -s "$LOG"
test -s "$SETUP_ENV"
test -s "$SOURCE_FINGERPRINT"
test -s "$DATABASE_STATE"

# shellcheck disable=SC1090
. "$SETUP_ENV"
: "${run_id:?setup.env is missing run_id}"
: "${setup_epoch:?setup.env is missing setup_epoch}"
grep -Fq "RUN_ID=$run_id" "$LOG"
expected_database_name="ei_p0046_${run_id//-/}"
grep -Fqx "run_id=$run_id" "$DATABASE_STATE"
grep -Fqx "isolated_database_name=$expected_database_name" "$DATABASE_STATE"
grep -Fqx 'residual=0' "$DATABASE_STATE"

"$ROOT/test/scenarios/_shared/scripts/frontend-real-backend-verify.sh" "$LOG" E2E.P0.046

for frontend_test in \
  PracticeScreen.test.tsx \
  PracticeI18n.test.ts \
  usePracticeMessages.test.tsx \
  usePracticeSessionLoader.test.tsx \
  useCompletePracticeSession.test.tsx; do
  grep -Fq "$frontend_test" "$LOG"
done

for frontend_marker in \
  'aborts the POST exactly at 95,000 ms' \
  'lets a later-started timeout reconcile win when an older loader refresh resolves first' \
  'lets a later-started loader refresh win when the older timeout reconcile resolves first' \
  'timeout reconciliation ends in missing-id' \
  'timeout reconciliation ends in read-failure' \
  'terminal state with one safe exact current-plan CTA'; do
  grep -Fq -- "$frontend_marker" "$OUT/frontend-contract.log"
done

for pass_marker in \
  '--- PASS: TestSendPracticeMessageMapsConflictAndIsolationErrors' \
  '--- PASS: TestSendPracticeMessageProviderFailureKeepsReservationUncommitted' \
  '--- PASS: TestSendPracticeMessagePersistsRetryableFailureWithDetachedBoundedContext' \
  '--- PASS: TestSendPracticeMessagePersistsTerminalFailure' \
  '--- PASS: TestSendPracticeMessageCommitFailurePersistsRetryableStateWithDetachedBoundedContext' \
  '--- PASS: TestSendPracticeMessageCommitFailureReturnsFinalizationError' \
  '--- PASS: TestSendPracticeMessageFailsClosedWithoutResumeContextAndSkipsAI' \
  '--- PASS: TestSendPracticeMessageExactReplayReturnsOriginalResultWithoutAICall' \
  '--- PASS: TestSendPracticeMessageMapsClientMismatchAndCrossUserAccess' \
  '--- PASS: TestSendPracticeMessagePendingSameIDDoesNotCallAI' \
  '--- PASS: TestSQLRepositoryGetSessionReturnsUserReplyRecoveryStateOnly' \
  '--- PASS: TestSQLRepositoryReservePracticeMessageRetriesOnlyRetryableFailure' \
  '--- PASS: TestSQLRepositoryReservePracticeMessageRejectsPendingAndTerminalSameID' \
  '--- PASS: TestSQLRepositoryReservePracticeMessageRejectsNewMessageWhileReplyPending' \
  '--- PASS: TestSQLRepositoryFailPracticeMessageTransitionsPendingAtomically' \
  '--- PASS: TestIntegrationPracticeReplyStateRecovery' \
  '--- PASS: TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce' \
  '--- PASS: TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce' \
  '--- PASS: TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration' \
  '--- PASS: TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery'; do
  grep -Fq -- "$pass_marker" "$LOG"
done

grep -Fq 'PRACTICE_REPLY_STATE_RECOVERY_PASS' "$LOG"
grep -Fq 'PRACTICE_PENDING_LEASE_RECOVERY_PASS lease_seconds=90 exact_boundary=true expired_same_id_generation=2' "$LOG"
grep -Fq 'PRACTICE_STALE_GENERATION_FENCED_PASS stale_generation=1 current_generation=2 stale_writes=0' "$LOG"
grep -Fq 'PRACTICE_CONCURRENT_RESERVATION_PASS new_ids=one_winner same_id=one_winner expired_same_id=one_generation_advance' "$LOG"
grep -Fq 'PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS timeout_ms=95000 same_id=true stale_read_directions=2 missing_id_fail_locked=true read_failure_fail_locked=true' "$LOG"
grep -Fq 'PRACTICE_TERMINAL_PLAN_RECOVERY_PASS route=parse target_job_id_only=true workspace=false plan_id=false row_retry=false' "$LOG"
grep -Fq 'PRACTICE_ISOLATED_POSTGRES_MIGRATIONS_PASS' "$LOG"
grep -Fq 'PRACTICE_ISOLATED_POSTGRES_CLEANUP_PASS residual=0' "$LOG"
grep -Eq '^version=[0-9]+ dirty=false$' "$OUT/migrations.log"
grep -Fq 'PRACTICE_P0046_SCREENSHOT_CAPTURE_PASS viewports=1440x900,390x844 states=retryable-failed,terminal-failed' "$LOG"

for browser_marker in \
  '[desktop]' \
  '[mobile]' \
  'retryable failure exposes one row-local retry and preserves the next draft' \
  'terminal failure has no retry escape hatch and keeps the interview locked'; do
  grep -Fq -- "$browser_marker" "$OUT/playwright.log"
done
grep -Eq '(^|[[:space:]])4 passed([[:space:]]|$)' "$OUT/playwright.log"

if grep -Eq -- '--- SKIP:|\[no tests to run\]|no tests to run|No tests found|--- FAIL:|^FAIL($|[[:space:]])|^[[:space:]]*[0-9]+ failed([[:space:]]|$)' "$LOG"; then
  echo "verify: skip, no-op, or failure marker found" >&2
  exit 1
fi

for forbidden in \
  '完整简历' \
  '完整 JD' \
  '我负责了迁移。' \
  'reply-state@example.test' \
  'ei_session=' \
  'SESSION_COOKIE_SECRET' \
  'AUTH_CHALLENGE_TOKEN_PEPPER'; do
  if grep -Fq -- "$forbidden" "$LOG"; then
    echo "verify: private value leaked into trigger log: $forbidden" >&2
    exit 1
  fi
done

python3 "$ROOT/test/scenarios/_shared/scripts/capture-source-fingerprint.py" \
  --repo-root "$ROOT" \
  --output "$VERIFY_FINGERPRINT" \
  --source-paths-from "$SOURCE_FINGERPRINT" >/dev/null
if ! cmp -s "$SOURCE_FINGERPRINT" "$VERIFY_FINGERPRINT"; then
  echo "verify: source fingerprint changed after screenshot capture" >&2
  exit 1
fi
rm -f "$VERIFY_FINGERPRINT"

python3 - "$OUT/screenshots" "$RESULT_FILE" "$run_id" "$SOURCE_FINGERPRINT" "$setup_epoch" "$LOG" <<'PY'
import hashlib
import json
import struct
import sys
from pathlib import Path

screenshot_dir = Path(sys.argv[1])
result_file = Path(sys.argv[2])
run_id = sys.argv[3]
source_fingerprint_path = Path(sys.argv[4])
setup_epoch = int(sys.argv[5])
trigger_log = Path(sys.argv[6])
source_fingerprint = json.loads(source_fingerprint_path.read_text(encoding="utf-8"))
for artifact in (source_fingerprint_path, trigger_log):
    if artifact.stat().st_mtime < setup_epoch:
        raise SystemExit(f"verify: stale evidence artifact: {artifact.name}")
expected = {
    "practice-retryable-failed-desktop.png": ("retryable-failed", [1440, 900], 1, [1440, 900]),
    "practice-retryable-failed-mobile.png": ("retryable-failed", [390, 844], 3, [1170, 2532]),
    "practice-terminal-failed-desktop.png": ("terminal-failed", [1440, 900], 1, [1440, 900]),
    "practice-terminal-failed-mobile.png": ("terminal-failed", [390, 844], 3, [1170, 2532]),
}
screenshots = []
for name, (state, css_viewport, dpr, png_size) in expected.items():
    path = screenshot_dir / name
    metadata_path = screenshot_dir / f"{path.stem}.metadata.json"
    data = path.read_bytes() if path.is_file() else b""
    if len(data) <= 10_000 or data[:8] != b"\x89PNG\r\n\x1a\n":
        raise SystemExit(f"verify: missing or invalid PNG evidence: {name}")
    if path.stat().st_mtime < setup_epoch:
        raise SystemExit(f"verify: stale evidence artifact: {name}")
    if not metadata_path.is_file():
        raise SystemExit(f"verify: missing screenshot metadata: {metadata_path.name}")
    if metadata_path.stat().st_mtime < setup_epoch:
        raise SystemExit(f"verify: stale evidence artifact: {metadata_path.name}")
    metadata = json.loads(metadata_path.read_text(encoding="utf-8"))
    if metadata.get("screenshot_file") != name:
        raise SystemExit(f"verify: screenshot metadata file mismatch: {name}")
    if metadata["css_viewport"] != css_viewport:
        raise SystemExit(f"verify: CSS viewport mismatch: {name}")
    if metadata["device_scale_factor"] != dpr:
        raise SystemExit(f"verify: DPR mismatch: {name}")
    width, height = struct.unpack(">II", data[16:24])
    if [width, height] != png_size:
        raise SystemExit(
            f"verify: {name} is {width}x{height}; expected {png_size[0]}x{png_size[1]}"
        )
    screenshots.append(
        {
            "state": state,
            "file": f"screenshots/{name}",
            "css_viewport": metadata["css_viewport"],
            "device_scale_factor": metadata["device_scale_factor"],
            "png_size": [width, height],
            "sha256": hashlib.sha256(data).hexdigest(),
        }
    )

payload = {
    "scenario_id": "E2E.P0.046",
    "suite_id": "e2e",
    "mode": "automated",
    "result": "PASS",
    "run_id": run_id,
    "source_fingerprint": source_fingerprint,
    "markers": [
        "PRACTICE_PENDING_LEASE_RECOVERY_PASS",
        "PRACTICE_STALE_GENERATION_FENCED_PASS",
        "PRACTICE_CONCURRENT_RESERVATION_PASS",
        "PRACTICE_POST_TIMEOUT_RECONCILIATION_PASS",
        "PRACTICE_TERMINAL_PLAN_RECOVERY_PASS",
        "PRACTICE_EVIDENCE_FINGERPRINT_PASS",
    ],
    "evidence": {
        "postgresql_tests": [
            "TestIntegrationPracticeReplyStateRecovery",
            "TestIntegrationPracticeReplyConcurrentNewIDsReserveOnce",
            "TestIntegrationPracticeReplyConcurrentSameIDInitialReserveOnce",
            "TestIntegrationPracticeReplyConcurrentExpiredSameIDRetryAdvancesOneGeneration",
            "TestIntegrationPracticeReplyStaleGenerationFencedAfterGETRecovery",
        ],
        "postgresql_marker": "PRACTICE_REPLY_STATE_RECOVERY_PASS",
        "postgresql_schema": "current-migrations-isolated-database",
        "postgresql_cleanup": "PRACTICE_ISOLATED_POSTGRES_CLEANUP_PASS residual=0",
        "commit_finalization_tests": [
            "TestSendPracticeMessageCommitFailurePersistsRetryableStateWithDetachedBoundedContext",
            "TestSendPracticeMessageCommitFailureReturnsFinalizationError",
        ],
        "isolation_test": "TestSendPracticeMessageMapsClientMismatchAndCrossUserAccess",
        "screenshots": screenshots,
    },
}
result_file.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY

echo 'PRACTICE_EVIDENCE_FINGERPRINT_PASS scenario=E2E.P0.046 screenshots=4 source=current'
echo 'PRACTICE_P0046_VIEWPORT_EVIDENCE_PASS css=1440x900,390x844 png=1440x900,1170x2532'
echo 'E2E.P0.046 PASS'
