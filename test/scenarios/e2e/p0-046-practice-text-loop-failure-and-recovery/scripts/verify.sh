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
  Transcript.test.tsx \
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
  'keeps hostile HTML, images, and unsafe links inert while hardening safe external links' \
  'sends and retries the exact raw Markdown bytes with one clientMessageId without replacing the next draft' \
  'uses runtime UTF-8 byte limits for the exact session boundary and blocks +1 before send' \
  'accepts exact UTF-8 bytes and rejects limit+1 before the generated client' \
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
  '--- PASS: TestSendPracticeMessageUsesConfiguredUTF8ByteLimitsBeforeAI' \
  '--- PASS: TestSQLRepositoryReservePracticeMessageStartsGenerationOneWithExactLease' \
  '--- PASS: TestSQLRepositoryReservePracticeMessageRejectsAggregateLimitPlusOneBeforeInsert' \
  '--- PASS: TestSQLRepositoryCommitPracticeMessageInsertsReplyAndCompletesUserAtomically' \
  '--- PASS: TestSQLRepositoryCommitPracticeMessageRejectsAssistantAggregateLimitPlusOneBeforeInsert' \
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
grep -Fq 'PRACTICE_MARKDOWN_SECURITY_PASS raw_html_inert=true remote_image_requests=0 unsafe_uri_rejected=true external_rel=noopener_noreferrer' "$LOG"
grep -Fq 'PRACTICE_RAW_RETRY_PASS exact_text=true same_client_message_id=true next_draft_preserved=true rendered_dom_payload=false' "$LOG"
grep -Fq 'PRACTICE_TERMINAL_PLAN_RECOVERY_PASS route=workspace target_job_id_only=true query_free=false parse=false plan_id=false row_retry=false' "$LOG"
grep -Fq 'PRACTICE_ISOLATED_POSTGRES_MIGRATIONS_PASS' "$LOG"
grep -Fq 'PRACTICE_ISOLATED_POSTGRES_CLEANUP_PASS residual=0' "$LOG"
grep -Eq '^version=[0-9]+ dirty=false$' "$OUT/migrations.log"
grep -Fq 'PRACTICE_P0046_SCREENSHOT_CAPTURE_PASS viewports=1440x900,390x844 states=retryable-failed,terminal-failed,hostile-markdown' "$LOG"

for browser_marker in \
  '[desktop]' \
  '[mobile]' \
  'hostile Markdown stays inert without image requests and hardens safe external links' \
  'row-local retry posts the exact original raw Markdown and clientMessageId while preserving the next draft' \
  'retryable failure exposes one row-local retry and preserves the next draft' \
  'terminal failure has no retry escape hatch and keeps the interview locked'; do
  grep -Fq -- "$browser_marker" "$OUT/playwright.log"
done
grep -Eq '(^|[[:space:]])8 passed([[:space:]]|$)' "$OUT/playwright.log"

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
    "practice-hostile-markdown-desktop.png": ("hostile-markdown", [1440, 900], 1, [1440, 900]),
    "practice-hostile-markdown-mobile.png": ("hostile-markdown", [390, 844], 3, [1170, 2532]),
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
        "PRACTICE_MARKDOWN_SECURITY_PASS",
        "PRACTICE_RAW_RETRY_PASS",
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
        "markdown_security_test": "hostile Markdown stays inert without image requests and hardens safe external links",
        "raw_retry_test": "row-local retry posts the exact original raw Markdown and clientMessageId while preserving the next draft",
        "terminal_route": "/workspace?targetJobId",
        "screenshots": screenshots,
    },
}
result_file.write_text(json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
PY

echo 'PRACTICE_EVIDENCE_FINGERPRINT_PASS scenario=E2E.P0.046 screenshots=6 source=current'
echo 'PRACTICE_P0046_VIEWPORT_EVIDENCE_PASS css=1440x900,390x844 png=1440x900,1170x2532'
echo 'E2E.P0.046 PASS'
