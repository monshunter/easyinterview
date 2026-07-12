# Expected Outcome

- The operation matrix exposes `/practice/sessions/{sessionId}/messages` and
  no append-event operation.
- Practice-plan persistence returns the exact canonical round pair; the real
  prompt reservation contains that round context, and wrong/legacy/equal-time
  plans cannot be reused.
- Session completion uses the current lifecycle-only event schema and replay
  produces no second completion fact or report job.
- TargetJob Get/List both move first → next → completed/null from persisted
  facts, hide out-of-order facts after the first gap, ignore report/lifecycle
  status, and select only an exact current-round/current-resume ready plan.
- Frontend mapping, cards, quick-start, and Report next-round consume the same
  projection and perform zero create/start calls for invalid or final progress.
- Browser persistence is limited to frontend display preferences; interview
  business state has no localStorage/sessionStorage/IndexedDB fallback.
- One conversation-level AI call produces dimensions, evidence, risks, and
  next actions.
- Report retry focus persists as PostgreSQL `text[]`, and a retried generating
  report can continue.
- Resume parse AI calls produce task-run and local raw-debug evidence through
  the shared observability wrapper.
- A real browser sees Workspace round states
  `current,pending,pending`, completes the persisted round-1 session through
  `POST /practice/sessions/{sessionId}/complete`, reloads, and sees
  `done,current,pending` with backend current round `round-2-technical`.
- Real Home and Parse reloads consume the same persisted TargetJob projection:
  Home keeps `done,current,pending`, while Parse remains launchable with
  `round-2-technical / 2` as the backend current round.
- Clicking the real Workspace start CTA sends a real
  `POST /practice/plans` whose request contains `round-2-technical`; the 201
  response and a follow-up real plan GET both contain
  `roundId=round-2-technical, roundSequence=2`.
- `POST /practice/sessions` alone is intercepted after the real plan is
  persisted, preventing an unrelated AI interviewer opening call from
  weakening or slowing the persistence/refresh regression gate.
