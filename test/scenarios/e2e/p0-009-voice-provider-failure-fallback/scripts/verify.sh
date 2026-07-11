#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel)"
OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-009-voice-provider-failure-fallback"
LOG_FILE="$OUTPUT_DIR/trigger.log"
VOICE_SERVICE="$REPO_ROOT/backend/internal/practice/voice_turn_service.go"
VOICE_TEST="$REPO_ROOT/backend/internal/practice/voice_turn_service_test.go"
FRONTEND_PRACTICE_DIR="$REPO_ROOT/frontend/src/app/screens/practice"
test -s "$LOG_FILE"

grep -Fq "RUNNER frontend-vitest E2E.P0.009" "$LOG_FILE" || { echo "E2E.P0.009: frontend runner marker missing" >&2; exit 1; }
grep -Fq "practiceVoiceTurn.test.tsx" "$LOG_FILE" || { echo "E2E.P0.009: practiceVoiceTurn.test.tsx did not run" >&2; exit 1; }
grep -Fq "devMockClient.test.ts" "$LOG_FILE" || { echo "E2E.P0.009: devMockClient.test.ts did not run" >&2; exit 1; }
grep -Fq "TTS_PROVIDER_FAILED" "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: frontend TTS fallback assertion missing" >&2; exit 1; }
grep -Fq 'scenarioByOp: { createPracticeVoiceTurn: "tts-failed" }' "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: frontend tts-failed fixture selection missing" >&2; exit 1; }
grep -Fq 'keeps the same session and localizes a double-invalid chat failure' "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: localized same-session failure assertion missing" >&2; exit 1; }
grep -Fq 'sessionId: SESSION_A' "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: exact same-session recovery assertion missing" >&2; exit 1; }
grep -Fq 'await user.click(screen.getByTestId("practice-phone-hangup"))' "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: error-state hang-up recovery assertion missing" >&2; exit 1; }
grep -Fq 'scenarioByOp: { createPracticeVoiceTurn: "chat-output-invalid" }' "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: chat-output-invalid fixture selection missing" >&2; exit 1; }
grep -Fq '暂时无法生成符合本场语言的问题' "$FRONTEND_PRACTICE_DIR/__tests__/practiceVoiceTurn.test.tsx" || { echo "E2E.P0.009: localized AI_OUTPUT_INVALID copy assertion missing" >&2; exit 1; }
grep -Fq 'unknown fixture scenario does-not-exist for operationId: createPracticeVoiceTurn' "$REPO_ROOT/frontend/src/api/devMockClient.test.ts" || { echo "E2E.P0.009: unknown voice fixture scenario guard missing" >&2; exit 1; }
grep -Eq 'Test Files +[0-9]+ passed \([0-9]+\)' "$LOG_FILE" || { echo "E2E.P0.009: no passing frontend test files found" >&2; exit 1; }

grep -Fq "RUNNER backend-practice-go-test E2E.P0.009" "$LOG_FILE" || { echo "E2E.P0.009: backend runner marker missing" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnStopsWhenSTTFails" "$LOG_FILE" || { echo "E2E.P0.009: STT fail-fast test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnStopsWhenChatFailsBeforeTTS" "$LOG_FILE" || { echo "E2E.P0.009: chat failure isolation test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnSecondLanguageMismatchSkipsTTSAndPersistence" "$LOG_FILE" || { echo "E2E.P0.009: second language mismatch isolation test did not run" >&2; exit 1; }
grep -Fq 'double-invalid voice turn must stop before TTS' "$VOICE_TEST" || { echo "E2E.P0.009: second-invalid no-TTS assertion missing" >&2; exit 1; }
grep -Fq 'double-invalid voice turn must not persist a result' "$VOICE_TEST" || { echo "E2E.P0.009: second-invalid no-persistence assertion missing" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnReturnsTranscriptAndAssistantTextWhenTTSFails" "$LOG_FILE" || { echo "E2E.P0.009: TTS fallback test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestCreatePracticeVoiceTurnPersistsBusinessTextOutsideAIMetadata" "$LOG_FILE" || { echo "E2E.P0.009: privacy metadata test did not run" >&2; exit 1; }

grep -Fq "RUNNER a3-go-test E2E.P0.009" "$LOG_FILE" || { echo "E2E.P0.009: A3 runner marker missing" >&2; exit 1; }
grep -Fq "=== RUN   TestTranscribe_RealtimeProfileFailsClosed" "$LOG_FILE" || { echo "E2E.P0.009: realtime fail-closed test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestSynthesize_UnsupportedCapabilityFailsClosedWithSharedError" "$LOG_FILE" || { echo "E2E.P0.009: TTS unsupported-capability test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestSynthesize_DisabledProfileFailsClosedWithSharedError" "$LOG_FILE" || { echo "E2E.P0.009: disabled TTS profile fail-closed test did not run" >&2; exit 1; }
grep -Fq "=== RUN   TestTrackedCatalogCoversF3AndProductUICapabilityProfiles" "$LOG_FILE" || { echo "E2E.P0.009: profile catalog coverage test did not run" >&2; exit 1; }
grep -Fq "frontend_status=0" "$LOG_FILE" || { echo "E2E.P0.009: frontend status was not zero" >&2; exit 1; }
grep -Fq "backend_status=0" "$LOG_FILE" || { echo "E2E.P0.009: backend status was not zero" >&2; exit 1; }
grep -Fq "a3_status=0" "$LOG_FILE" || { echo "E2E.P0.009: A3 status was not zero" >&2; exit 1; }
! grep -Eq -- '--- FAIL:|^FAIL($|[[:space:]])|no tests to run' "$LOG_FILE" || { echo "E2E.P0.009: failure marker found in trigger log" >&2; exit 1; }

grep -Fq '"stt-config-missing"' "$REPO_ROOT/openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json" || { echo "E2E.P0.009: stt-config-missing fixture missing" >&2; exit 1; }
grep -Fq '"tts-failed"' "$REPO_ROOT/openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json" || { echo "E2E.P0.009: tts-failed fixture missing" >&2; exit 1; }
grep -Fq '"chat-output-invalid"' "$REPO_ROOT/openapi/fixtures/PracticeSessions/createPracticeVoiceTurn.json" || { echo "E2E.P0.009: chat-output-invalid fixture missing" >&2; exit 1; }
! rg -n 'practice\.voice\.realtime\.default|stub' "$VOICE_SERVICE" "$FRONTEND_PRACTICE_DIR" -g '!*.test.*' -g '!__tests__/**' || { echo "E2E.P0.009: voice runtime contains realtime or stub shortcut" >&2; exit 1; }

echo "E2E.P0.009 PASS"
