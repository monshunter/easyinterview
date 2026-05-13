# Practice API Handlers

This package owns the HTTP adapter for the backend-practice operations. It maps generated OpenAPI request and response types to `backend/internal/practice` service contracts and keeps middleware ownership explicit at the route layer.

## 002 Event Loop Endpoints

- `POST /practice/sessions/{sessionId}/events` (`appendSessionEvent`) is not wrapped by the shared idempotency middleware. The request-level replay key is `clientEventId`; requests carrying `Idempotency-Key` are rejected with `400 VALIDATION_FAILED` and `details.policy=use_client_event_id`.
- `POST /practice/sessions/{sessionId}/complete` (`completePracticeSession`) is wrapped by `idempotency.Middleware` with `domain=practice` and `operation=completePracticeSession`. The handler returns `202 ReportWithJob` and marks the middleware resource as `feedback_report/{reportId}`.

## Idempotency Resource Handoff

Handlers wrapped by `idempotency.Middleware` must call `idempotency.SetResponseResource` after the domain service returns the committed resource. The middleware consumes those internal headers while buffering the 2xx response, stores `resource_type/resource_id` in `idempotency_records`, and strips the headers before the client response is flushed.

## Handoff Boundaries

- `003-mode-policies-and-provenance` owns assisted-mode hint behavior and any future `practice.turn.lightweight_observe` call. In 002, `hint_requested` intentionally remains strict-default `409 PRACTICE_SESSION_CONFLICT`.
- `004-derived-plans-debrief` owns retry, next-round, and debrief-derived plan behavior.
- `005-voice-turn-extension` owns voice/audio routes. 002 does not mount independent voice endpoints.
- `006-privacy-cascade-and-cleanup` owns account deletion cascade and timeout sweeps.
- backend-review/report generation owns report content, scoring, and readiness computation. 002 only creates the queued `feedback_report`, `report_generate` job, and `practice.session.completed` source event.
