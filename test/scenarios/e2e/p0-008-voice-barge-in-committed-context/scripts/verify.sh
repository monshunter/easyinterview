#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-008-voice-barge-in-committed-context"
LOG_FILE="$OUTPUT_DIR/trigger.log"
PRACTICE_TEST="$REPO_ROOT/frontend/src/app/screens/practice/__tests__/practiceVoiceTurn.test.tsx"
test -s "$LOG_FILE"

grep -Fq "RUNNER frontend-vitest E2E.P0.008" "$LOG_FILE" || { echo "E2E.P0.008: frontend runner marker missing" >&2; exit 1; }
grep -Fq "practiceVoiceTurn.test.tsx" "$LOG_FILE" || { echo "E2E.P0.008: practiceVoiceTurn.test.tsx did not run" >&2; exit 1; }
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.008: no passing frontend test files found" >&2; exit 1; }
grep -Fq 'reports partial playback before barge_in_detected when restarting during active assistant playback' "$PRACTICE_TEST" || { echo "E2E.P0.008: frontend barge-in assertion source missing" >&2; exit 1; }
grep -Fq 'expect(bargeInIndex).toBeGreaterThan(playedIndex)' "$PRACTICE_TEST" || { echo "E2E.P0.008: frontend played-before-barge-in assertion missing" >&2; exit 1; }
grep -Fq 'bargeInCall.headers.get("Idempotency-Key")).toBeNull()' "$PRACTICE_TEST" || { echo "E2E.P0.008: frontend idempotency boundary assertion missing" >&2; exit 1; }

grep -Fq "RUNNER backend-go-test E2E.P0.008" "$LOG_FILE" || { echo "E2E.P0.008: backend runner marker missing" >&2; exit 1; }
grep -Fq "=== RUN   TestBuildCommittedVoiceContextCompleteChunk" "$LOG_FILE" || { echo "E2E.P0.008: complete committed context test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestBuildCommittedVoiceContextPartialBargeIn" "$LOG_FILE" || { echo "E2E.P0.008: partial barge-in context test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestBuildCommittedVoiceContextNoPlayback" "$LOG_FILE" || { echo "E2E.P0.008: no-playback context test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestVoiceFollowUpPayloadInjectsCommittedContextWithoutUnplayedDraft" "$LOG_FILE" || { echo "E2E.P0.008: prompt committed context test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestSessionEventServiceRoutesVoicePlaybackEvents" "$LOG_FILE" || { echo "E2E.P0.008: voice event service test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestAppendSessionEventRejectsIdempotencyKeyHeader" "$LOG_FILE" || { echo "E2E.P0.008: handler idempotency boundary test did not run" >&2; exit 1; }
grep -Fq "frontend_status=0" "$LOG_FILE" || { echo "E2E.P0.008: frontend status was not zero" >&2; exit 1; }
grep -Fq "backend_status=0" "$LOG_FILE" || { echo "E2E.P0.008: backend status was not zero" >&2; exit 1; }
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG_FILE" || { echo "E2E.P0.008: failure marker found in trigger log" >&2; exit 1; }

echo "E2E.P0.008 PASS"
