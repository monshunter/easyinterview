#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-007-cascaded-voice-turn"
LOG_FILE="$OUTPUT_DIR/trigger.log"
test -s "$LOG_FILE"

grep -Fq "RUNNER frontend-vitest E2E.P0.007" "$LOG_FILE" || { echo "E2E.P0.007: frontend runner marker missing" >&2; exit 1; }
grep -Fq "practiceVoiceTurn.test.tsx" "$LOG_FILE" || { echo "E2E.P0.007: practiceVoiceTurn.test.tsx did not run" >&2; exit 1; }
grep -Fq "PracticeScreen.test.tsx" "$LOG_FILE" || { echo "E2E.P0.007: PracticeScreen.test.tsx did not run" >&2; exit 1; }
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.007: no passing frontend test files found" >&2; exit 1; }

grep -Fq "RUNNER backend-go-test E2E.P0.007" "$LOG_FILE" || { echo "E2E.P0.007: backend runner marker missing" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnRunsIndependentSTTChatTTSProfiles" "$LOG_FILE" || { echo "E2E.P0.007: domain voice test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestSQLRepositoryRecordPracticeVoiceTurnWritesBusinessEventWithoutAudioBytes" "$LOG_FILE" || { echo "E2E.P0.007: store voice test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnReturns200AndMapsRequest" "$LOG_FILE" || { echo "E2E.P0.007: handler voice test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestE2EP0007PracticeVoiceTurnHTTPRoute" "$LOG_FILE" || { echo "E2E.P0.007: HTTP route test did not run" >&2; exit 1; }
grep -Fq -- "--- PASS: TestE2EP0007PracticeVoiceTurnHTTPRoute" "$LOG_FILE" || { echo "E2E.P0.007: HTTP route test did not pass" >&2; exit 1; }
grep -Eq '^ok[[:space:]]+github.com/monshunter/easyinterview/backend/cmd/api([[:space:]]|$)' "$LOG_FILE" || { echo "E2E.P0.007: cmd/api go test did not pass" >&2; exit 1; }
grep -Fq "frontend_status=0" "$LOG_FILE" || { echo "E2E.P0.007: frontend status was not zero" >&2; exit 1; }
grep -Fq "backend_status=0" "$LOG_FILE" || { echo "E2E.P0.007: backend status was not zero" >&2; exit 1; }
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG_FILE" || { echo "E2E.P0.007: failure marker found in trigger log" >&2; exit 1; }

grep -Fq 'POST /api/v1/practice/sessions/{sessionId}/voice-turns' "$REPO_ROOT/backend/cmd/api/main.go" || { echo "E2E.P0.007: voice route is not mounted" >&2; exit 1; }
grep -Fq 'practice_voice_turn/{voiceTurnId}' "$REPO_ROOT/backend/internal/api/practice/README.md" || { echo "E2E.P0.007: idempotency resource handoff not documented" >&2; exit 1; }

echo "E2E.P0.007 PASS"
