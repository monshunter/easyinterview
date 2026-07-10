# Expected Outcome

- `startPracticeSession` returns a running session with the first turn.
- Exactly one observed AI task run row is captured for first-question generation.
- The AI task run row includes feature key, model profile, model family, fallback chain, validation status, route, feature flag, data-source version, user id, capability, resource type, and resource id.
- Metric labels match the A3 allowlist and do not include prompt-rubric provenance keys.
- Audit, outbox, AI logs, metric labels, and AI task metadata contain only allowed IDs and hash or length summaries.
- The backend-practice out-of-scope gate passes for the current implementation surface.
