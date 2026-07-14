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
  Ready Home/Workspace card bodies navigate directly to target-scoped Workspace
  detail rather than Parse.
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
- Real Home, Workspace-list, and Workspace-detail reloads consume the same
  persisted TargetJob projection. Clicking either ready card body yields exactly
  `/workspace?targetJobId=019f6098-0000-7000-8000-000000000003`; the detail
  performs one `getTargetJob` read per visit and zero `importTargetJob` / Parse
  poll calls.
- Before and after detail reload, the three detail cards expose
  `done,current,pending`, labels `已进行/即将进行/未进行`, and pairwise-distinct
  computed background colors and border colors. No Parse loading animation is
  mounted.
- Clicking the real Workspace start CTA sends a real
  `POST /practice/plans` whose request contains `round-2-technical`; the 201
  response and a follow-up real plan GET both contain
  `roundId=round-2-technical, roundSequence=2`.
- `POST /practice/sessions` alone is intercepted after the real plan is
  persisted, preventing an unrelated AI interviewer opening call from
  weakening or slowing the persistence/refresh regression gate.
