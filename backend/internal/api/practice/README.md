# Practice API Handlers

This package owns the HTTP adapter for the backend-practice operations. It maps generated OpenAPI request and response types to `backend/internal/practice` service contracts and keeps middleware ownership explicit at the route layer.

## 002 Event Loop Endpoints

- `POST /practice/sessions/{sessionId}/events` (`appendSessionEvent`) is not wrapped by the shared idempotency middleware. The request-level replay key is `clientEventId`; requests carrying `Idempotency-Key` are rejected with `400 VALIDATION_FAILED` and `details.policy=use_client_event_id`.
- `POST /practice/sessions/{sessionId}/complete` (`completePracticeSession`) is wrapped by `idempotency.Middleware` with `domain=practice` and `operation=completePracticeSession`. The handler returns `202 ReportWithJob` and marks the middleware resource as `feedback_report/{reportId}`.
- `POST /practice/sessions/{sessionId}/voice-turns` (`createPracticeVoiceTurn`) is wrapped by `idempotency.Middleware` with `domain=practice` and `operation=createPracticeVoiceTurn`. The handler decodes the small base64 audio payload, delegates the cascaded STT / chat / TTS orchestration to `backend/internal/practice`, returns `200 PracticeVoiceTurnResult`, and marks the middleware resource as `practice_voice_turn/{voiceTurnId}`.

## 003 Mode Policies and Provenance

- `handleHintRequested` dispatches only on `practice_plans.mode`: `assisted` routes to `applyHintAI`, while `strict` and unknown values return `409 PRACTICE_SESSION_CONFLICT` with `details.policy=hint_disabled_in_mode`.
- `applyHintAI` resolves `practice.turn.lightweight_observe`, calls the observed AI client, and uses `AITaskRunTaskHintGenerate` so `ai_task_runs.task_type='hint_generate'` records the hint path.
- `cmd/api` wraps the Practice AI client with the A3 observability decorator when a profile resolver is available and passes the SQL `ai_task_runs` writer into the service. F3/parse failures that occur before or after `AIClient.Complete` use the same writer for explicit failed `hint_generate` rows.
- Assisted hint success returns `AssistantAction{type=show_hint}` and writes `practice_turns.hint_text`; it does not advance `turn_count`, change turn status, emit `practice.turn.completed`, or write `audit_events`.
- D-36 graceful degrade keeps the session running and returns `session_wait` with non-AI provenance when F3/A3/parse failures occur. Degrade reasons stay in service-local metadata and `ai_task_runs.error_code`, not in the HTTP response body.

## Idempotency Resource Handoff

Handlers wrapped by `idempotency.Middleware` must call `idempotency.SetResponseResource` after the domain service returns the committed resource. The middleware consumes those internal headers while buffering the 2xx response, stores `resource_type/resource_id` in `idempotency_records`, and strips the headers before the client response is flushed.

## Handoff Boundaries

- `003-mode-policies-and-provenance` delivered assisted-mode hint behavior, `practice.turn.lightweight_observe` wiring, `hint_generate` task-run provenance, and strict-mode hint conflict replay.
- `004-derived-plans-debrief` owns retry, next-round, and debrief-derived plan behavior.
- `practice-voice-mvp/001-cascaded-stt-llm-tts` owns voice/audio routes and the `createPracticeVoiceTurn` handoff. 002 does not mount independent voice endpoints.
- `006-privacy-cascade-and-cleanup` owns account deletion cascade and timeout sweeps.
- backend-review/report generation owns report content, scoring, and readiness computation. 002 only creates the queued `feedback_report`, `report_generate` job, and `practice.session.completed` source event.
