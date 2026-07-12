# Expected Outcome

- The operation matrix exposes `/practice/sessions/{sessionId}/messages` and
  no append-event operation.
- Practice-plan persistence normalizes empty focus codes to a non-null empty
  PostgreSQL array.
- Session completion uses the current lifecycle-only event schema.
- One conversation-level AI call produces dimensions, evidence, risks, and
  next actions.
- Report retry focus persists as PostgreSQL `text[]`, and a retried generating
  report can continue.
- Resume parse AI calls produce task-run and local raw-debug evidence through
  the shared observability wrapper.
