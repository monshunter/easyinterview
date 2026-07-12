# Expected Outcome

- Practice outbox payloads contain lifecycle IDs only, never message content.
- AI logs, metrics, and task metadata do not expose plaintext prompt or response content.
- Metric labels match the allowlist and exclude provenance or raw model IDs.
- Report generation performs one conversation-level AI call over ordered messages.
- The backend-practice out-of-scope gate passes and every named test actually runs.
