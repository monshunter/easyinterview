# Expected Outcome

- The host-run frontend uses the real backend API transport.
- A real browser sees Workspace round states `current,pending,pending`, completes
  the persisted round-1 session through
  `POST /practice/sessions/{sessionId}/complete`, reloads, and sees
  `done,current,pending` with current round `round-2-technical`.
- Real Home, Workspace-list, and Workspace-detail reloads consume the same
  persisted TargetJob projection.
- Clicking either ready card body yields exactly
  `/workspace?targetJobId=019f6098-0000-7000-8000-000000000003`; the detail
  performs one real `getTargetJob` read per visit and no import or Parse polling.
- Before and after detail reload, the three detail cards expose
  `done,current,pending`, labels `已进行/即将进行/未进行`, and distinct computed
  background and border colors.
- Package tests, source checks, builds, direct database assertions, and
  intercepted requests are not accepted as E2E PASS evidence; this scenario
  performs no application request interception.
